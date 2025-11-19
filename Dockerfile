# Build stage
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum (if exists)
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY server.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server server.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for any HTTPS calls
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/server .

# Expose port 1234 for RPC server
EXPOSE 1234

# Run the server
CMD ["./server"]
