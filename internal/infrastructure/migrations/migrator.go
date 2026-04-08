package migrations

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func New(dbURL, schema, migrationsDir string) (*migrate.Migrate, error) {
	if err := ensureSchemaExists(dbURL, schema); err != nil {
		return nil, err
	}

	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("resolve migrations dir: %w", err)
	}

	sourceURL := "file://" + filepath.ToSlash(absDir)
	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return nil, fmt.Errorf("create migrator: %w", err)
	}

	return m, nil
}

func Up(dbURL, schema, migrationsDir string) error {
	m, err := New(dbURL, schema, migrationsDir)
	if err != nil {
		return err
	}
	defer closeMigrator(m)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func Down(dbURL, schema, migrationsDir string, steps int) error {
	m, err := New(dbURL, schema, migrationsDir)
	if err != nil {
		return err
	}
	defer closeMigrator(m)

	if steps <= 0 {
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("rollback migrations: %w", err)
		}
		return nil
	}

	if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rollback %d migration(s): %w", steps, err)
	}

	return nil
}

func Version(dbURL, schema, migrationsDir string) (uint, bool, error) {
	m, err := New(dbURL, schema, migrationsDir)
	if err != nil {
		return 0, false, err
	}
	defer closeMigrator(m)

	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("read migration version: %w", err)
	}

	return version, dirty, nil
}

func closeMigrator(m *migrate.Migrate) {
	sourceErr, databaseErr := m.Close()
	if sourceErr != nil || databaseErr != nil {
		return
	}
}

func ensureSchemaExists(dbURL, schema string) error {
	if schema == "" {
		return nil
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("open database for schema setup: %w", err)
	}
	defer db.Close()

	if _, err := db.Exec(schemaCreateStatement(schema)); err != nil {
		return fmt.Errorf("ensure schema %q exists: %w", schema, err)
	}

	return nil
}

func schemaCreateStatement(schema string) string {
	return fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", quoteIdentifier(schema))
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
