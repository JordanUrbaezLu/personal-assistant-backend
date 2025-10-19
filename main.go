package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"personal-assistant-backend/internal/config"
	"personal-assistant-backend/internal/handlers"
	"personal-assistant-backend/internal/middleware"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Only load .env in local dev (Fly sets secrets via env)
	if os.Getenv("FLY_APP_NAME") == "" {
		config.Load(".env")
		log.Println("✅ .env file loaded (local dev)")
	} else {
		log.Println("ℹ️ Running in Fly — using flyctl secrets")
	}

	// Database connection
	dsn := os.Getenv("USERS_DATABASE_URL")
	if dsn == "" {
		log.Fatal("❌ USERS_DATABASE_URL not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("❌ Failed to open DB:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("❌ Failed to ping DB:", err)
	}
	log.Println("✅ Connected to database")

	// API Key
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("❌ API_KEY not set")
	}

	// Setup Gin
	r := gin.Default()

	// Apply API key middleware globally (all routes require API key)
	r.Use(middleware.APIKeyAuthMiddleware(apiKey))

	// Handlers needing DB
	auth := handlers.NewAuthHandler(db)

	// Public routes
	r.POST("/signup", auth.Signup)
	r.POST("/login", auth.Login)
	r.POST("/token/refresh", auth.Refresh)

	// Protected routes (JWT required on top of API key)
	authGroup := r.Group("/")
	authGroup.Use(middleware.JWTAuthMiddleware())
	authGroup.GET("/me", auth.Me)

	// Misc routes
	r.GET("/hello", handlers.HelloHandler)
	r.GET("/test", handlers.TestHandler)

	// Run server
	log.Println("🚀 Server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
