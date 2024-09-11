package database

import (
	"sparrow/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func GetQueueItem(imdbID string) (item models.QueueItem, err error) {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		return
	}
	db.AutoMigrate(&models.QueueItem{})
	var queueItem models.QueueItem

	if err := db.Where("imdb_id = ?", imdbID).First(&queueItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return queueItem, nil
		}
		return queueItem, err
	}
	item = queueItem
	return
}

func PullQueueItem() (item models.QueueItem, err error) {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		return
	}
	db.AutoMigrate(&models.QueueItem{})
	var queueItem models.QueueItem

	if err := db.Where("status IN ('waiting', 'downloading')").First(&queueItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return queueItem, nil
		}
		return queueItem, err
	}

	item = queueItem

	return
}

func UpdateQueueItemStatus(imdbID string, status string) (err error) {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		return
	}
	db.AutoMigrate(&models.QueueItem{})
	var queueItem models.QueueItem
	if err := db.Where("imdb_id = ?", imdbID).First(&queueItem).Error; err != nil {
		return err
	}
	queueItem.Status = status
	if err := db.Model(&queueItem).Update("status", status).Error; err != nil {
		return err
	}
	return nil
}

func PushItemToQueue(item models.QueueItem) {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		return
	}
	db.AutoMigrate(&models.QueueItem{})
	db.Create(&item)
}

func DeleteQueueItem(imdbID string) (err error) {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		return
	}
	db.AutoMigrate(&models.QueueItem{})
	if err := db.Where("imdb_id = ?", imdbID).Delete(&models.QueueItem{}).Error; err != nil {
		return err
	}
	return nil
}
