package app

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type MediaConfig struct {
	ImdbID            string `json:"imdbID"`
	Title             string `json:"title"`
	TotalDuration     int64  `json:"totalDuration"`
	AvailableDuration int    `json:"availableDuration"`
	Status            string `json:"status"`
}

func UpdateMediaConfigFile(mediaConfig *MediaConfig, hashId string) error {
	// create empty file if not exists
	filePath := filepath.Join("tmp", hashId, "media.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath) // Create the file
		if err != nil {
			return err
		}
		defer file.Close() // Ensure the file is closed after writing

		if _, err := file.WriteString("{}"); err != nil {
			return err
		}
	}
	// get existent media config
	existentMediaConfig, err := LoadMediaConfigFromFile(hashId)
	if err != nil {
		return err
	}
	if mediaConfig.ImdbID != "" {
		existentMediaConfig.ImdbID = mediaConfig.ImdbID
	}
	if mediaConfig.Title != "" {
		existentMediaConfig.Title = mediaConfig.Title
	}
	if mediaConfig.TotalDuration != 0 {
		existentMediaConfig.TotalDuration = mediaConfig.TotalDuration
	}
	if mediaConfig.AvailableDuration != 0 {
		existentMediaConfig.AvailableDuration = mediaConfig.AvailableDuration
	}
	if mediaConfig.Status != "" {
		existentMediaConfig.Status = mediaConfig.Status
	}
	mediaConfigBytes, err := json.Marshal(existentMediaConfig)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filePath, mediaConfigBytes, 0644); err != nil {
		return err
	}
	return nil

}

func LoadMediaConfigFromFile(hashId string) (*MediaConfig, error) {
	filePath := filepath.Join("tmp", hashId, "media.json")
	// Read the file
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON into a MediaConfig struct
	var mediaConfig MediaConfig
	if err := json.Unmarshal(fileContent, &mediaConfig); err != nil {
		return nil, err
	}

	return &mediaConfig, nil
}
