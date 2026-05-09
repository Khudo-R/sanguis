# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /sanguis-server cmd/server/main.go

# Stage 2: Final
FROM alpine:latest

WORKDIR /

COPY --from=builder /sanguis-server /sanguis-server
COPY configs/config.yaml /configs/config.yaml

EXPOSE 50051 9090

ENTRYPOINT ["/sanguis-server"]
