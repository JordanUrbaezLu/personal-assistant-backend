package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HelloHandler godoc
// @Summary Simple hello endpoint
// @Description Returns a greeting message. If no name is provided, defaults to "World".
// @Tags General
// @Produce  json
// @Param name query string false "Name to greet"
// @Success 200 {object} map[string]string "Greeting message"
// @Router /hello [get]
func HelloHandler(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		name = "World"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, " + name + "!!!",
	})
}
