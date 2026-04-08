package config

import (
	"strings"
	"testing"
	"time"
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

func TestGetEmbeddingProviderDefaultsToLocalHash(t *testing.T) {
	t.Setenv("EMBEDDING_PROVIDER", "")

	if provider := GetEmbeddingProvider(); provider != "local-hash" {
		t.Fatalf("expected local-hash provider, got %q", provider)
	}
}

func TestGetEmbeddingRequestTimeoutFallsBackToDefault(t *testing.T) {
	t.Setenv("EMBEDDING_REQUEST_TIMEOUT", "not-a-duration")

	if timeout := GetEmbeddingRequestTimeout(); timeout != 15*time.Second {
		t.Fatalf("expected 15s default timeout, got %v", timeout)
	}
}
