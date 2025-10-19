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

# Set environment variables with defaults
ENV DB_HOST=103.13.30.32
ENV DB_PORT=5434
ENV DB_USER=chatbot
ENV DB_PASSWORD=chatbot123
ENV DB_NAME=testvector
ENV OLLAMA_HOST=http://host.docker.internal:11434
ENV OLLAMA_MODEL=bge-m3:latest
ENV PORT=8080

# Expose port for API
EXPOSE 8080

# Run the API server
CMD ["/app/api-server"]
