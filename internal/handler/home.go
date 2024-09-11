package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewEndpointHandler your new endpoint's handler function
func HomeHandler(c *gin.Context) {
	// Your handler logic here
	c.String(http.StatusOK, "Welcome to our Golang API!")
}
