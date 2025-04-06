# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go.mod first to leverage caching
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o whosay

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/whosay .

# Run the binary
ENTRYPOINT ["./whosay"]
CMD ["--all"]
