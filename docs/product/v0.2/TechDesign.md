# 技术设计说明

**项目名称**：Game 社交化游戏后端  
**文档版本**：v0.2  
**更新时间**：2026-05-09（基于实际代码状态修订）  
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

- WebSocket 接入层完整（`/ws` 端点，端口 8082）
- TCP 接入层完整（端口 8091）
- 二进制帧协议：16 字节大端 header（totalLen|headerLen|version|op|seq）+ JSON body
- JWT RS256 鉴权（platform API 公钥验证，提取 uid）
- 单聊消息路由：在线投递 + 离线 MySQL 存储
- 会话管理：bucket + ring 分片，心跳保活 30s
- etcd 服务发现注册
- Redis 会话存储 + MySQL 消息持久化
- 操作码：OpAuth(7)、OpAuthReply(8)、OpHeartbeat(2)、OpServerPush(4)、OpError(6)
- 前端 Web 客户端已接入，可实时收发消息

当前缺口：

- 群聊/频道
- 已读回执
- 消息搜索
- 多媒体消息
- TCP 客户端 SDK

## 2.4 friend 服务

目录：

- `code/common/service/friend`

当前状态：

- go-zero RPC 服务，提供好友 CRUD
- HTTP API 层已就绪（可通过 platform API 转发）
- 好友添加/删除/查询/在线状态
- 前端 API 客户端已封装（`packages/api/src/http/friend.ts`）

当前缺口：

- 好友分组/标签
- 黑名单
- 前端完整 UI 页面（API 层已就绪）

## 2.5 VDA 语音服务

目录：

- `code/common/service/vda`

当前状态：

- gRPC 服务（端口 9101）：token 签发、房间管理
- LiveKit 服务端集成（Docker 部署，端口 7880）
- 语音信令协议：call_invite / call_accept / call_reject / call_end
- 通话状态管理

当前缺口：

- 语音分钟数计量
- 通话录制

## 2.6 前端 Monorepo

目录：

- `frontend/`

当前状态：

- pnpm workspace monorepo
- `apps/web`：Vite + React SPA，Zustand 状态管理，React Router v7
- `apps/mobile`：Expo + React Native 壳工程
- `packages/api`：IM 协议编解码（二进制帧）+ HTTP API 客户端
- `packages/ui`：共享 UI 组件库（Card、StatusRow、ChatBubble 等）
- `packages/config`：共享配置
- `packages/shared`：共享类型和工具
- IM WebSocket 客户端已接入，可实时收发消息

当前缺口：

- Mobile 端功能实现
- E2E 测试（Playwright）
- 组件测试（Vitest + React Testing Library）

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

### 5.3 IM 服务

- 单节点部署，未验证水平扩展
- 离线消息量增长后 MySQL 性能风险
- WebSocket 重连时消息去重未完整覆盖

### 5.4 VDA 服务

- LiveKit 单实例部署，未验证多节点
- 通话分钟数未计量，无法计费

### 5.5 前端

- 无自动化测试覆盖
- Mobile 端为占位页面
- LiveKit Web SDK 未集成，语音链路未端到端打通

### 5.6 全局

- 测试体系仍明显不足
- 服务间无统一错误码
- 仓库里存在实验性服务和未完成模块，需要持续标注边界

---

## 6. 关联文档

- 产品需求文档：`docs/product/v0.2/PRD.md`
- 迭代开发任务清单：`docs/product/v0.2/IterationTaskList.md`
- 用户与评论专项计划：`plan/user-comment-service-plan.md`
- Docker 依赖说明：`deploy/docker/README.md`
