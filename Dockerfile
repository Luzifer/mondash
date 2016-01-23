FROM alpine

MAINTAINER Knut Ahlers <knut@luzifer.io>

ENV GOPATH /go:/go/src/github.com/Luzifer/mondash/Godeps/_workspace
EXPOSE 3000

ADD . /go/src/github.com/Luzifer/mondash
WORKDIR /go/src/github.com/Luzifer/mondash

RUN apk --update add git go ca-certificates \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" \
 && apk del --purge go git

ENTRYPOINT ["/go/bin/mondash"]
