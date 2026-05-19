# Stage 1: Build Go binary
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy dependency files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY configs/ ./configs/

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /app/oxmail-api ./cmd/oxmail-api

# Stage 2: Minimal runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates curl

WORKDIR /app

COPY --from=builder /app/oxmail-api .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./oxmail-api"]
