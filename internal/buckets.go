package internal

import (
	"path"
)

type Bucket struct {
	Id       string
	Name     string
	Formats  []string
	Channels []BucketChannel
}

type BucketChannel struct {
	Channel
	Stream Stream
	Epgs   []BucketEpg
}

type BucketEpg struct {
	SourceId  string
	ChannelId string
}

func ConvertToBuckets(config Config) (buckets []Bucket) {
	for _, pkg := range config.Packages {
		var chs []Channel
		if len(pkg.Channels) > 0 {
			// if the package has explicitly listed channels, only include those (in the order they are listed in the package)
			availableChannels := make(map[string]Channel) /* map[id]channel */
			for _, channel := range config.Channels {
				availableChannels[channel.Id] = channel
			}
			for _, channelId := range pkg.Channels {
				if channel, ok := availableChannels[channelId]; ok {
					chs = append(chs, channel)
				}
			}
		} else if len(pkg.ChannelLabels) > 0 {
			// allowed labels
			allowedLabels := make(map[string]bool) /* map[label]bool */
			for _, label := range pkg.ChannelLabels {
				allowedLabels[label] = true
			}
			// if the package has channel labels, only include channels that have at least one of those labels
			processedChannels := make(map[string]bool)      /* map[id]bool */
			availableChannels := make(map[string][]Channel) /* map[label]channels */
			for _, channel := range config.Channels {
				for _, label := range channel.Labels {
					availableChannels[label] = append(availableChannels[label], channel)
				}
			}
			for _, label := range pkg.ChannelLabels {
				if ok, _ := ArrayIn(label, pkg.ChannelLabels); !ok {
					continue
				}
				if channels, ok := availableChannels[label]; ok {
					for _, channel := range channels {
						if _, ok := processedChannels[channel.Id]; !ok {
							// ensure channel has only allowed labels
							labels := channel.Labels
							newLabels := make([]string, 0)
							for _, label := range labels {
								if _, ok := allowedLabels[label]; ok {
									newLabels = append(newLabels, label)
								}
							}
							channel.Labels = newLabels
							chs = append(chs, channel)
							processedChannels[channel.Id] = true
						}
					}
				}
			}
		} else {
			// if the package has no channel labels, include all channels
			chs = config.Channels
		}

		var chStreamMap = make(map[string][]Stream) /* map[channelId]streams */
		for _, stream := range config.Streams {
			chStreamMap[stream.Channel] = append(chStreamMap[stream.Channel], stream)
		}

		var channels []BucketChannel
		for _, channel := range chs {
			streams, ok := chStreamMap[channel.Id]
			if !ok {
				// skip channels without any streams
				continue
			}
			if len(pkg.ExcludedChannelLabels) > 0 && ArraysIntersect(pkg.ExcludedChannelLabels, channel.Labels) {
				// skip channels with excluded labels
				continue
			}

			var selectedStream Stream
			hasStream := false
			for _, stream := range streams {
				if len(pkg.Streams) > 0 && !ArraysIntersect(pkg.Streams, stream.Labels) {
					continue
				}

				// only apply provider if it is set
				if len(pkg.Providers) > 0 && stream.Provider != "" {
					// skip if it's not the wanted provider
					basePath, ok := pkg.Providers[stream.Provider]
					if !ok {
						continue
					}

					// replace the URL with the provider's base path
					stream.URL = basePath + stream.URL
					if stream.Catchup != nil && stream.Catchup.Mode == "default" {
						// deep copy to avoid modifying the original
						catchup := *stream.Catchup
						catchup.Source = basePath + catchup.Source
						stream.Catchup = &catchup
					}
				}
				selectedStream = stream
				hasStream = true
			}

			if !hasStream {
				// skip channels without selected streams
				continue
			}

			var epgs []BucketEpg
			for _, epg := range config.Epg {
				if len(pkg.Epg) > 0 && !ArraysIntersect(pkg.Epg, epg.Labels) {
					continue
				}

				for epgChId, channelId := range epg.ChannelsMap {
					if channelId != channel.Id {
						continue
					}

					epgs = append(epgs, BucketEpg{
						SourceId:  epg.Id,
						ChannelId: epgChId,
					})
				}
			}

			if pkg.Logos != "" {
				channel.Logo = pkg.Logos + path.Base(channel.Logo)
			}

			channels = append(channels, BucketChannel{
				Channel: channel,
				Stream:  selectedStream,
				Epgs:    epgs,
			})
		}

		buckets = append(buckets, Bucket{
			Id:       pkg.Id,
			Name:     pkg.Name,
			Formats:  pkg.Formats,
			Channels: channels,
		})
	}

	return buckets
}
