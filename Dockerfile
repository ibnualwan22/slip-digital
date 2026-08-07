# Stage 1: Build Frontend (Vite React)
FROM node:22-alpine AS node-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go Backend
FROM golang:alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o shift ./cmd/maintenance/shift.go

# Stage 3: Lightweight Runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Jakarta
WORKDIR /root/

# Copy Go binary
COPY --from=go-builder /app/server .
COPY --from=go-builder /app/shift .

# Copy built frontend assets
COPY --from=node-builder /app/web/dist ./web/dist

# Copy public static assets (images for slip generator etc.)
COPY web/public ./web/public

EXPOSE 8080

CMD ["./server"]
