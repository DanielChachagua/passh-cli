# Stage 1: Build the binary
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Compile static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o passh main.go

# Stage 2: Final runtime container
FROM alpine:3.19

# Add CA certificates for secure HTTPS API calls
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy compiled binary from builder
COPY --from=builder /app/passh /app/passh

# Define entrypoint to run CLI
ENTRYPOINT ["/app/passh"]
