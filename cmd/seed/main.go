package main

import (
	"context"
	"log"

	"github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/config"
	embeddingprovider "github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/embedding/provider"
	postgresdb "github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/persistence/postgres"
	"github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/seeds"
)

func main() {
	dsn := config.GetDatabaseDSN()
	sqlDB, err := postgresdb.New(dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer sqlDB.Close()

	embedder, embeddingProviderName := embeddingprovider.NewProductEmbedder()
	if closer, ok := embedder.(interface{ Close() error }); ok {
		defer closer.Close()
	}
	log.Printf("seed embedding provider: %s", embeddingProviderName)

	if err := seeds.Run(context.Background(), sqlDB, embedder); err != nil {
		log.Fatalf("seed run: %v", err)
	}

	log.Println("seed data applied")
}
