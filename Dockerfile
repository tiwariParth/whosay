# Stage 1: Build the application
FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o whosay

# stage 2: Create the runtime image
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/whosay /app/whosay

RUN chown -R appuser:appgroup /app

USER appuser

LABEL maintainer="Parth Tiwari <parth@example.com>"
LABEL description="Whosay - A Developer-Friendly System Monitor"
LABEL version="0.1.0"

ENTRYPOINT ["/app/whosay"]

CMD ["--all"]
