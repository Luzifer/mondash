package config // import "github.com/Luzifer/mondash/config"

import (
	"os"

	"github.com/spf13/pflag"
)

type Config struct {
	Storage  string
	BaseURL  string
	APIToken string

	S3 struct {
		Bucket string
	}
}

func Load() *Config {
	cfg := &Config{}
	pflag.StringVar(&cfg.Storage, "storage", "s3", "Storage engine to use")
	pflag.StringVar(&cfg.BaseURL, "baseurl", os.Getenv("BASE_URL"), "The Base-URL the application is running on for example https://mondash.org")
	pflag.StringVar(&cfg.APIToken, "api-token", os.Getenv("API_TOKEN"), "API Token used for the /welcome dashboard (you can choose your own)")

	// S3
	pflag.StringVar(&cfg.S3.Bucket, "s3Bucket", os.Getenv("S3Bucket"), "Bucket to use for S3 storage")

	pflag.Parse()
	return cfg
}
