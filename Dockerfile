# 1. Builder env
FROM golang:1.26-alpine AS builder
WORKDIR /app

# Copy and download dependencies only (for cache advantage)
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Compile the app
# CGO_ENABLED=0: Static binary with no dependence outside
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/exchange-api ./cmd/api/main.go

# 2. Prod env
FROM alpine:latest
WORKDIR /app

# OS time zone data for the time/date processes of Go
RUN apk --no-cache add tzdata

# Take compiled single file from step 1
COPY --from=builder /app/exchange-api .

EXPOSE 8080

# Start the app
ENTRYPOINT ["./exchange-api"]