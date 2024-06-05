package internal

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Url      string      `yaml:"url"`
	Epg      []EpgSource `yaml:"epg"`
	Channels []Channel   `yaml:"channels"`
	Streams  []Stream    `yaml:"streams"`
	Packages []Package   `yaml:"packages"`
}

type EpgSource struct {
	Id          string            `yaml:"id"`
	URL         string            `yaml:"url"`
	Labels      []string          `yaml:"labels,omitempty"`
	ChannelsMap map[string]string `yaml:"channels_map,omitempty"`
}

type Channel struct {
	Id     string   `yaml:"id"`
	Name   string   `yaml:"name"`
	Logo   string   `yaml:"logo"`
	Labels []string `yaml:"labels,omitempty"`
}

type Catchup struct {
	Mode   string `yaml:"mode"`
	Source string `yaml:"source"`
	Days   int    `yaml:"days"`
}

type Stream struct {
	Channel  string   `yaml:"channel"`
	URL      string   `yaml:"url"`
	Provider string   `yaml:"provider"`
	Labels   []string `yaml:"labels,omitempty"`
	Catchup  *Catchup `yaml:"catchup,omitempty"`
}

type Package struct {
	Id      string   `yaml:"id"`
	Name    string   `yaml:"name"`
	Formats []string `yaml:"formats"`
	// labels
	Channels      []string          `yaml:"channels,omitempty"`
	ChannelLabels []string          `yaml:"channel_labels,omitempty"`
	Streams       []string          `yaml:"streams,omitempty"`
	Epg           []string          `yaml:"epg,omitempty"`
	Logos         string            `yaml:"logos,omitempty"`
	Providers     map[string]string `yaml:"providers"` // ID -> URL
}

func (c *Config) Load(file string) error {
	// if provided file is directory, then load all files in it and merge them
	if fi, err := os.Stat(file); err == nil && fi.IsDir() {
		log.Printf("Loading config from directory: %s", file)

		files, err := os.ReadDir(file)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}

			filePath := filepath.Join(file, f.Name())

			if filepath.Ext(filePath) != ".yaml" && filepath.Ext(filePath) != ".yml" {
				continue
			}

			if err := c.load(filePath); err != nil {
				return err
			}
		}
	} else {
		if err := c.load(file); err != nil {
			return err
		}
	}

	log.Printf("Config loaded successfully: %d epg, %d channels, %d streams, %d packages", len(c.Epg), len(c.Channels), len(c.Streams), len(c.Packages))

	return nil
}

func (c *Config) load(file string) error {
	log.Printf("Loading config from file: %s", file)

	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal %q: %w", file, err)
	}

	if config.Url != "" {
		c.Url = config.Url
	}

	c.Epg = append(c.Epg, config.Epg...)
	c.Channels = append(c.Channels, config.Channels...)
	c.Streams = append(c.Streams, config.Streams...)
	c.Packages = append(c.Packages, config.Packages...)

	return nil
}

func (c Config) Save(file string) error {
	// if provided file is directory, then we can't save config and return error
	if fi, err := os.Stat(file); err == nil && fi.IsDir() {
		return fmt.Errorf("provided file is directory")
	}

	return c.save(file)
}

func (c Config) save(file string) error {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return err
	}

	return os.WriteFile(file, buf.Bytes(), 0644)
}
