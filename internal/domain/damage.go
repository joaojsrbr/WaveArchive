package domain

type DamageInput struct {
	ScalingStat                float64   `json:"scalingStat"`
	MotionValue                float64   `json:"motionValue"`
	FlatDamage                 float64   `json:"flatDamage"`
	FlatBonusDamage            float64   `json:"flatBonusDamage"`
	CharacterLevel             int       `json:"characterLevel"`
	EnemyLevel                 int       `json:"enemyLevel"`
	EnemyResistance            float64   `json:"enemyResistance"`
	ResistancePenetration      float64   `json:"resistancePenetration"`
	DefenseIgnore              float64   `json:"defenseIgnore"`
	DamageReduction            float64   `json:"damageReduction"`
	AdditionalDamageReduction  float64   `json:"additionalDamageReduction"`
	ElementReduction           float64   `json:"elementReduction"`
	AdditionalElementReduction float64   `json:"additionalElementReduction"`
	DamageBonuses              []float64 `json:"damageBonuses"`
	Amplifications             []float64 `json:"amplifications"`
	SpecialBonuses             []float64 `json:"specialBonuses"`
	CritRate                   float64   `json:"critRate"`
	CritDamage                 float64   `json:"critDamage"`
}

type DamageResult struct {
	FormulaVersion             string    `json:"formulaVersion"`
	FormulaConfidence          string    `json:"formulaConfidence"`
	BaseDamage                 float64   `json:"baseDamage"`
	EffectiveResistance        float64   `json:"effectiveResistance"`
	ResistanceMultiplier       float64   `json:"resistanceMultiplier"`
	DefenseMultiplier          float64   `json:"defenseMultiplier"`
	DamageReductionMultiplier  float64   `json:"damageReductionMultiplier"`
	ElementReductionMultiplier float64   `json:"elementReductionMultiplier"`
	DamageBonusMultiplier      float64   `json:"damageBonusMultiplier"`
	AmplificationMultiplier    float64   `json:"amplificationMultiplier"`
	SpecialDamageMultiplier    float64   `json:"specialDamageMultiplier"`
	NonCriticalDamage          float64   `json:"nonCriticalDamage"`
	CriticalDamage             float64   `json:"criticalDamage"`
	ExpectedDamage             float64   `json:"expectedDamage"`
	Insights                   []Insight `json:"insights"`
}

type Insight struct {
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Message  string `json:"message"`
}
