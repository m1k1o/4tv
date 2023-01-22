package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func downloadFile(filepath, url string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

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
	_, err = io.Copy(out, resp.Body)
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
