package models

import "time"

type QueueItem struct {
	ImdbID    string `gorm:"primaryKey"`
	Hash      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
