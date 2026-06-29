package main

import (
	"log"

	"{{MODULE_NAME}}/internal/config"
	"{{MODULE_NAME}}/internal/database"
	"{{MODULE_NAME}}/internal/routes"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	r := routes.Setup(cfg, db)

	log.Printf("🚀 Server running on %s", cfg.AppAddr)
	if err := r.Run(cfg.AppAddr); err != nil {
		log.Fatal(err)
	}
}
