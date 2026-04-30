# IM 客户端接入协议

## 1. 目标

本文档面向 IM 客户端、网关 SDK、联调方，说明当前长连接接入协议。

当前协议适用于：

- TCP 长连接
- WebSocket 长连接

两种传输层都使用同一套二进制帧格式，业务命令体仍为 JSON。

## 2. 帧格式

每个消息帧由固定 16 字节头和可变长 body 组成。

### 2.1 Header

按大端序编码：

| 字段 | 长度 | 说明 |
| --- | --- | --- |
| pack_len | 4 | 整个帧长度，包含 header + body |
| header_len | 2 | 固定为 `16` |
| ver | 2 | 当前固定为 `1` |
| op | 4 | 操作码 |
| seq | 4 | 请求序号，客户端自增即可 |

### 2.2 Body

- body 当前是 UTF-8 JSON 字节串
- 最大 body 长度当前限制为 `64KB`

## 3. 操作码

| 名称 | 值 | 方向 | 说明 |
| --- | --- | --- | --- |
| `OpHeartbeat` | `2` | C -> S | 心跳 |
| `OpHeartbeatReply` | `3` | S -> C | 心跳响应 |
| `OpSendMsg` | `4` | 双向 | 客户端命令 / 服务端推送 |
| `OpSendMsgReply` | `5` | S -> C | 命令应答 |
| `OpAuth` | `7` | C -> S | 登录鉴权 |
| `OpAuthReply` | `8` | S -> C | 登录成功应答 |
| `OpDisconnectReply` | `6` | S -> C | 错误应答 |

约束：

- 建连后首帧必须是 `OpAuth`
- 登录成功前不接受其他业务帧
- 心跳由客户端主动发送

## 4. 登录

### 4.1 请求

`op = OpAuth`

```json
{
  "token": "jwt-token",
  "domain": "platform",
  "scope": {
    "tenant_id": "",
    "project_id": "",
    "environment": ""
  }
}
```

字段说明：

- `token`: 现有 user 服务签发的 RSA JWT
- `domain`: `platform` 或 `tenant`
- `scope`: 当 `domain=tenant` 时必须完整提供

### 4.2 成功响应

`op = OpAuthReply`

```json
{
  "type": "login_ok",
  "user_id": 1001,
  "domain": "platform",
  "scope": {
    "tenant_id": "",
    "project_id": "",
    "environment": ""
  }
}
```

### 4.3 失败响应

`op = OpDisconnectReply`

```json
{
  "error": "invalid token"
}
```

## 5. 心跳

### 5.1 请求

`op = OpHeartbeat`

body 可以为空，也可以发送任意占位 JSON；当前服务端不依赖 body 内容。

### 5.2 响应

`op = OpHeartbeatReply`

body 为当前 UTC 时间字符串，例如：

```json
"2026-04-30T08:30:00Z"
```

建议：

- 客户端每 `30s` 发一次心跳
- 连续超过服务端允许的超时窗口后，连接可能被断开

## 6. 业务命令

所有业务命令都通过：

- `op = OpSendMsg`

统一承载，body 结构如下：

```json
{
  "action": "send",
  "data": {}
}
```

### 6.1 发送消息

```json
{
  "action": "send",
  "data": {
    "receiver": 2002,
    "msg_type": "direct_message",
    "seq": 101,
    "payload": {
      "text": "hello"
    }
  }
}
```

支持的 `msg_type`：

- `direct_message`
- `system_notice`
- `biz_push`

响应：

`op = OpSendMsgReply`

```json
{
  "type": "send_ack",
  "seq": 101,
  "online_recipients": 1,
  "stored_offline": false
}
```

### 6.2 拉取会话列表

```json
{
  "action": "list_conversations"
}
```

响应：

```json
{
  "type": "conversation_list",
  "conversations": []
}
```

### 6.3 拉取消息列表

```json
{
  "action": "list_messages",
  "data": {
    "peer_user_id": 2002,
    "limit": 20
  }
}
```

响应：

```json
{
  "type": "message_list",
  "messages": []
}
```

### 6.4 标记已读

```json
{
  "action": "mark_read",
  "data": {
    "peer_user_id": 2002,
    "seq": 101
  }
}
```

响应：

```json
{
  "type": "read_ack",
  "peer_user_id": 2002,
  "seq": 101
}
```

## 7. 服务端推送

服务端在线推送使用：

- `op = OpSendMsg`

body 结构：

```json
{
  "type": "message",
  "envelope": {
    "domain": "platform",
    "scope": {
      "tenant_id": "",
      "project_id": "",
      "environment": ""
    },
    "sender": 1001,
    "receiver": 2002,
    "msg_type": "direct_message",
    "seq": 101,
    "payload": {
      "text": "hello"
    },
    "sent_at": "2026-04-30T08:30:00Z"
  }
}
```

离线补发使用相同 opcode，body 类型为：

```json
{
  "type": "offline_batch",
  "messages": []
}
```

## 8. 错误处理

当前错误响应统一使用：

- `op = OpDisconnectReply`

body：

```json
{
  "error": "unsupported action \"xxx\""
}
```

客户端应按以下方式处理：

- 登录阶段错误：直接视为登录失败并断开
- 命令阶段错误：保留连接，但当前请求失败
- 读到未知 `op`：记录日志并忽略或关闭连接

## 9. 当前限制

- 业务 body 仍是 JSON，不是 protobuf
- 不支持群聊、房间、广播
- body 体积必须控制在 64KB 以内
- direct message、system notice、biz push 目前在协议层没有额外字段差异，只通过 `msg_type` 区分

## 10. 最小联调客户端

仓库内提供了一个最小 TCP 联调客户端：

- `cmd/tcpdemo`

示例：

```bash
rtk go run ./cmd/tcpdemo \
  -addr 127.0.0.1:8091 \
  -token '<jwt>' \
  -domain platform \
  -action list_conversations
```

发送私信示例：

```bash
rtk go run ./cmd/tcpdemo \
  -addr 127.0.0.1:8091 \
  -token '<jwt>' \
  -domain platform \
  -action send \
  -data '{"receiver":2002,"msg_type":"direct_message","seq":101,"payload":{"text":"hello"}}'
```
