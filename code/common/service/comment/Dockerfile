# ====== 第一阶段：构建 ======
FROM golang:1.22-alpine AS builder

# 构建参数
ARG SERVICE
ARG SERVICE_PATH
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# 先复制 go.mod/go.sum，下载依赖（利用缓存）
COPY go.mod go.sum ./
RUN go mod download

# 再复制代码
COPY . .

# 编译指定服务
RUN go build -o ${SERVICE} ${SERVICE_PATH}/*.go

# ====== 第二阶段：运行 ======
FROM alpine:latest

WORKDIR /root/

# 构建参数
ARG SERVICE

# 从 builder 拷贝对应的二进制
COPY --from=builder /app/${SERVICE} .

# 启动命令
CMD ["./${SERVICE} -f ${SERVICE_PATH}/etc/${SERVICE}.yaml"]
