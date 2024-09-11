package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type AvailableTorrent struct {
	ImdbID        string
	Provider      string
	InfoHash      string
	BehaviorHints BehaviorHints
}

type BehaviorHints struct {
	BingeGroup string `json:"bingeGroup"`
	Filename   string `json:"filename"`
}

func (a *AvailableTorrent) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Failed to unmarshal JSON value for AvailableTorrent")
	}

	result := json.Unmarshal(bytes, &a)
	return result
}

func (a AvailableTorrent) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type Media struct {
	ImdbID              string `gorm:"primaryKey"`
	Type                string
	Title               string
	Status              string
	Description         string
	Poster              string
	Year                string
	CurrentDownloadHash sql.NullString
	AvailableTorrents   []AvailableTorrent `gorm:"type:json"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type MediaIMDB struct {
	Title  string `json:"title"`
	Year   string `json:"year"`
	ImdbID string `json:"imdbID"`
	Type   string `json:"type"`
	Poster string `json:"poster"`
}

type QueueItem struct {
	ImdbID string
	Hash   string
}
