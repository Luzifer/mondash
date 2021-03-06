[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/mondash)](https://goreportcard.com/report/github.com/Luzifer/mondash)
![](https://badges.fyi/github/license/Luzifer/mondash)
![](https://badges.fyi/github/downloads/Luzifer/mondash)
![](https://badges.fyi/github/latest-release/Luzifer/mondash)
![](https://knut.in/project-status/mondash)

# Luzifer / mondash

MonDash is a service for everyone having to display monitoring results to people who have not the time or knowledge to get familar with Nagios / Icinga or similar monitoring systems. Therefore MonDash provides a [simple API](http://docs.mondash.apiary.io/) to submit monitoring results and a simple dashboard to view those results.

## Hosted

There is an instance of MonDash running on [mondash.org](https://mondash.org/) you can use for free. This means you can just head over there, create your own dashboard with one click and start to push your own metrics to your dashboard within 5 minutes. No registration, no fees, just your dashboard and you.

## Installation

However maybe you want to use MonDash for data you don't like to have public visible on the internet. As it is open source you also can host your own instance: The most simple way to install your own instance is to download a binary distribution from the [releases page](https://github.com/Luzifer/mondash/releases).

Additionally you will need a template for your dashboard to be displayed correctly. You can get the template from this repository in the `templates` folder. The template will be searched in a folder called `templates` inside the current working directory.

To start MonDash you will need to make sure you configured your instance correctly:

```bash
# mondash -h
Usage of mondash:
      --api-token string      API Token used for the /welcome dashboard (you can choose your own)
      --baseurl string        The Base-URL the application is running on for example https://mondash.org
      --frontend-dir string   Directory to serve frontend assets from (default "./frontend")
      --listen string         Address to listen on (default ":3000")
      --log-level string      Set log level (debug, info, warning, error) (default "info")
      --storage string        Storage engine to use (default "file:///data")
      --version               Prints current version and exits
```

1. If you want to store the data in S3:
  - Set AWS environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`)
  - Specify `--storage=s3://<yourbucket>/[optional prefix]`
2. If you want to store the data in local file system:
  - Ensure the data directory is writable
  - Specify `--storage=file:///absolute/path/to/directory`

In all cases you need to specify `--api-token` with a token containing more than 10 characters and `--baseurl` with the base-URL of your instance.

### Docker

To launch it, just replace the variables in following command and start the container:

```
docker run \
         -e AWS_ACCESS_KEY_ID=myaccesskeyid \
         -e AWS_SECRET_ACCESS_KEY=mysecretaccesskey \
         -e AWS_REGION=eu-west-1 \
         -e STORAGE=s3://mybucketname/ \
         -e BASE_URL=http://mondash.org \
         -e API_TOKEN=yourownrandomtoken \
         -p 80:3000 \
         luzifer/mondash
```

## Security

Just some words regarding security: MonDash was designed to be an open platform for creating dashboards without any hazzle. You just open a dashboard, send some data to it and you're already done. No need to think about OAuth or other authentication mechanisms.

The downpath of that concept is of course everyone can access every dashboard and see every data placed on it. So please don't use the public instances for private and/or secret data. You can just set up your own instance within 5 minutes (okay maybe 10 minutes if you want to do it right) and you can ensure that this instance is hidden from the internet.
