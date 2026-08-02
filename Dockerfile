# Stage 1: Build
FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go

# Stage 2: Runtime
FROM alpine:3.20

WORKDIR /root/

# Install ca-certificates and curl for HTTPS and healthchecks
RUN apk --no-cache add ca-certificates curl

# Copy binary from builder
COPY --from=builder /app/main .

# Expose API port
EXPOSE 8080

# Health check
RUN apk add --no-cache curl
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD curl -f http://localhost:8080/api/v1/health || exit 1

# Run the application
CMD ["./main"]
