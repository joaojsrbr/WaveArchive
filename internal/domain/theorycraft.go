package domain

import "context"

type BuildConfig struct {
	BuildID                int64   `json:"buildId"`
	ScalingType            string  `json:"scalingType"`
	BaseATK                float64 `json:"baseAtk"`
	BaseHP                 float64 `json:"baseHp"`
	BaseDEF                float64 `json:"baseDef"`
	MotionValue            float64 `json:"motionValue"`
	FlatDamage             float64 `json:"flatDamage"`
	EnemyLevel             int     `json:"enemyLevel"`
	EnemyResistance        float64 `json:"enemyResistance"`
	DefenseIgnore          float64 `json:"defenseIgnore"`
	DamageReduction        float64 `json:"damageReduction"`
	ElementReduction       float64 `json:"elementReduction"`
	ExtraDamageBonusesJSON string  `json:"extraDamageBonusesJson"`
}

type BuildStats struct {
	BaseATK       float64            `json:"baseAtk"`
	WeaponATK     float64            `json:"weaponAtk"`
	ATKPercent    float64            `json:"atkPercent"`
	FlatATK       float64            `json:"flatAtk"`
	TotalATK      float64            `json:"totalAtk"`
	BaseHP        float64            `json:"baseHp"`
	HPPercent     float64            `json:"hpPercent"`
	FlatHP        float64            `json:"flatHp"`
	TotalHP       float64            `json:"totalHp"`
	BaseDEF       float64            `json:"baseDef"`
	DEFPercent    float64            `json:"defPercent"`
	FlatDEF       float64            `json:"flatDef"`
	TotalDEF      float64            `json:"totalDef"`
	CritRate      float64            `json:"critRate"`
	CritDamage    float64            `json:"critDamage"`
	EnergyRegen   float64            `json:"energyRegen"`
	DamageBonuses map[string]float64 `json:"damageBonuses"`
	UnparsedStats []string           `json:"unparsedStats"`
	ScalingStat   float64            `json:"scalingStat"`
}

type BuildEvaluation struct {
	Build  Build        `json:"build"`
	Config BuildConfig  `json:"config"`
	Stats  BuildStats   `json:"stats"`
	Damage DamageResult `json:"damage"`
}

type Buff struct {
	ID            int64   `json:"id"`
	TeamID        int64   `json:"teamId"`
	SourceSlot    int     `json:"sourceSlot"`
	TargetSlot    int     `json:"targetSlot"`
	Name          string  `json:"name"`
	Group         string  `json:"group"`
	Value         float64 `json:"value"`
	Scope         string  `json:"scope"`
	Condition     string  `json:"condition"`
	Active        bool    `json:"active"`
	Duration      float64 `json:"duration"`
	TriggerAction string  `json:"triggerAction"`
}

type Rotation struct {
	ID       int64            `json:"id"`
	TeamID   int64            `json:"teamId"`
	Name     string           `json:"name"`
	Duration float64          `json:"duration"`
	Notes    string           `json:"notes"`
	Actions  []RotationAction `json:"actions"`
}

type RotationAction struct {
	ID          int64   `json:"id"`
	Order       int     `json:"order"`
	Slot        int     `json:"slot"`
	ActionType  string  `json:"actionType"`
	Name        string  `json:"name"`
	MotionValue float64 `json:"motionValue"`
	CastTime    float64 `json:"castTime"`
	Energy      float64 `json:"energy"`
	Concerto    float64 `json:"concerto"`
	Cooldown    float64 `json:"cooldown"`
	EnergyCost  float64 `json:"energyCost"`
	Notes       string  `json:"notes"`
}

type RotationActionResult struct {
	Action       RotationAction `json:"action"`
	Damage       float64        `json:"damage"`
	StartTime    float64        `json:"startTime"`
	EndTime      float64        `json:"endTime"`
	ActiveBuffs  []string       `json:"activeBuffs"`
	ExpiredBuffs []string       `json:"expiredBuffs"`
}

type RotationResult struct {
	Rotation        Rotation               `json:"rotation"`
	Actions         []RotationActionResult `json:"actions"`
	TotalDamage     float64                `json:"totalDamage"`
	Duration        float64                `json:"duration"`
	DPS             float64                `json:"dps"`
	EnergyBySlot    map[int]float64        `json:"energyBySlot"`
	ConcertoBySlot  map[int]float64        `json:"concertoBySlot"`
	FieldTimeBySlot map[int]float64        `json:"fieldTimeBySlot"`
	Warnings        []string               `json:"warnings"`
	Errors          []string               `json:"errors"`
}

type TeamTheorycraft struct {
	Team      Team       `json:"team"`
	Buffs     []Buff     `json:"buffs"`
	Rotations []Rotation `json:"rotations"`
	Warnings  []string   `json:"warnings"`
}

type TheorycraftRepository interface {
	GetBuildConfig(ctx context.Context, buildID int64) (BuildConfig, error)
	SaveBuildConfig(ctx context.Context, config BuildConfig) error
	ListBuffs(ctx context.Context, teamID int64) ([]Buff, error)
	SaveBuff(ctx context.Context, buff Buff) (Buff, error)
	DeleteBuff(ctx context.Context, id int64) error
	ListRotations(ctx context.Context, teamID int64) ([]Rotation, error)
	SaveRotation(ctx context.Context, rotation Rotation) (Rotation, error)
	DeleteRotation(ctx context.Context, id int64) error
}
