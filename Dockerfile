# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY *.go ./

# Build static binary
RUN CGO_ENABLED=0 go build -o runner .

# Runtime stage
FROM bash:4.4

COPY --from=builder /app/runner /runner

ENTRYPOINT ["/runner"]
