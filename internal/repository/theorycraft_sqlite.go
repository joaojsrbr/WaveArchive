package repository

import (
	"context"
	"database/sql"

	"wavearchive/internal/domain"
)

type TheorycraftSQLite struct{ db *sql.DB }

func NewTheorycraftSQLite(db *sql.DB) *TheorycraftSQLite { return &TheorycraftSQLite{db: db} }

func defaultBuildConfig(buildID int64) domain.BuildConfig {
	return domain.BuildConfig{
		BuildID: buildID, ScalingType: "ATK", BaseATK: 1000, BaseHP: 10000, BaseDEF: 1000,
		MotionValue: 2, EnemyLevel: 90, EnemyResistance: .1, ExtraDamageBonusesJSON: "[]",
	}
}

func (r *TheorycraftSQLite) GetBuildConfig(ctx context.Context, buildID int64) (domain.BuildConfig, error) {
	config := defaultBuildConfig(buildID)
	err := r.db.QueryRowContext(ctx, `SELECT build_id,scaling_type,base_atk,base_hp,base_def,motion_value,
		flat_damage,enemy_level,enemy_resistance,defense_ignore,damage_reduction,element_reduction,
		extra_damage_bonuses_json FROM build_configs WHERE build_id=?`, buildID).Scan(
		&config.BuildID, &config.ScalingType, &config.BaseATK, &config.BaseHP, &config.BaseDEF,
		&config.MotionValue, &config.FlatDamage, &config.EnemyLevel, &config.EnemyResistance,
		&config.DefenseIgnore, &config.DamageReduction, &config.ElementReduction,
		&config.ExtraDamageBonusesJSON,
	)
	if err == sql.ErrNoRows {
		return config, nil
	}
	return config, err
}

func (r *TheorycraftSQLite) SaveBuildConfig(ctx context.Context, config domain.BuildConfig) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO build_configs(build_id,scaling_type,base_atk,base_hp,
		base_def,motion_value,flat_damage,enemy_level,enemy_resistance,defense_ignore,damage_reduction,
		element_reduction,extra_damage_bonuses_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(build_id) DO UPDATE SET scaling_type=excluded.scaling_type,base_atk=excluded.base_atk,
		base_hp=excluded.base_hp,base_def=excluded.base_def,motion_value=excluded.motion_value,
		flat_damage=excluded.flat_damage,enemy_level=excluded.enemy_level,
		enemy_resistance=excluded.enemy_resistance,defense_ignore=excluded.defense_ignore,
		damage_reduction=excluded.damage_reduction,element_reduction=excluded.element_reduction,
		extra_damage_bonuses_json=excluded.extra_damage_bonuses_json,updated_at=CURRENT_TIMESTAMP`,
		config.BuildID, config.ScalingType, config.BaseATK, config.BaseHP, config.BaseDEF,
		config.MotionValue, config.FlatDamage, config.EnemyLevel, config.EnemyResistance,
		config.DefenseIgnore, config.DamageReduction, config.ElementReduction,
		config.ExtraDamageBonusesJSON)
	return err
}

func (r *TheorycraftSQLite) ListBuffs(ctx context.Context, teamID int64) ([]domain.Buff, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,team_id,source_slot,target_slot,name,modifier_group,
		value,scope,condition_text,assume_active,duration,trigger_action FROM buffs WHERE team_id=? ORDER BY source_slot,id`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	buffs := []domain.Buff{}
	for rows.Next() {
		var buff domain.Buff
		if err := rows.Scan(&buff.ID, &buff.TeamID, &buff.SourceSlot, &buff.TargetSlot, &buff.Name,
			&buff.Group, &buff.Value, &buff.Scope, &buff.Condition, &buff.Active, &buff.Duration, &buff.TriggerAction); err != nil {
			return nil, err
		}
		buffs = append(buffs, buff)
	}
	return buffs, rows.Err()
}

func (r *TheorycraftSQLite) SaveBuff(ctx context.Context, buff domain.Buff) (domain.Buff, error) {
	if buff.ID == 0 {
		result, err := r.db.ExecContext(ctx, `INSERT INTO buffs(team_id,source_slot,target_slot,name,
			modifier_group,value,scope,condition_text,assume_active,duration,trigger_action) VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
			buff.TeamID, buff.SourceSlot, buff.TargetSlot, buff.Name, buff.Group, buff.Value,
			buff.Scope, buff.Condition, buff.Active, buff.Duration, buff.TriggerAction)
		if err != nil {
			return buff, err
		}
		buff.ID, _ = result.LastInsertId()
		return buff, nil
	}
	_, err := r.db.ExecContext(ctx, `UPDATE buffs SET source_slot=?,target_slot=?,name=?,modifier_group=?,
		value=?,scope=?,condition_text=?,assume_active=?,duration=?,trigger_action=? WHERE id=? AND team_id=?`, buff.SourceSlot,
		buff.TargetSlot, buff.Name, buff.Group, buff.Value, buff.Scope, buff.Condition,
		buff.Active, buff.Duration, buff.TriggerAction, buff.ID, buff.TeamID)
	return buff, err
}

func (r *TheorycraftSQLite) DeleteBuff(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM buffs WHERE id=?", id)
	return err
}

func (r *TheorycraftSQLite) ListRotations(ctx context.Context, teamID int64) ([]domain.Rotation, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id,team_id,name,duration,notes FROM rotations WHERE team_id=? ORDER BY id DESC", teamID)
	if err != nil {
		return nil, err
	}
	rotations := []domain.Rotation{}
	for rows.Next() {
		var rotation domain.Rotation
		if err := rows.Scan(&rotation.ID, &rotation.TeamID, &rotation.Name, &rotation.Duration, &rotation.Notes); err != nil {
			rows.Close()
			return nil, err
		}
		rotations = append(rotations, rotation)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for index := range rotations {
		rotations[index].Actions, err = r.actions(ctx, rotations[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return rotations, nil
}

func (r *TheorycraftSQLite) SaveRotation(ctx context.Context, rotation domain.Rotation) (domain.Rotation, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return rotation, err
	}
	defer tx.Rollback()
	if rotation.ID == 0 {
		result, err := tx.ExecContext(ctx, "INSERT INTO rotations(team_id,name,duration,notes) VALUES(?,?,?,?)",
			rotation.TeamID, rotation.Name, rotation.Duration, rotation.Notes)
		if err != nil {
			return rotation, err
		}
		rotation.ID, _ = result.LastInsertId()
	} else {
		if _, err := tx.ExecContext(ctx, "UPDATE rotations SET name=?,duration=?,notes=? WHERE id=? AND team_id=?",
			rotation.Name, rotation.Duration, rotation.Notes, rotation.ID, rotation.TeamID); err != nil {
			return rotation, err
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM rotation_actions WHERE rotation_id=?", rotation.ID); err != nil {
			return rotation, err
		}
	}
	for index, action := range rotation.Actions {
		if _, err := tx.ExecContext(ctx, `INSERT INTO rotation_actions(rotation_id,sort_order,slot,
			action_type,name,motion_value,cast_time,energy,concerto,cooldown,energy_cost,notes) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
			rotation.ID, index+1, action.Slot, action.ActionType, action.Name,
			action.MotionValue, action.CastTime, action.Energy, action.Concerto, action.Cooldown, action.EnergyCost, action.Notes); err != nil {
			return rotation, err
		}
	}
	if err := tx.Commit(); err != nil {
		return rotation, err
	}
	rotation.Actions, err = r.actions(ctx, rotation.ID)
	return rotation, err
}

func (r *TheorycraftSQLite) DeleteRotation(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM rotations WHERE id=?", id)
	return err
}

func (r *TheorycraftSQLite) actions(ctx context.Context, rotationID int64) ([]domain.RotationAction, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,sort_order,slot,action_type,name,motion_value,
		cast_time,energy,concerto,cooldown,energy_cost,notes FROM rotation_actions WHERE rotation_id=? ORDER BY sort_order`, rotationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	actions := []domain.RotationAction{}
	for rows.Next() {
		var action domain.RotationAction
		if err := rows.Scan(&action.ID, &action.Order, &action.Slot, &action.ActionType,
			&action.Name, &action.MotionValue, &action.CastTime, &action.Energy, &action.Concerto, &action.Cooldown, &action.EnergyCost, &action.Notes); err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return actions, rows.Err()
}
