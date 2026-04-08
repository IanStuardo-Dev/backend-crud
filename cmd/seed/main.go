package main

import (
	"context"
	"log"

	"github.com/example/crud/internal/infrastructure/config"
	postgresdb "github.com/example/crud/internal/infrastructure/persistence/postgres"
	"github.com/example/crud/internal/infrastructure/seeds"
)

func main() {
	dsn := config.GetDatabaseDSN()
	sqlDB, err := postgresdb.New(dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer sqlDB.Close()

	if err := seeds.Run(context.Background(), sqlDB); err != nil {
		log.Fatalf("seed run: %v", err)
	}

	log.Println("seed data applied")
}
