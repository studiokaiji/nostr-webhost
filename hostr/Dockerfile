# Build stage
FROM golang:1.25-alpine AS builder

# Accept GOTOOLCHAIN as build argument
ARG GOTOOLCHAIN=auto
ENV GOTOOLCHAIN=${GOTOOLCHAIN}

# Install necessary build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hostr .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS support and netcat for connectivity checks
RUN apk --no-cache add ca-certificates netcat-openbsd

# Create a non-root user
RUN addgroup -g 1000 hostr && \
    adduser -D -u 1000 -G hostr hostr

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/hostr .

# Create directories for data and config
RUN mkdir -p /app/data /home/hostr/.nostr-webhost && \
    chown -R hostr:hostr /app /home/hostr

# Switch to non-root user
USER hostr

# Expose the default port
EXPOSE 3000

