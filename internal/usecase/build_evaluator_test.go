package usecase

import (
	"context"
	"math"
	"path/filepath"
	"testing"

	"wavearchive/internal/database"
	"wavearchive/internal/domain"
	"wavearchive/internal/repository"
)

func TestBuildEvaluatorParsesEchoesBuffsAndRotation(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(filepath.Join(t.TempDir(), "theorycraft.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	characters := repository.NewCharacterSQLite(db.SQL())
	if err := characters.ReplaceSynced(ctx, "3.6.1", []domain.CharacterProfile{
		{Character: domain.Character{ID: 1, Name: "Main", Rarity: 5}},
		{Character: domain.Character{ID: 2, Name: "Sub", Rarity: 5}},
		{Character: domain.Character{ID: 3, Name: "Support", Rarity: 5}},
	}); err != nil {
		t.Fatal(err)
	}
	weapons := repository.NewWeaponSQLite(db.SQL())
	if err := weapons.ReplaceSynced(ctx, "3.6.1", []domain.Weapon{
		{ID: 10, Name: "Test Sword", Rarity: 5, Type: 2, BaseATK: 500, SubStat: "CRIT Rate 24%"},
	}); err != nil {
		t.Fatal(err)
	}
	echoes := repository.NewEchoSQLite(db.SQL())
	if err := echoes.ReplaceSynced(ctx, "3.6.1", []domain.Echo{{ID: 20, Name: "Test Echo", Cost: 4}}, nil); err != nil {
		t.Fatal(err)
	}
	ownedEcho, err := echoes.SaveOwned(ctx, domain.OwnedEcho{
		EchoID: 20, MainStat: "ATK 33%", SubstatsJSON: `["ATK 10%","ATK 100","CRIT DMG 22%","Skill DMG Bonus 12%"]`,
	})
	if err != nil {
		t.Fatal(err)
	}
	builds := repository.NewBuildSQLite(db.SQL())
	weaponID := int64(10)
	build, err := builds.Save(ctx, domain.Build{
		Name: "Main build", CharacterID: 1, CharacterLevel: 90, Sequence: 0,
		WeaponID: &weaponID, WeaponLevel: 90, WeaponRank: 1, Echoes: []domain.OwnedEcho{ownedEcho},
	})
	if err != nil {
		t.Fatal(err)
	}
	theory := repository.NewTheorycraftSQLite(db.SQL())
	teams := repository.NewTeamSQLite(db.SQL())
	calculator := NewDamageCalculator()
	evaluator := NewBuildEvaluator(builds, weapons, theory, teams, calculator)
	evaluation, err := evaluator.SaveConfig(ctx, domain.BuildConfig{
		BuildID: build.ID, ScalingType: "ATK", BaseATK: 1000, BaseHP: 10000, BaseDEF: 1000,
		MotionValue: 2, EnemyLevel: 90, EnemyResistance: .1, ExtraDamageBonusesJSON: "[]",
	})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(evaluation.Stats.TotalATK-2245) > 1e-9 {
		t.Fatalf("total ATK = %v, want 2245", evaluation.Stats.TotalATK)
	}
	if math.Abs(evaluation.Stats.CritRate-.29) > 1e-9 || math.Abs(evaluation.Stats.CritDamage-1.72) > 1e-9 {
		t.Fatalf("unexpected crit stats: %#v", evaluation.Stats)
	}

	team, err := teams.Save(ctx, domain.Team{Name: "Test team", Members: []domain.TeamMember{
		{Slot: 1, CharacterID: 1, BuildID: &build.ID, Role: "main_dps"},
		{Slot: 2, CharacterID: 2, Role: "sub_dps"},
		{Slot: 3, CharacterID: 3, Role: "support"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := theory.SaveBuff(ctx, domain.Buff{
		TeamID: team.ID, SourceSlot: 3, TargetSlot: 1, Name: "Outro Amplify",
		Group: "amplification", Value: .2, Scope: "OUTRO", Duration: 14, Active: true,
	}); err != nil {
		t.Fatal(err)
	}
	rotation, err := theory.SaveRotation(ctx, domain.Rotation{
		TeamID: team.ID, Name: "Short rotation", Actions: []domain.RotationAction{
			{Slot: 1, ActionType: "SKILL", Name: "Skill", MotionValue: 2, CastTime: 1.5, Energy: 10, Concerto: 5, Cooldown: 12},
			{Slot: 1, ActionType: "SKILL", Name: "Skill again", MotionValue: 2, CastTime: 1.5, EnergyCost: 20, Cooldown: 12},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := evaluator.EvaluateRotation(ctx, rotation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalDamage <= 0 || result.DPS <= 0 || result.EnergyBySlot[1] != 10 {
		t.Fatalf("unexpected rotation result: %#v", result)
	}
	if len(result.Errors) != 2 || result.FieldTimeBySlot[1] != 3 {
		t.Fatalf("advanced rotation checks were not applied: %#v", result)
	}
}
