module github.com/Luzifer/mondash/cmd/mondash-nagios

go 1.15

replace github.com/Luzifer/mondash/client => ../../client

require (
	github.com/Luzifer/mondash/client v0.0.0-20201018014217-9635a0446be0
	github.com/Luzifer/rconfig/v2 v2.2.1
	github.com/gosimple/slug v1.9.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.3
	github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/objx v0.1.1 // indirect
	golang.org/x/sys v0.0.0-20201017003518-b09fb700fbb7
	gopkg.in/validator.v2 v2.0.0-20200605151824-2b28d334fa05
	gopkg.in/yaml.v2 v2.3.0
)
