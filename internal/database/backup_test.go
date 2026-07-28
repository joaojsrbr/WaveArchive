package database

import (
	"path/filepath"
	"testing"
)

func TestBackupAndRestore(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "current.db")
	snapshot := filepath.Join(dir, "snapshots", "snapshot.db")

	db, err := Open(current)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL().Exec("CREATE TABLE restore_probe(value TEXT NOT NULL)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL().Exec("INSERT INTO restore_probe(value) VALUES ('before')"); err != nil {
		t.Fatal(err)
	}
	if err := db.Backup(snapshot); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL().Exec("UPDATE restore_probe SET value = 'after'"); err != nil {
		t.Fatal(err)
	}
	if err := db.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if err := RestoreFile(snapshot, current); err != nil {
		t.Fatal(err)
	}
	restored, err := Open(current)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	var value string
	if err := restored.SQL().QueryRow("SELECT value FROM restore_probe").Scan(&value); err != nil {
		t.Fatal(err)
	}
	if value != "before" {
		t.Fatalf("restored value = %q", value)
	}
}
