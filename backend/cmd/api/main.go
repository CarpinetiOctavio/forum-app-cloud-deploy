package main

import (
	"log"
	"net/http"
	"os"

	"forum-app-cloud-deploy/internal/database"
	"forum-app-cloud-deploy/internal/handlers"
	"forum-app-cloud-deploy/internal/repository"
	"forum-app-cloud-deploy/internal/router"
	"forum-app-cloud-deploy/internal/services"
)

func main() {
	// Initialize database
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./database.db"
	}
	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Create repositories
	userRepo := repository.NewSQLiteUserRepository(db)
	postRepo := repository.NewSQLitePostRepository(db)

	// Create services
	authService := services.NewAuthService(userRepo)
	postService := services.NewPostService(postRepo, userRepo)

	// Create handlers
	authHandler := handlers.NewAuthHandler(authService)
	postHandler := handlers.NewPostHandler(postService)

	// Configure routes
	r := router.Setup(authHandler, postHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server running at http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
