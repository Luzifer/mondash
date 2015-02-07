# luzifer/mondash Dockerfile

This repository contains **Dockerfile** of [Luzifer/mondash](https://github.com/Luzifer/mondash) for [Docker](https://www.docker.com/)'s [automated build](https://registry.hub.docker.com/u/luzifer/mondash/) published to the public [Docker Hub Registry](https://registry.hub.docker.com/).

## Base Docker Image

- [golang](https://registry.hub.docker.com/_/golang/)

## Installation

1. Install [Docker](https://www.docker.com/).

2. Download [automated build](https://registry.hub.docker.com/u/luzifer/mondash/) from public [Docker Hub Registry](https://registry.hub.docker.com/): `docker pull luzifer/mondash`

## Usage

To launch it, just replace the variables in following command and start the container:

```
docker run \
         -e AWS_ACCESS_KEY_ID=myaccesskeyid \
         -e AWS_SECRET_ACCESS_KEY=mysecretaccesskey \
         -e S3Bucket=mybucketname \
         -e BASE_URL=http://www.mondash.org \
         -e API_TOKEN=yourownrandomtoken \
         -p 80:3000 \
         luzifer/mondash
```

Easy!

