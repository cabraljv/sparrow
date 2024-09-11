package mediaseach

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sparrow/models"
)

type MediaIMDBResponse struct {
	Search       []models.MediaIMDB `json:"Search"`
	TotalResults string             `json:"totalResults"`
	Response     string             `json:"Response"`
}

var API_KEY = ""
var BASE_URL = "http://www.omdbapi.com/?apikey=" + API_KEY

func SearchMedia(query string) (mediaIMDB []models.MediaIMDB, err error) {
	url := BASE_URL + "&s=" + query

	resp, err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("error when searching MediaIMDB")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	responseData := MediaIMDBResponse{}

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return
	}
	for _, media := range responseData.Search {
		if media.Type == "movie" {
			mediaIMDB = append(mediaIMDB, media)
		}
	}

	//only return movies by now filtering by type

	return
}

func GetMediaData(imdbID string) (media models.MediaIMDB, err error) {
	url := BASE_URL + "&i=" + imdbID

	resp, err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("error when searching MediaIMDB")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &media)
	if err != nil {
		return
	}

	return
}
