# IM 双域设计说明

## 1. 目标

IM 服务按两类业务域落地：

- `platform`：平台社交域，独立于租户体系，允许跨租户私信
- `tenant`：租户内 IM 域，强制绑定 `tenant_id + project_id + environment`

首期能力范围：

- 私信
- 系统通知
- 业务推送
- WebSocket + TCP 双接入
- RSA JWT 鉴权
- 服务发现抽象，Etcd 默认实现
- 持久化会话 + 离线箱接口预留

## 2. 核心边界

### 2.1 共享能力

平台域和租户域共享以下基础设施：

- 长连接接入
- JWT 鉴权
- 节点注册与服务发现
- 连接管理
- 节点间消息路由
- 统一消息信封

### 2.2 隔离能力

平台域和租户域必须在以下维度独立：

- 会话空间
- 路由键
- 消息存储
- 未读统计
- 离线箱
- 通知投递目标

## 3. 统一消息模型

所有消息统一使用信封：

- `domain`
- `scope`
- `sender`
- `receiver`
- `msg_type`
- `seq`
- `payload`
- `sent_at`

消息类型固定为：

- `direct_message`
- `system_notice`
- `biz_push`

## 4. 作用域规则

### 4.1 平台域

- `domain = platform`
- 不要求 `tenant/project/environment`
- 允许跨租户私信

会话键：

`platform:{uid_a}:{uid_b}`

### 4.2 租户域

- `domain = tenant`
- 必须提供 `tenant_id`
- 必须提供 `project_id`
- 必须提供 `environment`
- 严禁跨租户、跨项目、跨环境互通

会话键：

`tenant:{tenant}:{project}:{env}:{uid_a}:{uid_b}`

## 5. 鉴权与连接

长连接登录时复用现有 `user` 服务 RSA JWT：

1. 客户端携带 token 建立 TCP 或 WebSocket 连接
2. 服务端校验 RSA 签名和 `uid`
3. 客户端声明目标域与作用域
4. 服务端校验域规则
5. 绑定 `domain + scope + uid + session_id`

## 6. 路由与投递

消息发送流程：

1. 校验发送者上下文
2. 构建会话键
3. 查询目标用户在线节点
4. 本地节点直投或远端节点转发
5. 失败时写离线箱
6. 更新会话聚合

## 7. 服务接口

内部接口固定为：

- `AuthProvider`
- `Registry`
- `Router`
- `SessionManager`
- `MessageStore`
- `Notifier`
- `ScopeResolver`

代码实现必须围绕这些接口展开，不直接把 Etcd、JWT、离线箱、传输层写死到业务流里。

## 7.1 存储分层

为后续接入 ScyllaDB，消息存储拆成两层：

- `MessageArchive`
  - 负责消息归档写入
  - 负责消息历史查询
  - 当前实现：MySQL
  - 预留替换：ScyllaDB
- `SessionStateStore`
  - 负责会话聚合
  - 负责未读/已读
  - 负责离线箱
  - 当前实现：MySQL + Redis

组合后的 `MessageStore` 仍对上层暴露统一接口，服务层不直接感知底层是 MySQL 还是 ScyllaDB。

## 7.2 跨节点在线投递

跨节点在线投递按以下顺序处理：

1. 本地会话直投
2. Redis presence 查询目标用户所在节点
3. Etcd 解析目标节点内部 RPC 地址
4. 通过 `DeliverInternal` 转发到目标 IM 节点
5. 若目标节点不可达或未命中在线会话，则写离线箱

验证应至少覆盖：

- A 节点登录用户
- B 节点通过业务 RPC 发消息
- A 节点收到在线消息
- presence 失效后消息回落离线箱

## 8. 首期不做

- 群聊
- 频道/聊天室
- 厂商推送通道
- 独立 IM 账号体系
- 跨域会话复用
