package config

import (
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

var loadEnvOnce sync.Once

func LoadEnv() {
	loadEnvOnce.Do(func() {
		_ = godotenv.Load()
	})
}

func GetDatabaseDSN() string {
	LoadEnv()

	if value := os.Getenv("DATABASE_URL"); value != "" {
		return applySchemaToDSN(value, GetDatabaseSchema())
	}

	host := getenvDefault("DB_HOST", "localhost")
	port := getenvDefault("DB_PORT", "5432")
	user := getenvDefault("DB_USER", "postgres")
	password := getenvDefault("DB_PASSWORD", "password")
	name := getenvDefault("DB_NAME", "crud_db")
	sslMode := getenvDefault("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, name, sslMode)
	return applySchemaToDSN(dsn, GetDatabaseSchema())
}

func GetDatabaseSchema() string {
	LoadEnv()
	return getenvDefault("DB_SCHEMA", "public")
}

func GetJWTSecret() string {
	LoadEnv()
	return getenvDefault("JWT_SECRET", "dev-secret-change-me")
}

func GetJWTIssuer() string {
	LoadEnv()
	return getenvDefault("JWT_ISSUER", "crud-api")
}

func GetJWTDuration() time.Duration {
	LoadEnv()

	duration, err := time.ParseDuration(getenvDefault("JWT_TTL", "24h"))
	if err != nil {
		return 24 * time.Hour
	}

	return duration
}

func GetEmbeddingProvider() string {
	LoadEnv()
	return getenvDefault("EMBEDDING_PROVIDER", "local-hash")
}

func GetEmbeddingGRPCTarget() string {
	LoadEnv()
	return getenvDefault("EMBEDDING_GRPC_TARGET", "localhost:50051")
}

func GetEmbeddingRequestTimeout() time.Duration {
	LoadEnv()

	duration, err := time.ParseDuration(getenvDefault("EMBEDDING_REQUEST_TIMEOUT", "15s"))
	if err != nil {
		return 15 * time.Second
	}

	return duration
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func applySchemaToDSN(dsn, schema string) string {
	if schema == "" {
		return dsn
	}

	parsed, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}

	query := parsed.Query()
	if query.Get("search_path") == "" {
		query.Set("search_path", schema)
	}
	parsed.RawQuery = query.Encode()

	return parsed.String()
}
