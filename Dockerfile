FROM golang:alpine AS builder

RUN go version

RUN apk update && apk upgrade && apk add git zlib-dev gcc musl-dev

RUN mkdir -p /tmp/compile
WORKDIR /tmp/compile

COPY . /go/src/github.com/TicketsBot/logarchiver
WORKDIR  /go/src/github.com/TicketsBot/logarchiver

RUN set -Eeux && \
    go mod download && \
    go mod verify

RUN GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -o logarchiver cmd/logarchiver/main.go

FROM alpine:latest

RUN apk update && apk upgrade

COPY --from=builder /go/src/github.com/TicketsBot/logarchiver/logarchiver /srv/logarchiver/logarchiver
RUN chmod +x /srv/logarchiver/logarchiver

COPY --from=0 /go/src/github.com/TicketsBot/logarchiver/public /srv/logarchiver/public

RUN adduser container --disabled-password --no-create-home
USER container
WORKDIR /srv/logarchiver

CMD ["/srv/logarchiver/logarchiver"]