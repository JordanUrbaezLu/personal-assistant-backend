package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GreetHandler godoc
// @Summary Greet API to greet user
// @Description Returns a greeting using query parameters `first` and `last`.
// @Tags Misc
// @Accept  json
// @Produce  json
// @Param first query string true "First name"
// @Param last query string true "Last name"
// @Success 200 {object} map[string]string "Successful greeting"
// @Failure 400 {object} map[string]string "Missing first or last name"
// @Router /greet [get]
func GreetHandler(c *gin.Context) {
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
