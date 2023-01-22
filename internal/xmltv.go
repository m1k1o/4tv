package internal

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type xmltv struct {
	ChannelList   []xmlchannel   `xml:"channel"`
	ProgrammeList []xmlprogramme `xml:"programme"`
}

type xmlchannel struct {
	Id  string `xml:"id,attr"`
	Raw string `xml:",innerxml"`
}

type xmlprogramme struct {
	Start   string `xml:"start,attr"`
	Stop    string `xml:"stop,attr"`
	Channel string `xml:"channel,attr"`
	Raw     string `xml:",innerxml"`
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

// join xmltv files without duplicates
func JoinXmlTvs(inputs ...xmltv) (output xmltv, err error) {
	// check if already exists in output
	inputsMap := make(map[string]int)
	for i, input := range inputs {
		for _, channel := range input.ChannelList {
			if _, ok := inputsMap[channel.Id]; !ok {
				inputsMap[channel.Id] = i
				output.ChannelList = append(output.ChannelList, channel)
			}
		}
	}

	// join programmes based on inputsMap id
	for i, input := range inputs {
		for _, programme := range input.ProgrammeList {
			if j, ok := inputsMap[programme.Channel]; ok && j == i {
				output.ProgrammeList = append(output.ProgrammeList, programme)
			}
		}
	}

	return
}

func DownloadXmlTvByEpgSoruce(sources []EpgSource) (map[string][]*os.File, error) {
	epgs := make(map[string][]*os.File)
	for _, source := range sources {
		file, err := os.CreateTemp("./tmp/", fmt.Sprintf("xmltv-%s-*.xml", source.Provider))
		if err != nil {
			return nil, err
		}

		log.Printf("Downloading %s from %s\n", filepath.Base(file.Name()), source.URL)

		epgs[source.Provider] = append(epgs[source.Provider], file)

		// download file
		if err := downloadFile(file, source.URL); err != nil {
			for _, f := range epgs {
				for _, ff := range f {
					ff.Close()
				}
			}
			return nil, err
		}
	}

	return epgs, nil
}

func CreateXmlTvByBuckets(epg []EpgSource, buckets map[string][]Channel, outPath string) error {
	// create epg buckets
	epgBuckets := make(map[string]map[string][]string) // bucket -> provider -> channels
	epgProvides := make(map[string]struct{})           // providers
	for bucket, channels := range buckets {
		epgBuckets[bucket] = make(map[string][]string)
		for _, channel := range channels {
			if len(channel.Epg) > 0 {
				epg := channel.Epg[0] // we expect only one epg source per channel
				epgBuckets[bucket][epg.Provider] = append(epgBuckets[bucket][epg.Provider], epg.ID)
				epgProvides[epg.Provider] = struct{}{}
			}
		}
	}

	// get only used providers
	var epgSources []EpgSource
	for _, source := range epg {
		if _, ok := epgProvides[source.Provider]; ok {
			epgSources = append(epgSources, source)
		}
	}

	// download xmltv files
	files, err := DownloadXmlTvByEpgSoruce(epgSources)
	if err != nil {
		return err
	}

	// close all files
	defer func() {
		for _, f := range files {
			for _, ff := range f {
				ff.Close()
			}
		}
	}()

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
			// get file
			epgFiles := files[provider]

			for _, file := range epgFiles {
				_, err = file.Seek(0, io.SeekStart)
				if err != nil {
					return err
				}

				data, err := io.ReadAll(file)
				if err != nil {
					return err
				}

				// unmarshal xmltv
				xmltvFile, err := UnmarshalXmlTv(data)
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
