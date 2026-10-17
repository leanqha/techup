# 1️⃣ Builder
FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git

ENV GOPATH=/go
ENV PATH=$PATH:$GOPATH/bin

# Copy modules and download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build app
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Build goose (SQL migration tool) so the migrate service can run migrations.
# Pinned to a release that builds with this Go toolchain (newer goose needs Go 1.26+).
RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.24.3

# 2️⃣ Runtime
FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache tzdata ca-certificates bash

# Copy binary only; runtime secrets are injected via environment variables.
COPY --from=builder /app/server .
# goose binary + migrations, used by the `migrate` compose service.
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations ./migrations

EXPOSE 3000
CMD ["./server"]