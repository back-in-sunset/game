# k8s 交付文档

## 1. 架构总览

```
                          ┌─────────────────────────────────────────────┐
                          │  k8s cluster                                │
                          │                                             │
                          │  ┌─────── Ingress / APISIX (L7) ──────┐     │
                          │  │  user   comment  friend  history   │    │
                          │  │  platform  tag                     │    │
                          │  └────────────────────────────────────┘    │
                          │                                             │
                          │  ┌ Service (LoadBalancer) ──────────────┐   │
                          │  │  IM Pod1   IM Pod2   ...  (TCP/WS)  │   │
                          │  └──────────────────────────────────────┘   │
                          │     │ gRPC                                 │
                          │  ┌──▼────────────────────────────────────┐  │
                          │  │  VDA Pod1   VDA Pod2  (gRPC + Redis) │  │
                          │  └──┬────────────────────────────────────┘  │
                          │     │ HTTP API                              │
                          │  ┌──▼────────────────────────────────────┐  │
                          │  │  LiveKit (Helm)   (WebRTC + UDP)     │  │
                          │  └──────────────────────────────────────┘   │
                          │                                             │
                          │  Redis  etcd  MySQL  Jaeger  Prometheus    │
                          └─────────────────────────────────────────────┘
```

**分层策略**：

| 层 | 入口方式 | 服务 |
|----|---------|------|
| L7 API 网关 | APISIX (Ingress) | user, comment, friend, history, platform, tag |
| L4 长连接 | K8s Service LoadBalancer | IM (TCP 8091, WS 8081) |
| 集群内部 | ClusterIP Service | VDA, LiveKit, Redis, etcd, MySQL |
| 直连 | K8s Service LoadBalancer (UDP) | LiveKit WebRTC media |

---

## 2. 命名空间

```yaml
# namespaces.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: game-infra     # 基础组件
---
apiVersion: v1
kind: Namespace
metadata:
  name: game-svc       # 业务服务 (user/comment/friend/history/platform/tag)
---
apiVersion: v1
kind: Namespace
metadata:
  name: voice          # IM + VDA + LiveKit
```

```bash
kubectl apply -f namespaces.yaml
```

---

## 3. 基础组件

### 3.1 etcd 集群

```yaml
# infra/etcd.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: etcd
  namespace: game-infra
spec:
  serviceName: etcd
  replicas: 3
  selector:
    matchLabels:
      app: etcd
  template:
    metadata:
      labels:
        app: etcd
    spec:
      containers:
        - name: etcd
          image: bitnami/etcd:3.5
          env:
            - name: ALLOW_NONE_AUTHENTICATION
              value: "yes"
            - name: ETCD_ADVERTISE_CLIENT_URLS
              value: "http://$(MY_POD_NAME).etcd.game-infra.svc.cluster.local:2379"
            - name: ETCD_LISTEN_CLIENT_URLS
              value: "http://0.0.0.0:2379"
            - name: ETCD_INITIAL_CLUSTER
              value: "etcd-0=http://etcd-0.etcd.game-infra.svc.cluster.local:2380,etcd-1=http://etcd-1.etcd.game-infra.svc.cluster.local:2380,etcd-2=http://etcd-2.etcd.game-infra.svc.cluster.local:2380"
            - name: ETCD_INITIAL_CLUSTER_STATE
              value: "new"
            - name: MY_POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
          ports:
            - containerPort: 2379
              name: client
            - containerPort: 2380
              name: peer
          resources:
            requests: { cpu: 100m, memory: 256Mi }
            limits:   { cpu: 500m, memory: 512Mi }
          volumeMounts:
            - name: data
              mountPath: /bitnami/etcd
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 10Gi
---
apiVersion: v1
kind: Service
metadata:
  name: etcd
  namespace: game-infra
spec:
  clusterIP: None
  selector:
    app: etcd
  ports:
    - name: client
      port: 2379
    - name: peer
      port: 2380
```

**地址**: `etcd-0.etcd.game-infra.svc.cluster.local:2379` (集群内)

### 3.2 MySQL

```yaml
# infra/mysql.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql
  namespace: game-infra
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mysql
  template:
    metadata:
      labels:
        app: mysql
    spec:
      containers:
        - name: mysql
          image: mysql:8.0
          env:
            - name: MYSQL_ROOT_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: mysql-secret
                  key: root-password
          ports:
            - containerPort: 3306
          resources:
            requests: { cpu: 500m, memory: 512Mi }
            limits:   { cpu: 2000m, memory: 2Gi }
          volumeMounts:
            - name: data
              mountPath: /var/lib/mysql
            - name: init
              mountPath: /docker-entrypoint-initdb.d
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: mysql-pvc
        - name: init
          configMap:
            name: mysql-init
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-pvc
  namespace: game-infra
spec:
  accessModes: ["ReadWriteOnce"]
  resources:
    requests:
      storage: 50Gi
---
apiVersion: v1
kind: Service
metadata:
  name: mysql
  namespace: game-infra
spec:
  selector:
    app: mysql
  ports:
    - port: 3306
```

**地址**: `mysql.game-infra.svc.cluster.local:3306`

### 3.3 Redis

```yaml
# infra/redis.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: game-infra
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
          args: ["--save", "", "--appendonly", "no"]
          ports:
            - containerPort: 6379
          resources:
            requests: { cpu: 100m, memory: 128Mi }
            limits:   { cpu: 500m, memory: 512Mi }
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: game-infra
spec:
  selector:
    app: redis
  ports:
    - port: 6379
```

**地址**: `redis.game-infra.svc.cluster.local:6379`

### 3.4 创建 Secrets

```bash
kubectl create secret generic mysql-secret \
  --namespace game-infra \
  --from-literal=root-password='<your-mysql-password>'
```

---

## 4. APISIX 网关

### 4.1 安装 APISIX

```bash
helm repo add apisix https://charts.apiseven.com
helm repo update

helm install apisix apisix/apisix \
  --namespace apisix \
  --create-namespace \
  --set apisix.allow.ipList="{0.0.0.0/0}" \
  --set admin.allow.ipList="{0.0.0.0/0}"
```

### 4.2 路由配置

```yaml
# apisix/game-routes.yaml
apiVersion: apisix.apache.org/v2
kind: ApisixRoute
metadata:
  name: game-services
  namespace: apisix
spec:
  http:
    - name: user-rpc
      match:
        hosts:
          - api.game.example.com
        paths:
          - /user/*
      backends:
        - serviceName: user-rpc
          servicePort: 9001
          namespace: game-svc
      plugins:
        - name: jwt-auth
          enable: true
          config:
            key: token
            secret: <jwt-public-key>

    - name: comment-rpc
      match:
        hosts:
          - api.game.example.com
        paths:
          - /comment/*
      backends:
        - serviceName: comment-rpc
          servicePort: 9002
          namespace: game-svc

    - name: history-rpc
      match:
        hosts:
          - api.game.example.com
        paths:
          - /history/*
      backends:
        - serviceName: history-rpc
          servicePort: 9003
          namespace: game-svc

    - name: friend-rpc
      match:
        hosts:
          - api.game.example.com
        paths:
          - /friend/*
      backends:
        - serviceName: friend-rpc
          servicePort: 9102
          namespace: game-svc

    - name: platform-api
      match:
        hosts:
          - api.game.example.com
        paths:
          - /platform/*
      backends:
        - serviceName: platform-api
          servicePort: 8888
          namespace: game-svc
```

> **注意**: APISIX 只负责 HTTP/gRPC 短连接服务。IM 长连接和 LiveKit WebRTC 不经过 APISIX。

---

## 5. 业务服务

所有 HTTP/gRPC 服务使用统一的 Deployment 模板，通过 `SERVICE` 和 `SERVICE_PATH` 构建参数区分。

### 5.1 构建镜像

```dockerfile
# Dockerfile (项目根目录)
FROM golang:1.22-alpine AS builder

ARG SERVICE
ARG SERVICE_PATH
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o ${SERVICE} ${SERVICE_PATH}/*.go

FROM alpine:latest
WORKDIR /app/
ARG SERVICE
COPY --from=builder /app/${SERVICE} .
RUN chmod +x ./${SERVICE}
CMD ["./${SERVICE}", "-f", "/app/etc/${SERVICE}.yaml"]
```

```bash
# 构建 user 服务
docker build --build-arg SERVICE=user \
             --build-arg SERVICE_PATH=./common/service/user/rpc \
             -t registry.example.com/user-rpc:latest \
             -f Dockerfile .

# 类推: comment, history, friend, platform
```

### 5.2 user-rpc (port 9001)

```yaml
# svc/user.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: user-config
  namespace: game-svc
data:
  user.yaml: |
    Name: user.rpc
    ListenOn: 0.0.0.0:9001
    Mode: pro

    Etcd:
      Hosts:
        - etcd-0.etcd.game-infra.svc.cluster.local:2379
        - etcd-1.etcd.game-infra.svc.cluster.local:2379
        - etcd-2.etcd.game-infra.svc.cluster.local:2379
      Key: user.rpc

    Mysql:
      DataSource: root:<password>@tcp(mysql.game-infra.svc.cluster.local:3306)/user?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai

    CacheRedis:
      - Host: redis.game-infra.svc.cluster.local:6379
        Type: node

    ScyllaDB:
      Hosts:
        - <scylla-host>
      Keyspace: user
      Timeout: 10

    Telemetry:
      Name: user-rpc
      Endpoint: http://jaeger.game-infra.svc.cluster.local:14268/api/traces
      Sampler: 1.0
      Batcher: jaeger

    Identity:
      Timeout: 5
      PoolSize: 100000
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-rpc
  namespace: game-svc
spec:
  replicas: 2
  selector:
    matchLabels:
      app: user-rpc
  template:
    metadata:
      labels:
        app: user-rpc
    spec:
      containers:
        - name: user-rpc
          image: registry.example.com/user-rpc:latest
          ports:
            - containerPort: 9001
              name: rpc
          volumeMounts:
            - name: config
              mountPath: /app/etc
          resources:
            requests: { cpu: 100m, memory: 128Mi }
            limits:   { cpu: 500m, memory: 512Mi }
          livenessProbe:
            tcpSocket:
              port: 9001
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            tcpSocket:
              port: 9001
            initialDelaySeconds: 5
            periodSeconds: 5
      volumes:
        - name: config
          configMap:
            name: user-config
---
apiVersion: v1
kind: Service
metadata:
  name: user-rpc
  namespace: game-svc
spec:
  selector:
    app: user-rpc
  ports:
    - port: 9001
      targetPort: 9001
```

### 5.3 其他服务部署参数

| 服务 | ConfigMap | Port | 特点 |
|------|-----------|------|------|
| user-rpc | user.yaml | 9001 | ScyllaDB |
| comment-rpc | comment.yaml | 9002 | 依赖 IM etcd 发现 |
| history-rpc | history.yaml | 9003 | ReadFallbackToDB |
| friend-rpc | friend.yaml | 9102 | — |
| platform-api | platform.yaml | 8888 | HTTP API, JWT 鉴权 |

---

## 6. IM 服务 (长连接)

### 6.1 ConfigMap

```yaml
# im/im-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: im-config
  namespace: voice
data:
  im.yaml: |
    service_name: im
    node_id: ""  # Pod 启动时注入

    listen:
      websocket: ":8081"
      tcp: ":8091"
      rpc: ":8093"  # 集群内跨节点转发 gRPC

    auth:
      public_key_file: "/app/etc/public.key"

    discovery:
      endpoints:
        - etcd-0.etcd.game-infra.svc.cluster.local:2379
        - etcd-1.etcd.game-infra.svc.cluster.local:2379
        - etcd-2.etcd.game-infra.svc.cluster.local:2379
      service_prefix: "/services/im"
      lease_ttl_seconds: 10

    mysql:
      data_source: "root:<password>@tcp(mysql.game-infra.svc.cluster.local:3306)/game_im?parseTime=true&loc=Local"

    redis:
      addr: "redis.game-infra.svc.cluster.local:6379"
      password: ""
      db: 0
      key_prefix: "im"

    session:
      bucket_count: 64
      ring_size: 256
      reader_buffer_size: 4096
      writer_buffer_size: 4096
      frame_buffer_size: 8192
      heartbeat_interval_seconds: 30
      heartbeat_misses: 3
      write_flush_interval_ms: 5

    scope:
      default_environment: "prod"

    vda_endpoint: "vda.voice.svc.cluster.local:9101"
```

### 6.2 Deployment

```yaml
# im/im-deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: im
  namespace: voice
spec:
  replicas: 2
  selector:
    matchLabels:
      app: im
  template:
    metadata:
      labels:
        app: im
    spec:
      containers:
        - name: im
          image: registry.example.com/im:latest
          command:
            - /app/im
            - -f
            - /app/etc/im.yaml
          env:
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
          ports:
            - containerPort: 8081
              name: ws
            - containerPort: 8091
              name: tcp
            - containerPort: 8093
              name: rpc
          volumeMounts:
            - name: config
              mountPath: /app/etc
          resources:
            requests: { cpu: 250m, memory: 256Mi }
            limits:   { cpu: 1000m, memory: 1Gi }
          livenessProbe:
            tcpSocket:
              port: 8091
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            tcpSocket:
              port: 8091
            initialDelaySeconds: 3
            periodSeconds: 5
      volumes:
        - name: config
          configMap:
            name: im-config
```

### 6.3 Service (对外)

```yaml
# im/im-service.yaml
# TCP/WS 使用 LoadBalancer，不经过 APISIX
apiVersion: v1
kind: Service
metadata:
  name: im-external
  namespace: voice
spec:
  type: LoadBalancer
  selector:
    app: im
  ports:
    - name: ws
      port: 8081
      targetPort: 8081
      protocol: TCP
    - name: tcp
      port: 8091
      targetPort: 8091
      protocol: TCP
---
# 集群内 gRPC 通信使用 ClusterIP
apiVersion: v1
kind: Service
metadata:
  name: im
  namespace: voice
spec:
  type: ClusterIP
  selector:
    app: im
  ports:
    - name: rpc
      port: 8093
      targetPort: 8093
```

> **说明**：IM 跨节点转发已由 etcd + Redis presence + gRPC forward 自闭环，K8s Service 仅做连接入口。`sessionAffinity` 不需要设置，断连重连到不同 Pod 后 Redis presence 会自动更新绑定。

### 6.4 node_id 注入

IM 启动时 `node_id` 为空，需要写入 Pod 名称作为唯一节点标识。在 Deployment 的 `command` 中用 sed 注入：

```yaml
command:
  - sh
  - -c
  - |
    sed -i "s/node_id: \"\"/node_id: \"${NODE_ID}\"/" /app/etc/im.yaml
    exec /app/im -f /app/etc/im.yaml
```

---

## 7. VDA 服务

### 7.1 ConfigMap

```yaml
# vda/vda-config.yaml
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
      addr: "redis.game-infra.svc.cluster.local:6379"
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
        - etcd-0.etcd.game-infra.svc.cluster.local:2379
        - etcd-1.etcd.game-infra.svc.cluster.local:2379
        - etcd-2.etcd.game-infra.svc.cluster.local:2379
      service_prefix: "/services/vda"
      lease_ttl_seconds: 10

    log:
      level: info
```

### 7.2 Deployment

```yaml
# vda/vda-deploy.yaml
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
            requests: { cpu: 100m, memory: 64Mi }
            limits:   { cpu: 500m, memory: 256Mi }
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

---

## 8. LiveKit

### 8.1 values.yaml

```yaml
# livekit/livekit-values.yaml
livekit:
  port: 7880
  rtc_port_range_start: 50000
  rtc_port_range_end: 60000

  keys:
    devkey: secret  # 与 VDA livekit.api_key / api_secret 一致

  redis:
    address: redis.game-infra.svc.cluster.local:6379

  turn:
    enabled: true
    domain: turn.example.com
    certDomain: turn.example.com
    tlsPort: 443

  logging:
    level: info

  prometheus:
    enabled: true
    port: 9090

service:
  type: ClusterIP
  port: 7880

resources:
  requests: { cpu: 250m, memory: 256Mi }
  limits:   { cpu: 1000m, memory: 1Gi }
```

### 8.2 安装

```bash
helm repo add livekit https://helm.livekit.io
helm repo update

helm install livekit livekit/livekit \
  --namespace voice \
  --values livekit-values.yaml
```

### 8.3 WebRTC UDP 对外暴露

WebRTC 需要客户端直连 LiveKit 的 UDP 端口，必须用 LoadBalancer 或 NodePort 对外暴露：

```yaml
# livekit/livekit-udp.yaml
apiVersion: v1
kind: Service
metadata:
  name: livekit-udp
  namespace: voice
spec:
  type: LoadBalancer
  selector:
    app.kubernetes.io/name: livekit
  ports:
    - name: rtc-udp
      protocol: UDP
      port: 7881
      targetPort: 7881
    - name: rtc-udp-range
      protocol: UDP
      port: 50000
      targetPort: 50000
```

> **注意**: 每个 K8s Service 最多暴露一个端口范围，LiveKit 需要的 UDP 50000-60000 范围需要 NodePort + 防火墙方案，或使用 hostNetwork。

### 8.4 hostNetwork 方案（UDP 端口段）

如果 LoadBalancer 不支持大段 UDP，改用 hostNetwork + 节点选择器：

```yaml
# livekit/livekit-hostnet.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: livekit
  namespace: voice
spec:
  replicas: 2
  selector:
    matchLabels:
      app: livekit
  template:
    metadata:
      labels:
        app: livekit
    spec:
      hostNetwork: true
      nodeSelector:
        role: voice
      containers:
        - name: livekit
          image: livekit/livekit-server:v1.7
          env:
            - name: LIVEKIT_CONFIG
              valueFrom:
                configMapKeyRef:
                  name: livekit-config
                  key: config.yaml
          ports:
            - containerPort: 7880
            - containerPort: 7881
            - containerPort: 50000
            - containerPort: 50001
            # ... 根据 rtc_port_range 填全
```

---

## 9. 网络速查

### 9.1 端口矩阵

| 服务 | Pod 端口 | Service 端口 | 协议 | 对外 |
|------|---------|-------------|------|------|
| user-rpc | 9001 | 9001 | gRPC | APISIX |
| comment-rpc | 9002 | 9002 | gRPC | APISIX |
| history-rpc | 9003 | 9003 | gRPC | APISIX |
| friend-rpc | 9102 | 9102 | gRPC | APISIX |
| platform-api | 8888 | 8888 | HTTP | APISIX |
| **IM WS** | 8081 | 8081 | TCP | **LoadBalancer (直连)** |
| **IM TCP** | 8091 | 8091 | TCP | **LoadBalancer (直连)** |
| IM RPC | 8093 | 8093 | gRPC | ClusterIP (内部) |
| VDA RPC | 9101 | 9101 | gRPC | ClusterIP (内部) |
| LiveKit API | 7880 | 7880 | HTTP | ClusterIP (内部) |
| **LiveKit UDP** | 7881, 50000-60000 | — | UDP | **LoadBalancer/hostNetwork** |
| Redis | 6379 | 6379 | TCP | ClusterIP |
| etcd | 2379 | 2379 | TCP | ClusterIP |
| MySQL | 3306 | 3306 | TCP | ClusterIP |

### 9.2 依赖链

```
客户端
  ├── TCP/WS ──▶ IM ──gRPC──▶ VDA ──HTTP──▶ LiveKit
  │               │             │              │
  │               ├── etcd      ├── Redis      ├── Redis
  │               ├── MySQL     └── etcd       └── cert-manager (TLS)
  │               └── Redis
  │
  ├── HTTP/gRPC ──▶ APISIX ──▶ user / comment / friend / history / platform
  │                              │           │          │
  │                              └── MySQL ──┘──────────┘
  │                             他们都是 etcd + Redis
  │
  └── UDP (WebRTC) ──▶ LiveKit
```

---

## 10. 健康检查与启动顺序

```bash
# 1. 基础组件就绪
kubectl wait --for=condition=ready pod -l app=mysql -n game-infra --timeout=120s
kubectl wait --for=condition=ready pod -l app=redis -n game-infra --timeout=60s
kubectl wait --for=condition=ready pod -l app=etcd -n game-infra --timeout=120s

# 2. 业务服务
kubectl wait --for=condition=ready pod -l app=user-rpc -n game-svc --timeout=60s
kubectl wait --for=condition=ready pod -l app=comment-rpc -n game-svc --timeout=60s
# ... (类推)

# 3. LiveKit
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=livekit -n voice --timeout=60s

# 4. VDA
kubectl wait --for=condition=ready pod -l app=vda -n voice --timeout=60s

# 5. IM
kubectl wait --for=condition=ready pod -l app=im -n voice --timeout=60s
```

### 快速验证

```bash
# IM TCP 端口
nc -zv <loadbalancer-ip> 8091

# LiveKit API
curl http://livekit.voice.svc.cluster.local:7880/version

# VDA gRPC (需要 grpc-health-probe)
grpc-health-probe -addr=vda.voice.svc.cluster.local:9101

# Redis
kubectl exec -n game-infra deployment/redis -- redis-cli ping
# 预期: PONG
```

---

## 11. 部署检查清单

- [ ] 命名空间 `game-infra`, `game-svc`, `voice` 已创建
- [ ] MySQL Secret 已创建，数据库 `game_im` 已建表
- [ ] etcd 3 副本 Running，`etcdctl endpoint health` 全部正常
- [ ] Redis Running，`redis-cli ping` → PONG
- [ ] APISIX 已安装，路由规则已 apply
- [ ] 所有业务服务镜像已构建并推送至 registry
- [ ] user-rpc / comment-rpc / history-rpc / friend-rpc / platform-api 全部 Running
- [ ] LiveKit Helm chart 已安装，Pod Running
- [ ] LiveKit API `http://livekit.voice:7880/version` 正常返回
- [ ] VDA Deployment 已 apply，Pod Running
- [ ] VDA 日志无 LiveKit/Redis 连接错误
- [ ] IM Deployment 已 apply，Pod Running
- [ ] IM LoadBalancer Service 已分配 ExternalIP
- [ ] UDP 端口段 (50000-60000) 防火墙/NAT 已放行
- [ ] 客户端可建立 TCP/WS 连接，可完成 Auth + SendMsg
- [ ] 客户端可发起语音通话，WebRTC 连接成功
