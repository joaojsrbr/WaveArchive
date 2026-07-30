package repository

import (
	"context"
	"database/sql"
	"errors"

	"wavearchive/internal/domain"
)

type ConveneHistorySQLite struct{ db *sql.DB }

func NewConveneHistorySQLite(db *sql.DB) *ConveneHistorySQLite {
	return &ConveneHistorySQLite{db: db}
}

func (r *ConveneHistorySQLite) SaveImportedPulls(
	ctx context.Context,
	payload domain.ConveneImportPayload,
) (domain.ConveneProfile, int, int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ConveneProfile{}, 0, 0, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO convene_profiles(player_id,server_id,region,language_code,history_partial,last_imported_at)
		VALUES(?,?,?,?,1,CURRENT_TIMESTAMP)
		ON CONFLICT(player_id) DO UPDATE SET
			server_id=excluded.server_id,
			region=excluded.region,
			language_code=excluded.language_code,
			history_partial=1,
			last_imported_at=CURRENT_TIMESTAMP`,
		payload.PlayerID, payload.ServerID, payload.Region, payload.LanguageCode,
	)
	if err != nil {
		return domain.ConveneProfile{}, 0, 0, err
	}

	var profile domain.ConveneProfile
	if err := tx.QueryRowContext(ctx, `
		SELECT id,player_id,server_id,region,language_code,last_imported_at,history_partial
		FROM convene_profiles WHERE player_id=?`, payload.PlayerID,
	).Scan(
		&profile.ID,
		&profile.PlayerID,
		&profile.ServerID,
		&profile.Region,
		&profile.LanguageCode,
		&profile.LastImportedAt,
		&profile.HistoryPartial,
	); err != nil {
		return domain.ConveneProfile{}, 0, 0, err
	}

	imported := 0
	for _, pool := range payload.Pools {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO convene_pool_catalog(
				profile_id,pool_type,locale_key,name,short_name,kind,hard_pity,sort_order,updated_at
			) VALUES(?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP)
			ON CONFLICT(profile_id,pool_type) DO UPDATE SET
				locale_key=excluded.locale_key,
				name=excluded.name,
				short_name=excluded.short_name,
				kind=excluded.kind,
				hard_pity=excluded.hard_pity,
				sort_order=excluded.sort_order,
				updated_at=CURRENT_TIMESTAMP`,
			profile.ID,
			pool.PoolType,
			pool.LocaleKey,
			pool.Name,
			pool.ShortName,
			pool.Kind,
			pool.HardPity,
			pool.SortOrder,
		); err != nil {
			return domain.ConveneProfile{}, 0, 0, err
		}
	}
	for _, pull := range payload.Pulls {
		result, err := tx.ExecContext(ctx, `
			INSERT OR IGNORE INTO convene_pulls(
				profile_id,pool_type,resource_id,resource_type,item_name,rarity,
				quantity,obtained_at,source_index,fingerprint
			) VALUES(?,?,?,?,?,?,?,?,?,?)`,
			profile.ID,
			pull.PoolType,
			pull.ResourceID,
			pull.ResourceType,
			pull.ItemName,
			pull.Rarity,
			pull.Quantity,
			pull.ObtainedAt,
			pull.SourceIndex,
			pull.Fingerprint,
		)
		if err != nil {
			return domain.ConveneProfile{}, 0, 0, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return domain.ConveneProfile{}, 0, 0, err
		}
		imported += int(affected)
	}

	if err := tx.Commit(); err != nil {
		return domain.ConveneProfile{}, 0, 0, err
	}
	return profile, imported, len(payload.Pulls) - imported, nil
}

func (r *ConveneHistorySQLite) ListConvenePoolDefinitions(
	ctx context.Context,
	profileID int64,
) ([]domain.ConvenePoolDefinition, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT pool_type,locale_key,name,short_name,kind,hard_pity,sort_order
		FROM convene_pool_catalog
		WHERE profile_id=?
		ORDER BY sort_order,pool_type`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pools := make([]domain.ConvenePoolDefinition, 0)
	for rows.Next() {
		var pool domain.ConvenePoolDefinition
		if err := rows.Scan(
			&pool.PoolType,
			&pool.LocaleKey,
			&pool.Name,
			&pool.ShortName,
			&pool.Kind,
			&pool.HardPity,
			&pool.SortOrder,
		); err != nil {
			return nil, err
		}
		pools = append(pools, pool)
	}
	return pools, rows.Err()
}

func (r *ConveneHistorySQLite) GetConveneProfile(ctx context.Context) (*domain.ConveneProfile, error) {
	var profile domain.ConveneProfile
	err := r.db.QueryRowContext(ctx, `
		SELECT id,player_id,server_id,region,language_code,last_imported_at,history_partial
		FROM convene_profiles ORDER BY last_imported_at DESC,id DESC LIMIT 1`,
	).Scan(
		&profile.ID,
		&profile.PlayerID,
		&profile.ServerID,
		&profile.Region,
		&profile.LanguageCode,
		&profile.LastImportedAt,
		&profile.HistoryPartial,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *ConveneHistorySQLite) ListConvenePulls(ctx context.Context) ([]domain.ConvenePull, error) {
	profile, err := r.GetConveneProfile(ctx)
	if err != nil || profile == nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			p.id,p.profile_id,p.pool_type,p.resource_id,
			CASE
				WHEN EXISTS(SELECT 1 FROM characters c WHERE CAST(c.id AS TEXT)=p.resource_id) THEN 'character'
				WHEN EXISTS(SELECT 1 FROM weapons w WHERE CAST(w.id AS TEXT)=p.resource_id) THEN 'weapon'
				WHEN lower(trim(p.resource_type)) IN (
					'resonator','resonante','ressonante','ressonador','character','personagem'
				) THEN 'character'
				WHEN lower(trim(p.resource_type)) IN ('weapon','arma') THEN 'weapon'
				ELSE p.resource_type
			END,
			p.item_name,
			p.rarity,p.quantity,p.obtained_at,p.source_index,
			COALESCE(
				(SELECT icon_path FROM characters WHERE CAST(id AS TEXT)=p.resource_id LIMIT 1),
				(SELECT icon_path FROM weapons WHERE CAST(id AS TEXT)=p.resource_id LIMIT 1),
				(SELECT icon_path FROM characters WHERE lower(name)=lower(p.item_name) LIMIT 1),
				(SELECT icon_path FROM weapons WHERE lower(name)=lower(p.item_name) LIMIT 1),
				''
			)
		FROM convene_pulls p
		WHERE p.profile_id=?
		ORDER BY p.obtained_at DESC,p.source_index ASC,p.id DESC`,
		profile.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pulls := make([]domain.ConvenePull, 0)
	for rows.Next() {
		var pull domain.ConvenePull
		if err := rows.Scan(
			&pull.ID,
			&pull.ProfileID,
			&pull.PoolType,
			&pull.ResourceID,
			&pull.ResourceType,
			&pull.ItemName,
			&pull.Rarity,
			&pull.Quantity,
			&pull.ObtainedAt,
			&pull.SourceIndex,
			&pull.IconPath,
		); err != nil {
			return nil, err
		}
		pulls = append(pulls, pull)
	}
	return pulls, rows.Err()
}

func (r *ConveneHistorySQLite) DeleteConveneHistory(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM convene_profiles")
	return err
}

var _ domain.ConveneHistoryRepository = (*ConveneHistorySQLite)(nil)
