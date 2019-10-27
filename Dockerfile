FROM alpine:latest AS release
WORKDIR /app
COPY cmd/cli/cli .

RUN apk --no-cache add tzdata
RUN apk --no-cache add ca-certificates=20190108-r0

LABEL maintainer="David Douglas <david@onetwentyseven.dev>"