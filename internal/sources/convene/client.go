package convene

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"wavearchive/internal/domain"
)

const maxResponseBytes = 16 << 20

var allowedHistoryHost = regexp.MustCompile(`(?i)^aki-gm-resources(?:-oversea)?\.aki-game\.(?:net|com)$`)

type stringOrNumber string

func (value *stringOrNumber) UnmarshalJSON(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	if bytes.Equal(raw, []byte("null")) {
		*value = ""
		return nil
	}
	if len(raw) > 0 && raw[0] == '"' {
		var text string
		if err := json.Unmarshal(raw, &text); err != nil {
			return err
		}
		*value = stringOrNumber(text)
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(raw, &number); err != nil {
		return fmt.Errorf("esperado texto ou número: %w", err)
	}
	*value = stringOrNumber(number.String())
	return nil
}

type Credentials struct {
	PlayerID     string
	ServerID     string
	ResourceID   string
	RecordID     string
	LanguageCode string
	Region       string
	Endpoint     string
	LocaleURL    string
}

type Client struct{ http *http.Client }

func NewClient(client *http.Client) *Client { return &Client{http: client} }

func ParseHistoryURL(raw string) (Credentials, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return Credentials{}, errors.New("URL de Convene inválida")
	}
	if parsed.Scheme != "https" || !allowedHistoryHost.MatchString(parsed.Hostname()) {
		return Credentials{}, errors.New("a URL precisa ser do histórico oficial de Wuthering Waves")
	}
	queryText := parsed.RawQuery
	if index := strings.Index(parsed.Fragment, "?"); index >= 0 {
		queryText = parsed.Fragment[index+1:]
	}
	values, err := url.ParseQuery(queryText)
	if err != nil {
		return Credentials{}, errors.New("parâmetros da URL de Convene inválidos")
	}
	credentials := Credentials{
		PlayerID:     strings.TrimSpace(values.Get("player_id")),
		ServerID:     strings.TrimSpace(values.Get("svr_id")),
		ResourceID:   strings.TrimSpace(values.Get("resources_id")),
		RecordID:     strings.TrimSpace(values.Get("record_id")),
		LanguageCode: strings.TrimSpace(values.Get("lang")),
		Region:       "global",
		Endpoint:     "https://gmserver-api.aki-game2.net/gacha/record/query",
		LocaleURL:    parsed.Scheme + "://" + parsed.Host + "/aki/gacha/locales/",
	}
	if strings.HasSuffix(strings.ToLower(parsed.Hostname()), ".com") {
		credentials.Region = "cn"
		credentials.Endpoint = "https://gmserver-api.aki-game2.com/gacha/record/query"
	}
	if credentials.LanguageCode == "" {
		credentials.LanguageCode = "pt"
	}
	if credentials.PlayerID == "" || credentials.ServerID == "" || credentials.ResourceID == "" || credentials.RecordID == "" {
		return Credentials{}, errors.New("a URL não contém player_id, svr_id, resources_id e record_id")
	}
	if len(credentials.PlayerID) > 64 || len(credentials.ServerID) > 64 ||
		len(credentials.ResourceID) > 128 || len(credentials.RecordID) > 1024 ||
		len(credentials.LanguageCode) > 16 {
		return Credentials{}, errors.New("a URL de Convene contém parâmetros fora do limite")
	}
	return credentials, nil
}

func (c *Client) FetchPoolCatalog(ctx context.Context, credentials Credentials) ([]Pool, error) {
	language := strings.ToLower(credentials.LanguageCode)
	if index := strings.IndexAny(language, "-_"); index >= 0 {
		language = language[:index]
	}
	if !regexp.MustCompile(`^[a-z]{2,3}$`).MatchString(language) {
		language = "pt"
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, credentials.LocaleURL+language+".json", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("consultar lista oficial de banners: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("consultar lista oficial de banners: HTTP %d", response.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	var locale struct {
		SelectList map[string]string `json:"selectList"`
	}
	if err := json.Unmarshal(raw, &locale); err != nil {
		return nil, fmt.Errorf("decodificar lista oficial de banners: %w", err)
	}
	if len(locale.SelectList) == 0 {
		return nil, errors.New("a lista oficial de banners está vazia")
	}
	return PoolsFromSelectList(locale.SelectList), nil
}

func (c *Client) FetchAll(ctx context.Context, credentials Credentials, pools []Pool) ([]domain.ConvenePull, int, error) {
	type poolResult struct {
		pulls []domain.ConvenePull
		err   error
	}
	results := make(chan poolResult, len(pools))
	semaphore := make(chan struct{}, 4)
	var group sync.WaitGroup
	for _, pool := range pools {
		pool := pool
		group.Add(1)
		go func() {
			defer group.Done()
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				results <- poolResult{err: ctx.Err()}
				return
			}
			defer func() { <-semaphore }()
			pulls, err := c.fetchPool(ctx, credentials, pool)
			results <- poolResult{pulls: pulls, err: err}
		}()
	}
	group.Wait()
	close(results)

	all := make([]domain.ConvenePull, 0)
	updated := 0
	for result := range results {
		if result.err != nil {
			return nil, 0, result.err
		}
		if len(result.pulls) > 0 {
			updated++
			all = append(all, result.pulls...)
		}
	}
	return all, updated, nil
}

func (c *Client) fetchPool(
	ctx context.Context,
	credentials Credentials,
	pool Pool,
) ([]domain.ConvenePull, error) {
	body, err := json.Marshal(map[string]string{
		"playerId":     credentials.PlayerID,
		"serverId":     credentials.ServerID,
		"cardPoolId":   credentials.ResourceID,
		"cardPoolType": strconv.Itoa(pool.Type),
		"languageCode": credentials.LanguageCode,
		"recordId":     credentials.RecordID,
	})
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, credentials.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("consultar %s: %w", pool.Name, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("consultar %s: servidor retornou HTTP %d", pool.Name, response.StatusCode)
	}
	limited := io.LimitReader(response.Body, maxResponseBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(raw) > maxResponseBytes {
		return nil, errors.New("resposta do histórico excedeu o limite seguro")
	}
	var payload struct {
		Code    json.RawMessage `json:"code"`
		Message string          `json:"message"`
		Data    []struct {
			ResourceID   stringOrNumber `json:"resourceId"`
			ResourceType string         `json:"resourceType"`
			Name         string         `json:"name"`
			QualityLevel int            `json:"qualityLevel"`
			Count        int            `json:"count"`
			Time         string         `json:"time"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("decodificar histórico de %s: %w", pool.Name, err)
	}
	code := strings.Trim(string(payload.Code), `" `)
	if code != "" && code != "0" {
		if strings.TrimSpace(payload.Message) == "" {
			payload.Message = "o token do histórico pode ter expirado"
		}
		return nil, fmt.Errorf("consultar %s: %s", pool.Name, payload.Message)
	}
	pulls := make([]domain.ConvenePull, 0, len(payload.Data))
	occurrences := map[string]int{}
	for index, item := range payload.Data {
		if item.QualityLevel < 3 || item.QualityLevel > 5 || strings.TrimSpace(item.Name) == "" {
			continue
		}
		if item.Count <= 0 {
			item.Count = 1
		}
		base := strings.Join([]string{
			credentials.PlayerID,
			strconv.Itoa(pool.Type),
			string(item.ResourceID),
			strconv.Itoa(item.QualityLevel),
			normalizeResourceType(item.ResourceType),
			item.Name,
			item.Time,
		}, "\x1f")
		occurrence := occurrences[base]
		occurrences[base]++
		pulls = append(pulls, domain.ConvenePull{
			PoolType:     pool.Type,
			PoolName:     pool.Name,
			ResourceID:   string(item.ResourceID),
			ResourceType: normalizeResourceType(item.ResourceType),
			ItemName:     item.Name,
			Rarity:       item.QualityLevel,
			Quantity:     item.Count,
			ObtainedAt:   item.Time,
			SourceIndex:  index,
			Fingerprint:  fingerprint(base, occurrence),
		})
	}
	return pulls, nil
}

func normalizeResourceType(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch {
	case strings.Contains(normalized, "resonator"),
		strings.Contains(normalized, "resonante"),
		strings.Contains(normalized, "ressonante"),
		strings.Contains(normalized, "ressonador"),
		strings.Contains(normalized, "character"),
		strings.Contains(normalized, "personagem"):
		return "character"
	case strings.Contains(normalized, "weapon"),
		strings.Contains(normalized, "arma"):
		return "weapon"
	default:
		return normalized
	}
}
