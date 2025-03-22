FROM golang:1.23-bookworm AS base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o smarthome ./cmd/main.go

EXPOSE 8000

CMD ["./smarthome"]

