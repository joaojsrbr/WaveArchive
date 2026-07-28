package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wavearchive/internal/domain"
)

type CharacterSQLite struct {
	db *sql.DB
}

func NewCharacterSQLite(db *sql.DB) *CharacterSQLite {
	return &CharacterSQLite{db: db}
}

func (r *CharacterSQLite) List(ctx context.Context, filter domain.CharacterFilter) ([]domain.Character, error) {
	where := []string{"1 = 1"}
	args := make([]any, 0, 6)

	if query := ftsQuery(filter.Query); query != "" {
		where = append(where, "c.id IN (SELECT rowid FROM characters_fts WHERE characters_fts MATCH ?)")
		args = append(args, query)
	}
	if filter.Element > 0 {
		where = append(where, "c.element = ?")
		args = append(args, filter.Element)
	}
	if filter.Rarity > 0 {
		where = append(where, "c.rarity = ?")
		args = append(args, filter.Rarity)
	}
	if filter.OwnedOnly {
		where = append(where, "COALESCE(oc.owned, 0) = 1")
	}
	if filter.Favorites {
		where = append(where, "COALESCE(oc.favorite, 0) = 1")
	}

	order := "c.name COLLATE NOCASE ASC"
	switch filter.Sort {
	case "api":
		order = "c.api_order ASC, c.name COLLATE NOCASE ASC"
	case "rarity":
		order = "c.rarity DESC, c.name COLLATE NOCASE ASC"
	case "element":
		order = "c.element ASC, c.name COLLATE NOCASE ASC"
	case "id":
		order = "c.id ASC"
	}

	query := `
		SELECT c.id, c.name, c.nickname, c.rarity, c.element, c.weapon_type,
		       c.icon_path, c.background_path, c.gender, c.game_version, c.api_order,
		       COALESCE(oc.owned, 0), COALESCE(oc.level, 1),
		       COALESCE(oc.sequence, 0), COALESCE(oc.favorite, 0)
		FROM characters c
		LEFT JOIN owned_characters oc ON oc.character_id = c.id AND oc.profile_id = 1
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY ` + order

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list characters: %w", err)
	}
	defer rows.Close()

	characters := make([]domain.Character, 0)
	for rows.Next() {
		var c domain.Character
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Nickname, &c.Rarity, &c.Element, &c.WeaponType,
			&c.IconPath, &c.BackgroundPath, &c.Gender, &c.GameVersion, &c.APIOrder,
			&c.Owned, &c.Level, &c.Sequence, &c.Favorite,
		); err != nil {
			return nil, err
		}
		c.ElementName = c.Element.String()
		c.WeaponTypeName = weaponTypeName(c.WeaponType)
		characters = append(characters, c)
	}
	return characters, rows.Err()
}

func (r *CharacterSQLite) GetProfile(ctx context.Context, id int64) (domain.CharacterProfile, error) {
	var profile domain.CharacterProfile
	var signatureID sql.NullInt64
	var extrasJSON string
	err := r.db.QueryRowContext(ctx, `
		SELECT c.id, c.name, c.nickname, c.rarity, c.element, c.weapon_type,
		       c.icon_path, c.background_path, c.game_version, c.api_order,
		       COALESCE(oc.owned, 0), COALESCE(oc.level, 1),
		       COALESCE(oc.sequence, 0), COALESCE(oc.favorite, 0),
		       c.description, c.birthday, c.gender, c.region, c.faction,
		       c.talent_name, c.talent_description, c.signature_weapon_id,
		       c.detail_loaded, c.extras_json
		FROM characters c
		LEFT JOIN owned_characters oc ON oc.character_id = c.id AND oc.profile_id = 1
		WHERE c.id = ?
	`, id).Scan(
		&profile.Character.ID, &profile.Character.Name, &profile.Character.Nickname,
		&profile.Character.Rarity, &profile.Character.Element, &profile.Character.WeaponType,
		&profile.Character.IconPath, &profile.Character.BackgroundPath, &profile.Character.GameVersion, &profile.Character.APIOrder,
		&profile.Character.Owned, &profile.Character.Level, &profile.Character.Sequence, &profile.Character.Favorite,
		&profile.Description, &profile.Birthday, &profile.Gender, &profile.Region, &profile.Faction,
		&profile.TalentName, &profile.TalentDescription, &signatureID, &profile.DetailLoaded, &extrasJSON,
	)
	if err != nil {
		return profile, fmt.Errorf("get character %d: %w", id, err)
	}
	profile.Character.ElementName = profile.Character.Element.String()
	profile.Character.WeaponTypeName = weaponTypeName(profile.Character.WeaponType)
	profile.Character.Gender = profile.Gender
	_ = json.Unmarshal([]byte(extrasJSON), &profile.Extras)

	if signatureID.Valid {
		weapon := domain.Weapon{ID: signatureID.Int64}
		err := r.db.QueryRowContext(ctx, `
			SELECT name, rarity, weapon_type, description, effect_name, effect, icon_path, params_json
			FROM weapons WHERE id = ?
		`, signatureID.Int64).Scan(
			&weapon.Name, &weapon.Rarity, &weapon.Type, &weapon.Description,
			&weapon.EffectName, &weapon.Effect, &weapon.IconPath, &weapon.ParamsJSON,
		)
		if err != nil && err != sql.ErrNoRows {
			return profile, err
		}
		weapon.TypeName = weaponTypeName(weapon.Type)
		profile.SignatureWeapon = &weapon
	}

	skillRows, err := r.db.QueryContext(ctx, `
		SELECT node_id, skill_type, name, description, icon_path, levels_json, sort_order
		FROM skills WHERE character_id = ? ORDER BY sort_order, node_id
	`, id)
	if err != nil {
		return profile, err
	}
	defer skillRows.Close()
	profile.Skills = []domain.Skill{}
	for skillRows.Next() {
		var skill domain.Skill
		if err := skillRows.Scan(&skill.NodeID, &skill.Type, &skill.Name, &skill.Description, &skill.IconPath, &skill.LevelsJSON, &skill.SortOrder); err != nil {
			return profile, err
		}
		profile.Skills = append(profile.Skills, skill)
	}
	if err := skillRows.Err(); err != nil {
		return profile, err
	}
	if err := skillRows.Close(); err != nil {
		return profile, err
	}

	chainRows, err := r.db.QueryContext(ctx, `
		SELECT sequence, name, description, icon_path
		FROM resonance_chains WHERE character_id = ? ORDER BY sequence
	`, id)
	if err != nil {
		return profile, err
	}
	defer chainRows.Close()
	profile.Chains = []domain.ResonanceChain{}
	for chainRows.Next() {
		var chain domain.ResonanceChain
		if err := chainRows.Scan(&chain.Sequence, &chain.Name, &chain.Description, &chain.IconPath); err != nil {
			return profile, err
		}
		profile.Chains = append(profile.Chains, chain)
	}
	if err := chainRows.Err(); err != nil {
		return profile, err
	}
	if err := chainRows.Close(); err != nil {
		return profile, err
	}
	if err := r.loadProgression(ctx, id, &profile); err != nil {
		return profile, err
	}
	return profile, nil
}

func (r *CharacterSQLite) loadProgression(ctx context.Context, id int64, profile *domain.CharacterProfile) error {
	profile.Progression = domain.CharacterProgression{Ascensions: []domain.AscensionStage{}, Skills: []domain.SkillProgression{}, LevelEXP: []int{}, Stats: []domain.CharacterStat{}}
	rows, err := r.db.QueryContext(ctx, `
		SELECT ac.stage, m.id, m.name, m.rarity, m.material_type, m.description, m.icon_path, m.sources_json, m.game_version, ac.quantity
		FROM character_ascension_costs ac JOIN materials m ON m.id=ac.material_id
		WHERE ac.character_id=? ORDER BY ac.stage, m.id`, id)
	if err != nil {
		return err
	}
	stageMap := map[int]int{}
	for rows.Next() {
		var stage int
		var material domain.Material
		var sources string
		var quantity int
		if err := rows.Scan(&stage, &material.ID, &material.Name, &material.Rarity, &material.Type, &material.Description, &material.IconPath, &sources, &material.GameVersion, &quantity); err != nil {
			rows.Close()
			return err
		}
		_ = json.Unmarshal([]byte(sources), &material.Sources)
		index, ok := stageMap[stage]
		if !ok {
			index = len(profile.Progression.Ascensions)
			stageMap[stage] = index
			profile.Progression.Ascensions = append(profile.Progression.Ascensions, domain.AscensionStage{Stage: stage, UnlockLevel: ascensionLevel(stage), Costs: []domain.MaterialCost{}})
		}
		profile.Progression.Ascensions[index].Costs = append(profile.Progression.Ascensions[index].Costs, domain.MaterialCost{Material: material, Quantity: quantity})
	}
	if err := rows.Close(); err != nil {
		return err
	}

	expRows, err := r.db.QueryContext(ctx, "SELECT level, experience FROM character_level_exp WHERE character_id=? ORDER BY level", id)
	if err != nil {
		return err
	}
	for expRows.Next() {
		var level, value int
		if err := expRows.Scan(&level, &value); err != nil {
			expRows.Close()
			return err
		}
		for len(profile.Progression.LevelEXP) <= level {
			profile.Progression.LevelEXP = append(profile.Progression.LevelEXP, 0)
		}
		profile.Progression.LevelEXP[level] = value
	}
	expRows.Close()

	statRows, err := r.db.QueryContext(ctx, "SELECT ascension, level, hp, atk, def FROM character_stats WHERE character_id=? ORDER BY ascension, level", id)
	if err != nil {
		return err
	}
	for statRows.Next() {
		var stat domain.CharacterStat
		if err := statRows.Scan(&stat.Ascension, &stat.Level, &stat.HP, &stat.ATK, &stat.DEF); err != nil {
			statRows.Close()
			return err
		}
		profile.Progression.Stats = append(profile.Progression.Stats, stat)
	}
	statRows.Close()

	for _, skill := range profile.Skills {
		var nodeType, maxLevel int
		var valuesJSON string
		err := r.db.QueryRowContext(ctx, "SELECT node_type,max_level,values_json FROM skill_progression WHERE character_id=? AND node_id=?", id, skill.NodeID).Scan(&nodeType, &maxLevel, &valuesJSON)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return err
		}
		entry := domain.SkillProgression{NodeID: skill.NodeID, NodeType: nodeType, Type: skill.Type, Name: skill.Name, IconPath: skill.IconPath, MaxLevel: maxLevel, UnlockCosts: []domain.MaterialCost{}, LevelCosts: []domain.SkillLevelCost{}, Values: []domain.SkillValueRow{}}
		_ = json.Unmarshal([]byte(valuesJSON), &entry.Values)
		entry.UnlockCosts, err = r.materialCosts(ctx, "SELECT m.id,m.name,m.rarity,m.material_type,m.description,m.icon_path,m.sources_json,m.game_version,c.quantity FROM skill_unlock_costs c JOIN materials m ON m.id=c.material_id WHERE c.character_id=? AND c.node_id=? ORDER BY m.id", id, skill.NodeID)
		if err != nil {
			return err
		}
		levelRows, err := r.db.QueryContext(ctx, "SELECT DISTINCT level FROM skill_level_costs WHERE character_id=? AND node_id=? ORDER BY level", id, skill.NodeID)
		if err != nil {
			return err
		}
		levels := []int{}
		for levelRows.Next() {
			var level int
			if err := levelRows.Scan(&level); err != nil {
				levelRows.Close()
				return err
			}
			levels = append(levels, level)
		}
		if err := levelRows.Close(); err != nil {
			return err
		}
		for _, level := range levels {
			costs, err := r.materialCosts(ctx, "SELECT m.id,m.name,m.rarity,m.material_type,m.description,m.icon_path,m.sources_json,m.game_version,c.quantity FROM skill_level_costs c JOIN materials m ON m.id=c.material_id WHERE c.character_id=? AND c.node_id=? AND c.level=? ORDER BY m.id", id, skill.NodeID, level)
			if err != nil {
				return err
			}
			entry.LevelCosts = append(entry.LevelCosts, domain.SkillLevelCost{Level: level, Costs: costs})
		}
		profile.Progression.Skills = append(profile.Progression.Skills, entry)
	}
	return nil
}

func (r *CharacterSQLite) materialCosts(ctx context.Context, query string, args ...any) ([]domain.MaterialCost, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.MaterialCost{}
	for rows.Next() {
		var material domain.Material
		var sources string
		var quantity int
		if err := rows.Scan(&material.ID, &material.Name, &material.Rarity, &material.Type, &material.Description, &material.IconPath, &sources, &material.GameVersion, &quantity); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(sources), &material.Sources)
		result = append(result, domain.MaterialCost{Material: material, Quantity: quantity})
	}
	return result, rows.Err()
}

func ascensionLevel(stage int) int {
	values := []int{0, 20, 40, 50, 60, 70, 80}
	if stage > 0 && stage < len(values) {
		return values[stage]
	}
	return 0
}

func (r *CharacterSQLite) UpdateAccount(ctx context.Context, update domain.CharacterAccountUpdate) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO owned_characters(
			profile_id, character_id, level, sequence, favorite, owned
		) VALUES (1, ?, ?, ?, ?, ?)
		ON CONFLICT(profile_id, character_id) DO UPDATE SET
			level=excluded.level,
			sequence=excluded.sequence,
			favorite=excluded.favorite,
			owned=excluded.owned
	`, update.CharacterID, update.Level, update.Sequence, update.Favorite, update.Owned)
	if err != nil {
		return fmt.Errorf("update local character state: %w", err)
	}
	return nil
}

func (r *CharacterSQLite) ReplaceSynced(ctx context.Context, version string, profiles []domain.CharacterProfile) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO characters (
			id, name, nickname, rarity, element, weapon_type,
			icon_path, background_path, game_version, description, birthday,
			gender, region, faction, talent_name, talent_description,
			signature_weapon_id, detail_loaded, api_order, extras_json, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name,
			nickname=excluded.nickname,
			rarity=excluded.rarity,
			element=excluded.element,
			weapon_type=excluded.weapon_type,
			icon_path=excluded.icon_path,
			background_path=excluded.background_path,
			game_version=excluded.game_version,
			description=CASE WHEN excluded.detail_loaded=1 THEN excluded.description ELSE characters.description END,
			birthday=CASE WHEN excluded.detail_loaded=1 THEN excluded.birthday ELSE characters.birthday END,
			gender=CASE WHEN excluded.detail_loaded=1 THEN excluded.gender ELSE characters.gender END,
			region=CASE WHEN excluded.detail_loaded=1 THEN excluded.region ELSE characters.region END,
			faction=CASE WHEN excluded.detail_loaded=1 THEN excluded.faction ELSE characters.faction END,
			talent_name=CASE WHEN excluded.detail_loaded=1 THEN excluded.talent_name ELSE characters.talent_name END,
			talent_description=CASE WHEN excluded.detail_loaded=1 THEN excluded.talent_description ELSE characters.talent_description END,
			signature_weapon_id=CASE WHEN excluded.detail_loaded=1 THEN excluded.signature_weapon_id ELSE characters.signature_weapon_id END,
			detail_loaded=MAX(characters.detail_loaded, excluded.detail_loaded),
			api_order=excluded.api_order,
			extras_json=CASE WHEN excluded.detail_loaded=1 THEN excluded.extras_json ELSE characters.extras_json END,
			updated_at=CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	weaponStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO weapons (
			id, name, rarity, weapon_type, description, effect_name, effect,
			icon_path, params_json, game_version, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, rarity=excluded.rarity, weapon_type=excluded.weapon_type,
			description=excluded.description, effect_name=excluded.effect_name,
			effect=excluded.effect, icon_path=excluded.icon_path,
			params_json=excluded.params_json, game_version=excluded.game_version,
			updated_at=CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer weaponStmt.Close()

	skillStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO skills(character_id, node_id, skill_type, name, description, icon_path, levels_json, sort_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer skillStmt.Close()

	chainStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO resonance_chains(character_id, sequence, name, description, icon_path)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer chainStmt.Close()

	materialStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO materials(id,name,rarity,material_type,description,icon_path,sources_json,game_version)
		VALUES(?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name,rarity=excluded.rarity,material_type=excluded.material_type,
		description=excluded.description,icon_path=excluded.icon_path,sources_json=excluded.sources_json,game_version=excluded.game_version`)
	if err != nil {
		return err
	}
	defer materialStmt.Close()

	for _, profile := range profiles {
		c := profile.Character
		var signatureID any
		if profile.SignatureWeapon != nil {
			signatureID = profile.SignatureWeapon.ID
			if profile.SignatureWeapon.Name != "" {
				w := profile.SignatureWeapon
				if _, err := weaponStmt.ExecContext(ctx, w.ID, w.Name, w.Rarity, w.Type, w.Description, w.EffectName, w.Effect, w.IconPath, w.ParamsJSON, version); err != nil {
					return fmt.Errorf("upsert weapon %d: %w", w.ID, err)
				}
			}
		}
		extrasJSON, _ := json.Marshal(profile.Extras)
		if _, err := stmt.ExecContext(
			ctx, c.ID, c.Name, c.Nickname, c.Rarity, c.Element, c.WeaponType,
			c.IconPath, c.BackgroundPath, version, profile.Description, profile.Birthday,
			profile.Gender, profile.Region, profile.Faction, profile.TalentName,
			profile.TalentDescription, signatureID, profile.DetailLoaded, c.APIOrder, string(extrasJSON),
		); err != nil {
			return fmt.Errorf("upsert character %d: %w", c.ID, err)
		}
		if !profile.DetailLoaded {
			continue
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM skills WHERE character_id = ?", c.ID); err != nil {
			return err
		}
		for _, skill := range profile.Skills {
			if _, err := skillStmt.ExecContext(ctx, c.ID, skill.NodeID, skill.Type, skill.Name, skill.Description, skill.IconPath, skill.LevelsJSON, skill.SortOrder); err != nil {
				return fmt.Errorf("insert skill %s for character %d: %w", skill.NodeID, c.ID, err)
			}
		}
		for _, stage := range profile.Progression.Ascensions {
			for _, cost := range stage.Costs {
				if err := upsertMaterial(ctx, materialStmt, cost.Material, version); err != nil {
					return err
				}
			}
		}
		for _, skill := range profile.Progression.Skills {
			for _, cost := range skill.UnlockCosts {
				if err := upsertMaterial(ctx, materialStmt, cost.Material, version); err != nil {
					return err
				}
			}
			for _, level := range skill.LevelCosts {
				for _, cost := range level.Costs {
					if err := upsertMaterial(ctx, materialStmt, cost.Material, version); err != nil {
						return err
					}
				}
			}
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM character_ascension_costs WHERE character_id=?", c.ID); err != nil {
			return err
		}
		for _, stage := range profile.Progression.Ascensions {
			for _, cost := range stage.Costs {
				if _, err := tx.ExecContext(ctx, "INSERT INTO character_ascension_costs(character_id,stage,material_id,quantity) VALUES(?,?,?,?)", c.ID, stage.Stage, cost.Material.ID, cost.Quantity); err != nil {
					return err
				}
			}
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM character_level_exp WHERE character_id=?", c.ID); err != nil {
			return err
		}
		for level, experience := range profile.Progression.LevelEXP {
			if _, err := tx.ExecContext(ctx, "INSERT INTO character_level_exp(character_id,level,experience) VALUES(?,?,?)", c.ID, level, experience); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM character_stats WHERE character_id=?", c.ID); err != nil {
			return err
		}
		for _, stat := range profile.Progression.Stats {
			if _, err := tx.ExecContext(ctx, "INSERT INTO character_stats(character_id,ascension,level,hp,atk,def) VALUES(?,?,?,?,?,?)", c.ID, stat.Ascension, stat.Level, stat.HP, stat.ATK, stat.DEF); err != nil {
				return err
			}
		}
		for _, skill := range profile.Progression.Skills {
			values, _ := json.Marshal(skill.Values)
			if _, err := tx.ExecContext(ctx, "INSERT INTO skill_progression(character_id,node_id,node_type,max_level,values_json) VALUES(?,?,?,?,?)", c.ID, skill.NodeID, skill.NodeType, skill.MaxLevel, string(values)); err != nil {
				return err
			}
			for _, cost := range skill.UnlockCosts {
				if _, err := tx.ExecContext(ctx, "INSERT INTO skill_unlock_costs(character_id,node_id,material_id,quantity) VALUES(?,?,?,?)", c.ID, skill.NodeID, cost.Material.ID, cost.Quantity); err != nil {
					return err
				}
			}
			for _, level := range skill.LevelCosts {
				for _, cost := range level.Costs {
					if _, err := tx.ExecContext(ctx, "INSERT INTO skill_level_costs(character_id,node_id,level,material_id,quantity) VALUES(?,?,?,?,?)", c.ID, skill.NodeID, level.Level, cost.Material.ID, cost.Quantity); err != nil {
						return err
					}
				}
			}
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM resonance_chains WHERE character_id = ?", c.ID); err != nil {
			return err
		}
		for _, chain := range profile.Chains {
			if _, err := chainStmt.ExecContext(ctx, c.ID, chain.Sequence, chain.Name, chain.Description, chain.IconPath); err != nil {
				return fmt.Errorf("insert chain %d for character %d: %w", chain.Sequence, c.ID, err)
			}
		}
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO game_versions(version, synced_at) VALUES (?, ?)
		ON CONFLICT(version) DO UPDATE SET synced_at=excluded.synced_at
	`, version, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return err
	}
	return tx.Commit()
}

func upsertMaterial(ctx context.Context, stmt *sql.Stmt, material domain.Material, version string) error {
	sources, _ := json.Marshal(material.Sources)
	_, err := stmt.ExecContext(ctx, material.ID, material.Name, material.Rarity, material.Type, material.Description, material.IconPath, string(sources), version)
	return err
}

func (r *CharacterSQLite) Status(ctx context.Context) (domain.CatalogStatus, error) {
	var status domain.CatalogStatus
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM characters").Scan(&status.Count); err != nil {
		return status, err
	}
	var version, synced sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT version, synced_at FROM game_versions ORDER BY synced_at DESC LIMIT 1
	`).Scan(&version, &synced)
	if err != nil && err != sql.ErrNoRows {
		return status, err
	}
	if version.Valid {
		status.Version = version.String
	}
	if synced.Valid {
		if parsed, err := time.Parse(time.RFC3339, synced.String); err == nil {
			status.LastSyncAt = &parsed
		}
	}
	return status, nil
}

func ftsQuery(value string) string {
	fields := strings.Fields(strings.TrimSpace(value))
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.ReplaceAll(field, `"`, `""`)
		if field != "" {
			parts = append(parts, `"`+field+`"*`)
		}
	}
	return strings.Join(parts, " AND ")
}

func weaponTypeName(code int) string {
	switch code {
	case 1:
		return "Broadblade"
	case 2:
		return "Sword"
	case 3:
		return "Pistols"
	case 4:
		return "Gauntlets"
	case 5:
		return "Rectifier"
	default:
		return "Unknown"
	}
}
