FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/device_shadow ./rpc

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/device_shadow .
COPY --from=builder /app/rpc/etc /app/etc

EXPOSE 8080

CMD ["./device_shadow", "-f", "/app/etc/deviceshadow.yaml"]
