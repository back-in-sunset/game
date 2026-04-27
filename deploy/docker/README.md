# Docker 依赖拆分说明

## 目录

- `../../docker-compose.base.yaml`: 多文件组合时的基础入口
- `depends/core.yaml`: 核心依赖，包含 `etcd + mysql + redis`
- `depends/manage.yaml`: 管理界面，包含 `mysql-manage + redis-manage`
- `depends/observe.yaml`: 观测组件，包含 `prometheus + grafana + jaeger`
- `depends/txn.yaml`: 分布式事务组件 `dtm`

## 用法

全量启动：

```bash
docker compose up -d
```

只启动核心依赖：

```bash
docker compose -f docker-compose.base.yaml -f deploy/docker/depends/core.yaml up -d
```

核心依赖加管理界面：

```bash
docker compose \
  -f docker-compose.base.yaml \
  -f deploy/docker/depends/core.yaml \
  -f deploy/docker/depends/manage.yaml \
  up -d
```

核心依赖加观测组件：

```bash
docker compose \
  -f docker-compose.base.yaml \
  -f deploy/docker/depends/core.yaml \
  -f deploy/docker/depends/observe.yaml \
  up -d
```

核心依赖加 DTM：

```bash
docker compose \
  -f docker-compose.base.yaml \
  -f deploy/docker/depends/core.yaml \
  -f deploy/docker/depends/txn.yaml \
  up -d
```

## 说明

- `manage.yaml` 依赖 `core.yaml`
- `txn.yaml` 当前依赖 `etcd`，通常与 `core.yaml` 一起启动
- 根目录 `docker-compose.yaml` 仍然是全量启动入口
- 根目录 `docker-compose.base.yaml` 是按需组合时的基础入口，确保相对路径和 `.env` 都按仓库根目录解析
