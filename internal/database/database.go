package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Database struct {
	sql  *sql.DB
	path string
}

func Open(path string) (*Database, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Database{sql: db, path: path}, nil
}

func (d *Database) SQL() *sql.DB { return d.sql }
func (d *Database) Close() error { return d.sql.Close() }

func (d *Database) Backup(destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("snapshot already exists: %s", destination)
	} else if !os.IsNotExist(err) {
		return err
	}
	quoted := strings.ReplaceAll(destination, "'", "''")
	if _, err := d.sql.Exec("VACUUM INTO '" + quoted + "'"); err != nil {
		return fmt.Errorf("create sqlite snapshot: %w", err)
	}
	return nil
}

func (d *Database) Checkpoint() error {
	_, err := d.sql.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}

func RestoreFile(snapshot, destination string) error {
	source, err := os.Open(snapshot)
	if err != nil {
		return err
	}
	defer source.Close()
	temp, err := os.CreateTemp(filepath.Dir(destination), ".wavearchive-restore-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if _, err := io.Copy(temp, source); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}

	oldPath := destination + ".restore-old"
	_ = os.Remove(oldPath)
	if err := os.Rename(destination, oldPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(tempName, destination); err != nil {
		_ = os.Rename(oldPath, destination)
		return err
	}
	_ = os.Remove(oldPath)
	_ = os.Remove(destination + "-wal")
	_ = os.Remove(destination + "-shm")
	return nil
}

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}
	if err := reconcileLegacySchema(db); err != nil {
		return fmt.Errorf("reconcile legacy schema: %w", err)
	}

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		var applied int
		if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", entry.Name()).Scan(&applied); err != nil {
			return err
		}
		if applied > 0 {
			continue
		}
		sqlBytes, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err = tx.Exec(string(sqlBytes)); err == nil {
			_, err = tx.Exec("INSERT INTO schema_migrations(version) VALUES (?)", entry.Name())
		}
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func reconcileLegacySchema(db *sql.DB) error {
	hasCharacters, err := tableExists(db, "characters")
	if err != nil || !hasCharacters {
		return err
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS game_versions (
			version TEXT PRIMARY KEY,
			synced_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS owned_characters (
			profile_id INTEGER NOT NULL DEFAULT 1,
			character_id INTEGER NOT NULL,
			level INTEGER NOT NULL DEFAULT 1,
			sequence INTEGER NOT NULL DEFAULT 0 CHECK (sequence BETWEEN 0 AND 6),
			favorite INTEGER NOT NULL DEFAULT 0 CHECK (favorite IN (0, 1)),
			PRIMARY KEY (profile_id, character_id),
			FOREIGN KEY (character_id) REFERENCES characters(id)
		);
		CREATE VIRTUAL TABLE IF NOT EXISTS characters_fts USING fts5(
			name, nickname, content='characters', content_rowid='id',
			tokenize='unicode61 remove_diacritics 2'
		);
		CREATE TRIGGER IF NOT EXISTS characters_ai AFTER INSERT ON characters BEGIN
			INSERT INTO characters_fts(rowid, name, nickname) VALUES (new.id, new.name, new.nickname);
		END;
		CREATE TRIGGER IF NOT EXISTS characters_ad AFTER DELETE ON characters BEGIN
			INSERT INTO characters_fts(characters_fts, rowid, name, nickname)
			VALUES ('delete', old.id, old.name, old.nickname);
		END;
		CREATE TRIGGER IF NOT EXISTS characters_au AFTER UPDATE ON characters BEGIN
			INSERT INTO characters_fts(characters_fts, rowid, name, nickname)
			VALUES ('delete', old.id, old.name, old.nickname);
			INSERT INTO characters_fts(rowid, name, nickname) VALUES (new.id, new.name, new.nickname);
		END;
		CREATE INDEX IF NOT EXISTS characters_element_idx ON characters(element);
		CREATE INDEX IF NOT EXISTS characters_rarity_idx ON characters(rarity);
		INSERT INTO characters_fts(characters_fts) VALUES ('rebuild');
	`); err != nil {
		return err
	}
	if err := markMigration(db, "0001_catalog.sql"); err != nil {
		return err
	}

	hasDetails, err := columnExists(db, "characters", "detail_loaded")
	if err != nil {
		return err
	}
	if hasDetails {
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS weapons (
				id INTEGER PRIMARY KEY, name TEXT NOT NULL, rarity INTEGER NOT NULL,
				weapon_type INTEGER NOT NULL, description TEXT NOT NULL DEFAULT '',
				effect_name TEXT NOT NULL DEFAULT '', effect TEXT NOT NULL DEFAULT '',
				icon_path TEXT NOT NULL DEFAULT '', params_json TEXT NOT NULL DEFAULT '[]',
				game_version TEXT NOT NULL, updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			CREATE TABLE IF NOT EXISTS skills (
				character_id INTEGER NOT NULL, node_id TEXT NOT NULL,
				skill_type TEXT NOT NULL DEFAULT '', name TEXT NOT NULL,
				description TEXT NOT NULL DEFAULT '', icon_path TEXT NOT NULL DEFAULT '',
				levels_json TEXT NOT NULL DEFAULT '{}', sort_order INTEGER NOT NULL DEFAULT 0,
				PRIMARY KEY (character_id, node_id),
				FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
			);
			CREATE TABLE IF NOT EXISTS resonance_chains (
				character_id INTEGER NOT NULL, sequence INTEGER NOT NULL,
				name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
				icon_path TEXT NOT NULL DEFAULT '', PRIMARY KEY (character_id, sequence),
				FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS skills_character_order_idx ON skills(character_id, sort_order);
			CREATE INDEX IF NOT EXISTS chains_character_sequence_idx ON resonance_chains(character_id, sequence);
			CREATE INDEX IF NOT EXISTS characters_signature_weapon_idx ON characters(signature_weapon_id);
		`); err != nil {
			return err
		}
		if err := markMigration(db, "0002_character_details.sql"); err != nil {
			return err
		}
	}

	hasOwnedFlag, err := columnExists(db, "owned_characters", "owned")
	if err != nil {
		return err
	}
	if hasOwnedFlag {
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS accounts (
				id INTEGER PRIMARY KEY, name TEXT NOT NULL,
				created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			INSERT OR IGNORE INTO accounts(id, name) VALUES (1, 'Conta principal');
			CREATE INDEX IF NOT EXISTS owned_characters_owned_idx ON owned_characters(profile_id, owned);
			CREATE INDEX IF NOT EXISTS owned_characters_favorite_idx ON owned_characters(profile_id, favorite);
		`); err != nil {
			return err
		}
		if err := markMigration(db, "0003_local_accounts.sql"); err != nil {
			return err
		}
	}

	hasWeaponCatalog, err := columnExists(db, "weapons", "base_atk")
	if err != nil {
		return err
	}
	if hasWeaponCatalog {
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS owned_weapons (
				profile_id INTEGER NOT NULL DEFAULT 1, weapon_id INTEGER NOT NULL,
				level INTEGER NOT NULL DEFAULT 1, weapon_rank INTEGER NOT NULL DEFAULT 1,
				favorite INTEGER NOT NULL DEFAULT 0, owned INTEGER NOT NULL DEFAULT 1,
				PRIMARY KEY (profile_id, weapon_id),
				FOREIGN KEY (weapon_id) REFERENCES weapons(id)
			);
			CREATE VIRTUAL TABLE IF NOT EXISTS weapons_fts USING fts5(
				name, description, effect_name, content='weapons', content_rowid='id',
				tokenize='unicode61 remove_diacritics 2'
			);
			CREATE TRIGGER IF NOT EXISTS weapons_ai AFTER INSERT ON weapons BEGIN
				INSERT INTO weapons_fts(rowid, name, description, effect_name)
				VALUES (new.id, new.name, new.description, new.effect_name);
			END;
			CREATE TRIGGER IF NOT EXISTS weapons_ad AFTER DELETE ON weapons BEGIN
				INSERT INTO weapons_fts(weapons_fts, rowid, name, description, effect_name)
				VALUES ('delete', old.id, old.name, old.description, old.effect_name);
			END;
			CREATE TRIGGER IF NOT EXISTS weapons_au AFTER UPDATE ON weapons BEGIN
				INSERT INTO weapons_fts(weapons_fts, rowid, name, description, effect_name)
				VALUES ('delete', old.id, old.name, old.description, old.effect_name);
				INSERT INTO weapons_fts(rowid, name, description, effect_name)
				VALUES (new.id, new.name, new.description, new.effect_name);
			END;
			INSERT INTO weapons_fts(weapons_fts) VALUES ('rebuild');
			CREATE INDEX IF NOT EXISTS weapons_type_idx ON weapons(weapon_type);
			CREATE INDEX IF NOT EXISTS weapons_rarity_idx ON weapons(rarity);
			CREATE INDEX IF NOT EXISTS owned_weapons_owned_idx ON owned_weapons(profile_id, owned);
		`); err != nil {
			return err
		}
		if err := markMigration(db, "0004_weapon_catalog.sql"); err != nil {
			return err
		}
	}
	hasBuilds, err := tableExists(db, "builds")
	if err != nil {
		return err
	}
	if hasBuilds {
		if _, err := db.Exec(`
			CREATE INDEX IF NOT EXISTS builds_character_idx ON builds(character_id);
			CREATE INDEX IF NOT EXISTS builds_updated_idx ON builds(updated_at DESC);
			CREATE INDEX IF NOT EXISTS builds_deleted_idx ON builds(deleted_at);
		`); err != nil {
			return err
		}
		if err := markMigration(db, "0005_builds.sql"); err != nil {
			return err
		}
		hasSkillLevels, err := columnExists(db, "builds", "normal_attack_level")
		if err != nil {
			return err
		}
		if hasSkillLevels {
			if err := markMigration(db, "0017_build_skill_levels.sql"); err != nil {
				return err
			}
		}
	}
	hasEchoes, err := tableExists(db, "echoes")
	if err != nil {
		return err
	}
	if hasEchoes {
		if _, err := db.Exec(`
			CREATE INDEX IF NOT EXISTS echoes_cost_idx ON echoes(cost);
			CREATE INDEX IF NOT EXISTS owned_echoes_echo_idx ON owned_echoes(echo_id);
		`); err != nil {
			return err
		}
		if err := markMigration(db, "0006_echoes.sql"); err != nil {
			return err
		}
	}
	hasTeams, err := tableExists(db, "teams")
	if err != nil {
		return err
	}
	if hasTeams {
		if _, err := db.Exec(`
			CREATE INDEX IF NOT EXISTS teams_updated_idx ON teams(updated_at DESC);
			CREATE INDEX IF NOT EXISTS teams_deleted_idx ON teams(deleted_at);
			CREATE INDEX IF NOT EXISTS team_members_character_idx ON team_members(character_id);
		`); err != nil {
			return err
		}
		if err := markMigration(db, "0007_teams.sql"); err != nil {
			return err
		}
	}
	hasTheorycraft, err := tableExists(db, "build_configs")
	if err != nil {
		return err
	}
	if hasTheorycraft {
		if _, err := db.Exec(`
			CREATE INDEX IF NOT EXISTS buffs_team_idx ON buffs(team_id);
			CREATE INDEX IF NOT EXISTS rotations_team_idx ON rotations(team_id);
			CREATE INDEX IF NOT EXISTS rotation_actions_order_idx ON rotation_actions(rotation_id,sort_order);
			CREATE INDEX IF NOT EXISTS ai_messages_conversation_idx ON ai_messages(conversation_id,id);
		`); err != nil {
			return err
		}
		if err := markMigration(db, "0008_theorycraft_ai.sql"); err != nil {
			return err
		}
	}
	hasWorkspace, err := tableExists(db, "app_settings")
	if err != nil {
		return err
	}
	if hasWorkspace {
		if _, err := db.Exec(`
			CREATE INDEX IF NOT EXISTS build_versions_build_idx ON build_versions(build_id,id DESC);
			CREATE INDEX IF NOT EXISTS planner_goals_priority_idx ON planner_goals(completed,priority,id);
			CREATE INDEX IF NOT EXISTS convene_records_date_idx ON convene_records(obtained_at DESC,id DESC);
		`); err != nil {
			return err
		}
		if err := markMigration(db, "0009_workspace.sql"); err != nil {
			return err
		}
	}
	hasGuides, err := tableExists(db, "character_guides")
	if err != nil {
		return err
	}
	if hasGuides {
		if _, err := db.Exec("CREATE INDEX IF NOT EXISTS character_guides_character_idx ON character_guides(character_id,like_count DESC)"); err != nil {
			return err
		}
		if err := markMigration(db, "0010_guides_ai.sql"); err != nil {
			return err
		}
	}
	hasTimeline, err := columnExists(db, "rotation_actions", "cooldown")
	if err != nil {
		return err
	}
	if hasTimeline {
		if err := markMigration(db, "0011_rotation_timeline.sql"); err != nil {
			return err
		}
	}
	hasProgression, err := tableExists(db, "materials")
	if err != nil {
		return err
	}
	if hasProgression {
		if _, err := db.Exec(`
			CREATE INDEX IF NOT EXISTS character_ascension_stage_idx ON character_ascension_costs(character_id, stage);
			CREATE INDEX IF NOT EXISTS skill_level_cost_idx ON skill_level_costs(character_id, node_id, level);
		`); err != nil {
			return err
		}
		if err := markMigration(db, "0012_character_progression.sql"); err != nil {
			return err
		}
	}
	hasAPIOrder, err := columnExists(db, "characters", "api_order")
	if err != nil {
		return err
	}
	if hasAPIOrder {
		if _, err := db.Exec("CREATE INDEX IF NOT EXISTS characters_api_order_idx ON characters(api_order)"); err != nil {
			return err
		}
		if err := markMigration(db, "0013_character_api_order.sql"); err != nil {
			return err
		}
	}
	hasExtras, err := columnExists(db, "characters", "extras_json")
	if err != nil {
		return err
	}
	if hasExtras {
		if err := markMigration(db, "0014_character_extras.sql"); err != nil {
			return err
		}
	}
	hasConveneHistory, err := tableExists(db, "convene_profiles")
	if err != nil {
		return err
	}
	if hasConveneHistory {
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS convene_pool_catalog (
				profile_id INTEGER NOT NULL,
				pool_type INTEGER NOT NULL,
				locale_key TEXT NOT NULL,
				name TEXT NOT NULL,
				short_name TEXT NOT NULL,
				kind TEXT NOT NULL,
				hard_pity INTEGER NOT NULL DEFAULT 80,
				sort_order INTEGER NOT NULL DEFAULT 0,
				updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY(profile_id, pool_type),
				FOREIGN KEY(profile_id) REFERENCES convene_profiles(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS convene_pulls_profile_time_idx
				ON convene_pulls(profile_id, obtained_at DESC, source_index ASC, id DESC);
			CREATE INDEX IF NOT EXISTS convene_pulls_pool_time_idx
				ON convene_pulls(profile_id, pool_type, obtained_at DESC, source_index ASC);
			CREATE INDEX IF NOT EXISTS convene_pulls_rarity_idx
				ON convene_pulls(profile_id, rarity, obtained_at DESC);
			CREATE INDEX IF NOT EXISTS convene_pool_catalog_order_idx
				ON convene_pool_catalog(profile_id, sort_order, pool_type);
		`); err != nil {
			return err
		}
		if err := markMigration(db, "0016_convene_history.sql"); err != nil {
			return err
		}
	}
	return nil
}

func tableExists(db *sql.DB, name string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type IN ('table','view') AND name=?", name).Scan(&count)
	return count > 0, err
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

func markMigration(db *sql.DB, version string) error {
	_, err := db.Exec("INSERT OR IGNORE INTO schema_migrations(version) VALUES (?)", version)
	return err
}
