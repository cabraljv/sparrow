package main

import (
	"sparrow/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Setup routes
	router.SetupRoutes(r)

	// Start server
	r.Run(":3333") // adjust port as needed
}
