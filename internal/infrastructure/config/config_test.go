package config

import (
	"strings"
	"testing"
)

func TestGetDatabaseDSNIncludesSchemaSearchPath(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "postgres")
	t.Setenv("DB_NAME", "go-crud")
	t.Setenv("DB_SSLMODE", "disable")
	t.Setenv("DB_SCHEMA", "public")

	dsn := GetDatabaseDSN()

	if !strings.Contains(dsn, "search_path=public") {
		t.Fatalf("expected DSN to include search_path=public, got %q", dsn)
	}
}

func TestGetDatabaseDSNPreservesExistingSearchPath(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/go-crud?sslmode=disable&search_path=custom")
	t.Setenv("DB_SCHEMA", "public")

	dsn := GetDatabaseDSN()

	if !strings.Contains(dsn, "search_path=custom") {
		t.Fatalf("expected existing search_path to be preserved, got %q", dsn)
	}
	if strings.Contains(dsn, "search_path=public") {
		t.Fatalf("expected DB_SCHEMA not to overwrite existing search_path, got %q", dsn)
	}
}
