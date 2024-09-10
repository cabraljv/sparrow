package main

func GetDownloadedVideoParts(pieces []int16, videoLength int64) (downloadedParts map[int64]bool, err error) {
	// Initialize the map to store the download status of each second of the video
	downloadedParts = make(map[int64]bool)

	// Calculate the length of each piece in seconds
	if len(pieces) == 0 {
		return downloadedParts, nil
	}
	pieceLength := videoLength / int64(len(pieces))
	if videoLength%int64(len(pieces)) != 0 {
		pieceLength++
	}

	// Fill the map with the downloaded status
	for i, status := range pieces {
		for second := int64(i) * pieceLength; second < int64(i+1)*pieceLength && second < videoLength; second++ {
			downloadedParts[second] = (status == 2)
		}
	}

	return downloadedParts, nil
}
