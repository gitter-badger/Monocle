FROM golang:1.13.3 as builder
WORKDIR /app
COPY . .
WORKDIR /app/cmd/cli
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build


FROM alpine:latest AS release
WORKDIR /app

RUN apk --no-cache add tzdata
RUN apk --no-cache add ca-certificates=20190108-r0


COPY --from=builder /app/cmd/cli .

RUN adduser -S monocle
USER monocle

LABEL maintainer="David Douglas <david@onetwentyseven.dev>"
