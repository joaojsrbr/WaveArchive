package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"sync"
	"time"

	"wavearchive/internal/assets"
	"wavearchive/internal/domain"
	"wavearchive/internal/sources/nanoka/mapper"
)

type CharacterCatalog struct {
	repository domain.CharacterRepository
	source     CatalogSource
	assets     *assets.Cache
	logger     *slog.Logger
}

func NewCharacterCatalog(repository domain.CharacterRepository, source CatalogSource, assetCache *assets.Cache, logger *slog.Logger) *CharacterCatalog {
	return &CharacterCatalog{repository: repository, source: source, assets: assetCache, logger: logger}
}

func (c *CharacterCatalog) SetSource(source CatalogSource) { c.source = source }

func (c *CharacterCatalog) List(ctx context.Context, filter domain.CharacterFilter) ([]domain.Character, error) {
	return c.repository.List(ctx, filter)
}

func (c *CharacterCatalog) SearchContent(ctx context.Context, query string, limit int) ([]domain.CharacterContentSearchResult, error) {
	return c.repository.SearchContent(ctx, query, limit)
}

func (c *CharacterCatalog) Get(ctx context.Context, id int64) (domain.CharacterProfile, error) {
	return c.repository.GetProfile(ctx, id)
}

func (c *CharacterCatalog) CalculateProgression(ctx context.Context, request domain.ProgressionPlanRequest) (domain.ProgressionPlan, error) {
	if request.CharacterID <= 0 || request.CurrentLevel < 1 || request.TargetLevel > 90 || request.TargetLevel < request.CurrentLevel {
		return domain.ProgressionPlan{}, errors.New("invalid progression range")
	}
	profile, err := c.repository.GetProfile(ctx, request.CharacterID)
	if err != nil {
		return domain.ProgressionPlan{}, err
	}
	ascension := map[int64]domain.MaterialCost{}
	skills := map[int64]domain.MaterialCost{}
	for _, stage := range profile.Progression.Ascensions {
		if stage.UnlockLevel >= request.CurrentLevel && stage.UnlockLevel < request.TargetLevel {
			addCosts(ascension, stage.Costs)
		}
	}
	for _, skill := range profile.Progression.Skills {
		current := request.CurrentSkills[skill.NodeID]
		target := request.TargetSkills[skill.NodeID]
		if current <= 0 {
			current = 1
		}
		if target <= 0 {
			target = current
		}
		if target > skill.MaxLevel || target < current {
			return domain.ProgressionPlan{}, fmt.Errorf("invalid level range for %s", skill.Name)
		}
		if request.IncludeUnlocks && skill.MaxLevel <= 1 {
			addedOfficialCost := false
			for _, level := range skill.LevelCosts {
				if level.Level == 1 && len(level.Costs) > 0 {
					addCosts(skills, level.Costs)
					addedOfficialCost = true
					break
				}
			}
			if !addedOfficialCost && skill.NodeType == 4 {
				addCosts(skills, skill.UnlockCosts)
			}
		}
		for _, level := range skill.LevelCosts {
			if level.Level > current && level.Level <= target {
				addCosts(skills, level.Costs)
			}
		}
	}
	total := map[int64]domain.MaterialCost{}
	for _, cost := range ascension {
		addCosts(total, []domain.MaterialCost{cost})
	}
	for _, cost := range skills {
		addCosts(total, []domain.MaterialCost{cost})
	}
	return domain.ProgressionPlan{
		CharacterID: request.CharacterID,
		Ascensions:  sortedCosts(ascension), Skills: sortedCosts(skills), Total: sortedCosts(total),
	}, nil
}

func addCosts(target map[int64]domain.MaterialCost, costs []domain.MaterialCost) {
	for _, cost := range costs {
		current := target[cost.Material.ID]
		current.Material = cost.Material
		current.Quantity += cost.Quantity
		target[cost.Material.ID] = current
	}
}

func sortedCosts(source map[int64]domain.MaterialCost) []domain.MaterialCost {
	result := make([]domain.MaterialCost, 0, len(source))
	for _, cost := range source {
		result = append(result, cost)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Material.ID == 2 {
			return true
		}
		if result[j].Material.ID == 2 {
			return false
		}
		return result[i].Material.ID < result[j].Material.ID
	})
	return result
}

func (c *CharacterCatalog) UpdateAccount(ctx context.Context, update domain.CharacterAccountUpdate) error {
	if update.CharacterID <= 0 {
		return errors.New("invalid character id")
	}
	if update.Level < 1 || update.Level > 90 {
		return errors.New("character level must be between 1 and 90")
	}
	if update.Sequence < 0 || update.Sequence > 6 {
		return errors.New("sequence must be between 0 and 6")
	}
	return c.repository.UpdateAccount(ctx, update)
}

func (c *CharacterCatalog) Status(ctx context.Context) (domain.CatalogStatus, error) {
	return c.repository.Status(ctx)
}

func (c *CharacterCatalog) Sync(ctx context.Context, progress func(string, int)) (domain.SyncResult, error) {
	progress("detecting", 10)
	version, err := c.source.DetectVersion(ctx)
	if err != nil {
		return domain.SyncResult{}, fmt.Errorf("detect game version: %w", err)
	}

	return c.syncVersion(ctx, version, progress)
}

func (c *CharacterCatalog) SyncVersion(ctx context.Context, version string, progress func(string, int)) (domain.SyncResult, error) {
	if version == "" {
		return domain.SyncResult{}, errors.New("data version is required")
	}
	progress("detecting", 10)
	return c.syncVersion(ctx, version, progress)
}

func (c *CharacterCatalog) syncVersion(ctx context.Context, version string, progress func(string, int)) (domain.SyncResult, error) {
	progress("downloading", 35)
	index, err := c.source.CharacterIndex(ctx, version)
	if err != nil {
		return domain.SyncResult{}, fmt.Errorf("download character index: %w", err)
	}

	progress("normalizing", 40)
	profiles := make([]domain.CharacterProfile, 0, len(index))
	for id, entry := range index {
		if character, ok := mapper.CharacterIndex(id, entry, version); ok {
			profiles = append(profiles, character)
		} else {
			c.logger.Warn("ignored character with invalid id", "id", id)
		}
	}
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Character.ID < profiles[j].Character.ID
	})

	progress("details", 45)
	c.loadDetails(ctx, version, profiles, progress)

	progress("materials", 72)
	c.loadMaterials(ctx, version, profiles, progress)

	progress("weapons", 78)
	c.loadSignatureWeapons(ctx, version, profiles, progress)

	progress("assets", 85)
	c.loadAssets(ctx, profiles, progress)

	if err := ctx.Err(); err != nil {
		return domain.SyncResult{}, err
	}
	progress("saving", 90)
	if err := c.repository.ReplaceSynced(ctx, version, profiles); err != nil {
		return domain.SyncResult{}, fmt.Errorf("save character catalog: %w", err)
	}
	syncedAt := time.Now().UTC()
	progress("done", 100)
	return domain.SyncResult{
		Version: version,
		Count:   len(profiles),
		Synced:  syncedAt.Format(time.RFC3339),
	}, nil
}

func (c *CharacterCatalog) loadMaterials(ctx context.Context, version string, profiles []domain.CharacterProfile, progress func(string, int)) {
	rich, richErr := c.source.ItemCatalog(ctx, version, "en")
	compact, compactErr := c.source.ItemIndex(ctx, version, "en")
	if richErr != nil && compactErr != nil {
		c.logger.Warn("material catalogs unavailable", "item_all", richErr, "item", compactErr)
		return
	}

	materials := make(map[int64]domain.Material, len(rich)+len(compact))
	for idText, source := range compact {
		id, err := strconv.ParseInt(idText, 10, 64)
		if err == nil {
			materials[id] = domain.Material{ID: id, Name: source.Name, Rarity: source.Rank, Type: source.Type, IconPath: source.Icon, GameVersion: version, Sources: []string{}}
		}
	}
	for idText, source := range rich {
		id := source.ID
		if id == 0 {
			id, _ = strconv.ParseInt(idText, 10, 64)
		}
		if id > 0 {
			materials[id] = domain.Material{
				ID: id, Name: source.Name, Rarity: source.Rarity, Type: source.Type,
				Description: source.Description, IconPath: source.Icon,
				Sources: append([]string(nil), source.Source...), GameVersion: version,
			}
		}
	}

	used := map[int64]bool{}
	for profileIndex := range profiles {
		p := &profiles[profileIndex].Progression
		for stageIndex := range p.Ascensions {
			enrichMaterialCosts(p.Ascensions[stageIndex].Costs, materials, used)
		}
		for skillIndex := range p.Skills {
			enrichMaterialCosts(p.Skills[skillIndex].UnlockCosts, materials, used)
			for levelIndex := range p.Skills[skillIndex].LevelCosts {
				enrichMaterialCosts(p.Skills[skillIndex].LevelCosts[levelIndex].Costs, materials, used)
			}
		}
	}
	if c.assets != nil {
		completed := 0
		for id := range used {
			material := materials[id]
			if material.IconPath != "" {
				if local, err := c.assets.Ensure(ctx, material.IconPath, fmt.Sprintf("materials/%d/icon.webp", id)); err == nil {
					material.IconPath = local
					materials[id] = material
				} else {
					c.logger.Warn("material icon unavailable", "id", id, "error", err)
				}
			}
			completed++
			if len(used) > 0 {
				progress("materials", 72+completed*5/len(used))
			}
		}
		for profileIndex := range profiles {
			p := &profiles[profileIndex].Progression
			for stageIndex := range p.Ascensions {
				enrichMaterialCosts(p.Ascensions[stageIndex].Costs, materials, nil)
			}
			for skillIndex := range p.Skills {
				enrichMaterialCosts(p.Skills[skillIndex].UnlockCosts, materials, nil)
				for levelIndex := range p.Skills[skillIndex].LevelCosts {
					enrichMaterialCosts(p.Skills[skillIndex].LevelCosts[levelIndex].Costs, materials, nil)
				}
			}
		}
	}
}

func enrichMaterialCosts(costs []domain.MaterialCost, materials map[int64]domain.Material, used map[int64]bool) {
	for index := range costs {
		id := costs[index].Material.ID
		if material, ok := materials[id]; ok {
			costs[index].Material = material
		}
		if used != nil {
			used[id] = true
		}
	}
}

func (c *CharacterCatalog) loadAssets(
	ctx context.Context,
	profiles []domain.CharacterProfile,
	progress func(string, int),
) {
	if c.assets == nil {
		return
	}
	type task struct {
		profileIndex int
		itemIndex    int
		kind         string
		source       string
		relative     string
	}
	tasks := make([]task, 0, len(profiles)*9)
	seenWeapons := map[int64]bool{}
	for index, profile := range profiles {
		id := profile.Character.ID
		if profile.Character.IconPath != "" {
			tasks = append(tasks, task{profileIndex: index, kind: "icon", source: profile.Character.IconPath, relative: fmt.Sprintf("characters/%d/icon.webp", id)})
		}
		if profile.Character.BackgroundPath != "" {
			tasks = append(tasks, task{profileIndex: index, kind: "background", source: profile.Character.BackgroundPath, relative: fmt.Sprintf("characters/%d/background.webp", id)})
		}
		if profile.SignatureWeapon != nil && profile.SignatureWeapon.IconPath != "" && !seenWeapons[profile.SignatureWeapon.ID] {
			seenWeapons[profile.SignatureWeapon.ID] = true
			tasks = append(tasks, task{profileIndex: index, kind: "weapon", source: profile.SignatureWeapon.IconPath, relative: fmt.Sprintf("weapons/%d/icon.webp", profile.SignatureWeapon.ID)})
		}
		for skillIndex, skill := range profile.Skills {
			if skill.IconPath == "" {
				continue
			}
			tasks = append(tasks, task{
				profileIndex: index,
				itemIndex:    skillIndex,
				kind:         "skill",
				source:       skill.IconPath,
				relative:     fmt.Sprintf("characters/%d/skills/%03d.webp", id, skillIndex),
			})
		}
	}
	if len(tasks) == 0 {
		return
	}
	type result struct {
		task task
		url  string
		err  error
	}
	jobs := make(chan task)
	results := make(chan result)
	workers := 4
	if len(tasks) < workers {
		workers = len(tasks)
	}
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for item := range jobs {
				url, err := c.assets.Ensure(ctx, item.source, item.relative)
				results <- result{task: item, url: url, err: err}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, item := range tasks {
			select {
			case jobs <- item:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		group.Wait()
		close(results)
	}()

	completed := 0
	weaponURLs := map[int64]string{}
	for result := range results {
		completed++
		if result.err != nil {
			c.logger.Warn("asset unavailable", "source", result.task.source, "error", result.err)
		} else {
			switch result.task.kind {
			case "icon":
				profiles[result.task.profileIndex].Character.IconPath = result.url
			case "background":
				profiles[result.task.profileIndex].Character.BackgroundPath = result.url
			case "weapon":
				weapon := profiles[result.task.profileIndex].SignatureWeapon
				if weapon != nil {
					weaponURLs[weapon.ID] = result.url
				}
			case "skill":
				profile := &profiles[result.task.profileIndex]
				if result.task.itemIndex < 0 || result.task.itemIndex >= len(profile.Skills) {
					break
				}
				profile.Skills[result.task.itemIndex].IconPath = result.url
				nodeID := profile.Skills[result.task.itemIndex].NodeID
				for progressionIndex := range profile.Progression.Skills {
					if profile.Progression.Skills[progressionIndex].NodeID == nodeID {
						profile.Progression.Skills[progressionIndex].IconPath = result.url
					}
				}
			}
		}
		if len(tasks) > 0 {
			progress("assets", 85+completed*4/len(tasks))
		}
	}
	for index := range profiles {
		if profiles[index].SignatureWeapon != nil {
			if url := weaponURLs[profiles[index].SignatureWeapon.ID]; url != "" {
				profiles[index].SignatureWeapon.IconPath = url
			}
		}
	}
}

func (c *CharacterCatalog) loadDetails(
	ctx context.Context,
	version string,
	profiles []domain.CharacterProfile,
	progress func(string, int),
) {
	type result struct {
		index   int
		profile domain.CharacterProfile
		err     error
	}
	jobs := make(chan int)
	results := make(chan result)
	workers := 4
	if len(profiles) < workers {
		workers = len(profiles)
	}

	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range jobs {
				base := profiles[index]
				detail, err := c.source.CharacterDetail(ctx, version, "en", base.Character.ID)
				if err != nil {
					results <- result{index: index, profile: base, err: err}
					continue
				}
				results <- result{index: index, profile: mapper.MergeCharacterDetail(base, detail)}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for index := range profiles {
			select {
			case jobs <- index:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		group.Wait()
		close(results)
	}()

	completed := 0
	for result := range results {
		if result.err != nil {
			c.logger.Warn("character detail unavailable; preserving previous detail", "id", result.profile.Character.ID, "error", result.err)
		} else {
			profiles[result.index] = result.profile
		}
		completed++
		if len(profiles) > 0 {
			progress("details", 45+(completed*28/len(profiles)))
		}
	}
}

func (c *CharacterCatalog) loadSignatureWeapons(
	ctx context.Context,
	version string,
	profiles []domain.CharacterProfile,
	progress func(string, int),
) {
	ids := make([]int64, 0)
	seen := map[int64]bool{}
	for _, profile := range profiles {
		if profile.SignatureWeapon != nil && !seen[profile.SignatureWeapon.ID] {
			seen[profile.SignatureWeapon.ID] = true
			ids = append(ids, profile.SignatureWeapon.ID)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	weapons := make(map[int64]domain.Weapon, len(ids))
	for index, id := range ids {
		if ctx.Err() != nil {
			return
		}
		source, err := c.source.Weapon(ctx, version, "en", id)
		if err != nil {
			c.logger.Warn("signature weapon unavailable", "id", id, "error", err)
			continue
		}
		weapons[id] = mapper.Weapon(source)
		if len(ids) > 0 {
			progress("weapons", 78+(index+1)*10/len(ids))
		}
	}
	for index := range profiles {
		if profiles[index].SignatureWeapon == nil {
			continue
		}
		if weapon, ok := weapons[profiles[index].SignatureWeapon.ID]; ok {
			profiles[index].SignatureWeapon = &weapon
		}
	}
}
