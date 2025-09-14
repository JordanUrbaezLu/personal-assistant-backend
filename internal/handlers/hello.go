package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// HelloHandler handles GET /hello?name=XYZ
func HelloHandler(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		name = "World"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, " + name + "!",
	})
}
