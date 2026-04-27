# 技术设计说明

**项目名称**：Game 社交化游戏后端  
**文档版本**：v0.2  
**更新时间**：2026-04-27  
**文档目标**：描述当前仓库真实技术基线、服务现状、部署结构和主要技术风险。  

---

## 1. 当前技术基线

当前仓库真实采用的开发基线如下：

- 语言：Go
- 服务框架：go-zero
- 注册发现：etcd
- 数据库：MySQL
- 缓存：Redis
- 分布式事务：DTM
- 可观测性：Prometheus / Grafana / Jaeger
- 本地部署：Docker Compose

当前不应视为已落地基线的技术：

- APISIX Ingress
- ScyllaDB 作为主数据源
- Kafka 事件总线
- 媒体服务
- 信令服务

---

## 2. 服务设计现状

## 2.1 user 服务

目录：

- `code/common/service/user/api`
- `code/common/service/user/rpc`
- `code/common/service/user/model`

当前已实现：

- 注册
- 登录
- 用户信息查询
- JWT 鉴权
- 密码加密

本轮已完成：

- 修复 `user_id` 写入与返回不一致
- 补齐注册/登录基础参数校验
- 补齐 `userinfo` 的用户一致性校验

当前缺口：

- 刷新 token
- 修改资料
- 统一错误码
- 集成测试

## 2.2 comment 服务

目录：

- `code/common/service/comment/api`
- `code/common/service/comment/rpc`
- `code/common/service/comment/rpc/model`

当前已实现：

- 新增评论
- 查询评论详情
- 删除评论
- 评论列表

本轮已完成：

- 打通 API → RPC → Model 主链路
- 增加评论响应统一映射
- 增加逻辑删除能力
- 增加 `/api/comments/:id` 路径参数绑定

当前缺口：

- 点赞/取消点赞
- 屏蔽/取消屏蔽
- 置顶/取消置顶
- 审核能力
- 统计字段与删除一致性

## 2.3 IM 服务

目录：

- `code/common/service/im`

当前状态：

- 已有 TCP 服务基础代码
- 已有连接层、round、buffer、time 等基础工具
- 尚未形成完整消息业务闭环

当前缺口：

- 协议模型
- 身份认证
- 消息投递链路
- 存储边界
- 接入方式定稿

---

## 3. 分层设计

### 3.1 user/comment 服务

当前服务层次结构统一为：

- `api`：HTTP 入口
- `rpc`：内部业务服务
- `model`：数据访问层

设计原则：

- 先保证 API → RPC → Model 主链路完整
- 参数校验尽量前置到 API 或 RPC 入口
- 避免在未稳定前引入额外抽象层

### 3.2 comment 数据模型

评论服务当前采用：

- 评论主题
- 评论索引
- 评论内容

拆分设计。

目标是：

- 按对象分片
- 支持列表查询和内容分离
- 支持 Redis 侧评论 ID 缓存

当前问题：

- 删除、统计、缓存一致性还没有完全收敛

---

## 4. 部署设计

当前本地部署入口：

- `docker-compose.yaml`：全量启动入口
- `docker-compose.base.yaml`：按需组合基础入口

可选依赖拆分：

- `deploy/docker/depends/core.yaml`
  - `etcd + mysql + redis`
- `deploy/docker/depends/manage.yaml`
  - `mysql-manage + redis-manage`
- `deploy/docker/depends/observe.yaml`
  - `prometheus + grafana + jaeger`
- `deploy/docker/depends/txn.yaml`
  - `dtm`

设计原则：

- 本地默认只要求启动核心依赖
- 管理台、观测组件、事务组件按需叠加

---

## 5. 已知风险

### 5.1 user 服务

- `go test ./...` 仍受仓库已有测试和环境影响
- `user/utils/cachex` 存在原有失败测试

### 5.2 comment 服务

- 删除评论当前只改索引状态
- 评论主题统计未同步扣减
- 缓存、统计、删除的一致性仍需补齐

### 5.3 全局

- 文档和代码已基本对齐，但测试体系仍明显不足
- 仓库里存在实验性服务和未完成模块，需要持续标注边界

---

## 6. 关联文档

- 产品需求文档：`docs/product/v0.2/PRD.md`
- 迭代开发任务清单：`docs/product/v0.2/IterationTaskList.md`
- 用户与评论专项计划：`plan/user-comment-service-plan.md`
- Docker 依赖说明：`deploy/docker/README.md`
