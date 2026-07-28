package usecase

import (
	"errors"
	"math"

	"wavearchive/internal/domain"
)

type DamageCalculator struct{}

func NewDamageCalculator() *DamageCalculator { return &DamageCalculator{} }

func (c *DamageCalculator) Calculate(input domain.DamageInput) (domain.DamageResult, error) {
	if err := validateDamageInput(input); err != nil {
		return domain.DamageResult{}, err
	}
	effectiveResistance := input.EnemyResistance - input.ResistancePenetration
	resistanceMultiplier := resistanceMultiplier(effectiveResistance)
	defenseMultiplier := (800 + 8*float64(input.CharacterLevel)) /
		(800 + 8*float64(input.CharacterLevel) + (800+8*float64(input.EnemyLevel))*(1-input.DefenseIgnore))
	damageReductionMultiplier := math.Max(0, 1-input.DamageReduction-input.AdditionalDamageReduction)
	elementReductionMultiplier := math.Max(0, 1-input.ElementReduction-input.AdditionalElementReduction)
	damageBonusMultiplier := 1 + sum(input.DamageBonuses)
	amplificationMultiplier := 1 + sum(input.Amplifications)
	specialDamageMultiplier := 1 + sum(input.SpecialBonuses)
	baseDamage := input.ScalingStat*input.MotionValue + input.FlatDamage + input.FlatBonusDamage
	nonCritical := baseDamage * resistanceMultiplier * defenseMultiplier *
		damageReductionMultiplier * elementReductionMultiplier * damageBonusMultiplier *
		amplificationMultiplier * specialDamageMultiplier
	critical := nonCritical * input.CritDamage
	expected := nonCritical * (1 + input.CritRate*(input.CritDamage-1))
	result := domain.DamageResult{
		FormulaVersion:    "WaveArchive 1.0",
		FormulaConfidence: "community_tested",
		BaseDamage:        baseDamage, EffectiveResistance: effectiveResistance,
		ResistanceMultiplier: resistanceMultiplier, DefenseMultiplier: defenseMultiplier,
		DamageReductionMultiplier: damageReductionMultiplier, ElementReductionMultiplier: elementReductionMultiplier,
		DamageBonusMultiplier: damageBonusMultiplier, AmplificationMultiplier: amplificationMultiplier,
		SpecialDamageMultiplier: specialDamageMultiplier, NonCriticalDamage: nonCritical,
		CriticalDamage: critical, ExpectedDamage: expected,
	}
	result.Insights = damageInsights(input, result)
	return result, nil
}

func resistanceMultiplier(resistance float64) float64 {
	switch {
	case resistance < 0:
		return 1 - resistance/2
	case resistance < .8:
		return 1 - resistance
	default:
		return 1 / (1 + 5*resistance)
	}
}

func validateDamageInput(input domain.DamageInput) error {
	values := []float64{
		input.ScalingStat, input.MotionValue, input.FlatDamage, input.FlatBonusDamage,
		input.EnemyResistance, input.ResistancePenetration, input.DefenseIgnore,
		input.DamageReduction, input.AdditionalDamageReduction, input.ElementReduction,
		input.AdditionalElementReduction, input.CritRate, input.CritDamage,
	}
	values = append(values, input.DamageBonuses...)
	values = append(values, input.Amplifications...)
	values = append(values, input.SpecialBonuses...)
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return errors.New("damage input contains a non-finite number")
		}
	}
	if input.ScalingStat < 0 || input.MotionValue < 0 || input.FlatDamage < 0 || input.FlatBonusDamage < 0 {
		return errors.New("base damage values cannot be negative")
	}
	if input.CharacterLevel < 1 || input.CharacterLevel > 90 || input.EnemyLevel < 1 || input.EnemyLevel > 120 {
		return errors.New("character or enemy level is outside the supported range")
	}
	if input.DefenseIgnore < 0 || input.DefenseIgnore > 1 {
		return errors.New("defense ignore must be between 0 and 100%")
	}
	if input.CritRate < 0 || input.CritRate > 1 {
		return errors.New("critical rate must be between 0 and 100%")
	}
	if input.CritDamage < 1 {
		return errors.New("critical damage must include the base 100%")
	}
	if 1+sum(input.DamageBonuses) < 0 || 1+sum(input.Amplifications) < 0 || 1+sum(input.SpecialBonuses) < 0 {
		return errors.New("a player multiplier cannot be negative")
	}
	return nil
}

func damageInsights(input domain.DamageInput, result domain.DamageResult) []domain.Insight {
	insights := []domain.Insight{{
		Severity: "info", Title: "Resultado determinístico",
		Message: "A análise usa exatamente os multiplicadores exibidos; nenhuma IA altera o valor calculado.",
	}}
	critValue := input.CritRate * (input.CritDamage - 1)
	if input.CritRate < .5 {
		insights = append(insights, domain.Insight{Severity: "warning", Title: "Taxa crítica baixa",
			Message: "Aumentar a taxa crítica tende a estabilizar o dano esperado antes de investir apenas em dano crítico."})
	} else if input.CritDamage < 2 {
		insights = append(insights, domain.Insight{Severity: "tip", Title: "Espaço para dano crítico",
			Message: "A taxa crítica já é consistente; dano crítico pode oferecer um ganho eficiente."})
	}
	if result.EffectiveResistance > .4 {
		insights = append(insights, domain.Insight{Severity: "warning", Title: "Resistência é o principal freio",
			Message: "Penetração ou redução de resistência terá impacto multiplicativo relevante neste alvo."})
	} else if result.EffectiveResistance < 0 {
		insights = append(insights, domain.Insight{Severity: "tip", Title: "Resistência abaixo de zero",
			Message: "A penetração já ultrapassou a resistência do alvo e continua rendendo metade do excedente."})
	}
	if sum(input.DamageBonuses) > 1 && sum(input.Amplifications) < .1 {
		insights = append(insights, domain.Insight{Severity: "tip", Title: "Bônus concentrados no grupo aditivo",
			Message: "Uma fonte de Amplify/Deepen provavelmente vale mais que outro bônus aditivo de dano equivalente."})
	}
	if input.DefenseIgnore > 0 {
		insights = append(insights, domain.Insight{Severity: "info", Title: "DEF Ignore aplicado",
			Message: "A defesa ignorada foi aplicada somente ao termo defensivo do inimigo, conforme a fórmula."})
	}
	if critValue < .25 {
		insights = append(insights, domain.Insight{Severity: "tip", Title: "Baixa contribuição crítica",
			Message: "O multiplicador crítico médio está contribuindo menos de 25% ao dano esperado."})
	}
	return insights
}

func sum(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total
}
