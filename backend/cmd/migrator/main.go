package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/joho/godotenv"
)

// * Actions
// task migrate -- up
// task migrate -- down
// task migrate -- reset
func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	storageURL := os.Getenv("DATABASE_URL")

	var migrationsPath, migrationsTable, migrationAction string

	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migration")
	flag.StringVar(&migrationAction, "migrations-action", "", "action of migration")
	flag.Parse()

	if storageURL == "" {
		panic("storageURL is required")
	}

	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("%s&x-migrations-table=%s", storageURL, migrationsTable),
	)
	if err != nil {
		panic(err)
	}

	switch migrationAction {
	case "reset":
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")
				return
			}
		}
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")
				return
			}
		}
	case "down":
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")
				return
			}

			panic(err)
		}
	default:
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")
				return
			}

			panic(err)
		}
	}

	fmt.Println("migrations applied successfully")
}
