package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

const (
	qbittorrentURL = "http://localhost:8080"
	username       = "admin"
	password       = "adminadmin"
)

func parseCookies(cookieString string) []*http.Cookie {
	cookiePairs := strings.Split(cookieString, ";")
	cookies := []*http.Cookie{}

	for _, pair := range cookiePairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			cookie := &http.Cookie{
				Name:  strings.TrimSpace(parts[0]),
				Value: strings.TrimSpace(parts[1]),
			}
			cookies = append(cookies, cookie)
		}
	}

	return cookies
}

type TorrentData struct {
	Hash        string  `json:"hash"`
	Name        string  `json:"name"`
	MagnetURI   string  `json:"magnet_uri"`
	Progress    float32 `json:"progress"`
	ContentPath string  `json:"content_path"`
	Size        int64   `json:"size"`
}

func GetTorrentPiecesStatus(hash string) (pieces []int16, err error) {
	client, err := login()

	if err != nil {
		return
	}

	uri := qbittorrentURL + "/api/v2/torrents/pieceStates?hash=" + hash
	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		return
	}

	resp, err := client.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("error when get pieces")
		return
	}

	body, err := io.ReadAll(resp.Body)

	err = json.Unmarshal(body, &pieces)

	if err != nil {
		return
	}
	return
}

func GetTorrentDataFromHash(hash string) (torrent TorrentData, err error) {
	client, err := login()
	if err != nil {
		return
	}
	uri := qbittorrentURL + "/api/v2/torrents/info"

	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		return
	}
	resp, err := client.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("error when get torrents")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var torrents []TorrentData

	err = json.Unmarshal(body, &torrents)
	if err != nil {
		return
	}

	for _, t := range torrents {
		if t.Hash == hash {
			torrent = t
			return
		}
	}

	err = errors.New("torrent not found")
	return

}

func getTorrentHash(magnetURI string) (string, error) {
	u, err := url.Parse(magnetURI)
	if err != nil {
		return "", err
	}
	if u.Scheme != "magnet" {
		return "", fmt.Errorf("invalid scheme: %s", u.Scheme)
	}

	params := u.Query()
	xt := params.Get("xt")
	if xt == "" {
		return "", fmt.Errorf("missing xt parameter in magnet URI")
	}

	const (
		prefix = "urn:btih:"
	)
	if !strings.HasPrefix(xt, prefix) {
		return "", fmt.Errorf("invalid xt format: %s", xt)
	}

	hash := strings.ToLower(strings.TrimPrefix(xt, prefix))
	return hash, nil
}

// InitTorrentDownload initializes the download of a torrent via the qBittorrent API
func InitTorrentDownload(magnetURI string) (torrent TorrentData, err error) {

	client, err := login()
	if err != nil {
		return
	}
	// set cookie to client

	if !isTorrentAlreadyAdded(client, magnetURI) {
		addTorrent(client, magnetURI)
	}

	torrentHash, err := getTorrentHash(magnetURI)
	if err != nil {
		return
	}

	torrent, err = GetTorrentDataFromHash(torrentHash)

	return
}

func login() (client *http.Client, err error) {
	client = &http.Client{}

	uri := qbittorrentURL + "/api/v2/auth/login"
	data := fmt.Sprintf("username=%s&password=%s", username, password)

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer([]byte(data)))
	if err != nil {
		err = fmt.Errorf("failed to create login request: %w", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to login: %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("login failed with status code: %d, %v", resp.StatusCode, string(body))
		return
	}

	cookies := parseCookies(resp.Header.Get("set-cookie"))
	u, _ := url.Parse(qbittorrentURL) // Use the relevant URL
	jar, _ := cookiejar.New(nil)
	client.Jar = jar

	client.Jar.SetCookies(u, cookies)
	return
}

func isTorrentAlreadyAdded(client *http.Client, magnetURI string) bool {
	uri := qbittorrentURL + "/api/v2/torrents/info"

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error checking existing torrents:", err)
		return false
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return false
	}

	return bytes.Contains(bodyBytes, []byte(magnetURI))
}

func addTorrent(client *http.Client, magnetURI string) error {
	fmt.Println("Starting torrent download...")
	uri := qbittorrentURL + "/api/v2/torrents/add"
	data := fmt.Sprintf("urls=%s", magnetURI)

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return fmt.Errorf("failed to create add torrent request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add torrent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add torrent with status code: %d", resp.StatusCode)
	}

	return nil
}
