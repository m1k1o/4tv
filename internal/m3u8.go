package internal

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CreateM3U8ByBuckets(buckets map[string][]Channel, outPath string) error {
	// create m3u8 files
	for bucket, channles := range buckets {
		// create file name
		fileName := fmt.Sprintf("%s.m3u8", bucket)

		// create file path
		filePath := filepath.Join(outPath, fileName)

		// create m3u8 struct
		m3u8 := ChannelsToM3U8(fmt.Sprintf("%s.xml", bucket), channles)

		// write m3u8 file
		if err := os.WriteFile(filePath, []byte(m3u8), 0644); err != nil {
			return err
		}
	}

	return nil
}

func ChannelsToM3U8(xmlTvUrl string, channel []Channel) string {
	var buffer bytes.Buffer
	if xmlTvUrl != "" {
		buffer.WriteString(fmt.Sprintf("#EXTM3U x-tvg-url=\"%s\"\n", xmlTvUrl))
	} else {
		buffer.WriteString("#EXTM3U\n")
	}

	for i, channel := range channel {
		if len(channel.Streams) == 0 {
			continue
		}
		stream := channel.Streams[0] // we expect only one stream per channel

		buffer.WriteString(fmt.Sprintf("#EXTINF:-1 tvg-chno=\"%d\"", i))

		if channel.Logo != "" {
			buffer.WriteString(fmt.Sprintf(" tvg-logo=\"%s\"", channel.Logo))
		}

		if len(channel.Labels) > 0 {
			buffer.WriteString(fmt.Sprintf(" group-title=\"%s\"", strings.Join(channel.Labels, ";")))
		}

		if len(channel.Epg) > 0 {
			epg := channel.Epg[0] // we expect only one epg source per channel
			buffer.WriteString(fmt.Sprintf(" tvg-id=\"%s\"", epg.ID))
		}

		if stream.Catchup.Mode != "" {
			buffer.WriteString(fmt.Sprintf(" catchup=\"%s\"", stream.Catchup.Mode))
			if stream.Catchup.Days > 0 {
				buffer.WriteString(fmt.Sprintf(" catchup-days=\"%d\"", stream.Catchup.Days))
			}
			if stream.Catchup.Source != "" {
				buffer.WriteString(fmt.Sprintf(" catchup-source=\"%s\"", stream.Catchup.Source))
			}
		}

		buffer.WriteString(fmt.Sprintf(",%s\n%s\n", channel.Name, stream.URL))
	}

	return buffer.String()
}
