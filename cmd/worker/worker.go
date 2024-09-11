package main

import (
	"fmt"
	"sparrow/internal/app"
	"sparrow/models"
	"sparrow/pkg"
	"time"
)

func main() {

	// get the next item from the queue
	item, err := pkg.GetNextQueueItem()
	if err != nil {
		fmt.Println(err)
	}
	if item.Hash != "" {
		fmt.Println("No items in the queue")
		ProcessQueueItem(item)

		pkg.RemoveQueueItem(item)

	}

	time.Sleep(10 * time.Second)
	main()
}

func ProcessQueueItem(item models.QueueItem) {

	magnetUriToDownload := app.GenerateMagnetURI(item.Hash)

	torrent, err := app.InitTorrentDownload(magnetUriToDownload)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(torrent.Name)

	app.GenerateM3U8(torrent.Hash)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			pieces, err := app.GetTorrentPiecesStatus(torrent.Hash)
			if err != nil {
				fmt.Println(err)
				continue
			}
			videoLength, err := app.GetVideoLength("." + torrent.ContentPath)
			if err != nil {
				fmt.Println(err)
				continue
			}
			downloadedSeconds, err := app.GetDownloadedVideoParts(pieces, videoLength)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = app.GenerateFragments("."+torrent.ContentPath, downloadedSeconds, torrent.Hash, videoLength)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	// Keeping the main routine alive. You may replace this with your actual main routine logic.
	select {}
}
