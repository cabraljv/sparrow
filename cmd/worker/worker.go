package main

import (
	"fmt"
	"sparrow/internal/app"
	"sparrow/internal/app/database"
	"sparrow/models"
	"time"
)

func main() {

	// get the next item from the queue
	item, err := database.PullQueueItem()
	if err != nil {
		fmt.Println(err)
	}
	if item.Hash == "" {
		fmt.Println("No items in the queue")
	} else {
		ProcessQueueItem(item)
	}

	time.Sleep(10 * time.Second)
	main()
}

func ProcessQueueItem(item models.QueueItem) {
	fmt.Printf("Processing queue item %s\n", item.ImdbID)
	if item.Status != "downloading" {
		fmt.Printf("Queue item %s is \"%s\"\n", item.ImdbID, item.Status)
		database.UpdateQueueItemStatus(item.ImdbID, "downloading")
	}
	magnetUriToDownload := app.GenerateMagnetURI(item.Hash)

	torrent, err := app.InitTorrentDownload(magnetUriToDownload)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(torrent.Name)

	app.GenerateM3U8(torrent.Hash)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var videoLength int64
	go func() {
		for range ticker.C {
			fmt.Println("Checking torrent status")
			// verify if queue item is still in the queue
			queueItem, err := database.GetQueueItem(item.ImdbID)
			if err != nil {
				fmt.Println("Error getting queue item", err)
				return
			}
			if queueItem.ImdbID == "" {
				fmt.Println("Queue item not found")
				return
			}
			if queueItem.Status != "waiting" && queueItem.Status != "downloading" {
				fmt.Printf("Queue item %s is \"%s\"\n", queueItem.ImdbID, queueItem.Status)
				return
			}

			pieces, err := app.GetTorrentPiecesStatus(torrent.Hash)
			if err != nil {
				fmt.Println("Error getting pieces", err)
				continue
			}
			fmt.Println("Pieces got")
			if videoLength == 0 {
				videoLength, err = app.GetVideoLength("." + torrent.ContentPath)
				if err != nil {
					fmt.Println("Error getting video length", err)
					continue
				}
				fmt.Println("Video length got", videoLength)
			}
			app.UpdateMediaConfigFile(&app.MediaConfig{
				ImdbID:        torrent.Hash,
				Status:        "processing",
				TotalDuration: videoLength,
				Title:         torrent.Name,
			}, torrent.Hash)
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
