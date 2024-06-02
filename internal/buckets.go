package internal

type Bucket struct {
	Id       string
	Formats  []string
	Channels []BucketChannel
}

type BucketChannel struct {
	Channel
	Streams []Stream
	Epgs    []BucketEpg
}

type BucketEpg struct {
	SourceId  string
	ChannelId string
}

func ConvertToBuckets(config Config) (buckets []Bucket) {
	for _, pkg := range config.Packages {
		var channels []BucketChannel
		for _, channel := range config.Channels {
			if len(pkg.Channels) > 0 && !ArraysIntersect(pkg.Channels, channel.Labels) {
				continue
			}

			var streams []Stream
			for _, stream := range config.Streams {
				if len(pkg.Streams) > 0 && !ArraysIntersect(pkg.Streams, stream.Labels) {
					continue
				}
				streams = append(streams, stream)
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

			channels = append(channels, BucketChannel{
				Channel: channel,
				Streams: streams,
				Epgs:    epgs,
			})
		}

		buckets = append(buckets, Bucket{
			Id:       pkg.Id,
			Formats:  pkg.Formats,
			Channels: channels,
		})
	}

	return buckets
}
