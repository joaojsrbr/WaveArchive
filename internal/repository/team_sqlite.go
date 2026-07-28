package repository

import (
	"context"
	"database/sql"
	"fmt"

	"wavearchive/internal/domain"
)

type TeamSQLite struct{ db *sql.DB }

func NewTeamSQLite(db *sql.DB) *TeamSQLite { return &TeamSQLite{db: db} }

func (r *TeamSQLite) Get(ctx context.Context, id int64) (domain.Team, error) {
	return r.get(ctx, id)
}

func (r *TeamSQLite) List(ctx context.Context) ([]domain.Team, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,notes,favorite,locked,game_version,created_at,updated_at
		FROM teams WHERE deleted_at IS NULL ORDER BY favorite DESC,updated_at DESC,id DESC`)
	if err != nil {
		return nil, err
	}
	teams := []domain.Team{}
	for rows.Next() {
		var team domain.Team
		if err := rows.Scan(&team.ID, &team.Name, &team.Notes, &team.Favorite, &team.Locked,
			&team.GameVersion, &team.CreatedAt, &team.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		teams = append(teams, team)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for index := range teams {
		teams[index].Members, err = r.members(ctx, teams[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return teams, nil
}

func (r *TeamSQLite) Save(ctx context.Context, team domain.Team) (domain.Team, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Team{}, err
	}
	defer tx.Rollback()
	if team.ID == 0 {
		result, err := tx.ExecContext(ctx, `INSERT INTO teams(name,notes,favorite,locked,game_version) VALUES(?,?,?,?,?)`,
			team.Name, team.Notes, team.Favorite, team.Locked, team.GameVersion)
		if err != nil {
			return domain.Team{}, err
		}
		team.ID, err = result.LastInsertId()
		if err != nil {
			return domain.Team{}, err
		}
	} else {
		result, err := tx.ExecContext(ctx, `UPDATE teams SET name=?,notes=?,favorite=?,locked=?,game_version=?,
			updated_at=CURRENT_TIMESTAMP WHERE id=? AND deleted_at IS NULL`,
			team.Name, team.Notes, team.Favorite, team.Locked, team.GameVersion, team.ID)
		if err != nil {
			return domain.Team{}, err
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return domain.Team{}, sql.ErrNoRows
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM team_members WHERE team_id=?", team.ID); err != nil {
			return domain.Team{}, err
		}
	}
	for _, member := range team.Members {
		if _, err := tx.ExecContext(ctx, `INSERT INTO team_members(team_id,slot,character_id,build_id,role,custom_role)
			VALUES(?,?,?,?,?,?)`, team.ID, member.Slot, member.CharacterID, nullableID(member.BuildID), member.Role, member.CustomRole); err != nil {
			return domain.Team{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.Team{}, err
	}
	return r.get(ctx, team.ID)
}

func (r *TeamSQLite) Duplicate(ctx context.Context, id int64) (domain.Team, error) {
	team, err := r.get(ctx, id)
	if err != nil {
		return domain.Team{}, err
	}
	team.ID = 0
	team.Name += " — cópia"
	team.Favorite = false
	team.Locked = false
	return r.Save(ctx, team)
}

func (r *TeamSQLite) SoftDelete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "UPDATE teams SET deleted_at=CURRENT_TIMESTAMP WHERE id=? AND deleted_at IS NULL AND locked=0", id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *TeamSQLite) Restore(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "UPDATE teams SET deleted_at=NULL,updated_at=CURRENT_TIMESTAMP WHERE id=? AND deleted_at IS NOT NULL", id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *TeamSQLite) get(ctx context.Context, id int64) (domain.Team, error) {
	var team domain.Team
	err := r.db.QueryRowContext(ctx, `SELECT id,name,notes,favorite,locked,game_version,created_at,updated_at
		FROM teams WHERE id=?`, id).Scan(&team.ID, &team.Name, &team.Notes, &team.Favorite,
		&team.Locked, &team.GameVersion, &team.CreatedAt, &team.UpdatedAt)
	if err != nil {
		return domain.Team{}, fmt.Errorf("get team %d: %w", id, err)
	}
	team.Members, err = r.members(ctx, id)
	return team, err
}

func (r *TeamSQLite) members(ctx context.Context, teamID int64) ([]domain.TeamMember, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT tm.slot,tm.character_id,c.name,c.icon_path,tm.build_id,
		COALESCE(b.name,''),tm.role,tm.custom_role FROM team_members tm
		JOIN characters c ON c.id=tm.character_id LEFT JOIN builds b ON b.id=tm.build_id
		WHERE tm.team_id=? ORDER BY tm.slot`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	members := []domain.TeamMember{}
	for rows.Next() {
		var member domain.TeamMember
		var buildID sql.NullInt64
		if err := rows.Scan(&member.Slot, &member.CharacterID, &member.CharacterName,
			&member.CharacterIcon, &buildID, &member.BuildName, &member.Role, &member.CustomRole); err != nil {
			return nil, err
		}
		if buildID.Valid {
			value := buildID.Int64
			member.BuildID = &value
		}
		members = append(members, member)
	}
	return members, rows.Err()
}
