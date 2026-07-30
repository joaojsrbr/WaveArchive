package convene

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"

	"wavearchive/internal/domain"
)

type Pool struct {
	Type      int
	LocaleKey string
	Name      string
	ShortName string
	Kind      string
	HardPity  int
	SortOrder int
}

var Pools = []Pool{
	{Type: 1, LocaleKey: "characterEvent", Name: "Invocação de Ressonantes em Destaque", ShortName: "Ressonador", Kind: "featured_character", HardPity: 80, SortOrder: 0},
	{Type: 2, LocaleKey: "weaponEvent", Name: "Invocação de Armas em Destaque", ShortName: "Arma", Kind: "featured_weapon", HardPity: 80, SortOrder: 1},
	{Type: 3, LocaleKey: "characterPermanent", Name: "Invocação Permanente de Ressonantes", ShortName: "Permanente", Kind: "standard_character", HardPity: 80, SortOrder: 2},
	{Type: 4, LocaleKey: "weaponPermanent", Name: "Invocação Permanente de Armas", ShortName: "Arma padrão", Kind: "standard_weapon", HardPity: 80, SortOrder: 3},
	{Type: 5, LocaleKey: "beginner", Name: "Invocação de Novato", ShortName: "Novato", Kind: "novice", HardPity: 50, SortOrder: 4},
	{Type: 6, LocaleKey: "beginnerSelect", Name: "Invocação de Seleção de Novato", ShortName: "Escolha", Kind: "beginner_choice", HardPity: 80, SortOrder: 5},
	{Type: 8, LocaleKey: "characterNovice", Name: "Invocação de Ressonantes de Nova Viagem", ShortName: "Nova Viagem", Kind: "new_voyage_character", HardPity: 80, SortOrder: 6},
	{Type: 9, LocaleKey: "weaponNovice", Name: "Invocação de Armas de Nova Viagem", ShortName: "NV Arma", Kind: "new_voyage_weapon", HardPity: 80, SortOrder: 7},
	{Type: 10, LocaleKey: "characterCollaboration", Name: "Convocação de Ressonador Colab", ShortName: "Colab.", Kind: "collab_character", HardPity: 80, SortOrder: 8},
	{Type: 11, LocaleKey: "weaponCollaboration", Name: "Convocação de Arma Colab", ShortName: "Colab. Arma", Kind: "collab_weapon", HardPity: 80, SortOrder: 9},
	{Type: 12, LocaleKey: "characterMemory", Name: "Invocação de Ressonante de Reverberação", ShortName: "Reverberação", Kind: "reverb_character", HardPity: 80, SortOrder: 10},
	{Type: 13, LocaleKey: "weaponMemory", Name: "Invocação de Arma de Reverberação", ShortName: "Reverb. Arma", Kind: "reverb_weapon", HardPity: 80, SortOrder: 11},
}

var poolTypeByLocaleKey = func() map[string]int {
	result := make(map[string]int, len(Pools))
	for _, pool := range Pools {
		result[pool.LocaleKey] = pool.Type
	}
	return result
}()

func PoolsFromSelectList(selectList map[string]string) []Pool {
	knownOrder := make(map[string]int, len(Pools))
	for _, pool := range Pools {
		knownOrder[pool.LocaleKey] = pool.SortOrder
	}
	keys := make([]string, 0, len(selectList))
	for key := range selectList {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left, leftKnown := knownOrder[keys[i]]
		right, rightKnown := knownOrder[keys[j]]
		if leftKnown != rightKnown {
			return leftKnown
		}
		if leftKnown {
			return left < right
		}
		return keys[i] < keys[j]
	})

	nextDynamicType := 14
	result := make([]Pool, 0, len(keys))
	for index, key := range keys {
		poolType, known := poolTypeByLocaleKey[key]
		if !known {
			poolType = nextDynamicType
			nextDynamicType++
		}
		name := strings.TrimSpace(selectList[key])
		pool := Pool{
			Type:      poolType,
			LocaleKey: key,
			Name:      name,
			ShortName: shortPoolName(name),
			Kind:      poolKind(key),
			HardPity:  80,
			SortOrder: index,
		}
		if key == "beginner" {
			pool.HardPity = 50
		}
		result = append(result, pool)
	}
	return result
}

func (p Pool) Domain() domain.ConvenePoolDefinition {
	return domain.ConvenePoolDefinition{
		PoolType:  p.Type,
		LocaleKey: p.LocaleKey,
		Name:      p.Name,
		ShortName: p.ShortName,
		Kind:      p.Kind,
		HardPity:  p.HardPity,
		SortOrder: p.SortOrder,
	}
}

func poolKind(key string) string {
	lower := strings.ToLower(key)
	kind := "character"
	if strings.Contains(lower, "weapon") {
		kind = "weapon"
	}
	return "dynamic_" + kind
}

func shortPoolName(name string) string {
	replacer := strings.NewReplacer(
		"Invocação de ", "",
		"Invocação ", "",
		"Convocação de ", "",
		"Convene ", "",
	)
	short := strings.TrimSpace(replacer.Replace(name))
	runes := []rune(short)
	if len(runes) > 24 {
		return string(runes[:23]) + "…"
	}
	return short
}

func PoolByType(poolType int) (Pool, bool) {
	for _, pool := range Pools {
		if pool.Type == poolType {
			return pool, true
		}
	}
	return Pool{}, false
}

func fingerprint(base string, occurrence int) string {
	sum := sha256.Sum256([]byte(base + "\x1f" + strconv.Itoa(occurrence)))
	return hex.EncodeToString(sum[:])
}
