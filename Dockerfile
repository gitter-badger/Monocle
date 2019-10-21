FROM golang:1.12
WORKDIR /home/monocle/config
COPY /config /home/monocle/config
WORKDIR /home/monocle/src
COPY . /home/monocle/src
RUN rm -rf /home/monocle/src/config
RUN rm docker-rebuild.sh
WORKDIR /home/monocle/src/cmd/cli
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 MOD=vendor go build -o /home/monocle/bin/cli
WORKDIR /home/monocle/bin

LABEL maintainer="David Douglas <david@onetwentyseven.dev>"