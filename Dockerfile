# Phase 1: Build the Go application
FROM golang:alpine AS builder

WORKDIR /app

# Copy dependency files and download modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application files
COPY . .

# Build the executable
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Phase 2: Create a lightweight image for runtime
FROM alpine:latest

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/server .

# Copy static web assets (HTML, CSS, JS, Images, Slip Templates)
COPY --from=builder /app/web ./web

# Expose port (Railway overrides this with its dynamic PORT env var)
EXPOSE 8080

# Run the app
CMD ["./server"]
