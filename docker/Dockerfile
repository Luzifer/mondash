FROM golang

MAINTAINER Knut Ahlers <knut@ahlers.me>

RUN go get github.com/Luzifer/mondash && \
    go install github.com/Luzifer/mondash

EXPOSE 3000

WORKDIR /go/src/github.com/Luzifer/mondash
ENTRYPOINT ["/go/bin/mondash"]
