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

# 2️⃣ Runtime
FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache tzdata ca-certificates bash

# Copy binary only; runtime secrets are injected via environment variables.
COPY --from=builder /app/server .

EXPOSE 3000
CMD ["./server"]