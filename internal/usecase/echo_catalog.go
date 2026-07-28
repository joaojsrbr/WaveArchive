package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"

	"wavearchive/internal/assets"
	"wavearchive/internal/domain"
	"wavearchive/internal/sources/nanoka"
	"wavearchive/internal/sources/nanoka/mapper"
)

type EchoCatalog struct {
	repository domain.EchoRepository
	source     *nanoka.Client
	assets     *assets.Cache
	logger     *slog.Logger
}

func NewEchoCatalog(repository domain.EchoRepository, source *nanoka.Client, assetCache *assets.Cache, logger *slog.Logger) *EchoCatalog {
	return &EchoCatalog{repository: repository, source: source, assets: assetCache, logger: logger}
}

func (c *EchoCatalog) List(ctx context.Context, filter domain.EchoFilter) ([]domain.Echo, error) {
	return c.repository.List(ctx, filter)
}

func (c *EchoCatalog) Get(ctx context.Context, id int64) (domain.Echo, error) {
	return c.repository.Get(ctx, id)
}

func (c *EchoCatalog) ListSonatas(ctx context.Context) ([]domain.Sonata, error) {
	return c.repository.ListSonatas(ctx)
}

func (c *EchoCatalog) ListOwned(ctx context.Context) ([]domain.OwnedEcho, error) {
	return c.repository.ListOwned(ctx)
}

func (c *EchoCatalog) SaveOwned(ctx context.Context, item domain.OwnedEcho) (domain.OwnedEcho, error) {
	item.MainStat = strings.TrimSpace(item.MainStat)
	item.Note = strings.TrimSpace(item.Note)
	if item.EchoID <= 0 {
		return domain.OwnedEcho{}, errors.New("echo is required")
	}
	if item.Level < 0 || item.Level > 25 {
		return domain.OwnedEcho{}, errors.New("echo level must be between 0 and 25")
	}
	if item.SubstatsJSON == "" {
		item.SubstatsJSON = "[]"
	}
	return c.repository.SaveOwned(ctx, item)
}

func (c *EchoCatalog) DeleteOwned(ctx context.Context, id int64) error {
	return c.repository.DeleteOwned(ctx, id)
}

func (c *EchoCatalog) Sync(ctx context.Context, version string, progress func(string, int)) (int, error) {
	progress("echo_index", 3)
	index, err := c.source.EchoIndex(ctx, version)
	if err != nil {
		return 0, fmt.Errorf("download echo index: %w", err)
	}
	sonataIndex, err := c.source.SonataIndex(ctx, version)
	if err != nil {
		return 0, fmt.Errorf("download sonata index: %w", err)
	}

	echoes := make([]domain.Echo, 0, len(index))
	for id, entry := range index {
		if echo, ok := mapper.EchoIndex(id, entry, version); ok {
			echoes = append(echoes, echo)
		}
	}
	sort.Slice(echoes, func(i, j int) bool { return echoes[i].ID < echoes[j].ID })
	sonatas := make([]domain.Sonata, 0, len(sonataIndex))
	for _, entry := range sonataIndex {
		sonatas = append(sonatas, mapper.Sonata(entry, version))
	}
	sort.Slice(sonatas, func(i, j int) bool { return sonatas[i].ID < sonatas[j].ID })

	progress("echo_details", 10)
	c.loadEchoDetails(ctx, version, echoes, progress)
	progress("echo_assets", 70)
	c.loadEchoAssets(ctx, echoes, sonatas, progress)
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	progress("echo_saving", 94)
	if err := c.repository.ReplaceSynced(ctx, version, echoes, sonatas); err != nil {
		return 0, err
	}
	progress("echo_done", 100)
	return len(echoes), nil
}

func (c *EchoCatalog) loadEchoDetails(ctx context.Context, version string, echoes []domain.Echo, progress func(string, int)) {
	type result struct {
		index int
		echo  domain.Echo
		err   error
	}
	jobs := make(chan int)
	results := make(chan result)
	var group sync.WaitGroup
	workers := min(8, len(echoes))
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range jobs {
				detail, err := c.source.EchoDetail(ctx, version, "en", echoes[index].ID)
				if err != nil {
					results <- result{index: index, echo: echoes[index], err: err}
					continue
				}
				results <- result{index: index, echo: mapper.MergeEchoDetail(echoes[index], detail)}
			}
		}()
	}
	go sendIndexes(ctx, len(echoes), jobs)
	go func() {
		group.Wait()
		close(results)
	}()
	completed := 0
	for result := range results {
		if result.err != nil {
			c.logger.Warn("echo detail unavailable", "id", result.echo.ID, "error", result.err)
		} else {
			echoes[result.index] = result.echo
		}
		completed++
		if len(echoes) > 0 {
			progress("echo_details", 10+completed*58/len(echoes))
		}
	}
}

func (c *EchoCatalog) loadEchoAssets(ctx context.Context, echoes []domain.Echo, sonatas []domain.Sonata, progress func(string, int)) {
	if c.assets == nil {
		return
	}
	type assetJob struct {
		kind  string
		index int
	}
	type result struct {
		job assetJob
		url string
		err error
	}
	total := len(echoes) + len(sonatas)
	jobs := make(chan assetJob)
	results := make(chan result)
	var group sync.WaitGroup
	workers := min(8, total)
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for job := range jobs {
				if job.kind == "echo" {
					echo := echoes[job.index]
					url, err := c.assets.Ensure(ctx, echo.IconPath, fmt.Sprintf("echoes/%d/icon.webp", echo.ID))
					results <- result{job: job, url: url, err: err}
				} else {
					set := sonatas[job.index]
					url, err := c.assets.Ensure(ctx, set.IconPath, fmt.Sprintf("sonatas/%d/icon.webp", set.ID))
					results <- result{job: job, url: url, err: err}
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for index := range echoes {
			select {
			case jobs <- assetJob{kind: "echo", index: index}:
			case <-ctx.Done():
				return
			}
		}
		for index := range sonatas {
			select {
			case jobs <- assetJob{kind: "sonata", index: index}:
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
		if result.job.kind == "echo" {
			if result.err != nil {
				c.logger.Warn("echo asset unavailable", "id", echoes[result.job.index].ID, "error", result.err)
			} else if result.url != "" {
				echoes[result.job.index].IconPath = result.url
			}
		} else {
			if result.err != nil {
				c.logger.Warn("sonata asset unavailable", "id", sonatas[result.job.index].ID, "error", result.err)
			} else if result.url != "" {
				sonatas[result.job.index].IconPath = result.url
			}
		}
		completed++
		if total > 0 {
			progress("echo_assets", 70+completed*22/total)
		}
	}
}
