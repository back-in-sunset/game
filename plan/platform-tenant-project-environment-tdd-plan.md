# Tenant Project Environment TDD Plan

## 当前状态（2026-04-27）

### 已完成

- [x] 完成 `Tenant/Project/Environment` 领域模型与约束校验（`internal/domain`）
- [x] 完成内存仓储实现与测试（`model/memory_repository.go`）
- [x] 完成 MySQL 仓储并切换为 goctl 生成 model 适配实现（`model/mysql_repository.go`）
- [x] 完成 goctl 生成 model 与自定义扩展查询能力（`ListByTenantID/ListByProjectID`）
- [x] `internal` 重构为 go-zero 风格目录（`domain/logic/svc`）
- [x] 新增并打通完整 REST API（`api/platform.api` + `api/internal/{handler,logic,svc,types}`）
- [x] 已完成模块级测试通过（`rtk go test ./...`）
- [x] 已补充模块 README（`code/common/service/platform/README.md`）

### 进行中

- [ ] API 层统一错误码与错误语义（目前仍以通用 error 为主）
- [ ] API 请求参数的结构化校验（当前以基础字符串校验为主）
- [ ] API 层单元测试/集成测试（当前覆盖重点在 domain/model）

### 下一轮（P1）

- [ ] 增加统一错误返回（如 `code/message/requestId`）
- [ ] 给 6 个 API 逻辑补齐测试（成功、重复、非法归属、参数非法）
- [ ] 补齐 OpenAPI/接口示例（curl + 请求/响应样例）
- [ ] 增加分页能力（`list project/environment` 支持分页参数）
- [ ] 增加状态流转能力（`active/inactive/archived`）

## 背景

当前产品文档已经明确平台要向 SaaS 化游戏后端演进，但仓库仍处于“单项目基础服务”阶段。下一步不应直接扩散到更多业务模块，而应先落平台最小模型：

- `Tenant`
- `Project`
- `Environment`

本计划以 TDD 为核心，先定义最小行为，再通过测试驱动数据模型、服务接口和配置结构落地。

## 范围

本轮只覆盖平台模型最小闭环，不覆盖：

- `Workspace`
- 套餐计费
- 完整 RBAC
- 配置中心后台
- 服务实例调度

本轮目标：

- 定义 tenant/project/environment 的领域模型
- 明确三者关系与约束
- 为后续 `user/comment/im` 接入平台上下文预留字段
- 通过测试先行建立最小可信实现

## TDD 原则

### 1. 先测行为，不先堆结构

先写“系统应该做什么”的测试，再写表、模型和服务实现。

### 2. 先落最小闭环

不一次设计完整 SaaS 平台，只保证：

- 能创建租户
- 能在租户下创建项目
- 能在项目下创建环境
- 能验证基础唯一性和归属关系

### 3. 分层测试

测试按 3 层组织：

- 领域层测试
- 仓储层测试
- 应用服务层测试

---

## 里程碑

## M1. 领域模型与约束

目标：

- 先把模型行为跑通

测试先写：

- [ ] 创建 Tenant 成功
- [ ] Tenant 名称不能为空
- [ ] Tenant slug 唯一
- [ ] Project 必须归属于 Tenant
- [ ] 同一 Tenant 下 Project key 唯一
- [ ] Environment 必须归属于 Project
- [ ] 同一 Project 下 Environment name 唯一
- [ ] Environment 只允许预设枚举值或满足命名规则

实现后验收：

- 领域对象不依赖数据库即可通过核心规则测试

## M2. 仓储模型与持久化

目标：

- 把领域模型落到可存储结构

测试先写：

- [ ] 能持久化 Tenant
- [ ] 能按 ID 查询 Tenant
- [ ] 能按 Tenant 查询 Project 列表
- [ ] 能按 Project 查询 Environment 列表
- [ ] 重复 slug / key / name 写入时返回预期错误

实现后验收：

- 仓储层能表达最小 CRUD 和唯一性约束

## M3. 应用服务层

目标：

- 提供平台模型最小服务接口

测试先写：

- [ ] CreateTenant 返回 tenant_id
- [ ] CreateProject 会校验 tenant 是否存在
- [ ] CreateEnvironment 会校验 project 是否存在
- [ ] 禁止跨 tenant/project 的非法引用
- [ ] 列表接口能返回归属关系完整的数据

实现后验收：

- 应用服务层具备最小编排能力

## M4. 配置上下文接入预留

目标：

- 为现有服务未来接入平台上下文做准备

测试先写：

- [ ] 配置结构可携带 `tenant_id/project_id/environment`
- [ ] 运行上下文能传递项目环境标识
- [ ] 缺失平台上下文时，服务采用兼容降级策略

实现后验收：

- 后续 `user/comment/im` 可在不推翻现有结构的情况下逐步接入平台模型

## M5. API 完整化（已完成主链路）

目标：

- 基于 goctl 生成骨架打通平台 API 主链路

测试先写（下一轮补齐）：

- [ ] `CreateTenant/GetTenant` API 行为测试
- [ ] `CreateProject/ListProjects` API 行为测试
- [ ] `CreateEnvironment/ListEnvironments` API 行为测试
- [ ] API 错误语义测试（重复、不存在、参数不合法）

实现后验收：

- [x] 6 个 API 路由可用
- [x] API 逻辑已接入 MySQL 仓储（goctl model）
- [x] 配置已包含 `Mysql` 与 `CacheRedis`

---

## 测试策略

## 1. 领域层测试

位置建议：

- `internal/domain/...`
- `internal/domain/..._test.go`

关注点：

- 命名合法性
- 唯一性规则
- 归属关系
- 状态机和约束

特点：

- 不依赖数据库
- 跑得最快
- 最先写

## 2. 仓储层测试

位置建议：

- `internal/repository/...`
- `internal/repository/..._test.go`

关注点：

- 表结构映射
- 唯一索引
- 查询条件
- 关联读取

特点：

- 可以先用内存仓储做契约测试
- 再补真实数据库集成测试

## 3. 应用服务层测试

位置建议：

- `internal/application/...`
- `internal/application/..._test.go`

关注点：

- 创建流程
- 错误返回
- 上下文校验
- 组合查询

特点：

- 依赖 mock 或内存仓储
- 保证业务编排行为正确

## 4. 集成测试

本轮只做最小集成测试，不追求全套平台回归。

建议覆盖：

- CreateTenant → CreateProject → CreateEnvironment 全链路
- 重复创建失败
- 非法归属失败

---

## 实施顺序

1. 建立平台模型目录结构
2. 先写 Tenant 领域测试
3. 实现 Tenant 最小领域对象
4. 再写 Project 领域测试
5. 再写 Environment 领域测试
6. 补仓储契约测试
7. 补应用服务层测试
8. 最后接入配置上下文预留

原则：

- 每完成一层测试再进入下一层
- 不允许先把表和接口堆完再补测试

---

## 最小数据模型建议

### Tenant

- `tenant_id`
- `name`
- `slug`
- `status`
- `created_at`
- `updated_at`

### Project

- `project_id`
- `tenant_id`
- `name`
- `key`
- `status`
- `created_at`
- `updated_at`

### Environment

- `environment_id`
- `project_id`
- `name`
- `display_name`
- `status`
- `created_at`
- `updated_at`

---

## 验收标准

- 能通过测试创建 Tenant / Project / Environment 三层实体
- 领域规则测试先于实现落地
- 仓储层表达唯一性与归属关系
- 应用服务层具备最小创建与查询能力
- API 层具备最小 CRUD 查询闭环（创建+查询+列表）

## 本计划更新记录

- 2026-04-27：补充当前进展、M5 API 完整化、下一轮 P1 任务清单
- 形成可继续扩展到权限、配置、实例调度的模型基础

---

## 当前不做

- Workspace 落地
- 服务实例模型落地
- 配额引擎
- 计费模型
- 完整后台页面
- 多租户历史数据迁移

---

## 下一步衔接

本计划完成后，建议直接进入：

1. 平台模型代码骨架
2. MySQL 表结构草案
3. tenant/project/environment 的第一批测试文件
