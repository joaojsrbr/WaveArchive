package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wavearchive/internal/domain"
)

type WeaponSQLite struct {
	db *sql.DB
}

func NewWeaponSQLite(db *sql.DB) *WeaponSQLite {
	return &WeaponSQLite{db: db}
}

func (r *WeaponSQLite) List(ctx context.Context, filter domain.WeaponFilter) ([]domain.Weapon, error) {
	where := []string{"1 = 1"}
	args := make([]any, 0, 5)
	if query := ftsQuery(filter.Query); query != "" {
		where = append(where, "w.id IN (SELECT rowid FROM weapons_fts WHERE weapons_fts MATCH ?)")
		args = append(args, query)
	}
	if filter.Type > 0 {
		where = append(where, "w.weapon_type = ?")
		args = append(args, filter.Type)
	}
	if filter.Rarity > 0 {
		where = append(where, "w.rarity = ?")
		args = append(args, filter.Rarity)
	}
	if filter.SubStat != "" {
		where = append(where, "LOWER(w.sub_stat) LIKE LOWER(?)")
		args = append(args, "%"+filter.SubStat+"%")
	}
	if filter.Account == "owned" {
		where = append(where, "COALESCE(ow.owned, 0) = 1")
	} else if filter.Account == "missing" {
		where = append(where, "COALESCE(ow.owned, 0) = 0")
	}
	if filter.OwnedOnly {
		where = append(where, "COALESCE(ow.owned, 0) = 1")
	}
	if filter.Favorites {
		where = append(where, "COALESCE(ow.favorite, 0) = 1")
	}
	if filter.MinATK > 0 {
		where = append(where, "w.base_atk >= ?")
		args = append(args, filter.MinATK)
	}
	if filter.MaxATK > 0 {
		where = append(where, "w.base_atk <= ?")
		args = append(args, filter.MaxATK)
	}
	if filter.MinLevel > 0 {
		where = append(where, "COALESCE(ow.level, 1) >= ?")
		args = append(args, filter.MinLevel)
	}
	if filter.MaxLevel > 0 {
		where = append(where, "COALESCE(ow.level, 1) <= ?")
		args = append(args, filter.MaxLevel)
	}
	if filter.MinRank > 0 {
		where = append(where, "COALESCE(ow.weapon_rank, 1) >= ?")
		args = append(args, filter.MinRank)
	}
	if filter.MaxRank > 0 {
		where = append(where, "COALESCE(ow.weapon_rank, 1) <= ?")
		args = append(args, filter.MaxRank)
	}
	order := "w.name COLLATE NOCASE"
	switch filter.Sort {
	case "rarity":
		order = "w.rarity DESC, w.name COLLATE NOCASE"
	case "type":
		order = "w.weapon_type, w.name COLLATE NOCASE"
	case "atk":
		order = "w.base_atk DESC, w.rarity DESC, w.name COLLATE NOCASE"
	case "id":
		order = "w.id"
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT w.id, w.name, w.rarity, w.weapon_type, w.description,
		       w.effect_name, w.effect, w.icon_path, w.params_json,
		       w.base_atk, w.sub_stat, w.game_version,
		       COALESCE(ow.owned, 0), COALESCE(ow.level, 1),
		       COALESCE(ow.weapon_rank, 1), COALESCE(ow.favorite, 0)
		FROM weapons w
		LEFT JOIN owned_weapons ow ON ow.weapon_id=w.id AND ow.profile_id=1
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY `+order, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	weapons := []domain.Weapon{}
	for rows.Next() {
		weapon, err := scanWeapon(rows)
		if err != nil {
			return nil, err
		}
		weapons = append(weapons, weapon)
	}
	return weapons, rows.Err()
}

func (r *WeaponSQLite) Get(ctx context.Context, id int64) (domain.Weapon, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT w.id, w.name, w.rarity, w.weapon_type, w.description,
		       w.effect_name, w.effect, w.icon_path, w.params_json,
		       w.base_atk, w.sub_stat, w.game_version,
		       COALESCE(ow.owned, 0), COALESCE(ow.level, 1),
		       COALESCE(ow.weapon_rank, 1), COALESCE(ow.favorite, 0)
		FROM weapons w
		LEFT JOIN owned_weapons ow ON ow.weapon_id=w.id AND ow.profile_id=1
		WHERE w.id=?
	`, id)
	weapon, err := scanWeapon(row)
	if err != nil {
		return domain.Weapon{}, fmt.Errorf("get weapon %d: %w", id, err)
	}
	return weapon, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanWeapon(row scanner) (domain.Weapon, error) {
	var weapon domain.Weapon
	err := row.Scan(
		&weapon.ID, &weapon.Name, &weapon.Rarity, &weapon.Type, &weapon.Description,
		&weapon.EffectName, &weapon.Effect, &weapon.IconPath, &weapon.ParamsJSON,
		&weapon.BaseATK, &weapon.SubStat, &weapon.GameVersion,
		&weapon.Owned, &weapon.Level, &weapon.Rank, &weapon.Favorite,
	)
	weapon.TypeName = weaponTypeName(weapon.Type)
	return weapon, err
}

func (r *WeaponSQLite) ReplaceSynced(ctx context.Context, version string, weapons []domain.Weapon) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO weapons(
			id, name, rarity, weapon_type, description, effect_name, effect,
			icon_path, params_json, game_version, base_atk, sub_stat, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, rarity=excluded.rarity, weapon_type=excluded.weapon_type,
			description=excluded.description, effect_name=excluded.effect_name,
			effect=excluded.effect, icon_path=excluded.icon_path,
			params_json=excluded.params_json, game_version=excluded.game_version,
			base_atk=excluded.base_atk, sub_stat=excluded.sub_stat,
			updated_at=CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, weapon := range weapons {
		if _, err := stmt.ExecContext(
			ctx, weapon.ID, weapon.Name, weapon.Rarity, weapon.Type,
			weapon.Description, weapon.EffectName, weapon.Effect, weapon.IconPath,
			weapon.ParamsJSON, version, weapon.BaseATK, weapon.SubStat,
		); err != nil {
			return fmt.Errorf("upsert weapon %d: %w", weapon.ID, err)
		}
	}
	return tx.Commit()
}

func (r *WeaponSQLite) UpdateAccount(ctx context.Context, update domain.WeaponAccountUpdate) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO owned_weapons(profile_id, weapon_id, level, weapon_rank, favorite, owned)
		VALUES (1, ?, ?, ?, ?, ?)
		ON CONFLICT(profile_id, weapon_id) DO UPDATE SET
			level=excluded.level, weapon_rank=excluded.weapon_rank,
			favorite=excluded.favorite, owned=excluded.owned
	`, update.WeaponID, update.Level, update.Rank, update.Favorite, update.Owned)
	if err != nil {
		return fmt.Errorf("update local weapon state: %w", err)
	}
	return nil
}
