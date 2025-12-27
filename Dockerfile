# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
# COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -o /server cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

# Set GC tuning for Pi Zero
ENV GOGC=50

COPY --from=builder /server .
# Create directories for data and cache
RUN mkdir -p assets/photos kiosk_cache

EXPOSE 8080

CMD ["./server"]
