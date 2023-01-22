package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"go4tv/internal"

	"github.com/spf13/cobra"
)

var (
	configFile string
	dataFolder string
)

var rootCmd = &cobra.Command{
	Use:   "4tv",
	Short: "4tv - everything fo(u)r tv",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var playlistCmd = &cobra.Command{
	Use:   "playlist",
	Short: "Generate playlists.",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := internal.LoadConfig(configFile)
		if err != nil {
			log.Fatal(err)
		}

		m3u8 := map[string][]internal.Channel{}
		enigma2 := map[string][]internal.Channel{}
		for _, bucket := range config.Buckets {
			log.Printf("bucket: %s", bucket.Name)

			channels := internal.GetChannelsByBucket(config.Channels, bucket)
			for _, format := range bucket.Formats {
				log.Printf("format: %s", format)
				switch format {
				case "hls":
					fallthrough
				case "m3u":
					fallthrough
				case "m3u8":
					m3u8[bucket.Name] = channels
				case "enigma2":
					enigma2[bucket.Name] = channels
				default:
					log.Printf("format '%s' not supported", format)
				}
			}

			log.Printf("channels: %d", len(channels))
		}

		if err := internal.CreateM3U8ByBuckets(m3u8, dataFolder); err != nil {
			log.Fatal(err)
		}

		if err := internal.CreateEnigma2ByBuckets(enigma2, dataFolder); err != nil {
			log.Fatal(err)
		}
	},
}

var epgCmd = &cobra.Command{
	Use:   "epg",
	Short: "Generate epgs.",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := internal.LoadConfig(configFile)
		if err != nil {
			log.Fatal(err)
		}

		xmltv := map[string][]internal.Channel{}
		for _, bucket := range config.Buckets {
			log.Printf("bucket: %s", bucket.Name)

			channels := internal.GetChannelsByBucket(config.Channels, bucket)
			xmltv[bucket.Name] = channels

			log.Printf("channels: %d", len(channels))
		}

		if err := internal.CreateXmlTvByBuckets(config.Epg, xmltv, dataFolder); err != nil {
			log.Fatal(err)
		}
	},
}

var logosCmd = &cobra.Command{
	Use:   "logos",
	Short: "Add logos.",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := internal.LoadConfig(configFile)
		if err != nil {
			log.Fatal(err)
		}

		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}

		logos := map[string]string{}
		if err := json.Unmarshal(data, &logos); err != nil {
			log.Fatal(err)
		}

		for i, channel := range config.Channels {
			for _, epg := range channel.Epg {
				if logo, ok := logos[epg.ID]; ok {
					channel.Logo = logo
					break
				}
			}
			config.Channels[i] = channel
		}

		if err := internal.SaveConfig(configFile, config); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(playlistCmd)
	rootCmd.AddCommand(epgCmd)
	rootCmd.AddCommand(logosCmd)

	for _, cmd := range rootCmd.Commands() {
		cmd.Flags().StringVarP(&configFile, "config", "c", "./config.yaml", "Path to config file.")
		cmd.Flags().StringVarP(&dataFolder, "data", "d", "./data/", "Path to data folder.")
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
