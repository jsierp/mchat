package main

import (
	"flag"
	"fmt"
	"log"
	"mchat/internal/config"
	"mchat/internal/preview"
	"mchat/internal/storage"
	"os"
)

func main() {
	forceFlag := flag.Bool("force", false, "Overwrite the existing app database and config")
	flag.Parse()

	dbPath, err := storage.GetPath()
	if err != nil {
		log.Fatalf("failed to resolve database path: %v", err)
	}
	configPath, err := config.GetPath()
	if err != nil {
		log.Fatalf("failed to resolve config path: %v", err)
	}

	if err := ensureWritableTargets(*forceFlag, dbPath, configPath); err != nil {
		log.Fatal(err)
	}

	db, err := storage.GetDB()
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	count, err := preview.SeedDB(db)
	if err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	configJSON, err := preview.ConfigJSON()
	if err != nil {
		log.Fatalf("failed to build config: %v", err)
	}
	if err := os.WriteFile(configPath, configJSON, 0600); err != nil {
		log.Fatalf("failed to write config: %v", err)
	}

	fmt.Printf("Seeded %d fake messages into %s\n", count, dbPath)
	fmt.Printf("Wrote preview config to %s\n", configPath)
	fmt.Printf("Start the app normally with: go run ./cmd/mchat\n")
}

func ensureWritableTargets(force bool, paths ...string) error {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil && !force {
			return fmt.Errorf("%s already exists; rerun with --force to overwrite the app data", path)
		}
	}
	return nil
}
