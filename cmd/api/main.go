package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/v1-nce/threadtalk-backend/internal/router"
	"github.com/v1-nce/threadtalk-backend/internal/db"
)

func init() {
	// Load Dotenv
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	// Run Database Migrations
	db.RunDBMigrations(os.Getenv("DATABASE_URL"))

	// Connect Database
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Unable to Connect to Database due to: %v", err)
	}
	defer database.Close()
	log.Println("Connected to PostgreSQL Database")

	// Setup Router
	r := router.SetUpRouter(database)

	// Start Server
	r.Run(":" + os.Getenv("PORT"))
}