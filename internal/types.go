package internal

type Config struct {
	Epg      []EpgSource `yaml:"epg"`
	Channels []Channel   `yaml:"channels"`
	Buckets  []Bucket    `yaml:"buckets"`
}

type EpgSource struct {
	Provider string `yaml:"provider"`
	URL      string `yaml:"url"`
}

type Catchup struct {
	Mode   string `yaml:"mode"`
	Source string `yaml:"source"`
	Days   int    `yaml:"days"`
}

type Stream struct {
	URL     string   `yaml:"url"`
	Labels  []string `yaml:"labels,omitempty"`
	Catchup Catchup  `yaml:"catchup,omitempty"`
}

type ChannelEpg struct {
	Provider string   `yaml:"provider"`
	ID       string   `yaml:"id"`
	Labels   []string `yaml:"labels,omitempty"`
}

type Channel struct {
	Name    string       `yaml:"name"`
	Logo    string       `yaml:"logo"`
	Labels  []string     `yaml:"labels,omitempty"`
	Streams []Stream     `yaml:"streams"`
	Epg     []ChannelEpg `yaml:"epg"`
}

type Bucket struct {
	Name    string   `yaml:"name"`
	Formats []string `yaml:"formats"`
	// labels
	Channels []string `yaml:"channels,omitempty"`
	Streams  []string `yaml:"streams,omitempty"`
	Epg      []string `yaml:"epg,omitempty"`
}

func GetChannelsByBucket(channels []Channel, bucket Bucket) []Channel {
	var returnChannels []Channel

	for _, channel := range channels {
		// match channel labels
		if len(bucket.Channels) > 0 {
			var match bool
			for _, label := range bucket.Channels {
				if ok, _ := ArrayIn(label, channel.Labels); ok {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		// match channel streams
		if len(bucket.Streams) > 0 {
			var match *Stream
			for _, stream := range channel.Streams {
				// check if stream labels match
				var matchLabels bool
				for _, label := range bucket.Streams {
					if ok, _ := ArrayIn(label, stream.Labels); ok {
						matchLabels = true
						break
					}
				}
				if !matchLabels {
					continue
				}

				match = &stream
				break
			}

			// if no streams match, skip the channel
			if match == nil {
				continue
			}

			// only use the matching stream
			channel.Streams = []Stream{
				*match,
			}
		} else {
			// if no streams labels are specified, only use the first stream
			if len(channel.Streams) > 1 {
				channel.Streams = []Stream{
					channel.Streams[0],
				}
			} else if len(channel.Streams) == 0 {
				// if no streams are specified, skip the channel
				continue
			}
		}

		// match channel epgs
		if len(bucket.Epg) > 0 {
			var match *ChannelEpg
			for _, channelEpg := range channel.Epg {
				// check if stream labels match
				var matchLabels bool
				for _, label := range bucket.Epg {
					if ok, _ := ArrayIn(label, channelEpg.Labels); ok {
						matchLabels = true
						break
					}
				}
				if !matchLabels {
					continue
				}

				match = &channelEpg
				break
			}

			// if no epgs match, use empty epg
			if match != nil {
				channel.Epg = []ChannelEpg{}
			} else {
				// only use the matching epg
				channel.Epg = []ChannelEpg{
					*match,
				}
			}
		} else if len(channel.Epg) > 1 {
			// if no epgs labels are specified, only use the first stream
			channel.Epg = []ChannelEpg{
				channel.Epg[0],
			}
		}

		returnChannels = append(returnChannels, channel)
	}

	return returnChannels
}
