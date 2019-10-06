FROM golang:1.12 as builder
WORKDIR /go/src/github.com/ddouglas/monocle
COPY . /go/src/github.com/ddouglas/monocle
WORKDIR /go/src/github.com/ddouglas/monocle/cmd/cli
RUN mkdir /bin/monocle && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/monocle/cli && \
    mv ../../env.json /bin/monocle

WORKDIR /bin/monocle


LABEL maintainer="David Douglas <david@onetwentyseven.dev>"
