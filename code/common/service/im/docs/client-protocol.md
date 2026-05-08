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

## 7. 语音信令

语音信令复用 `OpSendMsg`，通过 `action` 字段区分命令类型。客户端发送语音 action → IM 转发至 VDA → 返回 LiveKit 凭证 → 客户端使用 LiveKit SDK 建立 WebRTC 媒体连接。

### 7.1 双层信令模型

```
客户端 ◀── IM 长连接（业务信令）──▶ IM Server ◀── gRPC ──▶ VDA ◀── HTTP ──▶ LiveKit
   │                                                                              │
   ╰══════════════════ WebRTC/Opus/UDP（媒体信令）══════════════════════════════════╯
```

- **业务信令**：通话建立/拆除/静音，通过 IM 的 TCP/WebSocket 连接
- **媒体信令**：音频编解码、网络穿透，通过 LiveKit SDK 直连 LiveKit 服务器

### 7.2 语音命令列表

| Action | 方向 | 说明 | data 字段 |
|--------|------|------|-----------|
| `call_invite` | C→S | 发起 1v1 语音通话 | `{"callee": 2002}` |
| `call_accept` | C→S | 接听来电 | `{"call_id": "call_xxx"}` |
| `call_reject` | C→S | 拒接来电 | `{"call_id": "call_xxx"}` |
| `call_end` | C→S | 挂断当前通话 | `{"call_id": "call_xxx"}` |
| `call_cancel` | C→S | 取消发出的通话（对方未接前） | `{"call_id": "call_xxx"}` |
| `room_join` | C→S | 加入语音房间 | `{"room_id": "room-xxx"}` |
| `room_leave` | C→S | 离开语音房间 | `{"room_id": "room-xxx"}` |
| `mute_toggle` | C→S | 切换麦克风静音 | `{"room_id": "room-xxx", "muted": true}` |

### 7.3 发起通话

请求：

```json
{
  "action": "call_invite",
  "data": {
    "callee": 2002
  }
}
```

响应 `OpSendMsgReply`：

```json
{
  "type": "call_invite_ack",
  "call_id": "call_a1b2c3d4",
  "state": "ringing",
  "livekit_token": "eyJhbGciOi...",
  "livekit_room": "call_platform_1001_2002_a1b2c3d4",
  "livekit_url": "http://livekit:7880"
}
```

主叫收到 ACK 后可直接使用 `livekit_token` + `livekit_url` 连接 LiveKit 房间等待对方加入。

### 7.4 被叫收到来电推送

服务端推送 `OpSendMsg`：

```json
{
  "type": "call_invite",
  "call_id": "call_a1b2c3d4",
  "caller": 1001,
  "livekit_room": "call_platform_1001_2002_a1b2c3d4"
}
```

被叫客户端收到后展示来电 UI。被叫接听后服务端会再推送 `call_accept`，其中包含被叫的 `livekit_token`。

### 7.5 接听通话

请求：

```json
{
  "action": "call_accept",
  "data": {
    "call_id": "call_a1b2c3d4"
  }
}
```

响应：

```json
{
  "type": "call_accept_ack",
  "call_id": "call_a1b2c3d4",
  "state": "connected",
  "livekit_token": "eyJhbGciOi...",
  "livekit_url": "http://livekit:7880"
}
```

### 7.6 拒接 / 取消 / 挂断

请求格式相同，`data` 携带 `call_id`：

```json
{ "action": "call_reject", "data": { "call_id": "call_a1b2c3d4" } }
{ "action": "call_cancel", "data": { "call_id": "call_a1b2c3d4" } }
{ "action": "call_end",    "data": { "call_id": "call_a1b2c3d4" } }
```

响应均为：

```json
{ "type": "call_reject_ack", "state": "rejected" }
{ "type": "call_cancel_ack", "state": "cancelled" }
{ "type": "call_end_ack",    "state": "ended"    }
```

### 7.7 加入 / 离开语音房间

加入请求：

```json
{
  "action": "room_join",
  "data": {
    "room_id": "room-lobby"
  }
}
```

加入响应：

```json
{
  "type": "room_join_ack",
  "livekit_token": "eyJhbGciOi...",
  "livekit_room": "room-lobby",
  "livekit_url": "http://livekit:7880"
}
```

加入房间后，服务端向房间内其他成员推送 `room_participant_joined`（当前通过 VDA push_targets 机制）。

离开请求：

```json
{
  "action": "room_leave",
  "data": {
    "room_id": "room-lobby"
  }
}
```

离开响应：

```json
{ "type": "room_leave_ack" }
```

### 7.8 静音切换

请求：

```json
{
  "action": "mute_toggle",
  "data": {
    "room_id": "room-lobby",
    "muted": true
  }
}
```

响应：

```json
{ "type": "mute_toggle_ack" }
```

### 7.9 语音通话时序图

```
主叫 App              IM Server           VDA              被叫 App
   │                     │                  │                   │
   │──call_invite───────▶│                  │                   │
   │                     │──HandleVoiceEvent▶│                   │
   │                     │◀─VoiceEventResp──│                   │
   │◀──call_invite_ack──│                  │                   │
   │  (token + url +     │                  │                   │
   │   room)             │                  │                   │
   │                     │──push: call_invite──────────────────▶│
   │                     │                  │                   │
   │══ 连接 LiveKit ═══════════════════════════════════════════│
   │                     │                  │                   │
   │                     │◀──call_accept───────────────────────│
   │                     │──HandleVoiceEvent▶│                   │
   │                     │◀─VoiceEventResp──│                   │
   │                     │──push: call_accept (token)─────────▶│
   │                     │                  │                   │
   │                     │                  │    被叫连接 LiveKit │
   │══════════════ Opus 音频 over UDP ═══════════════════════════│
   │                     │                  │                   │
   │──call_end─────────▶│                  │                   │
   │                     │──HandleVoiceEvent▶│                   │
   │◀──call_end_ack─────│                  │                   │
   │                     │──push: call_end────────────────────▶│
```

### 7.10 语音房间时序图

```
用户 A              IM Server           VDA              用户 B
   │                     │                  │                   │
   │──room_join────────▶│                  │                   │
   │                     │──HandleVoiceEvent▶│                   │
   │◀──room_join_ack────│                  │                   │
   │  (token + url)      │                  │                   │
   │                     │──push: room_participant_joined─────▶│
   │══ 连接 LiveKit ═══════════════════════════════════════════│
   │                     │                  │                   │
   │                     │◀──room_join─────────────────────────│
   │                     │──HandleVoiceEvent▶│                   │
   │◀──room_participant_joined (push)──────────│                   │
   │                     │──push: room_participant_joined─────▶│
   │                     │                  │     B 连接 LiveKit │
   │══════════════ Opus 音频 over UDP ═══════════════════════════│
   │                     │                  │                   │
   │──room_leave───────▶│                  │                   │
   │◀──room_leave_ack───│                  │                   │
   │                     │──push: room_participant_left───────▶│
```

## 8. 服务端推送

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

## 9. 错误处理

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

## 10. 当前限制

- 业务 body 仍是 JSON，不是 protobuf
- 文本消息不支持群聊、广播（仅 1v1 私信）
- body 体积必须控制在 64KB 以内
- direct message、system notice、biz push 目前在协议层没有额外字段差异，只通过 `msg_type` 区分
- 语音通话仅支持 1v1，不支持多人通话
- 语音房间参与者上限 20 人
- 语音信令推送（call_invite/room_participant_joined 等）当前通过 `OpSendMsg` 推送，push_targets 机制仍在完善中
- 语音房间 mute 状态不与 LiveKit server 同步，仅 VDA 内部记录

## 11. 最小联调客户端

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
