package config // import "github.com/Luzifer/mondash/config"

import (
	"os"

	"github.com/spf13/pflag"
)

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
	pflag.StringVar(&cfg.Storage, "storage", "s3", "Storage engine to use (s3, file)")
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
