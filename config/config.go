package config

import (
	"net/url"

	"github.com/pkg/errors"

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
func Load() (*Config, error) {
	cfg := &Config{}

	if err := rconfig.Parse(cfg); err != nil {
		return nil, errors.Wrap(err, "Unable to parse CLI config")
	}

	if err := cfg.isValid(); err != nil {
		return nil, errors.Wrap(err, "CLI config does not validate")
	}

	return cfg, nil
}

func (c Config) isValid() error {
	// Storage Driver check
	validStoragedriver := false
	for _, d := range storageDrivers {
		if c.Storage == d {
			validStoragedriver = true
			break
		}
	}
	if !validStoragedriver {
		return errors.Errorf("Storage driver %q is unknown", c.Storage)
	}

	// Minimum characters of API token
	if len(c.APIToken) < 10 {
		return errors.New("API-Token needs to have at least 10 characters")
	}

	// Base-URL check
	if _, err := url.Parse(c.BaseURL); err != nil {
		return errors.Wrapf(err, "The baseurl %q is not a valid URL", c.BaseURL)
	}

	return nil
}
