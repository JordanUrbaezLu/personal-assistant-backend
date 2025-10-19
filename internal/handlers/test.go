package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TestHandler godoc
// @Summary Test API to greet user
// @Description Returns a greeting using query parameters `first` and `last`.
// @Tags Test
// @Accept  json
// @Produce  json
// @Param first query string true "First name"
// @Param last query string true "Last name"
// @Success 200 {object} map[string]string "Successful greeting"
// @Failure 400 {object} map[string]string "Missing first or last name"
// @Router /test [get]
func TestHandler(c *gin.Context) {
	first := c.Query("first")
	last := c.Query("last")

	if first == "" || last == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing first or last name",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, " + first + " " + last,
	})
}
