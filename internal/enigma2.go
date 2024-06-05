package internal

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

func CreateEnigma2ByBuckets(buckets []Bucket, outPath string) error {
	// if having multiple buckets, we need to use different namespace
	var namespace = 1000

	// create enigma2 files
	for _, data := range buckets {
		// create file name
		tvFileName := fmt.Sprintf("%s.tv", data.Id)
		channelsFileName := fmt.Sprintf("%s.channels.xml", data.Id)

		// create file path
		tvFilePath := filepath.Join(outPath, tvFileName)
		channelsFilePath := filepath.Join(outPath, channelsFileName)

		// create enigma2 struct
		tv, channels := ChannelsToEnigma2(data.Name, namespace, data.Channels)

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

func ChannelsToEnigma2(bucketName string, namespace int, channel []BucketChannel) (tv, channels string) {
	var tvBuffer, channelsBuffer bytes.Buffer

	tvBuffer.WriteString(fmt.Sprintf("#NAME %s\n", bucketName))

	channelsBuffer.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\" ?>\n")
	channelsBuffer.WriteString("<channels>\n")

	for i, channel := range channel {
		id := i
		servicePrefix := fmt.Sprintf("4097:0:1:%d:%d:0:0:0:0:0", id, namespace)

		tvBuffer.WriteString(fmt.Sprintf("#SERVICE %s:%s:%s\n", servicePrefix, url.QueryEscape(channel.Stream.URL), channel.Name))
		tvBuffer.WriteString(fmt.Sprintf("#DESCRIPTION %s\n", channel.Name))

		if len(channel.Epgs) > 0 {
			channelsBuffer.WriteString(fmt.Sprintf("<channel id=\"%s\">%s:http%%3A//example.com</channel> <!-- %s -->\n", channel.Id, servicePrefix, channel.Name))
		}
	}

	channelsBuffer.WriteString("</channels>\n")

	return tvBuffer.String(), channelsBuffer.String()
}
