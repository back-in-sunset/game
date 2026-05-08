# LiveKit + VDA k3s 交付文档

## 架构概览

```
┌──────────────────────────────────────────────────┐
│  k3s cluster                                     │
│                                                  │
│  ┌──────────┐   gRPC    ┌──────────┐            │
│  │   IM     │◀────────▶│   VDA    │            │
│  │  :8093   │           │  :9101   │            │
│  └──────────┘           └────┬─────┘            │
│                              │ HTTP API          │
│                     ┌────────▼──────┐           │
│  ┌──────────┐       │   LiveKit     │           │
│  │  Redis   │       │  :7880 (API)  │           │
│  │  :6379   │       │  :7881 (WS)   │           │
│  └──────────┘       │ :443  (TURN)  │           │
│                     └───────────────┘           │
│                              │                   │
│           UDP 443-65535 (WebRTC/media)          │
│           TCP 7881     (WebSocket fallback)     │
└──────────────────────────────────────────────────┘
```

## 1. 前置条件

### 1.1 k3s 集群要求

| 组件 | 最低版本 | 说明 |
|------|---------|------|
| k3s | v1.27+ | 需要 `traefik`（自带）或替换为 nginx-ingress |
| Helm | v3.12+ | 用于安装 LiveKit chart |
| cert-manager | v1.12+ | TLS 证书自动签发（LiveKit 必须使用 TLS） |
| Redis | 7.x | VDA state backend |

### 1.2 域名与 TLS

LiveKit 要求 TLS（WebRTC 强制加密），需要准备：

- `livekit.<domain>` — LiveKit API 端点（VDA 连接此地址）
- `turn.<domain>` — TURN/TLS 端点（如果使用内置 TURN）

生产环境需使用有效证书，测试环境可用自签证书或 cert-manager + Let's Encrypt。

---

## 2. 安装 cert-manager

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.4/cert-manager.yaml
```

创建 Let's Encrypt ClusterIssuer（生产环境）：

```yaml
# cert-issuer.yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: ops@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            class: traefik
```

```bash
kubectl apply -f cert-issuer.yaml
```

---

## 3. 部署 Redis

```yaml
# redis-deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: voice
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
        - name: redis
          image: redis:7-alpine
          ports:
            - containerPort: 6379
          args:
            - "--save"
            - ""
            - "--appendonly"
            - "no"
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
          volumeMounts:
            - name: redis-data
              mountPath: /data
      volumes:
        - name: redis-data
          persistentVolumeClaim:
            claimName: redis-pvc
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-pvc
  namespace: voice
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: voice
spec:
  selector:
    app: redis
  ports:
    - port: 6379
      targetPort: 6379
```

---

## 4. 部署 LiveKit

### 4.1 添加 Helm repo

```bash
helm repo add livekit https://helm.livekit.io
helm repo update
```

### 4.2 values.yaml

```yaml
# livekit-values.yaml
livekit:
  # 服务配置
  port: 7880
  rtc_port_range_start: 50000
  rtc_port_range_end: 60000

  keys:
    # devkey 与 VDA etc/vda.yaml 中 livekit.api_key 一致
    devkey: secret

  # Redis — 使用集群内地址
  redis:
    address: redis.voice.svc.cluster.local:6379

  # TURN 配置（生产环境需要）
  turn:
    enabled: true
    domain: turn.example.com
    certDomain: turn.example.com
    tlsPort: 443

  # 日志
  logging:
    level: info

# Ingress 配置
ingress:
  enabled: true
  className: traefik
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: livekit.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - hosts:
        - livekit.example.com
      secretName: livekit-tls

# 内部 Service
service:
  type: ClusterIP
  port: 7880

# 资源限制
resources:
  requests:
    cpu: 250m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi
```

### 4.3 安装

```bash
kubectl create namespace voice

helm install livekit livekit/livekit \
  --namespace voice \
  --values livekit-values.yaml
```

### 4.4 验证

```bash
# 检查 Pod 状态
kubectl get pods -n voice -l app.kubernetes.io/name=livekit

# 检查 API 端点
kubectl port-forward -n voice svc/livekit 7880:7880
curl http://localhost:7880/version

# 预期输出: {"version":"1.x.x", ...}
```

---

## 5. 部署 VDA

### 5.1 Dockerfile

```dockerfile
# vda-deploy/Dockerfile
FROM golang:1.22-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o vda ./common/service/vda/cmd/*.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/vda .
RUN chmod +x ./vda
CMD ["./vda", "-f", "/app/etc/vda.yaml"]
```

### 5.2 ConfigMap

```yaml
# vda-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: vda-config
  namespace: voice
data:
  vda.yaml: |
    service_name: vda
    node_id: vda-k3s

    listen:
      rpc: ":9101"

    redis:
      addr: "redis.voice.svc.cluster.local:6379"
      password: ""
      db: 1
      key_prefix: "vda"

    livekit:
      host: "http://livekit.voice.svc.cluster.local:7880"
      api_key: "devkey"
      api_secret: "secret"

    im:
      endpoint: "im.voice.svc.cluster.local:8093"

    discovery:
      endpoints:
        - "etcd.voice.svc.cluster.local:2379"
      service_prefix: "/services/vda"
      lease_ttl_seconds: 10

    log:
      level: info
```

### 5.3 Deployment

```yaml
# vda-deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vda
  namespace: voice
spec:
  replicas: 2
  selector:
    matchLabels:
      app: vda
  template:
    metadata:
      labels:
        app: vda
    spec:
      containers:
        - name: vda
          image: registry.example.com/vda:latest
          ports:
            - containerPort: 9101
              name: rpc
          volumeMounts:
            - name: config
              mountPath: /app/etc
          resources:
            requests:
              cpu: 100m
              memory: 64Mi
            limits:
              cpu: 500m
              memory: 256Mi
          livenessProbe:
            tcpSocket:
              port: 9101
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            tcpSocket:
              port: 9101
            initialDelaySeconds: 3
            periodSeconds: 5
      volumes:
        - name: config
          configMap:
            name: vda-config
---
apiVersion: v1
kind: Service
metadata:
  name: vda
  namespace: voice
spec:
  selector:
    app: vda
  ports:
    - port: 9101
      targetPort: 9101
```

### 5.4 部署

```bash
kubectl apply -f vda-config.yaml
kubectl apply -f vda-deploy.yaml

# 验证
kubectl get pods -n voice -l app=vda
kubectl logs -n voice deployment/vda
```

---

## 6. 网络配置

### 6.1 k3s NodePort（测试环境）

如果集群没有 LoadBalancer，为 LiveKit 的 WebRTC UDP 端口暴露 NodePort：

```yaml
# livekit-nodeport.yaml
apiVersion: v1
kind: Service
metadata:
  name: livekit-udp
  namespace: voice
spec:
  type: NodePort
  selector:
    app.kubernetes.io/name: livekit
  ports:
    - name: rtc-udp
      protocol: UDP
      port: 7881
      nodePort: 30081
```

添加 k3s 启动参数允许 UDP 端口段：

```bash
# /etc/rancher/k3s/config.yaml 或 k3s 安装参数
k3s server \
  --kube-apiserver-arg=service-node-port-range=30000-32767
```

### 6.2 防火墙规则

| 协议 | 端口 | 用途 |
|------|------|------|
| TCP  | 443  | LiveKit TURN/TLS |
| TCP  | 7880 | LiveKit API（集群内部） |
| TCP  | 7881 | WebSocket fallback（可选） |
| UDP  | 50000-60000 | WebRTC media (SRTP/Opus) |
| UDP  | 443  | TURN/TLS (UDP) |

### 6.3 Traefik Ingress（可选）

k3s 自带 Traefik，如需对外暴露 LiveKit：

```yaml
# livekit-ingressroute.yaml
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: livekit
  namespace: voice
spec:
  entryPoints:
    - websecure
  routes:
    - match: Host(`livekit.example.com`)
      kind: Rule
      services:
        - name: livekit
          port: 7880
  tls:
    secretName: livekit-tls
```

---

## 7. 环境变量速查

VDA 与 LiveKit 之间的关键配置对照：

| VDA 配置项 | LiveKit 配置项 | 说明 |
|-----------|---------------|------|
| `livekit.host` | `livekit.port: 7880` | VDA 通过此地址调用 LiveKit API |
| `livekit.api_key` | `livekit.keys.devkey` | 必须一致 |
| `livekit.api_secret` | `livekit.keys.devkey` 的值 | API key = secret（LiveKit 默认模式） |
| — | `rtc_port_range_start/end` | WebRTC UDP 端口范围，需 NodePort/防火墙放行 |
| `redis.addr` | `livekit.redis.address` | VDA 和 LiveKit 可使用不同 Redis DB |

---

## 8. 健康检查

### 服务依赖链

```
IM ──gRPC──▶ VDA ──HTTP──▶ LiveKit
 │              │              │
 │              └─── Redis ◀───┘
 │
 └── Etcd (服务发现)
```

健康检查顺序：Redis → LiveKit → VDA → IM

### VDA

VDA 的 gRPC 端口即健康检查端点：

```bash
# TCP 探测（k3s liveness/readiness）
kubectl exec -n voice deployment/vda -- nc -zv localhost 9101

# gRPC 健康检查（需 grpc-health-probe 工具）
grpc-health-probe -addr=vda.voice.svc.cluster.local:9101
```

k3s 探针配置（已在 Deployment 中包含）：
```yaml
livenessProbe:
  tcpSocket:
    port: 9101
  initialDelaySeconds: 5
  periodSeconds: 10
readinessProbe:
  tcpSocket:
    port: 9101
  initialDelaySeconds: 3
  periodSeconds: 5
```

### LiveKit

```bash
# API 可用性
curl -s http://livekit.voice:7880/version
# 预期: {"version":"1.x.x", ...}

# WebRTC 端口监听
kubectl exec -n voice deployment/livekit -- ss -tuln | grep 7881
```

### Redis

```bash
kubectl exec -n voice deployment/redis -- redis-cli ping
# 预期: PONG
```

---

## 9. 监控指标

### VDA 指标（当前状态）

当前 VDA 未内置 Prometheus metrics endpoint。建议在后续版本中增加以下指标：

| 指标名 | 类型 | 说明 |
|--------|------|------|
| `vda_calls_active` | Gauge | 当前活跃通话数 |
| `vda_rooms_active` | Gauge | 当前活跃房间数 |
| `vda_call_duration_seconds` | Histogram | 通话时长分布 |
| `vda_token_generate_errors_total` | Counter | Token 签发失败次数 |
| `vda_redis_errors_total` | Counter | Redis 操作失败次数 |
| `vda_grpc_requests_total` | Counter | gRPC 请求总数（按 method） |
| `vda_grpc_request_duration_seconds` | Histogram | gRPC 请求延迟分布 |

实施方式：使用 `go-prometheus` interceptor 自动拦截 gRPC 请求，自定义指标通过 `prometheus.NewRegistry()` 注册。

### LiveKit 指标

LiveKit 内置 Prometheus metrics，可通过 Helm values 开启：

```yaml
# livekit-values.yaml
livekit:
  prometheus:
    enabled: true
    port: 9090
```

### k3s 基础监控

```bash
# Node 资源
kubectl top nodes

# Pod 资源
kubectl top pods -n voice

# LiveKit 日志异常检测
kubectl logs -n voice deployment/livekit | grep -E "error|ERROR|warn|WARN" | tail -50

# VDA 日志异常检测
kubectl logs -n voice deployment/vda | grep -E "error|ERROR|panic" | tail -50
```

---

## 10. 部署检查清单

- [ ] cert-manager 已安装并运行
- [ ] ClusterIssuer 已就绪（`kubectl get clusterissuer`）
- [ ] Redis 已部署，`redis.voice.svc.cluster.local:6379` 可达
- [ ] LiveKit Helm chart 已安装，Pod Running
- [ ] LiveKit API 可访问：`curl http://livekit.voice:7880`
- [ ] VDA ConfigMap 中 `livekit.host` 指向正确的 LiveKit Service 地址
- [ ] VDA 镜像已构建并推送至 registry
- [ ] VDA Deployment 已应用，2 副本 Running
- [ ] VDA 日志无 `livekit` 连接错误
- [ ] UDP 端口段在防火墙/NAT 已放行
- [ ] 客户端可通过 LiveKit SDK 连接并建立 WebRTC 连接

---

## 11. 常见问题

### 客户端无法连接 LiveKit

**排查**：
1. 确认 LiveKit TLS 证书有效
2. 确认 UDP 端口段已在 NAT/防火墙放行
3. 检查 LiveKit 日志：`kubectl logs -n voice deployment/livekit`
4. 确认客户端使用的 `livekit_url` 为 `wss://` 协议（WebSocket over TLS）

### VDA 无法签发 token

**排查**：
1. `livekit.host` 是否可达：`kubectl exec -n voice deployment/vda -- wget -qO- http://livekit.voice:7880`
2. `api_key` / `api_secret` 是否匹配 LiveKit keys 配置
3. 检查 VDA 日志：`kubectl logs -n voice deployment/vda | grep -i livekit`

### 音频卡顿/丢包

1. 确认 `rtc_port_range_start/end` 端口段在防火墙已放行
2. 检查 k3s 节点间网络延迟：`ping <node-ip>`
3. 考虑启用 TURN 中继（LiveKit 内置 TURN 或外部 coturn）

### Redis 连接失败

**排查**：
1. 检查 VDA 日志：`kubectl logs -n voice deployment/vda | grep -i redis`
2. 验证 Redis Service 可达：`kubectl exec -n voice deployment/vda -- nc -zv redis.voice 6379`
3. 检查 Redis 是否 OOM：`kubectl describe pod -n voice -l app=redis | grep OOM`
4. 确认 VDA config 中 `redis.addr` 与 Redis Service 名称一致

### IM 无法连接 VDA

**排查**：
1. 检查 VDA Pod 状态：`kubectl get pods -n voice -l app=vda`
2. 确认 VDA Service ClusterIP 可达：`kubectl exec -n voice deployment/im -- nc -zv vda.voice 9101`
3. 检查 IM 配置中 `vda_endpoint` 是否为 `vda.voice.svc.cluster.local:9101`
4. 查看 IM 日志中 gRPC 连接错误
5. 确认 VDA 没有达到资源限制被 throttling：`kubectl top pods -n voice -l app=vda`

### LiveKit 房间不释放

长时间空闲的 LiveKit 房间应自动销毁。如果房间堆积：

1. 检查 LiveKit 日志中 `room` 相关记录
2. 确认 LiveKit Redis 配置正确（房间状态存储在 Redis）
3. 手动清理：`kubectl exec -n voice deployment/redis -- redis-cli KEYS "lk:*" | wc -l`
