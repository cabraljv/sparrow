package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

var FRAG_TIME = int64(12)

func getVideoFileByDownloadPath(downloadPath string) (filePath string, err error) {
	info, err := os.Stat(downloadPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("O caminho não existe.")
		}
		fmt.Println("err", err)
		return
	}
	if info.IsDir() {
		// TO DO
	} else {
		filePath = downloadPath
	}
	return
}

func GetVideoLength(downloadPath string) (length int64, err error) {

	videoFilePath, err := getVideoFileByDownloadPath(downloadPath)

	if err != nil {
		return
	}
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoFilePath)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		fmt.Println("erraa", err)
		return 0, err
	}

	durationStr := strings.TrimSpace(out.String())
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}

	// Return the duration in seconds as an integer (you may adjust to milliseconds if needed)
	length = int64(durationFloat)
	return
}

func GenerateM3U8(hash string) (err error) {
	basePath := "./tmp/" + hash + "/"
	if err != nil {
		return fmt.Errorf("invalid duration: %v", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	playlistPath := basePath + "playlist.m3u8"
	file, err := os.Create(playlistPath)
	if err != nil {
		return fmt.Errorf("failed to create playlist file: %v", err)
	}
	defer file.Close()

	// Write M3U8 file
	if _, err := file.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n"); err != nil {
		return fmt.Errorf("failed to write headers: %v", err)
	}
	if _, err := file.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", FRAG_TIME)); err != nil {
		return fmt.Errorf("failed to write target duration: %v", err)
	}
	if _, err := file.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n"); err != nil {
		return fmt.Errorf("failed to write media sequence: %v", err)
	}
	// Write endlist tag
	// if _, err := file.WriteString("#EXT-X-ENDLIST\n"); err != nil {
	// 	return fmt.Errorf("failed to write endlist: %v", err)
	// }

	return nil
}

func GenerateFragments(downloadPath string, downloadedParts map[int64]bool, id string, length int64) error {
	videoFilePath, err := getVideoFileByDownloadPath(downloadPath)
	if err != nil {
		return fmt.Errorf("failed to get video file: %v", err)
	}

	startSec := int64(-1)
	maxConcurrentFragments := 1
	sem := make(chan struct{}, maxConcurrentFragments)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errFragment error

	processFragment := func(startSec, duration int64) {
		defer wg.Done()
		defer func() { <-sem }()
		localErr := generateFragment(id, videoFilePath, "./tmp/"+id+"/", startSec, duration)
		if localErr != nil {
			mu.Lock()
			errFragment = fmt.Errorf("error generating fragment: %v", localErr)
			mu.Unlock()
		}
		localErr = updatePlaylistM3U8(id)
		if localErr != nil {
			mu.Lock()
			errFragment = fmt.Errorf("error updating playlist.m3u8: %v", localErr)
			fmt.Println("errFragment", errFragment)
			mu.Unlock()
		}
	}

	for sec := int64(0); sec < length; sec++ {
		available := downloadedParts[sec]

		if available {
			if startSec == -1 {
				startSec = sec
			}

			if (sec+1)%FRAG_TIME == 0 {
				sem <- struct{}{}
				wg.Add(1)
				go processFragment(startSec, FRAG_TIME)
				startSec = -1
			}
		} else {
			if startSec != -1 && (sec-startSec)%FRAG_TIME != 0 {
				sem <- struct{}{}
				wg.Add(1)
				go processFragment(startSec, sec-startSec)
			}
			startSec = -1
		}
	}

	if startSec != -1 && (length-startSec)%FRAG_TIME != 0 {
		sem <- struct{}{}
		wg.Add(1)
		go processFragment(startSec, length-startSec)
	}

	wg.Wait()

	if errFragment != nil {
		return errFragment
	}
	return nil
}

func secondToTimeText(sec int64) string {
	hours := sec / 3600
	sec -= hours * 3600
	minutes := sec / 60
	sec -= minutes * 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, sec)
}

func getGeneratedFragIndexes(id string) (fragIndexes []int, err error) {
	// create dir if it doesn't exist
	if err := os.MkdirAll("./tmp/"+id, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}
	// create file if it doesn't exist
	if _, err := os.Stat("./tmp/" + id + "/generated.txt"); os.IsNotExist(err) {
		_, err := os.Create("./tmp/" + id + "/generated.txt")
		if err != nil {
			return nil, fmt.Errorf("failed to create generated.txt file: %v", err)
		}
	}

	generatedTxtPath := "./tmp/" + id + "/generated.txt"
	generatedFragIndexesTxt, err := os.ReadFile(generatedTxtPath)
	if err != nil {
		return nil, err
	}
	txtContent := string(generatedFragIndexesTxt)
	stringFragments := strings.Split(txtContent, "\n")
	for _, fragIndexTxt := range stringFragments {
		if fragIndexTxt == "" {
			continue
		}
		fragIndex, err := strconv.Atoi(fragIndexTxt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert fragment index to int: %v", err)
		}
		fragIndexes = append(fragIndexes, fragIndex)
	}
	return
}

// Function to run ffmpeg and generate the .ts fragment files and the .m3u8 playlist
func generateFragment(id, videoFilePath, downloadPath string, startTime, fragTime int64) error {
	generatedFragIndexes, err := getGeneratedFragIndexes(id)

	if err != nil {
		return fmt.Errorf("failed to get generated fragment indexes: %v", err)
	}

	endTime := startTime + fragTime - 1
	correctFragIndex := startTime / fragTime

	// verify if the fragment is already generated
	if contains(generatedFragIndexes, int(correctFragIndex)) {
		return nil
	}
	// verify if the fragment is already generated
	if _, err := os.Stat(downloadPath + fmt.Sprintf("video%d.ts", correctFragIndex)); err == nil {
		return nil
	}

	fmt.Printf("Generating fragment %v\n", correctFragIndex)

	tmpGenerationDir := "./tmp/generated/" + id + "/frag/" + fmt.Sprintf("%v", correctFragIndex) + "/"

	// create dir if it doesn't exist
	if err := os.MkdirAll(downloadPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.MkdirAll(tmpGenerationDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	// Construct the ffmpeg command
	ffmpegCmd := exec.Command("ffmpeg",
		"-i", videoFilePath,
		"-ss", secondToTimeText(startTime),
		"-to", secondToTimeText(endTime+1), // +1 to include the endTime second
		"-c:v", "libx264", // Re-encode video to ensure keyframes
		"-c:a", "aac",
		"-b:a", "192k",
		"-ac", "2",
		"-ar", "44100",
		"-f", "hls",
		"-force_key_frames", "expr:gte(t,n_forced*4)",
		"-hls_time", strconv.Itoa(int(fragTime)),
		"-hls_list_size", "0",
		"-hls_segment_filename", tmpGenerationDir+"video%d.ts",
		tmpGenerationDir+".m3u8",
	)
	// Execute the command
	out, err := ffmpegCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v, output: %s", err, string(out))
	}

	// add fragment index to generated fragments txt
	generatedTxtPath := "./tmp/" + id + "/generated.txt"
	generatedTxt, err := os.OpenFile(generatedTxtPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open generated.txt file: %v", err)
	}
	defer generatedTxt.Close()

	_, err = generatedTxt.WriteString(fmt.Sprintf("%v\n", float64(correctFragIndex)))
	if err != nil {
		return fmt.Errorf("failed to write fragment duration: %v", err)
	}

	// remove all files in the directory
	// for _, file := range files {
	// 	err = os.Remove(tmpGenerationDir + file.Name())
	// 	if err != nil {
	// 		return fmt.Errorf("failed to remove file: %v", err)
	// 	}
	// }

	// // remove the directory
	// err = os.Remove(tmpGenerationDir)
	// if err != nil {
	// 	return fmt.Errorf("failed to remove directory: %v", err)
	// }

	return nil
}

func getFragIndexesFromM3u8(m3u8Path string) (frags []int, err error) {
	file, err := os.Open(m3u8Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open m3u8 file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ".ts") {
			indexStr := strings.Split(line, "video")[1]
			indexStr = strings.Split(indexStr, ".ts")[0]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("failed to convert index to int: %v", err)
			}
			frags = append(frags, index)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan m3u8 file: %v", err)
	}

	return frags, nil
}

func transferAndAppendFrag(id string, fragIndex int) error {
	// verify if previous fragment exists on the playlist.m3u8 and is the fragIndex-1
	fragIndexes, err := getFragIndexesFromM3u8("./tmp/" + id + "/playlist.m3u8")
	if err != nil {
		return fmt.Errorf("failed to get fragment indexes from playlist.m3u8: %v", err)
	}

	if len(fragIndexes) > 0 {
		if fragIndexes[len(fragIndexes)-1] != fragIndex-1 {
			return fmt.Errorf("previous fragment is not the correct one")
		}
	}

	// get the fragment file
	tmpGenerationDir := "./tmp/generated/" + id + "/frag/" + fmt.Sprintf("%v", fragIndex) + "/"
	videoDir := "./tmp/" + id + "/"

	files, err := os.ReadDir(tmpGenerationDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".ts") {
			err = os.Rename(tmpGenerationDir+file.Name(), videoDir+fmt.Sprintf("video%d.ts", fragIndex))
			if err != nil {
				return fmt.Errorf("failed to move file: %v", err)
			}
		}
	}

	// get the fragment duration from .m3u8 EXTINF
	m3u8Path := tmpGenerationDir + ".m3u8"
	file, err := os.Open(m3u8Path)
	if err != nil {
		return fmt.Errorf("failed to open m3u8 file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentFragTime := float64(0)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "#EXTINF") {
			durationStr := strings.Split(line, ":")[1]
			durationStr = strings.Split(durationStr, ",")[0]
			duration, err := strconv.ParseFloat(durationStr, 64)
			if err != nil {
				return fmt.Errorf("failed to convert duration to float: %v", err)
			}
			currentFragTime = duration
			break
		}
	}

	// add the fragment to the playlist file if it's not already there

	playlistPath := "./tmp/" + id + "/playlist.m3u8"
	playlistFile, err := os.OpenFile(playlistPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open playlist file: %v", err)
	}
	defer playlistFile.Close()

	_, err = playlistFile.WriteString(fmt.Sprintf("#EXTINF:%f,\n", float64(currentFragTime)))
	if err != nil {
		return fmt.Errorf("failed to write fragment duration: %v", err)
	}
	_, err = playlistFile.WriteString(fmt.Sprintf("video%d.ts\n", fragIndex))
	if err != nil {
		return fmt.Errorf("failed to write fragment file name: %v", err)
	}
	_, err = playlistFile.WriteString("#EXT-X-DISCONTINUITY\n")
	if err != nil {
		return fmt.Errorf("failed to write fragment file name: %v", err)
	}

	return nil

}
func contains(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
func updatePlaylistM3U8(id string) (err error) {

	currentM3u8Fragments, err := getFragIndexesFromM3u8("./tmp/" + id + "/playlist.m3u8")

	if err != nil {
		return fmt.Errorf("failed to get fragment indexes from playlist.m3u8: %v", err)
	}

	generatedFragIndexes, err := getGeneratedFragIndexes(id)

	// get the fragments that were generated but are not in the playlist.m3u8

	toAppendFragments := []int{}
	for _, fragIndex := range generatedFragIndexes {
		if !contains(currentM3u8Fragments, fragIndex) {
			toAppendFragments = append(toAppendFragments, fragIndex)
		}
	}

	for _, fragIndex := range toAppendFragments {
		errTransfer := transferAndAppendFrag(id, fragIndex)
		if errTransfer != nil {
			continue
		}
	}

	return

}
