package config // import "github.com/Luzifer/mondash/config"

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

var storageDrivers = []string{"s3", "file"}

// Config is a storage struct for configuration parameters
type Config struct {
	Storage  string
	BaseURL  string
	APIToken string

	Listen string

	S3 struct {
		Bucket string
	}

	FileStorage struct {
		Directory string
	}
}

// Load parses arguments / ENV variable to load configuration
func Load() *Config {
	cfg := &Config{}
	pflag.StringVar(&cfg.Storage, "storage", "s3", fmt.Sprintf("Storage engine to use (%s)", strings.Join(storageDrivers, ", ")))
	pflag.StringVar(&cfg.BaseURL, "baseurl", os.Getenv("BASE_URL"), "The Base-URL the application is running on for example https://mondash.org")
	pflag.StringVar(&cfg.APIToken, "api-token", os.Getenv("API_TOKEN"), "API Token used for the /welcome dashboard (you can choose your own)")
	pflag.StringVar(&cfg.Listen, "listen", ":3000", "Address to listen on")

	// S3
	pflag.StringVar(&cfg.S3.Bucket, "s3Bucket", os.Getenv("S3Bucket"), "Bucket to use for S3 storage")

	// FileStorage
	pflag.StringVar(&cfg.FileStorage.Directory, "fileDirectory", "./", "Directory to use for plain text storage")

	pflag.Parse()
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
