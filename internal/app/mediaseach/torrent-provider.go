package mediaseach

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"sparrow/models"
	"strings"
)

type Torrent struct {
	Name          string               `json:"name"`
	Title         string               `json:"title"`
	InfoHash      string               `json:"infoHash"`
	FileIdx       int                  `json:"fileIdx"`
	BehaviorHints TorrentBehaviorHints `json:"behaviorHints"`
}

type TorrentBehaviorHints struct {
	BingeGroup string `json:"bingeGroup"`
	Filename   string `json:"filename"`
}

type TorrentProvider struct {
	TransportUrl  string   `json:"transportUrl"`
	TransportName string   `json:"transportName"`
	Manifest      Manifest `json:"manifest"`
}

type Manifest struct {
	Id            string         `json:"id"`
	Version       string         `json:"version"`
	Description   string         `json:"description"`
	Name          string         `json:"name"`
	Types         []string       `json:"types"`
	AddonCatalogs []AddonCatalog `json:"addonCatalogs"`
	Catalogs      []Catalog      `json:"catalogs"`
}

type AddonCatalog struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type Catalog struct {
	Type   string   `json:"type"`
	Id     string   `json:"id"`
	Genres []string `json:"genres"`
	Name   string   `json:"name"`
}

func loadTorrentProviders() (providers []TorrentProvider, err error) {

	file, err := os.Open("./configs/torrent-providers.json")
	if err != nil {
		return
	}
	body, err := io.ReadAll(file)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &providers)

	return
}

func GetTorrents(media models.Media) (torrents []models.AvailableTorrent, err error) {

	providers, err := loadTorrentProviders()

	if err != nil {
		return
	}

	for _, provider := range providers {
		torrents = append(torrents, searchTorrentOnProvider(provider, media)...)
	}

	return
}

type ProviderResponse struct {
	Streams []struct {
		Name          string `json:"name"`
		Title         string `json:"title"`
		InfoHash      string `json:"infoHash"`
		FileIdx       int    `json:"fileIdx,omitempty"`
		BehaviorHints struct {
			BingeGroup string `json:"bingeGroup"`
			Filename   string `json:"filename"`
		} `json:"behaviorHints"`
	} `json:"streams"`
	CacheMaxAge     int `json:"cacheMaxAge"`
	StaleRevalidate int `json:"staleRevalidate"`
	StaleError      int `json:"staleError"`
}

func searchTorrentOnProvider(provider TorrentProvider, media models.Media) (torrents []models.AvailableTorrent) {

	torrents = []models.AvailableTorrent{}
	fmt.Println("Searching for torrents on provider: ", provider.TransportUrl)
	baseUrl, _ := strings.CutSuffix(provider.TransportUrl, "/manifest.json")
	if !slices.Contains(provider.Manifest.Types, media.Type) {
		return
	}
	requestUri := baseUrl + "/stream/" + media.Type + "/" + media.ImdbID + ".json"

	resp, err := http.Get(requestUri)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	responseData := ProviderResponse{}

	json.Unmarshal(body, &responseData)

	torrents = make([]models.AvailableTorrent, len(responseData.Streams))

	for i, stream := range responseData.Streams {
		torrents[i] = models.AvailableTorrent{
			InfoHash:   stream.InfoHash,
			Provider:   provider.Manifest.Id,
			ImdbID:     media.ImdbID,
			BingeGroup: stream.BehaviorHints.BingeGroup,
			Filename:   stream.BehaviorHints.Filename,
		}
	}

	return
}
