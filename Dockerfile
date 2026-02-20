# Use the specific Go 1.26 version
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Leveraging Docker layer caching for dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 1. Added CGO_ENABLED=0 for a static binary
# 2. Added ldflags to strip debug information (reduces binary size)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server

# --- Multi-stage build for a tiny production image ---
FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
