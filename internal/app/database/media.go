package database

import (
	"database/sql"
	"sparrow/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SaveMedia(media models.Media) (err error) {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		return
	}
	db.AutoMigrate(&models.Media{})

	if err := db.Create(media).Error; err != nil {
		return err
	}
	return nil
}

func GetMedia(imdbID string) (media models.Media, err error) {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		return
	}
	db.AutoMigrate(&models.Media{})
	if err := db.Where("imdb_id = ?", imdbID).First(&media).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return media, nil
		}
		return media, err
	}
	return
}

type Torrent struct {
	Name          string
	Title         string
	InfoHash      string
	FileIdx       int
	BehaviorHints struct {
		BingeGroup string
		Filename   string
	}
}

func UpdateMediaAvailableTorrents(imdbID string, torrents []models.AvailableTorrent) (err error) {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		return
	}
	db.AutoMigrate(&models.Media{})
	var media models.Media
	if err := db.Where("imdb_id = ?", imdbID).First(&media).Error; err != nil {
		return err
	}
	media.AvailableTorrents = torrents

	if err := db.Save(&media).Error; err != nil {
		return err
	}
	return nil
}

func UpdateMediaDownloadHash(imdbID string, infoHash string) (err error) {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		return
	}
	db.AutoMigrate(&models.Media{})
	var media models.Media
	if err := db.Where("imdb_id = ?", imdbID).First(&media).Error; err != nil {
		return err
	}
	media.CurrentDownloadHash = sql.NullString{String: infoHash, Valid: true}

	if err := db.Save(&media).Error; err != nil {
		return err
	}
	return nil
}
