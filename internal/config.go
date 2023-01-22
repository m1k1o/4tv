package internal

import (
	"bytes"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig(file string) (Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func SaveConfig(file string, config Config) error {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(config); err != nil {
		return err
	}

	return os.WriteFile(file, buf.Bytes(), 0644)
}
