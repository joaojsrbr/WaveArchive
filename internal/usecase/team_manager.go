package usecase

import (
	"context"
	"errors"
	"strings"

	"wavearchive/internal/domain"
)

type TeamManager struct {
	repository domain.TeamRepository
	builds     domain.BuildRepository
}

func NewTeamManager(repository domain.TeamRepository, builds ...domain.BuildRepository) *TeamManager {
	manager := &TeamManager{repository: repository}
	if len(builds) > 0 {
		manager.builds = builds[0]
	}
	return manager
}

func (m *TeamManager) List(ctx context.Context) ([]domain.Team, error) {
	return m.repository.List(ctx)
}

func (m *TeamManager) Save(ctx context.Context, team domain.Team) (domain.Team, error) {
	team.Name = strings.TrimSpace(team.Name)
	team.Notes = strings.TrimSpace(team.Notes)
	if team.Name == "" {
		return domain.Team{}, errors.New("team name is required")
	}
	if len(team.Members) != 3 {
		return domain.Team{}, errors.New("a team must have exactly three characters")
	}
	seen := make(map[int64]struct{}, 3)
	for index := range team.Members {
		member := &team.Members[index]
		member.Slot = index + 1
		member.Role = strings.TrimSpace(member.Role)
		member.CustomRole = strings.TrimSpace(member.CustomRole)
		if member.CharacterID <= 0 {
			return domain.Team{}, errors.New("all three character slots are required")
		}
		if _, exists := seen[member.CharacterID]; exists {
			return domain.Team{}, errors.New("a character cannot occupy two team slots")
		}
		if member.BuildID != nil && m.builds != nil {
			build, err := m.builds.Get(ctx, *member.BuildID)
			if err != nil {
				return domain.Team{}, errors.New("linked build is unavailable")
			}
			if build.CharacterID != member.CharacterID {
				return domain.Team{}, errors.New("linked build belongs to another character")
			}
		}
		seen[member.CharacterID] = struct{}{}
	}
	return m.repository.Save(ctx, team)
}

func (m *TeamManager) Duplicate(ctx context.Context, id int64) (domain.Team, error) {
	return m.repository.Duplicate(ctx, id)
}
func (m *TeamManager) Delete(ctx context.Context, id int64) error {
	return m.repository.SoftDelete(ctx, id)
}
func (m *TeamManager) Restore(ctx context.Context, id int64) error {
	return m.repository.Restore(ctx, id)
}
