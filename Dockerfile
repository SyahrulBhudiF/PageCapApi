FROM golang:alpine AS builder
WORKDIR /app

COPY go.mod go.sum .env ./
RUN go mod download && go mod verify

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o page-capture ./cmd/main.go

FROM debian:bullseye-slim
WORKDIR /app

COPY --from=builder /app/page-capture .

RUN apt-get update && apt-get install -y \
    ca-certificates \
    wget \
    htop \
    chromium \
    --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

ENV ROD_BROWSER_PATH=/usr/bin/chromium

EXPOSE 8080

CMD ["./page-capture"]