# Build stage
FROM golang:1.25.3-alpine AS builder

# Install build tools for CGO
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary with CGO enabled
# -ldflags="-w -s" strips debug information and symbols, reducing the binary size
RUN go build -ldflags="-w -s" -o /isbetmf .

# Final stage
FROM alpine/curl:latest

WORKDIR /
COPY --from=builder /isbetmf /isbetmf
COPY www /www
COPY ./auth_policies.star /auth_policies.star
COPY --chmod=755 ./data/sqlite3_rsync /sqlite3_rsync

# Expose the port the server runs on
EXPOSE 9991

# Run the binary
ENTRYPOINT ["/isbetmf"]