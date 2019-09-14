FROM golang:1.12 as builder
WORKDIR /go/src/github.com/ddouglas/monocle
COPY . .
WORKDIR cmd/cli
RUN GO111MODULES=active CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

LABEL maintainer="David Douglas <david@onetwentyseven.dev>"