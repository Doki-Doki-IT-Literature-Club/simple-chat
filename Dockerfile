FROM golang:1.22.2-alpine AS dep-downloader

WORKDIR /simple-chat

COPY go.mod go.sum ./

RUN go mod download

FROM golang:1.22.2-alpine AS builder

WORKDIR /simple-chat

COPY . .

COPY --from=dep-downloader /go/pkg/mod /go/pkg/mod

RUN go build -o simple-chat

FROM alpine

WORKDIR /simple-chat

RUN apk add --no-cache curl

COPY --from=builder /simple-chat/simple-chat .

CMD ["./simple-chat"]