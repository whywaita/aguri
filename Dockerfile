FROM golang:1.13.0-buster as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /go/src/github.com/whywaita/aguri
COPY . .
RUN make build

ENTRYPOINT ["/go/src/github.com/whywaita/aguri/aguri"]
