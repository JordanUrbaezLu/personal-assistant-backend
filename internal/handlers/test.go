package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Test API handles GET /test?first=XYZ&last=ABC
func TestHandler(c *gin.Context) {
	first := c.Query("first")
	last := c.Query("last")

	if first == "" || last == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing first or last name",
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, " + first + " " + last,
	})
}
