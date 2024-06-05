package internal

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CreateM3U8ByBuckets(url string, buckets []Bucket, outPath string) error {
	// remove last slash
	url = strings.TrimSuffix(url, "/")

	// create m3u8 files
	for _, data := range buckets {
		// create file name
		fileName := fmt.Sprintf("%s.m3u8", data.Id)

		// create file path
		filePath := filepath.Join(outPath, fileName)

		// create m3u8 struct
		m3u8 := ChannelsToM3U8(fmt.Sprintf("%s/%s.xml", url, data.Id), data.Channels)

		// write m3u8 file
		if err := os.WriteFile(filePath, []byte(m3u8), 0644); err != nil {
			return err
		}
	}

	return nil
}

func ChannelsToM3U8(xmlTvUrl string, channel []BucketChannel) string {
	var buffer bytes.Buffer
	if xmlTvUrl != "" {
		buffer.WriteString(fmt.Sprintf("#EXTM3U x-tvg-url=\"%s\"\n", xmlTvUrl))
	} else {
		buffer.WriteString("#EXTM3U\n")
	}

	for i, ch := range channel {
		buffer.WriteString(fmt.Sprintf("#EXTINF:-1 tvg-chno=\"%d\"", i))

		if ch.Logo != "" {
			buffer.WriteString(fmt.Sprintf(" tvg-logo=\"%s\"", ch.Logo))
		}

		if len(ch.Labels) > 0 {
			buffer.WriteString(fmt.Sprintf(" group-title=\"%s\"", strings.Join(ch.Labels, ";")))
		}

		if len(ch.Epgs) > 0 {
			buffer.WriteString(fmt.Sprintf(" tvg-id=\"%s\"", ch.Id))
		}

		if ch.Stream.Catchup != nil && ch.Stream.Catchup.Mode != "" {
			buffer.WriteString(fmt.Sprintf(" catchup=\"%s\"", ch.Stream.Catchup.Mode))
			if ch.Stream.Catchup.Days > 0 {
				buffer.WriteString(fmt.Sprintf(" catchup-days=\"%d\"", ch.Stream.Catchup.Days))
			}
			if ch.Stream.Catchup.Source != "" {
				buffer.WriteString(fmt.Sprintf(" catchup-source=\"%s\"", ch.Stream.Catchup.Source))
			}
		}

		buffer.WriteString(fmt.Sprintf(",%s\n%s\n", ch.Name, ch.Stream.URL))
	}

	return buffer.String()
}
