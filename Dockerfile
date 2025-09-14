# Stage 1: build
FROM golang:1.25 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o personal-assistant-backend ./main.go

# Stage 2: run
FROM debian:bookworm-slim

WORKDIR /root/
COPY --from=builder /app/personal-assistant-backend .

EXPOSE 8080
CMD ["./personal-assistant-backend"]
