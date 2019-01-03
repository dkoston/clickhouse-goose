FROM golang:1.11.2-alpine3.8

WORKDIR /build
COPY . /build/

RUN apk update --no-cache && \
    apk add git build-base gcc && \
    cd /build/cmd/clickhouse-goose && \
    go build .

CMD ["/bin/ash", "-c", "sleep infinity"]