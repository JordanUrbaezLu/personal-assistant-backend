# =========================================================
# 🧱 Stage 1: Build Go binary
# =========================================================
FROM golang:1.25 AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN go build -o personal-assistant-backend ./main.go

# =========================================================
# 🚀 Stage 2: Runtime image (Debian slim)
# =========================================================
FROM debian:bookworm-slim

# ✅ Install CA certificates to fix TLS (for HTTPS APIs)
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/personal-assistant-backend .

# Expose app port
EXPOSE 8080

# Run the server
CMD ["./personal-assistant-backend"]
