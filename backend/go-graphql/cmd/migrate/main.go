package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/faizp/zenlist/backend/go-graphql/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	direction := flag.String("direction", "up", "one of: up, down, force")
	steps := flag.Int("steps", 0, "steps for down (0 means all)")
	version := flag.Int("version", 0, "version for force")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cwd error: %v\n", err)
		os.Exit(1)
	}

	migrationsPath := filepath.Join(cwd, "migrations")
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)

	m, err := migrate.New(sourceURL, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate init error: %v\n", err)
		os.Exit(1)
	}
	defer m.Close()

	switch *direction {
	case "up":
		err = m.Up()
	case "down":
		if *steps > 0 {
			err = m.Steps(-(*steps))
		} else {
			err = m.Down()
		}
	case "force":
		if *version <= 0 {
			fmt.Fprintln(os.Stderr, "-version must be set and > 0 for force")
			os.Exit(1)
		}
		err = m.Force(*version)
	default:
		fmt.Fprintf(os.Stderr, "unsupported direction: %s\n", *direction)
		os.Exit(1)
	}

	if err != nil && err != migrate.ErrNoChange {
		fmt.Fprintf(os.Stderr, "migration error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("migration complete")
}
