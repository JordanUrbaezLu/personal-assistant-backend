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
	chatHandler "personal-assistant-backend/internal/handlers/chat"
	"personal-assistant-backend/internal/middleware"
	"personal-assistant-backend/docs"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// @title Personal Assistant Backend API
// @version 1.0
// @description REST API for authentication, user management, and assistant chat features.
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
	// =====================================================
	// ⚙️ Environment Configuration
	// =====================================================
	isLocal := os.Getenv("FLY_APP_NAME") == ""
	if isLocal {
		config.Load(".env")
		log.Println("✅ .env file loaded (local dev)")
	} else {
		log.Println("ℹ️ Running in Fly — using flyctl secrets")
	}

	// =====================================================
	// 🗄️ Database Connection
	// =====================================================
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

	// =====================================================
	// 🔑 API Key
	// =====================================================
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("❌ API_KEY not set")
	}

	// =====================================================
	// 🌐 Gin Setup + Swagger Config
	// =====================================================
	r := gin.Default()

	swaggerHost := "localhost:8080"
	swaggerSchemes := []string{"http"}
	swaggerURL := "http://localhost:8080/swagger/doc.json"

	if os.Getenv("FLY_APP_NAME") != "" {
		swaggerHost = "personal-assistant-backend-fly.fly.dev"
		swaggerSchemes = []string{"https"}
		swaggerURL = "https://personal-assistant-backend-fly.fly.dev/swagger/doc.json"
	}

	docs.SwaggerInfo.Host = swaggerHost
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = swaggerSchemes

	// =====================================================
	// 🧱 Middleware
	// =====================================================
	if isLocal {
		r.Use(func(c *gin.Context) {
			path := c.Request.URL.Path
			// Allow Swagger & /hello without API key
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

	// =====================================================
	// 📜 Swagger Docs Route
	// =====================================================
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL(swaggerURL)))

	// =====================================================
	// 🧩 Initialize Handlers
	// =====================================================
	auth := handlers.NewAuthHandler(db)
	chats := chatHandler.NewChatHandler(db)

	// =====================================================
	// 🚪 Public Auth Routes
	// =====================================================
	r.POST("/signup", auth.Signup)
	r.POST("/login", auth.Login)
	r.POST("/token/refresh", auth.Refresh)

	// =====================================================
	// 🔒 Protected Routes (JWT)
	// =====================================================
	authGroup := r.Group("/")
	authGroup.Use(middleware.JWTAuthMiddleware())

	// --- Auth check (on app startup)
	authGroup.GET("/auth", auth.AuthCheck)

	// --- Current user info
	authGroup.GET("/me", auth.Me)

	// --- Chat routes
	authGroup.POST("/chats", chats.CreateChat)
	authGroup.GET("/chats", chats.ListChats)
	authGroup.POST("/chats/:chat_id/messages", chats.SendMessage)
	authGroup.GET("/chats/:chat_id/messages", chats.ListMessages)

	// =====================================================
	// 🧩 Misc Routes
	// =====================================================
	r.GET("/hello", handlers.HelloHandler)
	r.GET("/greet", handlers.GreetHandler)

	// =====================================================
	// 🚀 Start Server
	// =====================================================
	log.Println("🚀 Server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
