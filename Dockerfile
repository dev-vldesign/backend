# Use the official Golang image as a builder
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates for go mod download
RUN apk add --no-cache git ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go app with static linking and optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o main .

# Use Alpine for the final image
FROM alpine:3.19

# Install only ca-certificates
RUN apk add --no-cache ca-certificates

# Create a non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set the working directory
WORKDIR /app

# Copy the built binary
COPY --from=builder /app/main /app/main

# Change ownership to the app user
RUN chown -R appuser:appgroup /app

# Switch to the app user
USER appuser

# Expose the application port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Command to run the application
CMD ["/app/main"]