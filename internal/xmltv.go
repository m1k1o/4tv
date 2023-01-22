package internal

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type xmltv struct {
	ChannelList   []xmlchannel   `xml:"channel"`
	ProgrammeList []xmlprogramme `xml:"programme"`
}

type xmlchannel struct {
	Id   string `xml:"id,attr"`
	Name string `xml:"display-name"`
}

type xmlprogramme struct {
	Start       string   `xml:"start,attr"`
	Stop        string   `xml:"stop,attr"`
	Channel     string   `xml:"channel,attr"`
	Title       string   `xml:"title"`
	SubTitle    string   `xml:"sub-title"`
	Description string   `xml:"desc"`
	Credits     string   `xml:"credits"`
	Date        string   `xml:"date"`
	Categories  []string `xml:"category"`
	Rating      string   `xml:"rating>value"`
}

func UnmarshalXmlTv(s []byte) (x xmltv, err error) {
	err = xml.Unmarshal(s, &x)
	return
}

func MarshalXmlTv(x xmltv) ([]byte, error) {
	return xml.MarshalIndent(x, "", "  ")
}

func FilterXmlTvByChannels(input xmltv, channels []string) (output xmltv, err error) {
	// filter channels list
	for _, v := range input.ChannelList {
		for _, c := range channels {
			if v.Id == c {
				// add channel to output list if it is in the channels list
				output.ChannelList = append(output.ChannelList, v)
				break
			}
		}
	}

	// filter programme list
	for _, v := range input.ProgrammeList {
		for _, c := range channels {
			if v.Channel == c {
				// add channel to output list if it is in the channels list
				output.ProgrammeList = append(output.ProgrammeList, v)
				break
			}
		}
	}

	return
}

func JoinXmlTvs(inputs ...xmltv) (output xmltv, err error) {
	// join channels
	for _, input := range inputs {
		output.ChannelList = append(output.ChannelList, input.ChannelList...)
	}

	// join programmes
	for _, input := range inputs {
		output.ProgrammeList = append(output.ProgrammeList, input.ProgrammeList...)
	}

	return
}

func DownloadXmlTvByEpgSoruce(sources []EpgSource, dataPath string) error {
	for _, source := range sources {
		// create file name
		fileName := fmt.Sprintf("%s.xml", source.Provider)

		// create file path
		filePath := filepath.Join(dataPath, fileName)

		// download file
		if err := downloadFile(filePath, source.URL); err != nil {
			return err
		}
	}

	return nil
}

func CreateXmlTvByBuckets(epg []EpgSource, buckets map[string][]Channel, outPath string) error {
	tmpDir, err := os.MkdirTemp("./tmp/", "xmltv-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// download xmltv files
	if err := DownloadXmlTvByEpgSoruce(epg, tmpDir); err != nil {
		return err
	}

	// create epg buckets
	epgBuckets := make(map[string]map[string][]string) // bucket -> provider -> channels
	for bucket, channels := range buckets {
		epgBuckets[bucket] = make(map[string][]string)
		for _, channel := range channels {
			if len(channel.Epg) > 0 {
				epg := channel.Epg[0] // we expect only one epg source per channel
				epgBuckets[bucket][epg.Provider] = append(epgBuckets[bucket][epg.Provider], epg.ID)
			}
		}
	}

	// create xmltv files
	for bucket, providers := range epgBuckets {
		// create file name
		fileName := fmt.Sprintf("%s.xml", bucket)

		// create file path
		filePath := filepath.Join(outPath, fileName)

		// create xmltv struct
		var xmltv xmltv

		// join xmltv files
		for provider, channels := range providers {
			// create file name
			fileName := fmt.Sprintf("%s.xml", provider)

			// create file path
			filePath := filepath.Join(tmpDir, fileName)

			// read file
			file, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}

			// unmarshal xmltv
			xmltvFile, err := UnmarshalXmlTv(file)
			if err != nil {
				return err
			}

			// filter xmltv
			xmltvFile, err = FilterXmlTvByChannels(xmltvFile, channels)
			if err != nil {
				return err
			}

			// join xmltv
			xmltv, err = JoinXmlTvs(xmltv, xmltvFile)
			if err != nil {
				return err
			}
		}

		// marshal xmltv
		xmltvFile, err := MarshalXmlTv(xmltv)
		if err != nil {
			return err
		}

		// write file
		if err := os.WriteFile(filePath, xmltvFile, 0644); err != nil {
			return err
		}
	}

	return nil
}
