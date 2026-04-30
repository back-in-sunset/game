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

## 7.3 接入协议与连接模型

当前长连接层借鉴 `goim` 的 comet 思路，但不直接复刻其 logic/job/broadcast 体系。

客户端接入细节见：

- `docs/client-protocol.md`

### 协议帧

TCP 和 WebSocket 统一使用 goim 风格二进制帧：

- 固定 16 字节头
- `ver`
- `op`
- `seq`
- `body`

当前使用的操作码：

- `OpAuth`
- `OpAuthReply`
- `OpHeartbeat`
- `OpHeartbeatReply`
- `OpSendMsg`
- `OpSendMsgReply`

约束：

- 首帧必须是 `OpAuth`
- 心跳使用 `OpHeartbeat`
- 业务命令仍然放在 `body` 中，格式仍为 JSON
- 服务端推送消息使用 `OpSendMsg`
- 命令应答使用 `OpSendMsgReply`

这样做的目的，是把传输层协议统一，同时保留现有业务命令结构，避免一次性重写全部客户端命令体。

### 连接模型

每条连接按以下方式工作：

1. reader goroutine 负责收包、鉴权、心跳、命令解析
2. writer goroutine 负责消费发送缓冲并刷 socket
3. 登录成功后把连接绑定到 `SessionManager`
4. 连接关闭时解绑会话并清理 presence

连接发送不再直接写 socket，而是先进入连接级发送缓冲，再由 writer goroutine 批量 flush。

### 发送缓冲

每条连接都维护一个固定大小的 Ring：

- 容量按 `2^N` 归整
- 通过 `rp/wp/mask` 读写
- 只允许单 writer 消费

当前行为：

- 本地在线投递先把推送帧写入目标连接 Ring
- writer goroutine 被 signal 唤醒后批量 flush
- 只要任一连接成功入队，即视为在线送达
- 所有连接都未能入队，才回落离线箱

## 7.4 内存池与分桶

为降低长连接热点路径上的分配和锁竞争，接入层做了两类优化。

### Session 分桶

`SessionManager` 使用固定数量 bucket：

- 每个 bucket 独立维护连接表和用户连接索引
- `principalKey` 哈希后落到 bucket
- `Bind / Unbind / SendToPrincipal` 只命中单 bucket

目的：

- 避免全局大锁
- 降低热点用户和普通用户之间的互相阻塞

### Buffer Pool

接入层引入轻量 buffer pool：

- 帧编码使用 `frame buffer`
- TCP 读写使用固定大小 reader/writer buffer
- buffer 通过池复用，减少高频分配

当前池化范围仅限接入层热点路径，不扩展到业务对象或存储对象。

## 7.5 接入配置

当前接入层配置如下：

- `session.bucket_count`
- `session.ring_size`
- `session.reader_buffer_size`
- `session.writer_buffer_size`
- `session.frame_buffer_size`
- `session.heartbeat_interval_seconds`
- `session.heartbeat_misses`
- `session.write_flush_interval_ms`

默认值：

```yaml
session:
  bucket_count: 64
  ring_size: 256
  reader_buffer_size: 4096
  writer_buffer_size: 4096
  frame_buffer_size: 8192
  heartbeat_interval_seconds: 30
  heartbeat_misses: 3
  write_flush_interval_ms: 5
```

默认原则：

- `ring_size` 应保持为 `2^N`
- heartbeat 超过允许次数后直接断链
- write loop 用小间隔批量 flush，优先吞吐而不是逐条即时刷写

## 8. 当前限制

当前实现已经完成接入层重构，但仍有明确边界：

- 业务命令体仍使用 JSON，而不是 protobuf
- direct message / system notice / biz push 还没有按消息类型区分不同的满队列策略
- frame buffer 已经引入 pool，但 payload 仍会在部分路径复制
- 尚未引入 room、broadcast、watch op 等 goim 高阶能力

这些限制是当前阶段的刻意取舍，优先解决连接承载和投递路径一致性。

## 9. 首期不做

- 群聊
- 频道/聊天室
- 厂商推送通道
- 独立 IM 账号体系
- 跨域会话复用
