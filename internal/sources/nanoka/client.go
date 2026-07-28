package nanoka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"wavearchive/internal/sources/nanoka/dto"
)

const baseURL = "https://static.nanoka.cc"

type Client struct {
	http *http.Client
}

func NewClient(client *http.Client) *Client {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{http: client}
}

func (c *Client) DetectVersion(ctx context.Context) (string, error) {
	var manifest dto.Manifest
	if err := c.getJSON(ctx, baseURL+"/manifest.json", &manifest); err != nil {
		return "", err
	}
	version := strings.Split(manifest.WW.Latest, "+")[0]
	if version == "" {
		return "", errors.New("Nanoka manifest did not provide ww.latest")
	}
	return version, nil
}

func (c *Client) CharacterIndex(ctx context.Context, version string) (map[string]dto.CharacterIndexEntry, error) {
	var ordered orderedCharacterIndex
	url := fmt.Sprintf("%s/ww/%s/character.json", baseURL, version)
	if err := c.getJSON(ctx, url, &ordered); err != nil {
		return nil, err
	}
	return map[string]dto.CharacterIndexEntry(ordered), nil
}

type orderedCharacterIndex map[string]dto.CharacterIndexEntry

func (index *orderedCharacterIndex) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return errors.New("character index must be a JSON object")
	}
	result := orderedCharacterIndex{}
	order := 0
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := keyToken.(string)
		if !ok {
			return errors.New("character index contains a non-string key")
		}
		var entry dto.CharacterIndexEntry
		if err := decoder.Decode(&entry); err != nil {
			return err
		}
		entry.APIOrder = order
		result[key] = entry
		order++
	}
	if _, err := decoder.Token(); err != nil {
		return err
	}
	*index = result
	return nil
}

func (c *Client) ItemCatalog(ctx context.Context, version, language string) (map[string]dto.Item, error) {
	var items map[string]dto.Item
	url := fmt.Sprintf("%s/ww/%s/%s/item_all.json", baseURL, version, language)
	if err := c.getJSON(ctx, url, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (c *Client) ItemIndex(ctx context.Context, version, language string) (map[string]dto.ItemIndexEntry, error) {
	var items map[string]dto.ItemIndexEntry
	url := fmt.Sprintf("%s/ww/%s/%s/item.json", baseURL, version, language)
	if err := c.getJSON(ctx, url, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (c *Client) WeaponIndex(ctx context.Context, version string) (map[string]dto.WeaponIndexEntry, error) {
	var index map[string]dto.WeaponIndexEntry
	url := fmt.Sprintf("%s/ww/%s/weapon.json", baseURL, version)
	if err := c.getJSON(ctx, url, &index); err != nil {
		return nil, err
	}
	return index, nil
}

func (c *Client) EchoIndex(ctx context.Context, version string) (map[string]dto.EchoIndexEntry, error) {
	var index map[string]dto.EchoIndexEntry
	if err := c.getJSON(ctx, fmt.Sprintf("%s/ww/%s/echo.json", baseURL, version), &index); err != nil {
		return nil, err
	}
	return index, nil
}

func (c *Client) SonataIndex(ctx context.Context, version string) (map[string]dto.SonataIndexEntry, error) {
	var index map[string]dto.SonataIndexEntry
	if err := c.getJSON(ctx, fmt.Sprintf("%s/ww/%s/sonata.json", baseURL, version), &index); err != nil {
		return nil, err
	}
	return index, nil
}

func (c *Client) EchoDetail(ctx context.Context, version, language string, id int64) (dto.EchoDetail, error) {
	var detail dto.EchoDetail
	url := fmt.Sprintf("%s/ww/%s/%s/echo/%d.json", baseURL, version, language, id)
	if err := c.getJSON(ctx, url, &detail); err != nil {
		return dto.EchoDetail{}, err
	}
	return detail, nil
}

func (c *Client) CharacterDetail(ctx context.Context, version, language string, id int64) (dto.CharacterDetail, error) {
	var detail dto.CharacterDetail
	url := fmt.Sprintf("%s/ww/%s/%s/character/%d.json", baseURL, version, language, id)
	if err := c.getJSON(ctx, url, &detail); err != nil {
		return dto.CharacterDetail{}, err
	}
	return detail, nil
}

func (c *Client) Weapon(ctx context.Context, version, language string, id int64) (dto.Weapon, error) {
	var weapon dto.Weapon
	url := fmt.Sprintf("%s/ww/%s/%s/weapon/%d.json", baseURL, version, language, id)
	if err := c.getJSON(ctx, url, &weapon); err != nil {
		return dto.Weapon{}, err
	}
	return weapon, nil
}

func (c *Client) getJSON(ctx context.Context, url string, target any) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			timer := time.NewTimer(time.Duration(attempt) * 350 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "WaveArchive/0.1 (+https://static.nanoka.cc)")
		req.Header.Set("Accept", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			resp.Body.Close()
			return fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
			continue
		}
		err = json.NewDecoder(io.LimitReader(resp.Body, 32<<20)).Decode(target)
		resp.Body.Close()
		if err == nil {
			return nil
		}
		lastErr = err
	}
	return fmt.Errorf("fetch JSON: %w", lastErr)
}
