package main

import (
	"log"

	"gameCore/config"
	"gameCore/internal/storage"
)

// main.go
func main() {
	cfg, err := config.LoadConfig("/home/oksuide/GoProjects/gameCore/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Init DB
	if err := storage.Connect(cfg.Database); err != nil {
		log.Fatal("DB error:", err)
	}
	defer storage.Close()
	// Migration
	if err := storage.InitTables(); err != nil {
		log.Fatal("Migration error:", err)
	}
	// Init Redis
	if err := storage.InitRedis(cfg.Redis); err != nil {
		log.Fatal(err)
	}
}
