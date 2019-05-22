package config

import (
	"log"
	"net/url"

	"github.com/Luzifer/rconfig"
)

var storageDrivers = []string{"s3", "file"}

// Config is a storage struct for configuration parameters
type Config struct {
	Storage  string `flag:"storage" default:"s3" description:"Storage engine to use"`
	BaseURL  string `flag:"baseurl" env:"BASE_URL" description:"The Base-URL the application is running on for example https://mondash.org"`
	APIToken string `flag:"api-token" env:"API_TOKEN" description:"API Token used for the /welcome dashboard (you can choose your own)"`

	Listen string `flag:"listen" default:":3000" description:"Address to listen on"`

	S3 struct {
		Bucket string `flag:"s3Bucket" env:"S3Bucket" description:"Bucket to use for S3 storage"`
	}

	FileStorage struct {
		Directory string `flag:"fileDirectory" default:"./" description:"Directory to use for plain text storage"`
	}
}

// Load parses arguments / ENV variable to load configuration
func Load() *Config {
	cfg := &Config{}
	rconfig.Parse(cfg)
	return cfg
}

func (c Config) isValid() bool {
	// Storage Driver check
	validStoragedriver := false
	for _, d := range storageDrivers {
		if c.Storage == d {
			validStoragedriver = true
			break
		}
	}
	if !validStoragedriver {
		log.Printf("You specified a wrong storage driver: %s\n\n", c.Storage)
		return false
	}

	// Minimum characters of API token
	if len(c.APIToken) < 10 {
		log.Printf("You need to specify an api-token with more than 9 characters.\n\n")
		return false
	}

	// Base-URL check
	if _, err := url.Parse(c.BaseURL); err != nil {
		log.Printf("The baseurl '%s' does not look like a valid URL: %s.\n\n", c.BaseURL, err)
		return false
	}

	return true
}
