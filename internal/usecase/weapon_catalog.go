package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"sync"

	"wavearchive/internal/assets"
	"wavearchive/internal/domain"
	"wavearchive/internal/sources/nanoka"
	"wavearchive/internal/sources/nanoka/mapper"
)

type WeaponCatalog struct {
	repository domain.WeaponRepository
	source     *nanoka.Client
	assets     *assets.Cache
	logger     *slog.Logger
}

func NewWeaponCatalog(repository domain.WeaponRepository, source *nanoka.Client, assetCache *assets.Cache, logger *slog.Logger) *WeaponCatalog {
	return &WeaponCatalog{repository: repository, source: source, assets: assetCache, logger: logger}
}

func (c *WeaponCatalog) List(ctx context.Context, filter domain.WeaponFilter) ([]domain.Weapon, error) {
	return c.repository.List(ctx, filter)
}

func (c *WeaponCatalog) Get(ctx context.Context, id int64) (domain.Weapon, error) {
	return c.repository.Get(ctx, id)
}

func (c *WeaponCatalog) UpdateAccount(ctx context.Context, update domain.WeaponAccountUpdate) error {
	if update.WeaponID <= 0 {
		return errors.New("invalid weapon id")
	}
	if update.Level < 1 || update.Level > 90 {
		return errors.New("weapon level must be between 1 and 90")
	}
	if update.Rank < 1 || update.Rank > 5 {
		return errors.New("weapon rank must be between 1 and 5")
	}
	return c.repository.UpdateAccount(ctx, update)
}

func (c *WeaponCatalog) Sync(ctx context.Context, version string, progress func(string, int)) (int, error) {
	progress("weapon_index", 5)
	index, err := c.source.WeaponIndex(ctx, version)
	if err != nil {
		return 0, fmt.Errorf("download weapon index: %w", err)
	}
	weapons := make([]domain.Weapon, 0, len(index))
	for id, entry := range index {
		if weapon, ok := mapper.WeaponIndex(id, entry, version); ok {
			weapons = append(weapons, weapon)
		}
	}
	sort.Slice(weapons, func(i, j int) bool { return weapons[i].ID < weapons[j].ID })

	progress("weapon_details", 15)
	c.loadDetails(ctx, version, weapons, progress)
	progress("weapon_assets", 70)
	c.loadAssets(ctx, weapons, progress)
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	progress("weapon_saving", 92)
	if err := c.repository.ReplaceSynced(ctx, version, weapons); err != nil {
		return 0, err
	}
	progress("weapon_done", 100)
	return len(weapons), nil
}

func (c *WeaponCatalog) loadDetails(ctx context.Context, version string, weapons []domain.Weapon, progress func(string, int)) {
	type result struct {
		index  int
		weapon domain.Weapon
		err    error
	}
	jobs := make(chan int)
	results := make(chan result)
	var group sync.WaitGroup
	workers := min(6, len(weapons))
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range jobs {
				source, err := c.source.Weapon(ctx, version, "en", weapons[index].ID)
				if err != nil {
					results <- result{index: index, weapon: weapons[index], err: err}
					continue
				}
				results <- result{index: index, weapon: mapper.MergeWeaponDetail(weapons[index], source)}
			}
		}()
	}
	go sendIndexes(ctx, len(weapons), jobs)
	go func() {
		group.Wait()
		close(results)
	}()
	completed := 0
	for result := range results {
		if result.err != nil {
			c.logger.Warn("weapon detail unavailable", "id", result.weapon.ID, "error", result.err)
		} else {
			weapons[result.index] = result.weapon
		}
		completed++
		if len(weapons) > 0 {
			progress("weapon_details", 15+completed*53/len(weapons))
		}
	}
}

func (c *WeaponCatalog) loadAssets(ctx context.Context, weapons []domain.Weapon, progress func(string, int)) {
	if c.assets == nil {
		return
	}
	type result struct {
		index int
		url   string
		err   error
	}
	jobs := make(chan int)
	results := make(chan result)
	var group sync.WaitGroup
	workers := min(6, len(weapons))
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range jobs {
				weapon := weapons[index]
				url, err := c.assets.Ensure(ctx, weapon.IconPath, fmt.Sprintf("weapons/%d/icon.webp", weapon.ID))
				results <- result{index: index, url: url, err: err}
			}
		}()
	}
	go sendIndexes(ctx, len(weapons), jobs)
	go func() {
		group.Wait()
		close(results)
	}()
	completed := 0
	for result := range results {
		if result.err != nil {
			c.logger.Warn("weapon asset unavailable", "id", weapons[result.index].ID, "error", result.err)
		} else if result.url != "" {
			weapons[result.index].IconPath = result.url
		}
		completed++
		if len(weapons) > 0 {
			progress("weapon_assets", 70+completed*20/len(weapons))
		}
	}
}

func sendIndexes(ctx context.Context, count int, jobs chan<- int) {
	defer close(jobs)
	for index := 0; index < count; index++ {
		select {
		case jobs <- index:
		case <-ctx.Done():
			return
		}
	}
}
