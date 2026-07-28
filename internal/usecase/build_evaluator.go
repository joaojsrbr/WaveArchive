package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"wavearchive/internal/domain"
)

var statNumber = regexp.MustCompile(`([0-9]+(?:[.,][0-9]+)?)`)

type BuildEvaluator struct {
	builds     domain.BuildRepository
	weapons    domain.WeaponRepository
	theory     domain.TheorycraftRepository
	calculator *DamageCalculator
	teams      domain.TeamRepository
}

func NewBuildEvaluator(builds domain.BuildRepository, weapons domain.WeaponRepository, theory domain.TheorycraftRepository, teams domain.TeamRepository, calculator *DamageCalculator) *BuildEvaluator {
	return &BuildEvaluator{builds: builds, weapons: weapons, theory: theory, teams: teams, calculator: calculator}
}

func (e *BuildEvaluator) SaveConfig(ctx context.Context, config domain.BuildConfig) (domain.BuildEvaluation, error) {
	config.ScalingType = strings.ToUpper(strings.TrimSpace(config.ScalingType))
	if config.ScalingType != "ATK" && config.ScalingType != "HP" && config.ScalingType != "DEF" {
		return domain.BuildEvaluation{}, errors.New("scaling type must be ATK, HP or DEF")
	}
	if config.BuildID <= 0 || config.BaseATK < 0 || config.BaseHP < 0 || config.BaseDEF < 0 || config.MotionValue < 0 {
		return domain.BuildEvaluation{}, errors.New("invalid build configuration")
	}
	if config.ExtraDamageBonusesJSON == "" {
		config.ExtraDamageBonusesJSON = "[]"
	}
	var bonuses []float64
	if err := json.Unmarshal([]byte(config.ExtraDamageBonusesJSON), &bonuses); err != nil {
		return domain.BuildEvaluation{}, errors.New("extra damage bonuses must be a JSON number array")
	}
	if err := e.theory.SaveBuildConfig(ctx, config); err != nil {
		return domain.BuildEvaluation{}, err
	}
	return e.Evaluate(ctx, config.BuildID, nil)
}

func (e *BuildEvaluator) Evaluate(ctx context.Context, buildID int64, buffs []domain.Buff) (domain.BuildEvaluation, error) {
	build, err := e.builds.Get(ctx, buildID)
	if err != nil {
		return domain.BuildEvaluation{}, err
	}
	config, err := e.theory.GetBuildConfig(ctx, buildID)
	if err != nil {
		return domain.BuildEvaluation{}, err
	}
	return e.evaluate(ctx, build, config, buffs)
}

func (e *BuildEvaluator) evaluate(ctx context.Context, build domain.Build, config domain.BuildConfig, buffs []domain.Buff) (domain.BuildEvaluation, error) {
	stats := domain.BuildStats{
		BaseATK: config.BaseATK, BaseHP: config.BaseHP, BaseDEF: config.BaseDEF,
		CritRate: .05, CritDamage: 1.5, EnergyRegen: 1,
		DamageBonuses: map[string]float64{}, UnparsedStats: []string{},
	}
	if build.WeaponID != nil {
		if weapon, err := e.weapons.Get(ctx, *build.WeaponID); err == nil {
			stats.WeaponATK = float64(weapon.BaseATK)
			parseStat(weapon.SubStat, &stats)
		}
	}
	for _, echo := range build.Echoes {
		parseStat(echo.MainStat, &stats)
		var substats []string
		if err := json.Unmarshal([]byte(echo.SubstatsJSON), &substats); err != nil {
			stats.UnparsedStats = append(stats.UnparsedStats, echo.SubstatsJSON)
		} else {
			for _, substat := range substats {
				parseStat(substat, &stats)
			}
		}
	}
	var amplifications, special []float64
	resistancePenetration := 0.0
	defenseIgnore := config.DefenseIgnore
	for _, buff := range buffs {
		if !buff.Active {
			continue
		}
		switch strings.ToLower(buff.Group) {
		case "atk_percent":
			stats.ATKPercent += buff.Value
		case "hp_percent":
			stats.HPPercent += buff.Value
		case "def_percent":
			stats.DEFPercent += buff.Value
		case "crit_rate":
			stats.CritRate += buff.Value
		case "crit_damage":
			stats.CritDamage += buff.Value
		case "damage_bonus":
			stats.DamageBonuses[buff.Name] += buff.Value
		case "amplification":
			amplifications = append(amplifications, buff.Value)
		case "special":
			special = append(special, buff.Value)
		case "resistance_penetration":
			resistancePenetration += buff.Value
		case "defense_ignore":
			defenseIgnore += buff.Value
		}
	}
	stats.TotalATK = (stats.BaseATK+stats.WeaponATK)*(1+stats.ATKPercent) + stats.FlatATK
	stats.TotalHP = stats.BaseHP*(1+stats.HPPercent) + stats.FlatHP
	stats.TotalDEF = stats.BaseDEF*(1+stats.DEFPercent) + stats.FlatDEF
	switch strings.ToUpper(config.ScalingType) {
	case "HP":
		stats.ScalingStat = stats.TotalHP
	case "DEF":
		stats.ScalingStat = stats.TotalDEF
	default:
		stats.ScalingStat = stats.TotalATK
	}
	damageBonuses := make([]float64, 0, len(stats.DamageBonuses))
	for _, value := range stats.DamageBonuses {
		damageBonuses = append(damageBonuses, value)
	}
	var extra []float64
	_ = json.Unmarshal([]byte(config.ExtraDamageBonusesJSON), &extra)
	damageBonuses = append(damageBonuses, extra...)
	damage, err := e.calculator.Calculate(domain.DamageInput{
		ScalingStat: stats.ScalingStat, MotionValue: config.MotionValue, FlatDamage: config.FlatDamage,
		CharacterLevel: build.CharacterLevel, EnemyLevel: config.EnemyLevel,
		EnemyResistance: config.EnemyResistance, ResistancePenetration: resistancePenetration,
		DefenseIgnore: min(1, defenseIgnore), DamageReduction: config.DamageReduction,
		ElementReduction: config.ElementReduction, DamageBonuses: damageBonuses,
		Amplifications: amplifications, SpecialBonuses: special,
		CritRate: min(1, max(0, stats.CritRate)), CritDamage: max(1, stats.CritDamage),
	})
	if err != nil {
		return domain.BuildEvaluation{}, err
	}
	if len(stats.UnparsedStats) > 0 {
		damage.Insights = append(damage.Insights, domain.Insight{
			Severity: "warning", Title: "Atributos não interpretados",
			Message: fmt.Sprintf("%d texto(s) de atributo precisam de revisão manual.", len(stats.UnparsedStats)),
		})
	}
	return domain.BuildEvaluation{Build: build, Config: config, Stats: stats, Damage: damage}, nil
}

func parseStat(raw string, stats *domain.BuildStats) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}
	match := statNumber.FindStringSubmatch(raw)
	if len(match) < 2 {
		stats.UnparsedStats = append(stats.UnparsedStats, raw)
		return
	}
	value, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", "."), 64)
	if err != nil {
		stats.UnparsedStats = append(stats.UnparsedStats, raw)
		return
	}
	upper := strings.ToUpper(raw)
	percent := strings.Contains(raw, "%")
	decimal := value
	if percent {
		decimal /= 100
	}
	switch {
	case strings.Contains(upper, "CRIT") && (strings.Contains(upper, "RATE") || strings.Contains(upper, "TAXA")):
		stats.CritRate += decimal
	case strings.Contains(upper, "CRIT") && (strings.Contains(upper, "DMG") || strings.Contains(upper, "DANO")):
		stats.CritDamage += decimal
	case strings.Contains(upper, "ENERGY") || strings.Contains(upper, "REGEN"):
		stats.EnergyRegen += decimal
	case strings.Contains(upper, "DMG") || strings.Contains(upper, "DANO") || strings.Contains(upper, "BONUS"):
		stats.DamageBonuses[statLabel(upper)] += decimal
	case strings.Contains(upper, "ATK"):
		if percent {
			stats.ATKPercent += decimal
		} else {
			stats.FlatATK += value
		}
	case strings.Contains(upper, "HP") || strings.Contains(upper, "VIDA"):
		if percent {
			stats.HPPercent += decimal
		} else {
			stats.FlatHP += value
		}
	case strings.Contains(upper, "DEF"):
		if percent {
			stats.DEFPercent += decimal
		} else {
			stats.FlatDEF += value
		}
	default:
		stats.UnparsedStats = append(stats.UnparsedStats, raw)
	}
}

func statLabel(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 48 {
		return value[:48]
	}
	return value
}

func (e *BuildEvaluator) TeamTheorycraft(ctx context.Context, teamID int64) (domain.TeamTheorycraft, error) {
	team, err := e.teams.Get(ctx, teamID)
	if err != nil {
		return domain.TeamTheorycraft{}, err
	}
	buffs, err := e.theory.ListBuffs(ctx, teamID)
	if err != nil {
		return domain.TeamTheorycraft{}, err
	}
	rotations, err := e.theory.ListRotations(ctx, teamID)
	if err != nil {
		return domain.TeamTheorycraft{}, err
	}
	warnings := []string{}
	for _, member := range team.Members {
		if member.BuildID == nil {
			warnings = append(warnings, fmt.Sprintf("%s não possui build associada.", member.CharacterName))
		}
	}
	for _, buff := range buffs {
		if buff.Condition != "" {
			status := "ativa no cenário"
			if !buff.Active {
				status = "inativa no cenário"
			}
			warnings = append(warnings, fmt.Sprintf("%s depende da condição (%s): %s", buff.Name, status, buff.Condition))
		}
	}
	return domain.TeamTheorycraft{Team: team, Buffs: buffs, Rotations: rotations, Warnings: warnings}, nil
}

func (e *BuildEvaluator) EvaluateRotation(ctx context.Context, rotationID int64) (domain.RotationResult, error) {
	selected, team, err := e.findRotation(ctx, rotationID)
	if err != nil {
		return domain.RotationResult{}, err
	}
	buffs, err := e.theory.ListBuffs(ctx, team.ID)
	if err != nil {
		return domain.RotationResult{}, err
	}
	result := domain.RotationResult{Rotation: selected, Actions: []domain.RotationActionResult{},
		EnergyBySlot: map[int]float64{}, ConcertoBySlot: map[int]float64{}, FieldTimeBySlot: map[int]float64{},
		Warnings: []string{}, Errors: []string{}}
	cursor := 0.0
	lastUse := map[string]float64{}
	activatedAt := map[int64]float64{}
	for _, buff := range buffs {
		if buff.Active && strings.TrimSpace(buff.TriggerAction) == "" {
			activatedAt[buff.ID] = 0
		}
	}
	for _, action := range selected.Actions {
		start, end := cursor, cursor+action.CastTime
		cursor = end
		if action.Slot < 1 || action.Slot > len(team.Members) {
			result.Errors = append(result.Errors, fmt.Sprintf("Ação %s possui slot inválido.", action.Name))
			continue
		}
		key := fmt.Sprintf("%d:%s", action.Slot, strings.ToUpper(action.ActionType))
		if previous, ok := lastUse[key]; ok && action.Cooldown > 0 && start-previous < action.Cooldown {
			result.Errors = append(result.Errors, fmt.Sprintf("%s ainda possui %.2fs de cooldown.", action.Name, action.Cooldown-(start-previous)))
		}
		lastUse[key] = start
		for _, buff := range buffs {
			trigger := strings.TrimSpace(strings.ToUpper(buff.TriggerAction))
			if buff.Active && buff.SourceSlot == action.Slot && trigger != "" && (trigger == strings.ToUpper(action.ActionType) || trigger == strings.ToUpper(action.Name)) {
				activatedAt[buff.ID] = end
			}
		}
		if action.EnergyCost > result.EnergyBySlot[action.Slot] {
			result.Errors = append(result.Errors, fmt.Sprintf("%s requer %.0f de energia; disponível: %.0f.", action.Name, action.EnergyCost, result.EnergyBySlot[action.Slot]))
		} else {
			result.EnergyBySlot[action.Slot] -= action.EnergyCost
		}
		result.EnergyBySlot[action.Slot] += action.Energy
		result.ConcertoBySlot[action.Slot] += action.Concerto
		result.FieldTimeBySlot[action.Slot] += action.CastTime
		actionResult := domain.RotationActionResult{Action: action, StartTime: start, EndTime: end, ActiveBuffs: []string{}, ExpiredBuffs: []string{}}
		member := team.Members[action.Slot-1]
		if member.BuildID == nil {
			if action.MotionValue > 0 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s não tem build para calcular %s.", member.CharacterName, action.Name))
			}
			result.Actions = append(result.Actions, actionResult)
			continue
		}
		config, err := e.theory.GetBuildConfig(ctx, *member.BuildID)
		if err != nil {
			return result, err
		}
		config.MotionValue = action.MotionValue
		build, err := e.builds.Get(ctx, *member.BuildID)
		if err != nil {
			return result, err
		}
		active := []domain.Buff{}
		for _, buff := range buffs {
			activation, triggered := activatedAt[buff.ID]
			inWindow := triggered && (buff.Duration <= 0 || start <= activation+buff.Duration)
			if buff.Active && inWindow && (buff.TargetSlot == 0 || buff.TargetSlot == action.Slot) {
				active = append(active, buff)
				actionResult.ActiveBuffs = append(actionResult.ActiveBuffs, buff.Name)
			} else if buff.Active && triggered && buff.Duration > 0 && start > activation+buff.Duration {
				actionResult.ExpiredBuffs = append(actionResult.ExpiredBuffs, buff.Name)
			}
		}
		evaluation, err := e.evaluate(ctx, build, config, active)
		if err != nil {
			return result, err
		}
		actionResult.Damage = evaluation.Damage.ExpectedDamage
		result.TotalDamage += actionResult.Damage
		result.Actions = append(result.Actions, actionResult)
	}
	result.Duration = cursor
	if selected.Duration > 0 {
		if selected.Duration < cursor {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Duração fixa %.2fs menor que ações %.2fs.", selected.Duration, cursor))
		}
		result.Duration = selected.Duration
	}
	if result.Duration > 0 {
		result.DPS = result.TotalDamage / result.Duration
	} else {
		result.Errors = append(result.Errors, "Informe duração para calcular DPS.")
	}
	return result, nil
}

func (e *BuildEvaluator) findRotation(ctx context.Context, id int64) (domain.Rotation, domain.Team, error) {
	teams, err := e.teams.List(ctx)
	if err != nil {
		return domain.Rotation{}, domain.Team{}, err
	}
	for _, team := range teams {
		rotations, err := e.theory.ListRotations(ctx, team.ID)
		if err != nil {
			return domain.Rotation{}, domain.Team{}, err
		}
		for _, rotation := range rotations {
			if rotation.ID == id {
				return rotation, team, nil
			}
		}
	}
	return domain.Rotation{}, domain.Team{}, errors.New("rotation not found")
}
