FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# Copy vda module (for the e2e replace directive)
COPY go.mod go.sum ./
RUN go mod download

COPY rpc/        rpc/
COPY internal/   internal/
COPY cmd/        cmd/

# Build e2e test
WORKDIR /app/deploy/e2e
COPY deploy/e2e/go.mod deploy/e2e/go.sum* ./
RUN go mod download 2>/dev/null || true
COPY deploy/e2e/main.go .
RUN go build -o e2e .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/deploy/e2e/e2e .
RUN chmod +x ./e2e
CMD ["./e2e"]
