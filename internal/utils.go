package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func downloadFile(file *os.File, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(file, resp.Body)
	return err
}

func ArrayIn[T comparable](val T, array []T) (exists bool, index int) {
	exists, index = false, -1
	for i, a := range array {
		if a == val {
			exists, index = true, i
			return
		}
	}
	return
}
