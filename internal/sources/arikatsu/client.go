package arikatsu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"wavearchive/internal/sources/nanoka"
	"wavearchive/internal/sources/nanoka/dto"
)

const (
	repositoryBase = "https://raw.githubusercontent.com/Arikatsu/WutheringWaves_Data"
	maxSourceFile  = 80 << 20
)

var supportedVersions = map[string]bool{"3.5": true, "3.4": true, "3.3": true}

type Client struct {
	version  string
	root     string
	http     *http.Client
	fallback *nanoka.Client

	mu           sync.Mutex
	texts        map[string]string
	roles        []roleInfo
	weapons      []weaponConf
	weaponGrowth []weaponPropertyGrowth
	properties   []propertyIndex
	weaponResons []weaponReson
	echoes       []phantomItem
	echoSkills   []phantomSkill
}

func NewClient(version, cacheRoot string, httpClient *http.Client, fallback *nanoka.Client) (*Client, error) {
	if !supportedVersions[version] {
		return nil, fmt.Errorf("unsupported Arikatsu version %q", version)
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 90 * time.Second}
	}
	if fallback == nil {
		fallback = nanoka.NewClient(httpClient)
	}
	return &Client{version: version, root: cacheRoot, http: httpClient, fallback: fallback}, nil
}

func (c *Client) DetectVersion(context.Context) (string, error) { return c.version, nil }

func (c *Client) CharacterIndex(ctx context.Context, version string) (map[string]dto.CharacterIndexEntry, error) {
	if err := c.checkVersion(version); err != nil {
		return nil, err
	}
	roles, err := c.roleCatalog(ctx)
	if err != nil {
		return nil, err
	}
	texts, err := c.textMap(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]dto.CharacterIndexEntry)
	order := 0
	for _, role := range roles {
		if role.RoleType != 1 || !role.ShowInBag || (role.QualityID != 4 && role.QualityID != 5) {
			continue
		}
		result[strconv.FormatInt(role.ID, 10)] = dto.CharacterIndexEntry{
			Name:       localized(texts, role.Name),
			Nickname:   localized(texts, role.Nickname),
			Rank:       role.QualityID,
			Element:    role.ElementID,
			Weapon:     role.WeaponType,
			Icon:       role.RoleHeadIconLarge,
			Background: role.FormationRoleCard,
			Gender:     roleGender(role.RoleBody),
			APIOrder:   order,
		}
		order++
	}
	if len(result) == 0 {
		return nil, errors.New("Arikatsu character catalog is empty")
	}
	return result, nil
}

func (c *Client) CharacterDetail(ctx context.Context, version, language string, id int64) (dto.CharacterDetail, error) {
	detail, fallbackErr := c.fallback.CharacterDetail(ctx, version, language, id)
	roles, err := c.roleCatalog(ctx)
	if err != nil {
		if fallbackErr != nil {
			return dto.CharacterDetail{}, err
		}
		return detail, nil
	}
	texts, err := c.textMap(ctx)
	if err != nil {
		if fallbackErr != nil {
			return dto.CharacterDetail{}, err
		}
		return detail, nil
	}
	for _, role := range roles {
		if role.ID != id {
			continue
		}
		if fallbackErr != nil {
			return dto.CharacterDetail{}, fmt.Errorf(
				"normalized details for Arikatsu character %d are unavailable: %w",
				id,
				fallbackErr,
			)
		}
		detail.ID = role.ID
		detail.Name = localized(texts, role.Name)
		detail.Nickname = localized(texts, role.Nickname)
		detail.Description = localized(texts, role.Introduction)
		detail.Rarity = role.QualityID
		detail.Element = role.ElementID
		detail.Weapon = role.WeaponType
		detail.Icon = role.RoleHeadIconLarge
		detail.Background = role.FormationRoleCard
		detail.CharaInfo.Sex = roleGender(role.RoleBody)
		return detail, nil
	}
	if fallbackErr != nil {
		return dto.CharacterDetail{}, fmt.Errorf("character %d not found in Arikatsu", id)
	}
	return detail, nil
}

func (c *Client) ItemCatalog(ctx context.Context, version, language string) (map[string]dto.Item, error) {
	return c.fallback.ItemCatalog(ctx, version, language)
}

func (c *Client) ItemIndex(ctx context.Context, version, language string) (map[string]dto.ItemIndexEntry, error) {
	return c.fallback.ItemIndex(ctx, version, language)
}

func (c *Client) WeaponIndex(ctx context.Context, version string) (map[string]dto.WeaponIndexEntry, error) {
	if err := c.checkVersion(version); err != nil {
		return nil, err
	}
	weapons, err := c.weaponCatalog(ctx)
	if err != nil {
		return nil, err
	}
	texts, err := c.textMap(ctx)
	if err != nil {
		return nil, err
	}
	growth, err := c.weaponPropertyGrowth(ctx)
	if err != nil {
		return nil, err
	}
	properties, err := c.propertyCatalog(ctx)
	if err != nil {
		return nil, err
	}
	propertyByID := make(map[int64]propertyIndex, len(properties))
	for _, property := range properties {
		propertyByID[property.ID] = property
	}
	result := make(map[string]dto.WeaponIndexEntry)
	for _, weapon := range weapons {
		if !weapon.IsShow || !weapon.ShowInBag || weapon.QualityID < 3 || weapon.QualityID > 5 {
			continue
		}
		result[strconv.FormatInt(weapon.ItemID, 10)] = dto.WeaponIndexEntry{
			Icon:        weapon.Icon,
			Rank:        weapon.QualityID,
			Type:        weapon.WeaponType,
			Name:        localized(texts, weapon.WeaponName),
			Description: localized(texts, weapon.AttributesDescription),
			BaseATK:     maxWeaponATK(weapon, growth),
			SubStat:     weaponSubStat(weapon, growth, propertyByID, texts),
		}
	}
	if len(result) == 0 {
		return nil, errors.New("Arikatsu weapon catalog is empty")
	}
	return result, nil
}

func (c *Client) Weapon(ctx context.Context, version, language string, id int64) (dto.Weapon, error) {
	if err := c.checkVersion(version); err != nil {
		return dto.Weapon{}, err
	}
	weapons, err := c.weaponCatalog(ctx)
	if err != nil {
		return dto.Weapon{}, err
	}
	texts, err := c.textMap(ctx)
	if err != nil {
		return dto.Weapon{}, err
	}
	resons, resonErr := c.weaponResonCatalog(ctx)
	for _, weapon := range weapons {
		if weapon.ItemID != id {
			continue
		}
		detail := dto.Weapon{
			ID:          weapon.ItemID,
			Name:        localized(texts, weapon.WeaponName),
			Description: localized(texts, weapon.AttributesDescription),
			Rarity:      weapon.QualityID,
			Type:        weapon.WeaponType,
			Icon:        weapon.Icon,
			Effect:      localized(texts, weapon.Desc),
			Params:      parameterRows(weapon.DescParams),
		}
		if resonErr == nil {
			for _, reson := range resons {
				if reson.ResonID == weapon.ResonID && reson.Level == 1 {
					detail.EffectName = localized(texts, reson.Name)
					break
				}
			}
		}
		return detail, nil
	}
	return dto.Weapon{}, fmt.Errorf("weapon %d not found in Arikatsu", id)
}

func (c *Client) EchoIndex(ctx context.Context, version string) (map[string]dto.EchoIndexEntry, error) {
	if err := c.checkVersion(version); err != nil {
		return nil, err
	}
	items, texts, err := c.echoItems(ctx)
	if err != nil {
		return nil, err
	}
	grouped := make(map[int64][]phantomItem)
	for _, item := range items {
		if isCatalogEcho(item) {
			grouped[item.MonsterID] = append(grouped[item.MonsterID], item)
		}
	}
	result := make(map[string]dto.EchoIndexEntry, len(grouped))
	for monsterID, variants := range grouped {
		sort.Slice(variants, func(i, j int) bool { return variants[i].QualityID < variants[j].QualityID })
		best := variants[len(variants)-1]
		ranks := make([]int, 0, len(variants))
		seen := map[int]bool{}
		for _, variant := range variants {
			if variant.QualityID > 0 && !seen[variant.QualityID] {
				seen[variant.QualityID] = true
				ranks = append(ranks, variant.QualityID)
			}
		}
		result[strconv.FormatInt(monsterID, 10)] = dto.EchoIndexEntry{
			Icon: best.Icon, Ranks: ranks, Groups: append([]int64(nil), best.FetterGroup...),
			Name: localized(texts, best.MonsterName), Intensity: best.Rarity,
		}
	}
	if len(result) == 0 {
		return nil, errors.New("Arikatsu Echo catalog is empty")
	}
	return result, nil
}

func (c *Client) EchoDetail(ctx context.Context, version, language string, id int64) (dto.EchoDetail, error) {
	detail, fallbackErr := c.fallback.EchoDetail(ctx, version, language, id)
	items, texts, err := c.echoItems(ctx)
	if err != nil {
		if fallbackErr != nil {
			return dto.EchoDetail{}, err
		}
		return detail, nil
	}
	skills, _ := c.phantomSkills(ctx)
	var matches []phantomItem
	for _, item := range items {
		if item.MonsterID == id && isCatalogEcho(item) {
			matches = append(matches, item)
		}
	}
	if len(matches) == 0 {
		if fallbackErr != nil {
			return dto.EchoDetail{}, fmt.Errorf("Echo %d not found in Arikatsu", id)
		}
		return detail, nil
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].QualityID < matches[j].QualityID })
	best := matches[len(matches)-1]
	detail.ID = id
	detail.Name = localized(texts, best.MonsterName)
	detail.Icon = best.Icon
	detail.IntensityCode = best.Rarity
	detail.Rarity = detail.Rarity[:0]
	for _, item := range matches {
		if item.QualityID > 0 {
			detail.Rarity = appendUnique(detail.Rarity, item.QualityID)
		}
	}
	detail.Group = map[string]any{}
	for _, group := range best.FetterGroup {
		detail.Group[strconv.FormatInt(group, 10)] = true
	}
	for _, skill := range skills {
		if skill.ID == best.SkillID || skill.PhantomSkillID == best.SkillID {
			detail.Skill.Description = localized(texts, skill.Description)
			detail.Skill.SimpleDesc = localized(texts, skill.SimpleDescription)
			detail.Skill.Params = parameterRows(skill.LevelDesc)
			break
		}
	}
	return detail, nil
}

func (c *Client) SonataIndex(ctx context.Context, version string) (map[string]dto.SonataIndexEntry, error) {
	if err := c.checkVersion(version); err != nil {
		return nil, err
	}
	var groups []phantomFetterGroup
	var effects []phantomFetter
	if err := c.readJSON(ctx, "BinData/phantom/phantomfettergroup.json", &groups); err != nil {
		return nil, err
	}
	if err := c.readJSON(ctx, "BinData/phantom/phantomfetter.json", &effects); err != nil {
		return nil, err
	}
	texts, err := c.textMap(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]phantomFetter, len(effects))
	for _, effect := range effects {
		byID[effect.ID] = effect
	}
	result := make(map[string]dto.SonataIndexEntry, len(groups))
	for _, group := range groups {
		entry := dto.SonataIndexEntry{ID: group.ID, Icon: group.Icon, Set: map[string]map[string]struct {
			Description string `json:"desc"`
			Params      []any  `json:"param"`
		}{}}
		entry.Name.English = localized(texts, group.Name)
		for _, link := range group.FetterMap {
			effect, ok := byID[link.Value]
			if !ok {
				continue
			}
			piece := strconv.Itoa(link.Key)
			entry.Set[piece] = map[string]struct {
				Description string `json:"desc"`
				Params      []any  `json:"param"`
			}{"en": {Description: localized(texts, effect.Description), Params: stringsToAny(effect.Params)}}
		}
		result[strconv.FormatInt(group.ID, 10)] = entry
	}
	return result, nil
}

func (c *Client) echoItems(ctx context.Context) ([]phantomItem, map[string]string, error) {
	items, err := c.phantomCatalog(ctx)
	if err != nil {
		return nil, nil, err
	}
	texts, err := c.textMap(ctx)
	return items, texts, err
}

func (c *Client) roleCatalog(ctx context.Context) ([]roleInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.roles != nil {
		return c.roles, nil
	}
	if err := c.readJSONUnlocked(ctx, "BinData/role/roleinfo.json", &c.roles); err != nil {
		return nil, err
	}
	return c.roles, nil
}

func (c *Client) weaponCatalog(ctx context.Context) ([]weaponConf, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.weapons != nil {
		return c.weapons, nil
	}
	if err := c.readJSONUnlocked(ctx, "BinData/weapon/weaponconf.json", &c.weapons); err != nil {
		return nil, err
	}
	return c.weapons, nil
}

func (c *Client) weaponPropertyGrowth(ctx context.Context) ([]weaponPropertyGrowth, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.weaponGrowth != nil {
		return c.weaponGrowth, nil
	}
	if err := c.readJSONUnlocked(ctx, "BinData/property/weaponpropertygrowth.json", &c.weaponGrowth); err != nil {
		return nil, err
	}
	return c.weaponGrowth, nil
}

func (c *Client) propertyCatalog(ctx context.Context) ([]propertyIndex, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.properties != nil {
		return c.properties, nil
	}
	if err := c.readJSONUnlocked(ctx, "BinData/property/propertyindex.json", &c.properties); err != nil {
		return nil, err
	}
	return c.properties, nil
}

func (c *Client) weaponResonCatalog(ctx context.Context) ([]weaponReson, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.weaponResons != nil {
		return c.weaponResons, nil
	}
	if err := c.readJSONUnlocked(ctx, "BinData/weapon/weaponreson.json", &c.weaponResons); err != nil {
		return nil, err
	}
	return c.weaponResons, nil
}

func (c *Client) phantomCatalog(ctx context.Context) ([]phantomItem, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.echoes != nil {
		return c.echoes, nil
	}
	if err := c.readJSONUnlocked(ctx, "BinData/phantom/phantomitem.json", &c.echoes); err != nil {
		return nil, err
	}
	return c.echoes, nil
}

func (c *Client) phantomSkills(ctx context.Context) ([]phantomSkill, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.echoSkills != nil {
		return c.echoSkills, nil
	}
	if err := c.readJSONUnlocked(ctx, "BinData/phantom/phantomskill.json", &c.echoSkills); err != nil {
		return nil, err
	}
	return c.echoSkills, nil
}

func (c *Client) textMap(ctx context.Context) (map[string]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.texts != nil {
		return c.texts, nil
	}
	var entries []textEntry
	if err := c.readJSONUnlocked(ctx, "Textmaps/pt/multi_text/MultiText.json", &entries); err != nil {
		return nil, err
	}
	texts := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.ID != "" && entry.Content != "" {
			texts[entry.ID] = entry.Content
		}
	}
	c.texts = texts
	return texts, nil
}

func (c *Client) readJSON(ctx context.Context, path string, target any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.readJSONUnlocked(ctx, path, target)
}

func (c *Client) readJSONUnlocked(ctx context.Context, path string, target any) error {
	localPath, err := c.ensureFile(ctx, path)
	if err != nil {
		return err
	}
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(target); err != nil {
		return fmt.Errorf("decode Arikatsu %s: %w", path, err)
	}
	return nil
}

func (c *Client) ensureFile(ctx context.Context, path string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(path))
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", errors.New("invalid Arikatsu source path")
	}
	versionRoot := filepath.Join(c.root, c.version)
	destination := filepath.Join(versionRoot, clean)
	relative, err := filepath.Rel(versionRoot, destination)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("Arikatsu source path escapes cache")
	}
	if info, err := os.Stat(destination); err == nil && info.Size() > 0 {
		return destination, nil
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, repositoryBase+"/"+c.version+"/"+filepath.ToSlash(clean), nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "WaveArchive/0.1")
	response, err := c.http.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download Arikatsu %s: HTTP %d", path, response.StatusCode)
	}
	if response.ContentLength > maxSourceFile {
		return "", fmt.Errorf("Arikatsu source %s exceeds size limit", path)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return "", err
	}
	temp, err := os.CreateTemp(filepath.Dir(destination), ".arikatsu-*")
	if err != nil {
		return "", err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	written, copyErr := io.Copy(temp, io.LimitReader(response.Body, maxSourceFile+1))
	if copyErr == nil && written > maxSourceFile {
		copyErr = errors.New("Arikatsu source exceeds size limit")
	}
	if closeErr := temp.Close(); copyErr == nil {
		copyErr = closeErr
	}
	if copyErr != nil {
		return "", copyErr
	}
	if err := os.Rename(tempName, destination); err != nil {
		return "", err
	}
	return destination, nil
}

func (c *Client) checkVersion(version string) error {
	if version != c.version {
		return fmt.Errorf("Arikatsu client configured for %s, received %s", c.version, version)
	}
	return nil
}

func localized(texts map[string]string, key string) string {
	if value := strings.TrimSpace(texts[key]); value != "" {
		return value
	}
	return strings.TrimSpace(key)
}

func roleGender(body string) string {
	switch {
	case strings.HasPrefix(strings.ToLower(body), "female"):
		return "Female"
	case strings.HasPrefix(strings.ToLower(body), "male"):
		return "Male"
	default:
		return ""
	}
}

// The 35xxxxxx MonsterInfo namespace contains Phantom cosmetic variants of
// existing Echoes. They are valid raw inventory records, but are not separate
// entries in the Echo Data Bank/catalog.
func isCatalogEcho(item phantomItem) bool {
	return item.PhantomType == 1 &&
		item.MonsterID > 0 &&
		!strings.HasPrefix(item.MonsterName, "MonsterInfo_35")
}

func maxWeaponATK(weapon weaponConf, growth []weaponPropertyGrowth) int {
	return int(math.Round(maxWeaponProperty(weapon.FirstProp.Value, weapon.FirstCurve, growth)))
}

func weaponSubStat(
	weapon weaponConf,
	growth []weaponPropertyGrowth,
	properties map[int64]propertyIndex,
	texts map[string]string,
) string {
	if weapon.SecondProp.ID == 0 || weapon.SecondProp.Value == 0 {
		return ""
	}
	property, ok := properties[weapon.SecondProp.ID]
	if ok && property.ConvertToWhiteID != 0 {
		if base, exists := properties[property.ConvertToWhiteID]; exists {
			property = base
		}
	}
	name := localized(texts, property.Name)
	if property.Name == "" || name == property.Name {
		name = property.Key
	}
	value := maxWeaponProperty(weapon.SecondProp.Value, weapon.SecondCurve, growth)
	if weapon.SecondProp.IsRatio {
		return strings.TrimSpace(name + " " + formatDecimal(value*100) + "%")
	}
	return strings.TrimSpace(name + " " + formatDecimal(value))
}

func maxWeaponProperty(base float64, curveID int, growth []weaponPropertyGrowth) float64 {
	bestLevel := -1
	bestBreach := -1
	multiplier := float64(10000)
	for _, point := range growth {
		if point.CurveID != curveID {
			continue
		}
		if point.Level > bestLevel || (point.Level == bestLevel && point.BreachLevel > bestBreach) {
			bestLevel = point.Level
			bestBreach = point.BreachLevel
			multiplier = point.CurveValue
		}
	}
	return base * multiplier / 10000
}

func formatDecimal(value float64) string {
	formatted := strconv.FormatFloat(value, 'f', 1, 64)
	formatted = strings.TrimSuffix(formatted, ".0")
	return strings.ReplaceAll(formatted, ".", ",")
}

func parameterRows(rows []stringArray) []any {
	result := make([]any, 0, len(rows))
	for _, row := range rows {
		values := make([]any, len(row.Values))
		for index, value := range row.Values {
			values[index] = value
		}
		result = append(result, values)
	}
	return result
}

func stringsToAny(values []string) []any {
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = value
	}
	return result
}

func appendUnique(values []int, value int) []int {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
