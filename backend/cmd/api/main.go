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
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./database.db"
	}

	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatal("failed to initialize database:", err)
	}
	defer db.Close()

	userRepo := repository.NewSQLiteUserRepository(db)
	postRepo := repository.NewSQLitePostRepository(db)

	authService := services.NewAuthService(userRepo)
	postService := services.NewPostService(postRepo, userRepo)

	authHandler := handlers.NewAuthHandler(authService)
	postHandler := handlers.NewPostHandler(postService)

	r := router.Setup(authHandler, postHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal("failed to start server:", err)
	}
}
