package handler

import (
	"fmt"
	"sparrow/internal/app"
	"sparrow/internal/app/database"
	"sparrow/internal/app/mediaseach"
	"sparrow/models"
	"time"

	"github.com/gin-gonic/gin"
)

func SearchMoviesHandler(c *gin.Context) {

	query := c.Query("s")

	if query == "" {
		c.JSON(400, gin.H{
			"error": "missing query parameter",
		})
		return
	}

	// Call the SearchMedia function from the moviesearch package
	// and pass the query parameter
	media, err := mediaseach.SearchMedia(query)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, media)
}

func GetMovieDataHandler(c *gin.Context) {
	//get the imdbID from the path parameter
	imdbID := c.Param("imdbID")

	if imdbID == "" {
		c.JSON(400, gin.H{
			"error": "missing imdbID parameter",
		})
		return
	}

	existsOnDB, err := database.GetMedia(imdbID)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	if existsOnDB.ImdbID != "" {

		if len(existsOnDB.AvailableTorrents) > 0 {
			c.JSON(200, existsOnDB)
			return
		}

		availableTorrents, err := mediaseach.GetTorrents(existsOnDB)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		if len(availableTorrents) == 0 {
			c.JSON(200, existsOnDB)
			return
		}

		database.UpdateMediaAvailableTorrents(existsOnDB.ImdbID, availableTorrents)

		existsOnDB.AvailableTorrents = availableTorrents

		c.JSON(200, existsOnDB)
		return
	}

	media, err := mediaseach.GetMediaData(imdbID)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	mediaToSave := models.Media{
		ImdbID:            media.ImdbID,
		Type:              media.Type,
		Title:             media.Title,
		Year:              media.Year,
		AvailableTorrents: []models.AvailableTorrent{},
		Poster:            media.Poster,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	availableTorrents, err := mediaseach.GetTorrents(mediaToSave)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	mediaToSave.AvailableTorrents = availableTorrents

	database.SaveMedia(mediaToSave)

	c.JSON(200, mediaToSave)

}

func StartMediaWatcher(c *gin.Context) {
	imdbId := c.Param("imdbID")

	if imdbId == "" {
		c.JSON(400, gin.H{
			"error": "missing imdbID parameter",
		})
		return
	}

	media, err := database.GetMedia(imdbId)

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	if media.ImdbID == "" {
		c.JSON(404, gin.H{
			"error": "media not found",
		})
		return
	}

	// verify if media is already in the queue
	existentQueueItem, err := database.GetQueueItem(media.ImdbID)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	if existentQueueItem.ImdbID != "" {
		if existentQueueItem.Status == "downloading" {
			c.JSON(200, gin.H{
				"message": "media already downloading",
				"hash":    existentQueueItem.Hash,
			})
			return
		}
		database.UpdateQueueItemStatus(existentQueueItem.ImdbID, "waiting")
		c.JSON(200, gin.H{
			"message": "media added to queue",
			"hash":    existentQueueItem.Hash,
		})
		return
	}

	bestTorrent := app.VerifyBestTorrent(media.AvailableTorrents)

	fmt.Println("Best torrent found: ", media)

	if bestTorrent.ImdbID == "" {
		c.JSON(404, gin.H{
			"error": "no torrents found",
		})
		return
	}

	existentQueueItem, err = database.GetQueueItem(bestTorrent.ImdbID)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	queueItem := models.QueueItem{
		ImdbID:    bestTorrent.ImdbID,
		Hash:      bestTorrent.InfoHash,
		Status:    "waiting",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	database.PushItemToQueue(queueItem)

	c.JSON(200, gin.H{
		"message": "media added to queue",
		"hash":    bestTorrent.InfoHash,
	})

}
