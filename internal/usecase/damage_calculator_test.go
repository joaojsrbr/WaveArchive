package usecase

import (
	"math"
	"testing"

	"wavearchive/internal/domain"
)

func TestDamageCalculatorMatchesFormulaGroups(t *testing.T) {
	calculator := NewDamageCalculator()
	result, err := calculator.Calculate(domain.DamageInput{
		ScalingStat: 2000, MotionValue: 2, FlatDamage: 100,
		CharacterLevel: 90, EnemyLevel: 90, EnemyResistance: .1,
		DamageBonuses: []float64{.4, .2}, Amplifications: []float64{.15},
		SpecialBonuses: []float64{.1}, CritRate: .75, CritDamage: 2.5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.BaseDamage != 4100 {
		t.Fatalf("base damage = %v", result.BaseDamage)
	}
	if math.Abs(result.ResistanceMultiplier-.9) > 1e-9 || math.Abs(result.DefenseMultiplier-.5) > 1e-9 {
		t.Fatalf("unexpected enemy multipliers: %#v", result)
	}
	wantNonCrit := 4100 * .9 * .5 * 1.6 * 1.15 * 1.1
	if math.Abs(result.NonCriticalDamage-wantNonCrit) > 1e-6 {
		t.Fatalf("non-critical = %v, want %v", result.NonCriticalDamage, wantNonCrit)
	}
	wantExpected := wantNonCrit * (1 + .75*(2.5-1))
	if math.Abs(result.ExpectedDamage-wantExpected) > 1e-6 {
		t.Fatalf("expected = %v, want %v", result.ExpectedDamage, wantExpected)
	}
}

func TestResistanceMultiplierBranches(t *testing.T) {
	tests := []struct{ resistance, want float64 }{
		{-.2, 1.1}, {.1, .9}, {.8, .2},
	}
	for _, test := range tests {
		if got := resistanceMultiplier(test.resistance); math.Abs(got-test.want) > 1e-9 {
			t.Fatalf("resistanceMultiplier(%v) = %v, want %v", test.resistance, got, test.want)
		}
	}
}
