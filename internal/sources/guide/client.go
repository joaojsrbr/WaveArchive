package guide

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"wavearchive/internal/domain"
)

const baseURL = "https://guide-server.aki-game.net"

type Client struct{ http *http.Client }

func NewClient(client *http.Client) *Client {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{http: client}
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}
type listItem struct {
	ID                 any         `json:"id"`
	LikeCount          int         `json:"likeCount"`
	Texts              []localized `json:"texts"`
	TeammateRecommends []teamSlot  `json:"teammateRecommends"`
}
type localized struct {
	Language           string `json:"language"`
	IntroductionName   string `json:"introductionName"`
	IntroductionSource string `json:"introductionSource"`
}
type teamSlot struct {
	Main   *role  `json:"main"`
	Spares []role `json:"spares"`
}
type role struct {
	RoleGbID any `json:"roleGbId"`
}

func (c *Client) Fetch(ctx context.Context, characterID int64, language string) ([]domain.CharacterGuide, error) {
	items, usedLanguage, err := c.list(ctx, characterID, language)
	if err != nil && language != "en" {
		items, usedLanguage, err = c.list(ctx, characterID, "en")
	}
	if err != nil {
		return nil, err
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].LikeCount > items[j].LikeCount })
	if len(items) > 5 {
		items = items[:5]
	}
	guides := make([]domain.CharacterGuide, 0, len(items))
	for _, item := range items {
		id := stringID(item.ID)
		text := localizedText(item.Texts, usedLanguage)
		rawDetail, _ := c.info(ctx, id, usedLanguage)
		teams := extractTeams(characterID, item.TeammateRecommends)
		wrapped, _ := json.Marshal(map[string]any{"teams": teams, "detail": json.RawMessage(rawDetail)})
		name := text.IntroductionName
		if name == "" {
			name = "Guide #" + id
		}
		source := text.IntroductionSource
		if source == "" {
			source = "Aki Game"
		}
		guides = append(guides, domain.CharacterGuide{ID: id, CharacterID: characterID, Name: name, Source: source, LikeCount: item.LikeCount, Language: usedLanguage, Teams: teams, DataJSON: string(wrapped)})
	}
	return guides, nil
}
func (c *Client) list(ctx context.Context, characterID int64, language string) ([]listItem, string, error) {
	var env envelope
	path := fmt.Sprintf("%s/introduction/list?roleGbId=%d", baseURL, characterID)
	if err := c.get(ctx, path, language, &env); err != nil {
		return nil, language, err
	}
	if env.Code != 200 || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil, language, errors.New("guide list is empty")
	}
	var items []listItem
	if err := json.Unmarshal(env.Data, &items); err != nil {
		return nil, language, err
	}
	if len(items) == 0 {
		return nil, language, errors.New("guide list is empty")
	}
	return items, language, nil
}
func (c *Client) info(ctx context.Context, id, language string) ([]byte, error) {
	var env envelope
	if err := c.get(ctx, baseURL+"/introduction/info?id="+url.QueryEscape(id), language, &env); err != nil {
		return nil, err
	}
	if env.Code != 200 || len(env.Data) == 0 {
		return nil, errors.New("guide detail is empty")
	}
	return env.Data, nil
}
func (c *Client) get(ctx context.Context, path, language string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-language", language)
	req.Header.Set("User-Agent", "WaveArchive/0.2")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("guide API HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(target)
}
func localizedText(texts []localized, language string) localized {
	for _, t := range texts {
		if t.Language == language {
			return t
		}
	}
	for _, t := range texts {
		if t.Language == "en" {
			return t
		}
	}
	if len(texts) > 0 {
		return texts[0]
	}
	return localized{}
}
func extractTeams(characterID int64, slots []teamSlot) [][]int64 {
	main := []int64{characterID}
	for _, slot := range slots {
		if slot.Main != nil {
			if id := intID(slot.Main.RoleGbID); id > 0 {
				main = append(main, id)
			}
		}
	}
	teams := [][]int64{}
	if len(main) == 3 {
		teams = append(teams, append([]int64{}, main...))
	}
	for index, slot := range slots {
		for _, spare := range slot.Spares {
			if id := intID(spare.RoleGbID); id > 0 && len(main) == 3 {
				alt := append([]int64{}, main...)
				alt[index+1] = id
				teams = append(teams, alt)
			}
		}
	}
	return teams
}
func stringID(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatInt(int64(v), 10)
	case json.Number:
		return v.String()
	}
	return fmt.Sprint(value)
}
func intID(value any) int64 { id, _ := strconv.ParseInt(stringID(value), 10, 64); return id }
