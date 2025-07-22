package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/models"
	"forum/routes"
	"forum/utils"
)

func main() {
	err := utils.LoadEnv(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	
	// Initialize database
	db, err := models.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Setup routes
	handler := routes.SetupRoutes(db)

	// Start server
	port := 8080
	fmt.Printf("Server is running on http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
}
