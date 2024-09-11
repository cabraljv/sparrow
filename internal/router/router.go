package router

import (
	"sparrow/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	router.GET("/newendpoint", handler.HomeHandler)
	router.GET("/movies", handler.SearchMoviesHandler)
	router.GET("/movies/:imdbID", handler.GetMovieDataHandler)
	router.GET("/movies/:imdbID/start-watch", handler.StartMediaWatcher)
	router.Static("/static", "./tmp")
}
