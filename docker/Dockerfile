FROM golang

MAINTAINER Knut Ahlers <knut@ahlers.me>

RUN go get github.com/Luzifer/mondash && \
    go install github.com/Luzifer/mondash

EXPOSE 3000

ENTRYPOINT ["/go/bin/mondash"]
