FROM golang:1.22.2-bookworm

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
