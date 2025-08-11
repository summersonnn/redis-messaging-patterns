# ---------- Build Stage ----------
FROM golang:1.21-alpine AS builder

# Install git (needed for `go mod download`)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o subscriber main.go

# ---------- Final Stage ----------
FROM alpine:latest

# Install CA certificates for HTTPS support
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/subscriber .

# Set environment variables with sensible defaults
ENV REDIS_HOST=redis \
    REDIS_PORT=6379 \
    PATTERN=both

# Define the entrypoint
CMD ["./subscriber"]
