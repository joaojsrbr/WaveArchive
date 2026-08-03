package database

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestOpenReconcilesLegacyDatabaseWithoutMigrationHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	migration, err := migrationFiles.ReadFile("migrations/0001_catalog.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(string(migration)); err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(`
		INSERT INTO characters(
			id, name, nickname, rarity, element, weapon_type,
			icon_path, background_path, game_version
		) VALUES (1102, 'Sanhua', 'Snow Waltz', 4, 1, 2, '', '', '3.6.1')
	`); err != nil {
		t.Fatal(err)
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() should reconcile legacy schema: %v", err)
	}
	defer db.Close()

	var name string
	if err := db.SQL().QueryRow("SELECT name FROM characters WHERE id=1102").Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Sanhua" {
		t.Fatalf("character was not preserved: %q", name)
	}
	var migrations int
	if err := db.SQL().QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrations); err != nil {
		t.Fatal(err)
	}
	if migrations != 17 {
		t.Fatalf("migration count = %d, want 17", migrations)
	}
	hasOwned, err := columnExists(db.SQL(), "owned_characters", "owned")
	if err != nil || !hasOwned {
		t.Fatalf("owned column missing after upgrade: %v", err)
	}
}

func TestOpenRebuildsDeletedMigrationHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL().Exec("DELETE FROM schema_migrations"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("Open() should rebuild migration history: %v", err)
	}
	defer reopened.Close()
	var migrations int
	if err := reopened.SQL().QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrations); err != nil {
		t.Fatal(err)
	}
	if migrations != 17 {
		t.Fatalf("migration count = %d, want 17", migrations)
	}
}

func TestOpenReconcilesPartialHistoryWithMissingFirstMigration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "partial-history.db")
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	migration, err := migrationFiles.ReadFile("migrations/0001_catalog.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(string(migration)); err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(`
		CREATE TABLE schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO schema_migrations(version) VALUES ('legacy-import');
	`); err != nil {
		t.Fatal(err)
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() should reconcile a partial non-empty history: %v", err)
	}
	defer db.Close()
	var firstMigration int
	if err := db.SQL().QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version='0001_catalog.sql'").Scan(&firstMigration); err != nil {
		t.Fatal(err)
	}
	if firstMigration != 1 {
		t.Fatalf("first migration was not reconciled")
	}
}

func TestOpenRepairsPartiallyCreatedFirstMigration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "partial-first.db")
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(`
		CREATE TABLE game_versions (
			version TEXT PRIMARY KEY,
			synced_at TEXT NOT NULL
		);
		CREATE TABLE schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO schema_migrations(version) VALUES ('legacy-import');
	`); err != nil {
		t.Fatal(err)
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() should finish a partially created first migration: %v", err)
	}
	defer db.Close()
	hasCharacters, err := tableExists(db.SQL(), "characters")
	if err != nil || !hasCharacters {
		t.Fatalf("characters table was not repaired: %v", err)
	}
}
