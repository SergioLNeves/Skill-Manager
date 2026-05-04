package persistence

import (
	"database/sql"
	"embed"
	"fmt"
	"sync/atomic"

	_ "modernc.org/sqlite" // register sqlite driver
)

//go:embed migrations/*.sql
var migrations embed.FS

// Open opens (or creates) a SQLite database at dsn, enables WAL mode, and
// runs any pending migrations.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open %s: %w", dsn, err)
	}

	if _, err = db.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite: pragma: %w", err)
	}

	if err = migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

var memCounter atomic.Int64

// OpenMemory returns a named in-memory database for tests.
// Each call returns an isolated instance via a unique name so parallel tests
// do not share connections or schema state.
func OpenMemory() (*sql.DB, error) {
	n := memCounter.Add(1)
	dsn := fmt.Sprintf("file:testdb%d?mode=memory&cache=shared&_fk=1", n)
	db, err := Open(dsn)
	if err != nil {
		return nil, err
	}
	// Pin to one connection: the shared-cache in-memory DB exists only while at
	// least one connection is open, and all readers must use the same process cache.
	db.SetMaxOpenConns(1)
	return db, nil
}

func migrate(db *sql.DB) error {
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("sqlite: read user_version: %w", err)
	}

	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("sqlite: read migrations dir: %w", err)
	}

	for i, entry := range entries {
		if i < version {
			continue
		}
		content, readErr := migrations.ReadFile("migrations/" + entry.Name())
		if readErr != nil {
			return fmt.Errorf("sqlite: read migration %s: %w", entry.Name(), readErr)
		}
		if _, execErr := db.Exec(string(content)); execErr != nil {
			return fmt.Errorf("sqlite: apply migration %s: %w", entry.Name(), execErr)
		}
		if _, execErr := db.Exec(fmt.Sprintf("PRAGMA user_version=%d", i+1)); execErr != nil {
			return fmt.Errorf("sqlite: set user_version after %s: %w", entry.Name(), execErr)
		}
	}

	return nil
}
