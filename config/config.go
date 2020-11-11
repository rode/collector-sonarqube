package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// Configuration creates a Global Conifguration to be passed throughout packages
var Configuration *Config

// Config struct for collector config
type Config struct {
	RodeAPI struct {
		Address string `yaml:"address"`
	} `yaml:"rode-api"`
}

// NewConfig creates a config structure
func NewConfig(configPath string) (*Config, error) {
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func init() {
	var err error
	Configuration, err = NewConfig("./config.yaml")
	if err != nil {
		log.Fatal(err)
	}
}
