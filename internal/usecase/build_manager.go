package usecase

import (
	"context"
	"errors"
	"strings"

	"wavearchive/internal/domain"
)

type BuildManager struct {
	repository domain.BuildRepository
}

func NewBuildManager(repository domain.BuildRepository) *BuildManager {
	return &BuildManager{repository: repository}
}

func (m *BuildManager) List(ctx context.Context) ([]domain.Build, error) {
	return m.repository.List(ctx)
}

func (m *BuildManager) Save(ctx context.Context, build domain.Build) (domain.Build, error) {
	build.Name = strings.TrimSpace(build.Name)
	if build.Name == "" {
		return domain.Build{}, errors.New("build name is required")
	}
	if build.CharacterID <= 0 {
		return domain.Build{}, errors.New("character is required")
	}
	if build.CharacterLevel < 1 || build.CharacterLevel > 90 {
		return domain.Build{}, errors.New("character level must be between 1 and 90")
	}
	if build.Sequence < 0 || build.Sequence > 6 {
		return domain.Build{}, errors.New("sequence must be between 0 and 6")
	}
	if build.WeaponLevel < 1 || build.WeaponLevel > 90 {
		return domain.Build{}, errors.New("weapon level must be between 1 and 90")
	}
	if build.WeaponRank < 1 || build.WeaponRank > 5 {
		return domain.Build{}, errors.New("weapon rank must be between 1 and 5")
	}
	if len(build.Echoes) > 5 {
		return domain.Build{}, errors.New("a build can have at most five echoes")
	}
	echoIDs := make(map[int64]struct{}, len(build.Echoes))
	totalEchoCost := 0
	for _, echo := range build.Echoes {
		if echo.ID <= 0 {
			return domain.Build{}, errors.New("invalid owned echo")
		}
		if echo.Cost <= 0 {
			return domain.Build{}, errors.New("invalid echo cost")
		}
		totalEchoCost += echo.Cost
		if _, exists := echoIDs[echo.ID]; exists {
			return domain.Build{}, errors.New("the same echo cannot occupy two build slots")
		}
		echoIDs[echo.ID] = struct{}{}
	}
	if totalEchoCost > 12 {
		return domain.Build{}, errors.New("echo cost cannot exceed 12")
	}
	return m.repository.Save(ctx, build)
}

func (m *BuildManager) Duplicate(ctx context.Context, id int64) (domain.Build, error) {
	return m.repository.Duplicate(ctx, id)
}

func (m *BuildManager) Delete(ctx context.Context, id int64) error {
	return m.repository.SoftDelete(ctx, id)
}

func (m *BuildManager) Restore(ctx context.Context, id int64) error {
	return m.repository.Restore(ctx, id)
}

func (m *BuildManager) History(ctx context.Context, id int64) ([]domain.BuildVersion, error) {
	return m.repository.History(ctx, id)
}
