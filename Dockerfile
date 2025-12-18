# 1️⃣ Builder
FROM golang:1.24-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git

ENV GOPATH=/go
ENV PATH=$PATH:$GOPATH/bin

# Install goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy modules and download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build app
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# 2️⃣ Runtime
FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache tzdata ca-certificates bash

# Copy goose binary
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy server, migrations, env
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/.env .env

EXPOSE 3000
CMD ["./server"]