# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the API server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-server .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/api-server /app/api-server

# Copy doc folder for document processing
COPY doc/ /app/doc/

# Copy .env file if exists
COPY .env* ./

# Note: Environment variables are loaded from .env file at runtime
# Defaults are handled in config.go

# Expose port for API
EXPOSE 8080

# Run the API server
CMD ["/app/api-server"]
