FROM golang:alpine as builder

ADD . /go/src/github.com/Luzifer/mondash
WORKDIR /go/src/github.com/Luzifer/mondash

RUN set -ex \
 && apk add --update git \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)"

FROM alpine:latest

LABEL maintainer "Knut Ahlers <knut@ahlers.me>"

RUN set -ex \
 && apk --no-cache add ca-certificates

COPY --from=builder /go/bin/mondash /usr/local/bin/mondash
COPY --from=builder /go/src/github.com/Luzifer/mondash/templates /usr/local/share/mondash/templates

WORKDIR /usr/local/share/mondash
EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/mondash"]
CMD ["--"]

# vim: set ft=Dockerfile:
