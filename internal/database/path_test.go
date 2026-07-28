package database

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestResolveApplicationPathKeepsCompatibleDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wavearchive.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	resolved, err := ResolveApplicationPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != path {
		t.Fatalf("resolved path = %q, want %q", resolved, path)
	}
}

func TestResolveApplicationPathSeparatesPrototypeDatabase(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "wavearchive.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		CREATE TABLE schema_migrations(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL);
		INSERT INTO schema_migrations(version,applied_at) VALUES(1,CURRENT_TIMESTAMP);
		CREATE TABLE game_versions(
			id TEXT PRIMARY KEY, version TEXT NOT NULL UNIQUE,
			release_date TEXT, is_current INTEGER DEFAULT 0, synced_at TEXT
		);
		CREATE TABLE characters(
			id TEXT PRIMARY KEY, name TEXT NOT NULL, slug TEXT NOT NULL UNIQUE,
			element TEXT NOT NULL, rarity INTEGER NOT NULL, weapon_type TEXT NOT NULL
		);
	`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	resolved, err := ResolveApplicationPath(path)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(directory, "wavearchive-v2.db")
	if resolved != want {
		t.Fatalf("resolved path = %q, want %q", resolved, want)
	}
	newDB, err := Open(resolved)
	if err != nil {
		t.Fatalf("new database should migrate cleanly: %v", err)
	}
	if err := newDB.Close(); err != nil {
		t.Fatal(err)
	}
}
