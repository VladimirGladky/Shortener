FROM golang:1.25 AS builder

WORKDIR /newApp

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o shortener ./cmd/main.go

EXPOSE 4052

CMD ["./shortener"]