# Stage 1: Build Go binary
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.work go.work.sum ./
COPY go.mod go.sum ./
COPY cmd/oxmail-api/go.mod ./cmd/oxmail-api/
COPY cmd/oxmail/go.mod cmd/oxmail/go.sum ./cmd/oxmail/
RUN go work sync

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY configs/ ./configs/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/oxmail-api ./cmd/oxmail-api

# Stage 2: Minimal runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates curl docker-cli \
    && mkdir -p /etc/oxmail/postfix /etc/oxmail/dovecot /etc/oxmail/dkim

WORKDIR /app

COPY --from=builder /app/oxmail-api .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./oxmail-api"]
