package main

import (
	"log"

	"distribution-commission/internal/config"
	"distribution-commission/internal/database"
	"distribution-commission/internal/router"
)

func main() {
	config.LoadConfig()

	database.InitPostgreSQL()
	database.InitRedis()

	database.RunMigrations()

	database.SeedData()

	r := router.SetupRouter()

	log.Println("Server starting on port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
