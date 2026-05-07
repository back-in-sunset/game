FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o vda ./cmd/*.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/vda .
RUN chmod +x ./vda
CMD ["./vda", "-f", "/app/etc/vda.yaml"]
