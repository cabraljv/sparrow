package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type BehaviorHints struct {
	BingeGroup string `json:"bingeGroup"`
	Filename   string `json:"filename"`
}

type AvailableTorrent struct {
	ImdbID     string
	Provider   string
	InfoHash   string
	BingeGroup string
	Filename   string
}

type JSONAvailableTorrents []AvailableTorrent

func (j JSONAvailableTorrents) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONAvailableTorrents) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONAvailableTorrents, 0)
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &j)
}

type Media struct {
	ImdbID              string `gorm:"primaryKey"`
	Type                string
	Title               string
	Status              string
	Description         string
	Poster              string
	Year                string
	CurrentDownloadHash sql.NullString        `gorm:"type:varchar(255)"`
	AvailableTorrents   JSONAvailableTorrents `gorm:"type:json"`
	CreatedAt           time.Time             `gorm:"autoCreateTime"`
	UpdatedAt           time.Time             `gorm:"autoUpdateTime"`
}

type MediaIMDB struct {
	Title  string `json:"title"`
	Year   string `json:"year"`
	ImdbID string `json:"imdbID"`
	Type   string `json:"type"`
	Poster string `json:"poster"`
}
