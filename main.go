package main

import (
	"github.com/gin-gonic/gin"
	"personal-assistant-backend/internal/handlers"
)

func main() {
	r := gin.Default()

	// Register routes
	r.GET("/hello", handlers.HelloHandler)

	// Run server
	r.Run(":8080")
}
