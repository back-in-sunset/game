# Platform Service

平台服务（`code/common/service/platform`）提供租户、项目、环境三层资源管理能力，采用 go-zero/goctl 结构实现，包含可运行的 REST API、领域模型、内存仓储与 MySQL 仓储。

## 已实现功能

- 租户管理：创建租户、按 `tenantId` 查询租户
- 项目管理：创建项目、按 `tenantId` 查询项目列表
- 环境管理：创建环境、按 `projectId` 查询环境列表
- 领域约束校验：
  - `tenant.slug` 唯一
  - `project.(tenant_id, project_key)` 唯一
  - `environment.(project_id, name)` 唯一
  - 项目必须归属已存在租户，环境必须归属已存在项目
- 仓储层：
  - `model/memory_repository.go`：本地内存实现（便于单测）
  - `model/mysql_repository.go`：基于 goctl 生成 model 的 MySQL 适配实现

## 目录说明

- `api/`：平台 API 服务（goctl 生成骨架 + 业务逻辑实现）
- `internal/domain/`：领域模型与业务规则
- `internal/logic/`：模块内逻辑（当前主要用于模块级测试）
- `internal/svc/`：模块内服务上下文（保留 `servicecontext.go`）
- `model/`：goctl 生成 model 与仓储实现、SQL DDL

## API 列表

- `POST /api/platform/tenant/create`
- `GET /api/platform/tenant/my`（需要登录态，基于 JWT `uid`）
- `GET /api/platform/tenant/:tenantId`
- `POST /api/platform/project/create`
- `GET /api/platform/project/list?tenantId=...`
- `POST /api/platform/environment/create`
- `GET /api/platform/environment/list?projectId=...`

API 定义文件：`api/platform.api`

## 配置说明

配置文件：`api/etc/platform.yaml`

关键配置项：

- `Host` / `Port`：API 启动监听地址
- `Auth.PublicKeyFile`：JWT 公钥文件（用于解析 `Authorization`）
- `Mysql.DataSource`：MySQL 连接串
- `CacheRedis`：go-zero model 缓存配置

## 本地运行

在 `code/common/service/platform` 目录执行：

```bash
rtk go run ./api -f api/etc/platform.yaml
```

## 测试与校验

```bash
rtk go test ./...
```

## 集成测试（testcontainers-go）

为避免污染主模块依赖，Docker 集成测试放在独立子模块：

- `integration/`

运行方式：

```bash
cd code/common/service/platform/integration
GOMODCACHE=/tmp/go-mod-platform-int GOCACHE=/tmp/go-build-platform-int rtk proxy go test -v . -count=1
```

说明：

- 测试使用 `testcontainers-go` 启动 MySQL 容器，并用 `miniredis` 提供缓存节点
- 通用测试启动逻辑已抽到 `core/testkit`（可供其他服务复用）
- 当前机器无 Docker socket 权限时，测试会自动 `skip`

## 数据库初始化

执行 `model/platform.sql` 建表。该 DDL 含外键约束，业务层也会进行存在性校验与唯一性校验映射。

## goctl 生成说明

- `model/*_gen.go` 为 goctl 生成文件，不要手改
- 自定义扩展在 `model/platformtenantmodel.go`、`model/platformprojectmodel.go`、`model/platformenvironmentmodel.go`
- 当前项目实践中，若使用 `goctl model mysql ddl` 直接解析含外键 DDL 可能失败，可先用去 FK 的临时 DDL 生成，再保留正式 DDL 到仓库
