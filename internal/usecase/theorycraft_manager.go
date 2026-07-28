package usecase

import (
	"context"
	"errors"
	"math"
	"strings"

	"wavearchive/internal/domain"
)

type TheorycraftManager struct {
	repository domain.TheorycraftRepository
	evaluator  *BuildEvaluator
}

func NewTheorycraftManager(repository domain.TheorycraftRepository, evaluator *BuildEvaluator) *TheorycraftManager {
	return &TheorycraftManager{repository: repository, evaluator: evaluator}
}

func (m *TheorycraftManager) Team(ctx context.Context, teamID int64) (domain.TeamTheorycraft, error) {
	return m.evaluator.TeamTheorycraft(ctx, teamID)
}

func (m *TheorycraftManager) SaveBuff(ctx context.Context, buff domain.Buff) (domain.Buff, error) {
	buff.Name = strings.TrimSpace(buff.Name)
	buff.Group = strings.TrimSpace(strings.ToLower(buff.Group))
	buff.Scope = strings.TrimSpace(strings.ToUpper(buff.Scope))
	buff.Condition = strings.TrimSpace(buff.Condition)
	if buff.TeamID <= 0 || buff.SourceSlot < 1 || buff.SourceSlot > 3 || buff.TargetSlot < 0 || buff.TargetSlot > 3 {
		return buff, errors.New("invalid buff team or slots")
	}
	if buff.Name == "" || !allowedBuffGroup(buff.Group) || math.IsNaN(buff.Value) || math.IsInf(buff.Value, 0) {
		return buff, errors.New("invalid buff name, group or value")
	}
	if buff.Duration < 0 {
		return buff, errors.New("buff duration cannot be negative")
	}
	if buff.Scope == "" {
		buff.Scope = "TEAM"
	}
	return m.repository.SaveBuff(ctx, buff)
}

func (m *TheorycraftManager) DeleteBuff(ctx context.Context, id int64) error {
	return m.repository.DeleteBuff(ctx, id)
}

func (m *TheorycraftManager) SaveRotation(ctx context.Context, rotation domain.Rotation) (domain.Rotation, error) {
	rotation.Name = strings.TrimSpace(rotation.Name)
	rotation.Notes = strings.TrimSpace(rotation.Notes)
	if rotation.TeamID <= 0 || rotation.Name == "" {
		return rotation, errors.New("rotation team and name are required")
	}
	if rotation.Duration < 0 {
		return rotation, errors.New("rotation duration cannot be negative")
	}
	for index := range rotation.Actions {
		action := &rotation.Actions[index]
		action.Order = index + 1
		action.Name = strings.TrimSpace(action.Name)
		action.ActionType = strings.ToUpper(strings.TrimSpace(action.ActionType))
		if action.Slot < 1 || action.Slot > 3 || action.Name == "" || action.MotionValue < 0 || action.CastTime < 0 {
			return rotation, errors.New("invalid rotation action")
		}
	}
	return m.repository.SaveRotation(ctx, rotation)
}

func (m *TheorycraftManager) DeleteRotation(ctx context.Context, id int64) error {
	return m.repository.DeleteRotation(ctx, id)
}

func (m *TheorycraftManager) EvaluateRotation(ctx context.Context, id int64) (domain.RotationResult, error) {
	return m.evaluator.EvaluateRotation(ctx, id)
}

func allowedBuffGroup(group string) bool {
	switch group {
	case "atk_percent", "hp_percent", "def_percent", "crit_rate", "crit_damage",
		"damage_bonus", "amplification", "special", "resistance_penetration", "defense_ignore":
		return true
	default:
		return false
	}
}
