package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"personal-assistant-backend/internal/config"
	"personal-assistant-backend/internal/handlers"
	"personal-assistant-backend/internal/middleware"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "personal-assistant-backend/docs" // 👈 Import generated Swagger docs
)

// @title Personal Assistant Backend API
// @version 1.0
// @description REST API for authentication, user management, and assistant features.
// @termsOfService http://swagger.io/terms/

// @contact.name Jordan Urbaez
// @contact.email jordana.urbaez@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Only load .env in local dev (Fly sets secrets via env)
	isLocal := os.Getenv("FLY_APP_NAME") == ""
	if isLocal {
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

	// ✅ Conditionally apply middleware
	if isLocal {
		r.Use(func(c *gin.Context) {
			path := c.Request.URL.Path
	
			// ✅ Allow all Swagger UI assets and /hello without API key
			if path == "/hello" || 
			   path == "/swagger" || 
			   path == "/swagger/" || 
			   len(path) >= 9 && path[:9] == "/swagger/" {
				c.Next()
				return
			}
	
			middleware.APIKeyAuthMiddleware(apiKey)(c)
		})
		log.Println("🧩 Running in local mode: Swagger and /hello are open (no API key needed)")
	} else {
		r.Use(middleware.APIKeyAuthMiddleware(apiKey))
		log.Println("🔒 Running in production: All routes protected by API key")
	}
	

	// Swagger route — accessible at /swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("http://localhost:8080/swagger/doc.json")))

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
