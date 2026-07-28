package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"wavearchive/internal/domain"
)

type GuideSQLite struct{ db *sql.DB }

func NewGuideSQLite(db *sql.DB) *GuideSQLite { return &GuideSQLite{db: db} }
func (r *GuideSQLite) List(ctx context.Context, characterID int64) ([]domain.CharacterGuide, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,character_id,name,source,like_count,language,data_json,synced_at FROM character_guides WHERE character_id=? ORDER BY like_count DESC,id`, characterID)
	if err != nil {
		return nil, err
	}
	return scanGuides(rows)
}
func (r *GuideSQLite) ListAll(ctx context.Context) ([]domain.CharacterGuide, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,character_id,name,source,like_count,language,data_json,synced_at FROM character_guides ORDER BY like_count DESC,character_id,id`)
	if err != nil {
		return nil, err
	}
	return scanGuides(rows)
}
func scanGuides(rows *sql.Rows) ([]domain.CharacterGuide, error) {
	defer rows.Close()
	items := []domain.CharacterGuide{}
	for rows.Next() {
		var g domain.CharacterGuide
		if err := rows.Scan(&g.ID, &g.CharacterID, &g.Name, &g.Source, &g.LikeCount, &g.Language, &g.DataJSON, &g.SyncedAt); err != nil {
			return nil, err
		}
		var wrapped struct {
			Teams [][]int64 `json:"teams"`
		}
		_ = json.Unmarshal([]byte(g.DataJSON), &wrapped)
		g.Teams = wrapped.Teams
		items = append(items, g)
	}
	return items, rows.Err()
}
func (r *GuideSQLite) Replace(ctx context.Context, characterID int64, guides []domain.CharacterGuide) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "DELETE FROM character_guides WHERE character_id=?", characterID); err != nil {
		return err
	}
	for _, g := range guides {
		if _, err := tx.ExecContext(ctx, `INSERT INTO character_guides(id,character_id,name,source,like_count,language,data_json) VALUES(?,?,?,?,?,?,?)`, g.ID, characterID, g.Name, g.Source, g.LikeCount, g.Language, g.DataJSON); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (r *GuideSQLite) Search(ctx context.Context, query string, limit int) ([]domain.KnowledgeSource, error) {
	match := ftsQuery(query)
	if match == "" {
		return []domain.KnowledgeSource{}, nil
	}
	if limit < 1 || limit > 20 {
		limit = 8
	}
	rows, err := r.db.QueryContext(ctx, `SELECT entity_type,entity_id,title,snippet(local_knowledge_fts,3,'<mark>','</mark>',' … ',18)
		FROM local_knowledge_fts WHERE local_knowledge_fts MATCH ? ORDER BY bm25(local_knowledge_fts) LIMIT ?`, match, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.KnowledgeSource{}
	for rows.Next() {
		var s domain.KnowledgeSource
		if err := rows.Scan(&s.EntityType, &s.EntityID, &s.Title, &s.Snippet); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}
