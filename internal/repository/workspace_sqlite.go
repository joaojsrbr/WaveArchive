package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"wavearchive/internal/domain"
)

type WorkspaceSQLite struct{ db *sql.DB }

func NewWorkspaceSQLite(db *sql.DB) *WorkspaceSQLite { return &WorkspaceSQLite{db: db} }

func (r *WorkspaceSQLite) GetSettings(ctx context.Context) (domain.AppSettings, error) {
	values := map[string]string{}
	rows, err := r.db.QueryContext(ctx, "SELECT key,value FROM app_settings")
	if err != nil {
		return domain.AppSettings{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return domain.AppSettings{}, err
		}
		values[k] = v
	}
	return domain.AppSettings{
		Density: values["density"], SidebarCollapsed: values["sidebar_collapsed"] == "true",
		AIProvider: values["ai_provider"], AIEndpoint: values["ai_endpoint"], AIModel: values["ai_model"],
		AIMode: values["ai_mode"], ReduceMotion: values["reduce_motion"] == "true",
	}, rows.Err()
}

func (r *WorkspaceSQLite) SaveSettings(ctx context.Context, s domain.AppSettings) (domain.AppSettings, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return s, err
	}
	defer tx.Rollback()
	values := map[string]string{
		"density": s.Density, "sidebar_collapsed": boolText(s.SidebarCollapsed), "ai_provider": s.AIProvider,
		"ai_endpoint": s.AIEndpoint, "ai_model": s.AIModel, "ai_mode": s.AIMode, "reduce_motion": boolText(s.ReduceMotion),
	}
	for key, value := range values {
		if _, err := tx.ExecContext(ctx, `INSERT INTO app_settings(key,value,updated_at) VALUES(?,?,CURRENT_TIMESTAMP)
			ON CONFLICT(key) DO UPDATE SET value=excluded.value,updated_at=CURRENT_TIMESTAMP`, key, value); err != nil {
			return s, err
		}
	}
	if err := tx.Commit(); err != nil {
		return s, err
	}
	return r.GetSettings(ctx)
}

func (r *WorkspaceSQLite) GetAccount(ctx context.Context) (domain.AccountSummary, error) {
	var a domain.AccountSummary
	err := r.db.QueryRowContext(ctx, `SELECT a.id,a.name,a.notes,a.astrite,a.radiant_tides,
		(SELECT COUNT(*) FROM owned_characters WHERE profile_id=a.id AND owned=1),
		(SELECT COUNT(*) FROM owned_weapons WHERE profile_id=a.id AND owned=1),
		(SELECT COUNT(*) FROM owned_echoes WHERE profile_id=a.id)
		FROM accounts a WHERE a.active=1 ORDER BY a.id LIMIT 1`).Scan(
		&a.ID, &a.Name, &a.Notes, &a.Astrite, &a.RadiantTides, &a.OwnedCharacters, &a.OwnedWeapons, &a.OwnedEchoes)
	return a, err
}

func (r *WorkspaceSQLite) SaveAccount(ctx context.Context, a domain.AccountSummary) (domain.AccountSummary, error) {
	if a.ID == 0 {
		a.ID = 1
	}
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET name=?,notes=?,astrite=?,radiant_tides=?,updated_at=CURRENT_TIMESTAMP WHERE id=?",
		strings.TrimSpace(a.Name), a.Notes, a.Astrite, a.RadiantTides, a.ID)
	if err != nil {
		return a, err
	}
	return r.GetAccount(ctx)
}

func (r *WorkspaceSQLite) ListGoals(ctx context.Context) ([]domain.PlannerGoal, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,title,goal_type,target_name,required_amount,owned_amount,shell_credits,
		priority,due_date,completed,notes,created_at,updated_at FROM planner_goals ORDER BY completed,priority,id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.PlannerGoal{}
	for rows.Next() {
		var g domain.PlannerGoal
		if err := rows.Scan(&g.ID, &g.Title, &g.GoalType, &g.TargetName, &g.RequiredAmount,
			&g.OwnedAmount, &g.ShellCredits, &g.Priority, &g.DueDate, &g.Completed, &g.Notes, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, g)
	}
	return items, rows.Err()
}

func (r *WorkspaceSQLite) SaveGoal(ctx context.Context, g domain.PlannerGoal) (domain.PlannerGoal, error) {
	if g.ID == 0 {
		res, err := r.db.ExecContext(ctx, `INSERT INTO planner_goals(title,goal_type,target_name,required_amount,owned_amount,shell_credits,priority,due_date,completed,notes)
			VALUES(?,?,?,?,?,?,?,?,?,?)`, g.Title, g.GoalType, g.TargetName, g.RequiredAmount, g.OwnedAmount, g.ShellCredits, g.Priority, g.DueDate, g.Completed, g.Notes)
		if err != nil {
			return g, err
		}
		g.ID, err = res.LastInsertId()
		return g, err
	}
	_, err := r.db.ExecContext(ctx, `UPDATE planner_goals SET title=?,goal_type=?,target_name=?,required_amount=?,owned_amount=?,
		shell_credits=?,priority=?,due_date=?,completed=?,notes=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		g.Title, g.GoalType, g.TargetName, g.RequiredAmount, g.OwnedAmount, g.ShellCredits, g.Priority, g.DueDate, g.Completed, g.Notes, g.ID)
	return g, err
}
func (r *WorkspaceSQLite) DeleteGoal(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM planner_goals WHERE id=?", id)
	return err
}

func (r *WorkspaceSQLite) ListConvenes(ctx context.Context) ([]domain.ConveneRecord, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,banner,banner_type,item_name,rarity,pull_number,guaranteed,obtained_at,notes
		FROM convene_records ORDER BY obtained_at DESC,id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.ConveneRecord{}
	for rows.Next() {
		var c domain.ConveneRecord
		if err := rows.Scan(&c.ID, &c.Banner, &c.BannerType, &c.ItemName, &c.Rarity, &c.PullNumber, &c.Guaranteed, &c.ObtainedAt, &c.Notes); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}
func (r *WorkspaceSQLite) SaveConvene(ctx context.Context, c domain.ConveneRecord) (domain.ConveneRecord, error) {
	if c.ID == 0 {
		res, err := r.db.ExecContext(ctx, `INSERT INTO convene_records(banner,banner_type,item_name,rarity,pull_number,guaranteed,obtained_at,notes) VALUES(?,?,?,?,?,?,?,?)`, c.Banner, c.BannerType, c.ItemName, c.Rarity, c.PullNumber, c.Guaranteed, c.ObtainedAt, c.Notes)
		if err != nil {
			return c, err
		}
		c.ID, err = res.LastInsertId()
		return c, err
	}
	_, err := r.db.ExecContext(ctx, `UPDATE convene_records SET banner=?,banner_type=?,item_name=?,rarity=?,pull_number=?,guaranteed=?,obtained_at=?,notes=? WHERE id=?`, c.Banner, c.BannerType, c.ItemName, c.Rarity, c.PullNumber, c.Guaranteed, c.ObtainedAt, c.Notes, c.ID)
	return c, err
}
func (r *WorkspaceSQLite) DeleteConvene(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM convene_records WHERE id=?", id)
	return err
}

func (r *WorkspaceSQLite) ListEnemies(ctx context.Context) ([]domain.Enemy, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id,name,level,resistance,damage_reduction,element_reduction,notes FROM enemies ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Enemy{}
	for rows.Next() {
		var e domain.Enemy
		if err := rows.Scan(&e.ID, &e.Name, &e.Level, &e.Resistance, &e.DamageReduction, &e.ElementReduction, &e.Notes); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}
func (r *WorkspaceSQLite) ListFormulaVersions(ctx context.Context) ([]domain.FormulaVersion, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,game_version,defense_constant,level_factor,confidence,references_text,rounding_policy,active FROM formula_versions ORDER BY active DESC,id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.FormulaVersion{}
	for rows.Next() {
		var f domain.FormulaVersion
		if err := rows.Scan(&f.ID, &f.Name, &f.GameVersion, &f.DefenseConstant, &f.LevelFactor, &f.Confidence, &f.References, &f.RoundingPolicy, &f.Active); err != nil {
			return nil, err
		}
		items = append(items, f)
	}
	return items, rows.Err()
}
func (r *WorkspaceSQLite) Dashboard(ctx context.Context) (domain.DashboardSummary, error) {
	var d domain.DashboardSummary
	err := r.db.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM characters),(SELECT COUNT(*) FROM weapons),(SELECT COUNT(*) FROM echoes),
		(SELECT COUNT(*) FROM builds WHERE deleted_at IS NULL),(SELECT COUNT(*) FROM teams WHERE deleted_at IS NULL)`).Scan(&d.Characters, &d.Weapons, &d.Echoes, &d.Builds, &d.Teams)
	if err != nil {
		return d, err
	}
	return d, nil
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

var _ = errors.Is
