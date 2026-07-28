package repository

import (
	"context"
	"database/sql"
	"strings"

	"wavearchive/internal/domain"
)

type EchoSQLite struct{ db *sql.DB }

func NewEchoSQLite(db *sql.DB) *EchoSQLite { return &EchoSQLite{db: db} }

func (r *EchoSQLite) List(ctx context.Context, filter domain.EchoFilter) ([]domain.Echo, error) {
	where := []string{"1=1"}
	args := []any{}
	if query := ftsQuery(filter.Query); query != "" {
		where = append(where, "e.id IN (SELECT rowid FROM echoes_fts WHERE echoes_fts MATCH ?)")
		args = append(args, query)
	}
	if filter.Cost > 0 {
		where = append(where, "e.cost=?")
		args = append(args, filter.Cost)
	}
	if filter.SonataID > 0 {
		where = append(where, "EXISTS(SELECT 1 FROM json_each(e.sonata_ids_json) WHERE value=?)")
		args = append(args, filter.SonataID)
	}
	if filter.OwnedOnly {
		where = append(where, "EXISTS(SELECT 1 FROM owned_echoes oe WHERE oe.echo_id=e.id)")
	}
	order := "e.name COLLATE NOCASE"
	if filter.Sort == "cost" {
		order = "e.cost DESC, e.name COLLATE NOCASE"
	} else if filter.Sort == "id" {
		order = "e.id"
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.id,e.name,e.code,e.echo_type,e.class,e.cost,e.place,e.icon_path,
		       e.skill,e.rarities_json,e.sonata_ids_json,e.game_version,
		       COUNT(oe.id),COALESCE(MAX(oe.favorite),0)
		FROM echoes e LEFT JOIN owned_echoes oe ON oe.echo_id=e.id
		WHERE `+strings.Join(where, " AND ")+`
		GROUP BY e.id ORDER BY `+order, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	echoes := []domain.Echo{}
	for rows.Next() {
		echo, err := scanEcho(rows)
		if err != nil {
			return nil, err
		}
		echoes = append(echoes, echo)
	}
	return echoes, rows.Err()
}

func (r *EchoSQLite) Get(ctx context.Context, id int64) (domain.Echo, error) {
	return scanEcho(r.db.QueryRowContext(ctx, `
		SELECT e.id,e.name,e.code,e.echo_type,e.class,e.cost,e.place,e.icon_path,
		       e.skill,e.rarities_json,e.sonata_ids_json,e.game_version,
		       COUNT(oe.id),COALESCE(MAX(oe.favorite),0)
		FROM echoes e LEFT JOIN owned_echoes oe ON oe.echo_id=e.id WHERE e.id=? GROUP BY e.id
	`, id))
}

func scanEcho(row scanner) (domain.Echo, error) {
	var echo domain.Echo
	err := row.Scan(&echo.ID, &echo.Name, &echo.Code, &echo.Type, &echo.Class, &echo.Cost,
		&echo.Place, &echo.IconPath, &echo.Skill, &echo.RaritiesJSON,
		&echo.SonataIDsJSON, &echo.GameVersion, &echo.OwnedCount, &echo.Favorite)
	return echo, err
}

func (r *EchoSQLite) ListSonatas(ctx context.Context) ([]domain.Sonata, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id,name,icon_path,two_piece,five_piece,game_version FROM echo_sets ORDER BY name COLLATE NOCASE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sets := []domain.Sonata{}
	for rows.Next() {
		var set domain.Sonata
		if err := rows.Scan(&set.ID, &set.Name, &set.IconPath, &set.TwoPiece, &set.FivePiece, &set.GameVersion); err != nil {
			return nil, err
		}
		sets = append(sets, set)
	}
	return sets, rows.Err()
}

func (r *EchoSQLite) ReplaceSynced(ctx context.Context, version string, echoes []domain.Echo, sonatas []domain.Sonata) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, set := range sonatas {
		if _, err := tx.ExecContext(ctx, `INSERT INTO echo_sets(id,name,icon_path,two_piece,five_piece,game_version)
			VALUES(?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,icon_path=excluded.icon_path,
			two_piece=excluded.two_piece,five_piece=excluded.five_piece,game_version=excluded.game_version`,
			set.ID, set.Name, set.IconPath, set.TwoPiece, set.FivePiece, version); err != nil {
			return err
		}
	}
	for _, echo := range echoes {
		if _, err := tx.ExecContext(ctx, `INSERT INTO echoes(id,name,code,echo_type,class,cost,place,icon_path,skill,rarities_json,sonata_ids_json,game_version)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,code=excluded.code,
			echo_type=excluded.echo_type,class=excluded.class,cost=excluded.cost,place=excluded.place,
			icon_path=excluded.icon_path,skill=excluded.skill,rarities_json=excluded.rarities_json,
			sonata_ids_json=excluded.sonata_ids_json,game_version=excluded.game_version`,
			echo.ID, echo.Name, echo.Code, echo.Type, echo.Class, echo.Cost, echo.Place,
			echo.IconPath, echo.Skill, echo.RaritiesJSON, echo.SonataIDsJSON, version); err != nil {
			return err
		}
	}
	echoIDs := make([]int64, 0, len(echoes))
	for _, echo := range echoes {
		echoIDs = append(echoIDs, echo.ID)
	}
	if err := deleteStaleEchoes(ctx, tx, echoIDs); err != nil {
		return err
	}
	sonataIDs := make([]int64, 0, len(sonatas))
	for _, sonata := range sonatas {
		sonataIDs = append(sonataIDs, sonata.ID)
	}
	if err := deleteStaleSonatas(ctx, tx, sonataIDs); err != nil {
		return err
	}
	return tx.Commit()
}

func deleteStaleEchoes(ctx context.Context, tx *sql.Tx, active []int64) error {
	query := `DELETE FROM echoes
		WHERE NOT EXISTS (SELECT 1 FROM owned_echoes oe WHERE oe.echo_id=echoes.id)`
	args := make([]any, len(active))
	if len(active) > 0 {
		query += " AND id NOT IN (" + strings.TrimSuffix(strings.Repeat("?,", len(active)), ",") + ")"
		for index, id := range active {
			args[index] = id
		}
	}
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func deleteStaleSonatas(ctx context.Context, tx *sql.Tx, active []int64) error {
	query := `DELETE FROM echo_sets
		WHERE NOT EXISTS (SELECT 1 FROM owned_echoes oe WHERE oe.sonata_id=echo_sets.id)`
	args := make([]any, len(active))
	if len(active) > 0 {
		query += " AND id NOT IN (" + strings.TrimSuffix(strings.Repeat("?,", len(active)), ",") + ")"
		for index, id := range active {
			args[index] = id
		}
	}
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *EchoSQLite) ListOwned(ctx context.Context) ([]domain.OwnedEcho, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT oe.id,oe.echo_id,e.name,e.icon_path,e.cost,oe.main_stat,
		oe.substats_json,oe.level,oe.sonata_id,COALESCE(es.name,''),oe.character_id,
		COALESCE(c.name,''),oe.locked,oe.favorite,oe.note
		FROM owned_echoes oe JOIN echoes e ON e.id=oe.echo_id
		LEFT JOIN echo_sets es ON es.id=oe.sonata_id LEFT JOIN characters c ON c.id=oe.character_id
		ORDER BY oe.favorite DESC,oe.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.OwnedEcho{}
	for rows.Next() {
		item, err := scanOwnedEcho(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *EchoSQLite) SaveOwned(ctx context.Context, item domain.OwnedEcho) (domain.OwnedEcho, error) {
	if item.ID == 0 {
		result, err := r.db.ExecContext(ctx, `INSERT INTO owned_echoes(echo_id,main_stat,substats_json,level,sonata_id,character_id,locked,favorite,note)
			VALUES(?,?,?,?,?,?,?,?,?)`, item.EchoID, item.MainStat, item.SubstatsJSON, item.Level,
			nullableID(item.SonataID), nullableID(item.CharacterID), item.Locked, item.Favorite, item.Note)
		if err != nil {
			return item, err
		}
		item.ID, _ = result.LastInsertId()
	} else {
		_, err := r.db.ExecContext(ctx, `UPDATE owned_echoes SET echo_id=?,main_stat=?,substats_json=?,level=?,
			sonata_id=?,character_id=?,locked=?,favorite=?,note=? WHERE id=?`,
			item.EchoID, item.MainStat, item.SubstatsJSON, item.Level, nullableID(item.SonataID),
			nullableID(item.CharacterID), item.Locked, item.Favorite, item.Note, item.ID)
		if err != nil {
			return item, err
		}
	}
	return r.getOwned(ctx, item.ID)
}

func (r *EchoSQLite) DeleteOwned(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM owned_echoes WHERE id=? AND locked=0", id)
	if err != nil {
		return err
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *EchoSQLite) getOwned(ctx context.Context, id int64) (domain.OwnedEcho, error) {
	return scanOwnedEcho(r.db.QueryRowContext(ctx, `SELECT oe.id,oe.echo_id,e.name,e.icon_path,e.cost,oe.main_stat,
		oe.substats_json,oe.level,oe.sonata_id,COALESCE(es.name,''),oe.character_id,
		COALESCE(c.name,''),oe.locked,oe.favorite,oe.note
		FROM owned_echoes oe JOIN echoes e ON e.id=oe.echo_id
		LEFT JOIN echo_sets es ON es.id=oe.sonata_id LEFT JOIN characters c ON c.id=oe.character_id WHERE oe.id=?`, id))
}

func scanOwnedEcho(row scanner) (domain.OwnedEcho, error) {
	var item domain.OwnedEcho
	var sonata, character sql.NullInt64
	err := row.Scan(&item.ID, &item.EchoID, &item.EchoName, &item.IconPath, &item.Cost,
		&item.MainStat, &item.SubstatsJSON, &item.Level, &sonata, &item.SonataName,
		&character, &item.CharacterName, &item.Locked, &item.Favorite, &item.Note)
	if sonata.Valid {
		value := sonata.Int64
		item.SonataID = &value
	}
	if character.Valid {
		value := character.Int64
		item.CharacterID = &value
	}
	return item, err
}
