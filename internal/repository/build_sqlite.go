package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"wavearchive/internal/domain"
)

type BuildSQLite struct {
	db *sql.DB
}

func NewBuildSQLite(db *sql.DB) *BuildSQLite {
	return &BuildSQLite{db: db}
}

func (r *BuildSQLite) Get(ctx context.Context, id int64) (domain.Build, error) {
	return r.get(ctx, id)
}

func (r *BuildSQLite) List(ctx context.Context) ([]domain.Build, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT b.id, b.name, b.character_id, c.name, c.icon_path,
		       b.character_level, b.sequence, b.weapon_id,
		       COALESCE(w.name, ''), COALESCE(w.icon_path, ''),
		       b.weapon_level, b.weapon_rank,
		       b.normal_attack_level, b.resonance_skill_level, b.forte_level,
		       b.liberation_level, b.intro_level,
		       b.notes, b.favorite, b.locked,
		       b.game_version, b.created_at, b.updated_at,
		       b.target_enemy_id,b.rotation_id,b.conditions
		FROM builds b
		JOIN characters c ON c.id=b.character_id
		LEFT JOIN weapons w ON w.id=b.weapon_id
		WHERE b.deleted_at IS NULL
		ORDER BY b.favorite DESC, b.updated_at DESC, b.id DESC
	`)
	if err != nil {
		return nil, err
	}
	builds := []domain.Build{}
	for rows.Next() {
		var build domain.Build
		var weaponID sql.NullInt64
		if err := rows.Scan(
			&build.ID, &build.Name, &build.CharacterID, &build.CharacterName,
			&build.CharacterIcon, &build.CharacterLevel, &build.Sequence, &weaponID,
			&build.WeaponName, &build.WeaponIcon, &build.WeaponLevel, &build.WeaponRank,
			&build.NormalAttackLevel, &build.ResonanceSkillLevel, &build.ForteLevel,
			&build.LiberationLevel, &build.IntroLevel,
			&build.Notes, &build.Favorite, &build.Locked, &build.GameVersion,
			&build.CreatedAt, &build.UpdatedAt, &build.TargetEnemyID, &build.RotationID, &build.Conditions,
		); err != nil {
			return nil, err
		}
		if weaponID.Valid {
			value := weaponID.Int64
			build.WeaponID = &value
		}
		builds = append(builds, build)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for index := range builds {
		builds[index].Echoes, err = r.loadEchoes(ctx, builds[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return builds, nil
}

func (r *BuildSQLite) Save(ctx context.Context, build domain.Build) (domain.Build, error) {
	if build.ID == 0 {
		result, err := r.db.ExecContext(ctx, `
			INSERT INTO builds(
				name, character_id, character_level, sequence, weapon_id,
				weapon_level, weapon_rank, normal_attack_level, resonance_skill_level,
				forte_level, liberation_level, intro_level,
				notes, favorite, locked, game_version,
				target_enemy_id,rotation_id,conditions
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, build.Name, build.CharacterID, build.CharacterLevel, build.Sequence,
			nullableID(build.WeaponID), build.WeaponLevel, build.WeaponRank,
			build.NormalAttackLevel, build.ResonanceSkillLevel, build.ForteLevel,
			build.LiberationLevel, build.IntroLevel,
			build.Notes, build.Favorite, build.Locked, build.GameVersion,
			nullableID(build.TargetEnemyID), nullableID(build.RotationID), build.Conditions)
		if err != nil {
			return domain.Build{}, err
		}
		build.ID, err = result.LastInsertId()
		if err != nil {
			return domain.Build{}, err
		}
	} else {
		if previous, getErr := r.get(ctx, build.ID); getErr == nil {
			if snapshot, marshalErr := json.Marshal(previous); marshalErr == nil {
				_, _ = r.db.ExecContext(ctx, "INSERT INTO build_versions(build_id,snapshot_json) VALUES(?,?)", build.ID, string(snapshot))
			}
		}
		result, err := r.db.ExecContext(ctx, `
			UPDATE builds SET
				name=?, character_id=?, character_level=?, sequence=?, weapon_id=?,
				weapon_level=?, weapon_rank=?, normal_attack_level=?, resonance_skill_level=?,
				forte_level=?, liberation_level=?, intro_level=?,
				notes=?, favorite=?, locked=?,
				game_version=?, target_enemy_id=?,rotation_id=?,conditions=?,updated_at=CURRENT_TIMESTAMP
			WHERE id=? AND deleted_at IS NULL
		`, build.Name, build.CharacterID, build.CharacterLevel, build.Sequence,
			nullableID(build.WeaponID), build.WeaponLevel, build.WeaponRank,
			build.NormalAttackLevel, build.ResonanceSkillLevel, build.ForteLevel,
			build.LiberationLevel, build.IntroLevel,
			build.Notes, build.Favorite, build.Locked, build.GameVersion,
			nullableID(build.TargetEnemyID), nullableID(build.RotationID), build.Conditions, build.ID)
		if err != nil {
			return domain.Build{}, err
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return domain.Build{}, sql.ErrNoRows
		}
	}
	if err := r.saveEchoes(ctx, build.ID, build.Echoes); err != nil {
		return domain.Build{}, err
	}
	return r.get(ctx, build.ID)
}

func (r *BuildSQLite) Duplicate(ctx context.Context, id int64) (domain.Build, error) {
	build, err := r.get(ctx, id)
	if err != nil {
		return domain.Build{}, err
	}
	build.ID = 0
	build.Name += " — cópia"
	build.Locked = false
	build.Favorite = false
	return r.Save(ctx, build)
}

func (r *BuildSQLite) SoftDelete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "UPDATE builds SET deleted_at=CURRENT_TIMESTAMP WHERE id=? AND deleted_at IS NULL AND locked=0", id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *BuildSQLite) Restore(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "UPDATE builds SET deleted_at=NULL, updated_at=CURRENT_TIMESTAMP WHERE id=? AND deleted_at IS NOT NULL", id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *BuildSQLite) get(ctx context.Context, id int64) (domain.Build, error) {
	var build domain.Build
	var weaponID sql.NullInt64
	err := r.db.QueryRowContext(ctx, `
		SELECT b.id, b.name, b.character_id, c.name, c.icon_path,
		       b.character_level, b.sequence, b.weapon_id,
		       COALESCE(w.name, ''), COALESCE(w.icon_path, ''),
		       b.weapon_level, b.weapon_rank,
		       b.normal_attack_level, b.resonance_skill_level, b.forte_level,
		       b.liberation_level, b.intro_level,
		       b.notes, b.favorite, b.locked,
		       b.game_version, b.created_at, b.updated_at,
		       b.target_enemy_id,b.rotation_id,b.conditions
		FROM builds b JOIN characters c ON c.id=b.character_id
		LEFT JOIN weapons w ON w.id=b.weapon_id
		WHERE b.id=?
	`, id).Scan(
		&build.ID, &build.Name, &build.CharacterID, &build.CharacterName,
		&build.CharacterIcon, &build.CharacterLevel, &build.Sequence, &weaponID,
		&build.WeaponName, &build.WeaponIcon, &build.WeaponLevel, &build.WeaponRank,
		&build.NormalAttackLevel, &build.ResonanceSkillLevel, &build.ForteLevel,
		&build.LiberationLevel, &build.IntroLevel,
		&build.Notes, &build.Favorite, &build.Locked, &build.GameVersion,
		&build.CreatedAt, &build.UpdatedAt, &build.TargetEnemyID, &build.RotationID, &build.Conditions,
	)
	if err != nil {
		return domain.Build{}, fmt.Errorf("get build %d: %w", id, err)
	}
	if weaponID.Valid {
		value := weaponID.Int64
		build.WeaponID = &value
	}
	build.Echoes, err = r.loadEchoes(ctx, build.ID)
	if err != nil {
		return domain.Build{}, fmt.Errorf("get build echoes %d: %w", id, err)
	}
	return build, nil
}

func (r *BuildSQLite) History(ctx context.Context, id int64) ([]domain.BuildVersion, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id,build_id,snapshot_json,created_at FROM build_versions WHERE build_id=? ORDER BY id DESC LIMIT 30", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.BuildVersion{}
	for rows.Next() {
		var item domain.BuildVersion
		if err := rows.Scan(&item.ID, &item.BuildID, &item.Snapshot, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *BuildSQLite) saveEchoes(ctx context.Context, buildID int64, echoes []domain.OwnedEcho) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "DELETE FROM build_echoes WHERE build_id=?", buildID); err != nil {
		return err
	}
	for index, echo := range echoes {
		if index >= 5 || echo.ID <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO build_echoes(build_id,slot,owned_echo_id) VALUES(?,?,?)", buildID, index+1, echo.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *BuildSQLite) loadEchoes(ctx context.Context, buildID int64) ([]domain.OwnedEcho, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT oe.id,oe.echo_id,e.name,e.icon_path,e.cost,oe.main_stat,
		oe.substats_json,oe.level,oe.sonata_id,COALESCE(es.name,''),oe.character_id,
		COALESCE(c.name,''),oe.locked,oe.favorite,oe.note
		FROM build_echoes be JOIN owned_echoes oe ON oe.id=be.owned_echo_id
		JOIN echoes e ON e.id=oe.echo_id
		LEFT JOIN echo_sets es ON es.id=oe.sonata_id LEFT JOIN characters c ON c.id=oe.character_id
		WHERE be.build_id=? ORDER BY be.slot`, buildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	echoes := []domain.OwnedEcho{}
	for rows.Next() {
		echo, err := scanOwnedEcho(rows)
		if err != nil {
			return nil, err
		}
		echoes = append(echoes, echo)
	}
	return echoes, rows.Err()
}

func nullableID(id *int64) any {
	if id == nil || *id == 0 {
		return nil
	}
	return *id
}
