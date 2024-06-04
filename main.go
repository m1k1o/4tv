package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

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

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check config.",
	Run: func(cmd *cobra.Command, args []string) {
		config := internal.Config{}
		err := config.Load(configFile)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("---")

		fmt.Printf("Found %d channels\n", len(config.Channels))
		var channels = map[string]bool{}
		for _, channel := range config.Channels {
			if _, ok := channels[channel.Id]; ok {
				log.Fatalf("Channel with id %q already exists", channel.Id)
			}
			channels[channel.Id] = true
		}

		fmt.Println("---")

		fmt.Printf("Found %d epg sources:\n", len(config.Epg))
		var channelsWithEpg = map[string]bool{}
		for _, epg := range config.Epg {
			fmt.Printf(" |- %q with %d channels\n", epg.Id, len(epg.ChannelsMap))
			for epgId, chId := range epg.ChannelsMap {
				if _, ok := channels[chId]; !ok {
					fmt.Printf(" | |- epg %q mapped to a non-existent channel id %q\n", epgId, chId)
				}
				channelsWithEpg[chId] = true
			}
		}
		noEpgChannels := []string{}
		for chId := range channels {
			if _, ok := channelsWithEpg[chId]; !ok {
				noEpgChannels = append(noEpgChannels, chId)
			}
		}
		if len(noEpgChannels) > 0 {
			fmt.Printf("There are %d channels without epg: \n", len(noEpgChannels))
			sort.Strings(noEpgChannels)
			for _, chId := range noEpgChannels {
				fmt.Printf(" |- %q\n", chId)
			}
		}

		fmt.Println("---")

		fmt.Printf("Found %d streams\n", len(config.Streams))
		var channelsWithSources = map[string]bool{}
		for _, stream := range config.Streams {
			if _, ok := channels[stream.Channel]; !ok {
				fmt.Printf(" |- stream %q mapped to a non-existent channel id %q\n", stream.URL, stream.Channel)
			}
			channelsWithSources[stream.Channel] = true
		}
		noStreamChannels := []string{}
		for chId := range channels {
			if _, ok := channelsWithSources[chId]; !ok {
				noStreamChannels = append(noStreamChannels, chId)
			}
		}
		if len(noStreamChannels) > 0 {
			fmt.Printf("There are %d channels without stream: \n", len(noStreamChannels))
			sort.Strings(noStreamChannels)
			for _, chId := range noStreamChannels {
				fmt.Printf(" |- %q\n", chId)
			}
		}

		fmt.Println("---")

		fmt.Printf("Found %d packages\n", len(config.Packages))

		fmt.Println("---")
	},
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
	rootCmd.AddCommand(checkCmd)
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
