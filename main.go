package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Hello, World!")

	magnetUriToDownload := `magnet:?xt=urn:btih:B39E457EE2A4EE81BF2FCE4925CF3EF064F02BE1&dn=Ratatouille+2007+REPACK+1080p+BluRay+DD%2B+5.1+x265-edge2020&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.torrent.eu.org%3A451&tr=udp%3A%2F%2Fopen.stealth.si%3A80%2Fannounce&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A6969&tr=udp%3A%2F%2Ftracker.tiny-vps.com%3A6969%2Fannounce&tr=udp%3A%2F%2Fopen.demonii.com%3A1337&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=http%3A%2F%2Ftracker.openbittorrent.com%3A80%2Fannounce&tr=udp%3A%2F%2Fopentracker.i2p.rocks%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.internetwarriors.net%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969%2Fannounce&tr=udp%3A%2F%2Fcoppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.zer0day.to%3A1337%2Fannounce`

	torrent, err := InitTorrentDownload(magnetUriToDownload)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(torrent.Name)

	GenerateM3U8(torrent.Hash)
	// err = GenerateFragments("."+torrent.ContentPath, downloadedSeconds, torrent.Hash, videoLength)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			pieces, err := GetTorrentPiecesStatus(torrent.Hash)
			if err != nil {
				fmt.Println(err)
				continue
			}
			videoLength, err := GetVideoLength("." + torrent.ContentPath)
			if err != nil {
				fmt.Println(err)
				continue
			}
			downloadedSeconds, err := GetDownloadedVideoParts(pieces, videoLength)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = GenerateFragments("."+torrent.ContentPath, downloadedSeconds, torrent.Hash, videoLength)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	// Keeping the main routine alive. You may replace this with your actual main routine logic.
	select {}

}
