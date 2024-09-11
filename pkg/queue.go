package pkg

import (
	"io"
	"os"
	"sparrow/models"
	"strings"
)

func GetNextQueueItem() (item models.QueueItem, err error) {
	// read txt file
	// get first line

	filePath := "queue.txt"
	// Open file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	// Read the file
	data, err := io.ReadAll(file)
	if err != nil {
		return
	}

	// Get the first line
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 {
		line := strings.Split(lines[0], ",")

		item = models.QueueItem{
			Hash:   line[1],
			ImdbID: line[0],
		}
		return
	}

	return

}
func PushItemToQueue(item models.QueueItem) {
	filePath := "queue.txt"
	// Open file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	// Read the file
	data, err := io.ReadAll(file)
	if err != nil {
		return
	}

	// Get the first line
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 {
		lines = append(lines, item.ImdbID+","+item.Hash)
	}

	// Write the new data
	newData := strings.Join(lines, "\n")
	err = os.WriteFile(filePath, []byte(newData), 0644)
	if err != nil {
		return
	}
}

func RemoveQueueItem(item models.QueueItem) {
	filePath := "queue.txt"
	// Open file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	// Read the file
	data, err := io.ReadAll(file)
	if err != nil {
		return
	}

	// Get the first line
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 {
		lines = lines[1:]
	}

	// Write the new data
	newData := strings.Join(lines, "\n")
	err = os.WriteFile(filePath, []byte(newData), 0644)
	if err != nil {
		return
	}
}
