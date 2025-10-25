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
	"personal-assistant-backend/docs" // ✅ Import generated Swagger docs

	_ "github.com/jackc/pgx/v5/stdlib"
)

// @title Personal Assistant Backend API
// @version 1.0
// @description REST API for authentication, user management, and assistant features.
// @termsOfService http://swagger.io/terms/

// @contact.name Jordan Urbaez
// @contact.email jordana.urbaez@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load environment configuration
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

	// API Key setup
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("❌ API_KEY not set")
	}

	// Setup Gin
	r := gin.Default()

	// Determine correct Swagger host dynamically
	swaggerHost := "localhost:8080"
	swaggerSchemes := []string{"http"}
	swaggerURL := "http://localhost:8080/swagger/doc.json"

	if os.Getenv("FLY_APP_NAME") != "" {
		swaggerHost = "personal-assistant-backend-fly.fly.dev"
		swaggerSchemes = []string{"https"}
		swaggerURL = "https://personal-assistant-backend-fly.fly.dev/swagger/doc.json"
	}

	// ✅ Set Swagger runtime info
	docs.SwaggerInfo.Host = swaggerHost
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = swaggerSchemes

	// ✅ Conditionally apply middleware
	if isLocal {
		r.Use(func(c *gin.Context) {
			path := c.Request.URL.Path
			// ✅ Allow Swagger & hello without API key
			if path == "/hello" ||
				path == "/swagger" ||
				path == "/swagger/" ||
				len(path) >= 9 && path[:9] == "/swagger/" {
				c.Next()
				return
			}
			middleware.APIKeyAuthMiddleware(apiKey)(c)
		})
		log.Println("🧩 Local mode: Swagger + /hello are open (no API key needed)")
	} else {
		r.Use(middleware.APIKeyAuthMiddleware(apiKey))
		log.Println("🔒 Production mode: All routes protected by API key")
	}

	// ✅ Swagger route — uses environment-appropriate URL
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL(swaggerURL)))

	// Initialize handlers
	auth := handlers.NewAuthHandler(db)

	// ========================================
	// 🚪 Public Auth Routes
	// ========================================
	r.POST("/signup", auth.Signup)
	r.POST("/login", auth.Login)
	r.POST("/token/refresh", auth.Refresh)

	// ========================================
	// 🔒 Protected Auth Routes
	// ========================================
	authGroup := r.Group("/")
	authGroup.Use(middleware.JWTAuthMiddleware())
	authGroup.GET("/auth", auth.AuthCheck)

	// ========================================
	// 🧩 Misc Routes
	// ========================================
	r.GET("/hello", handlers.HelloHandler)
	r.GET("/test", handlers.TestHandler)

	// ========================================
	// 🚀 Start Server
	// ========================================
	log.Println("🚀 Server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
