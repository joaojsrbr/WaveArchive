package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// ResolveApplicationPath keeps databases created by the original prototype
// untouched. That schema used the same filename, but its tables and migration
// history are not compatible with the current local-first catalog.
func ResolveApplicationPath(preferred string) (string, error) {
	compatible, exists, err := compatibleDatabase(preferred)
	if err != nil {
		return "", fmt.Errorf("inspect existing database: %w", err)
	}
	if !exists || compatible {
		return preferred, nil
	}
	directory := filepath.Dir(preferred)
	for generation := 2; generation < 100; generation++ {
		candidate := filepath.Join(directory, fmt.Sprintf("wavearchive-v%d.db", generation))
		compatible, exists, err := compatibleDatabase(candidate)
		if err != nil {
			return "", fmt.Errorf("inspect database generation %d: %w", generation, err)
		}
		if !exists || compatible {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no compatible database path is available in %s", directory)
}

func compatibleDatabase(path string) (compatible bool, exists bool, err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true, false, nil
	} else if err != nil {
		return false, false, err
	}
	db, err := sql.Open("sqlite", path+"?mode=ro&_pragma=query_only(1)")
	if err != nil {
		return false, true, err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return false, true, err
	}

	hasCharacters, err := tableExists(db, "characters")
	if err != nil {
		return false, true, err
	}
	if hasCharacters {
		hasNickname, err := columnExists(db, "characters", "nickname")
		if err != nil {
			return false, true, err
		}
		hasGameVersion, err := columnExists(db, "characters", "game_version")
		if err != nil {
			return false, true, err
		}
		return hasNickname && hasGameVersion, true, nil
	}

	hasVersions, err := tableExists(db, "game_versions")
	if err != nil {
		return false, true, err
	}
	if hasVersions {
		hasLegacyID, err := columnExists(db, "game_versions", "id")
		if err != nil {
			return false, true, err
		}
		return !hasLegacyID, true, nil
	}

	var applicationTables int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table'
		  AND name NOT LIKE 'sqlite_%'
		  AND name <> 'schema_migrations'
	`).Scan(&applicationTables)
	if err != nil {
		return false, true, err
	}
	return applicationTables == 0, true, nil
}
