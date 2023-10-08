FROM golang:1.16-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY logger/ ./logger
RUN go mod download

COPY ./producer .

RUN go build -o producer .

ENTRYPOINT [ "./producer" ]
