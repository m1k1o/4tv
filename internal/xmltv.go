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
	XMLName       xml.Name       `xml:"tv"`
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
	data, err := xml.MarshalIndent(x, "", "  ")
	if err != nil {
		return nil, err
	}

	return []byte("<?xml version=\"1.0\" encoding=\"UTF-8\" ?>\n" + string(data)), nil
}

func FilterXmlTvByChannels(input xmltv, channels map[string]string /* channel_id -> epg_channel_id */) (output xmltv, err error) {
	// filter channels list
	for _, v := range input.ChannelList {
		for chid, epg_chid := range channels {
			if v.Id == epg_chid {
				// add channel to output list if it is in the channels list
				output.ChannelList = append(output.ChannelList, xmlchannel{
					Id:  chid, // use channel id instead of epg channel id
					Raw: v.Raw,
				})
				break
			}
		}
	}

	// filter programme list
	for _, v := range input.ProgrammeList {
		for chid, epg_chid := range channels {
			if v.Channel == epg_chid {
				// add channel to output list if it is in the channels list
				output.ProgrammeList = append(output.ProgrammeList, xmlprogramme{
					Start:   v.Start,
					Stop:    v.Stop,
					Channel: chid, // use channel id instead of epg channel id
					Raw:     v.Raw,
				})
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
		file, err := os.CreateTemp("", fmt.Sprintf("xmltv-%s-*.xml", source.Id))
		if err != nil {
			return nil, err
		}

		log.Printf("Downloading %s from %s\n", filepath.Base(file.Name()), source.URL)

		epgs[source.Id] = append(epgs[source.Id], file)

		// download file
		if err := downloadFile(file, source.URL); err != nil {
			for _, f := range epgs {
				for _, ff := range f {
					ff.Close()
					os.Remove(ff.Name())
				}
			}
			return nil, err
		}
	}

	return epgs, nil
}

func CreateXmlTvByBuckets(epg []EpgSource, buckets []Bucket, outPath string) error {
	// create epg buckets
	epgBuckets := make(map[string]map[string]map[string]string) // bucket_ID -> epg_source_ID -> channel_ID -> epg_channel_ID}
	epgProvides := make(map[string]struct{})                    // epg_source_ID
	for _, b := range buckets {
		epgBuckets[b.Id] = make(map[string]map[string]string)
		for _, channel := range b.Channels {
			for _, epg := range channel.Epgs {
				if _, ok := epgBuckets[b.Id][epg.SourceId]; !ok {
					epgBuckets[b.Id][epg.SourceId] = make(map[string]string)
				}
				epgBuckets[b.Id][epg.SourceId][channel.Channel.Id] = epg.ChannelId
				epgProvides[epg.SourceId] = struct{}{}
			}
		}
	}

	// get only used providers
	var epgSources []EpgSource
	for _, source := range epg {
		if _, ok := epgProvides[source.Id]; ok {
			epgSources = append(epgSources, source)
		}
	}

	// download xmltv files
	files, err := DownloadXmlTvByEpgSoruce(epgSources)
	if err != nil {
		return err
	}

	// close and remove all files
	defer func() {
		for _, f := range files {
			for _, ff := range f {
				ff.Close()
				os.Remove(ff.Name())
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
