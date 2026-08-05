FROM golang:1.24.13-alpine3.23 AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/api

FROM alpine:3.23.5
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
RUN chown appuser:appgroup /app
COPY --from=builder --chown=appuser:appgroup /app/server .
USER appuser
EXPOSE 8080
CMD ["./server"]
