FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod init traffic-generator
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o traffic-generator main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/traffic-generator .
EXPOSE 8081
CMD ["./traffic-generator", "-metrics-port", "8081"]

