package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go4tv/internal"

	"github.com/spf13/cobra"
)

var (
	configFile  string
	dataFolder  string
	logosFolder string
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
		config := internal.Config{}
		err := config.Load(configFile)
		if err != nil {
			log.Fatal(err)
		}

		buckets := internal.ConvertToBuckets(config)

		m3u8 := map[string][]internal.BucketChannel{}
		enigma2 := map[string][]internal.BucketChannel{}
		for _, bucket := range buckets {
			log.Printf("bucket: %s", bucket.Id)

			for _, format := range bucket.Formats {
				log.Printf("format: %s", format)
				switch format {
				case "hls":
					fallthrough
				case "m3u":
					fallthrough
				case "m3u8":
					m3u8[bucket.Id] = bucket.Channels
				case "enigma2":
					enigma2[bucket.Id] = bucket.Channels
				default:
					log.Printf("format '%s' not supported", format)
				}
			}

			log.Printf("channels: %d", len(bucket.Channels))
		}

		if err := internal.CreateM3U8ByBuckets(config.Url, m3u8, dataFolder); err != nil {
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
		config := internal.Config{}
		err := config.Load(configFile)
		if err != nil {
			log.Fatal(err)
		}

		buckets := internal.ConvertToBuckets(config)

		if err := internal.CreateXmlTvByBuckets(config.Epg, buckets, dataFolder); err != nil {
			log.Fatal(err)
		}
	},
}

var logosCmd = &cobra.Command{
	Use:   "logos",
	Short: "Download logos.",
	Run: func(cmd *cobra.Command, args []string) {
		config := internal.Config{}
		err := config.Load(configFile)
		if err != nil {
			log.Fatal(err)
		}

		// create a map of channels that are already in the config
		channels := map[string]string{}
		for _, channel := range config.Channels {
			channels[channel.Id] = channel.Logo
		}

		// list all files in logos folder
		files, err := os.ReadDir(logosFolder)
		if err != nil {
			log.Fatal(err)
		}

		// create a map of logos that are already downloaded
		channelsWithLogos := map[string]bool{}
		for _, file := range files {
			// remove file extension (everything after the last dot)
			nameWithExtension := file.Name()
			name := nameWithExtension[:len(nameWithExtension)-len(filepath.Ext(nameWithExtension))]
			channelsWithLogos[name] = true
		}

		// download logos that are missing
		for channel, logo := range channels {
			if _, ok := channelsWithLogos[channel]; ok {
				continue
			}

			// check if its png/jpg
			ext := filepath.Ext(logo)
			ext = fmt.Sprintf(".%s", ext)
			if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".svg" {
				ext = ".png" // default to png
			}

			log.Printf("downloading logo for channel: %s from: %s", channel, logo)

			filePath := filepath.Join(logosFolder, fmt.Sprintf("%s.png", channel))
			file, err := os.Create(filePath)
			if err != nil {
				log.Fatal(err)
			}

			if err := internal.DownloadFile(file, logo); err != nil {
				file.Close()
				log.Fatal(err)
			}

			file.Close()
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
		cmd.Flags().StringVarP(&logosFolder, "logos", "l", "./logos/", "Path to store logos.")
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
