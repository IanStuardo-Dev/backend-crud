package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/config"
	"github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/migrations"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	dbURL := config.GetDatabaseDSN()
	schema := config.GetDatabaseSchema()
	migrationsDir := filepath.Join(".", "migrations")

	switch os.Args[1] {
	case "up":
		if err := migrations.Up(dbURL, schema, migrationsDir); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		fmt.Println("migrations applied")
	case "down":
		steps := 0
		if len(os.Args) > 2 {
			parsed, err := strconv.Atoi(os.Args[2])
			if err != nil {
				log.Fatalf("invalid down steps %q: %v", os.Args[2], err)
			}
			steps = parsed
		}
		if err := migrations.Down(dbURL, schema, migrationsDir, steps); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
		if steps > 0 {
			fmt.Printf("rolled back %d migration(s)\n", steps)
			return
		}
		fmt.Println("all migrations rolled back")
	case "version":
		version, dirty, err := migrations.Version(dbURL, schema, migrationsDir)
		if err != nil {
			log.Fatalf("migrate version: %v", err)
		}
		if version == 0 && !dirty {
			fmt.Println("no migrations applied")
			return
		}
		fmt.Printf("version=%d dirty=%t\n", version, dirty)
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [up|down [steps]|version]\n", filepath.Base(os.Args[0]))
	os.Exit(1)
}
