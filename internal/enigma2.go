package internal

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

func CreateEnigma2ByBuckets(buckets map[string][]Channel, outPath string) error {
	// if having multiple buckets, we need to use different namespace
	var namespace = 1000

	// create enigma2 files
	for bucket, channles := range buckets {
		// create file name
		tvFileName := fmt.Sprintf("%s.tv", bucket)
		channelsFileName := fmt.Sprintf("%s.channels.xml", bucket)

		// create file path
		tvFilePath := filepath.Join(outPath, tvFileName)
		channelsFilePath := filepath.Join(outPath, channelsFileName)

		// create enigma2 struct
		tv, channels := ChannelsToEnigma2(bucket, namespace, channles)

		// write enigma2 file
		if err := os.WriteFile(tvFilePath, []byte(tv), 0644); err != nil {
			return err
		}

		if err := os.WriteFile(channelsFilePath, []byte(channels), 0644); err != nil {
			return err
		}

		namespace++
	}

	return nil
}

func ChannelsToEnigma2(bucketName string, namespace int, channel []Channel) (tv, channels string) {
	var tvBuffer, channelsBuffer bytes.Buffer

	tvBuffer.WriteString(fmt.Sprintf("#NAME %s\n", bucketName))

	channelsBuffer.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\" ?>\n")
	channelsBuffer.WriteString("<channels>\n")

	for i, channel := range channel {
		if len(channel.Streams) == 0 {
			continue
		}
		stream := channel.Streams[0] // we expect only one stream per channel

		id := i
		servicePrefix := fmt.Sprintf("4097:0:1:%d:%d:0:0:0:0:0", id, namespace)

		tvBuffer.WriteString(fmt.Sprintf("#SERVICE %s:%s:%s\n", servicePrefix, url.QueryEscape(stream.URL), channel.Name))
		tvBuffer.WriteString(fmt.Sprintf("#DESCRIPTION %s\n", channel.Name))

		if len(channel.Epg) > 0 {
			epg := channel.Epg[0] // we expect only one epg source per channel
			channelsBuffer.WriteString(fmt.Sprintf("<channel id=\"%s\">%s:http%%3A//example.com</channel> <!-- %s -->\n", epg.ID, servicePrefix, channel.Name))
		}
	}

	channelsBuffer.WriteString("</channels>\n")

	return tvBuffer.String(), channelsBuffer.String()
}
