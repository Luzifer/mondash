FROM node:alpine as node_builder

COPY . /src/mondash
WORKDIR /src/mondash/src

RUN set -ex \
 && npm ci \
 && npm run build


FROM golang:alpine as builder

COPY . /go/src/github.com/Luzifer/mondash
WORKDIR /go/src/github.com/Luzifer/mondash

RUN set -ex \
 && apk add --update git \
 && go install -ldflags "-X main.version=$(git describe --tags --always || echo dev)"


FROM alpine:latest

ENV FRONTEND_DIR=/usr/local/share/mondash/frontend \
    STORAGE=file:///data

LABEL maintainer "Knut Ahlers <knut@ahlers.me>"

RUN set -ex \
 && apk --no-cache add ca-certificates

COPY --from=builder /go/bin/mondash /usr/local/bin/mondash
COPY --from=node_builder /src/mondash/frontend /usr/local/share/mondash/frontend

EXPOSE 3000
VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/mondash"]
CMD ["--"]

# vim: set ft=Dockerfile:
