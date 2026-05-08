# 语音客户端集成指南

本文档面向客户端开发者（iOS / Android / Web / Flutter），说明如何从零开始集成语音功能。

## 架构概览

语音功能采用**双层信令**架构：

```
移动端 / Web 端
    │
    ├── IM SDK（业务信令）
    │   └── TCP/WebSocket 长连接 → IM Server → gRPC → VDA
    │       作用：发起通话、接听、拒接、挂断、房间加入/离开、静音
    │
    └── LiveKit SDK（媒体信令）
        └── WebRTC/UDP 直连 → LiveKit Server
            作用：音频采集、Opus 编码、网络传输、音频播放
```

**核心理念**：IM 长连接负责"打电话"，LiveKit 负责"通电话"。两者互不依赖传输层。

---

## 集成步骤

### Step 1: 建立 IM 长连接

首先需要完成 IM 登录认证，建立 TCP 或 WebSocket 长连接。详见 [IM 客户端接入协议](../im/docs/client-protocol.md)。

```pseudocode
// 1. 建立 TCP/WebSocket 连接
conn = connect("tcp://im.example.com:8091")

// 2. 发送登录鉴权
send({
  op: OpAuth,
  body: {
    token: userJWT,
    domain: "platform",
    scope: { tenant_id: "", project_id: "", environment: "" }
  }
})

// 3. 等待登录成功
resp = recv()
assert resp.op == OpAuthReply
assert resp.body.type == "login_ok"
```

### Step 2: 通过 IM 发送语音信令

登录成功后，所有语音信令通过 `OpSendMsg` 发送，`action` 字段区分命令类型。

```pseudocode
function sendVoiceCommand(action, data):
    send({
        op: OpSendMsg,
        body: {
            action: action,
            data: data
        }
    })
    resp = recv()  // 等待 OpSendMsgReply
    return resp.body
```

### Step 3: 提取 LiveKit 凭证

语音命令的 ACK 响应中包含 LiveKit 连接所需的三要素：

```json
{
  "livekit_url": "wss://livekit.example.com",
  "livekit_token": "eyJhbGciOiJIUzI1NiIs...",
  "livekit_room": "call_platform_1001_2002_a1b2c3d4"
}
```

### Step 4: 使用 LiveKit SDK 连接媒体房间

```pseudocode
import LiveKit  // 官方 SDK: livekit-client-sdk

lkRoom = LiveKit.Room()
await lkRoom.connect(
    url:    ack.livekit_url,
    token:  ack.livekit_token
)
// 连接成功后自动开始音频采集和播放
```

---

## 完整流程示例

### 1v1 语音通话

```pseudocode
// ===== 主叫方 (User 1001) =====

// Step 1: 发起通话
ack = sendVoiceCommand("call_invite", { callee: 2002 })
// ack = {
//   type: "call_invite_ack",
//   call_id: "call_abc",
//   state: "ringing",
//   livekit_token: "<caller-token>",
//   livekit_room: "call_platform_1001_2002_abc",
//   livekit_url: "wss://livekit.example.com"
// }

// Step 2: 主叫立刻连接 LiveKit 等待
lkRoom = LiveKit.Room()
await lkRoom.connect(ack.livekit_url, ack.livekit_token)
// 此时房间内只有主叫，等待被叫加入

// Step 3: 等待被叫接听（通过 IM push 通知）
push = recv()  // type: "call_accept"

// Step 4: LiveKit 自动检测到新参与者，音频开始传输
lkRoom.onParticipantConnected = (p) => {
    showToast("对方已加入")
}
// Opus 音频通过 WebRTC 自动传输，无需手动处理

// Step 5: 挂断
sendVoiceCommand("call_end", { call_id: "call_abc" })
lkRoom.disconnect()


// ===== 被叫方 (User 2002) =====

// Step 1: 收到来电推送
push = recv()
// push.type = "call_invite"
// push.call_id = "call_abc"
// push.caller = 1001
showIncomingCallUI(caller: 1001, callID: "call_abc")

// Step 2: 用户点击接听
ack = sendVoiceCommand("call_accept", { call_id: "call_abc" })
// ack.livekit_token = "<callee-token>"
// ack.livekit_url = "wss://livekit.example.com"

// Step 3: 连接 LiveKit
lkRoom = LiveKit.Room()
await lkRoom.connect(ack.livekit_url, ack.livekit_token)
// 音频自动开始传输

// 如果用户点击拒接
sendVoiceCommand("call_reject", { call_id: "call_abc" })
// 主叫方会收到 push: { type: "call_reject" }
```

### 语音房间

```pseudocode
// Step 1: 加入房间
ack = sendVoiceCommand("room_join", { room_id: "room-lobby" })
// ack = {
//   type: "room_join_ack",
//   livekit_token: "<token>",
//   livekit_room: "room-lobby",
//   livekit_url: "wss://livekit.example.com"
// }

// Step 2: 连接 LiveKit
lkRoom = LiveKit.Room()
await lkRoom.connect(ack.livekit_url, ack.livekit_token)

// Step 3: 其他参与者加入时会收到 push
lkRoom.onParticipantConnected = (p) => {
    updateParticipantList()
}

// Step 4: 静音切换
sendVoiceCommand("mute_toggle", {
    room_id: "room-lobby",
    muted: true
})
// 本地麦克风也需要同步静音
lkRoom.localParticipant.setMicrophoneEnabled(false)

// Step 5: 离开房间
sendVoiceCommand("room_leave", { room_id: "room-lobby" })
lkRoom.disconnect()
```

---

## LiveKit SDK 最小接入

### iOS (Swift)

```swift
import LiveKit

let room = Room()
try await room.connect(
    url: ack.livekitUrl,
    token: ack.livekitToken
)
// 自动开始音频采集和播放
```

依赖：`https://github.com/livekit/client-sdk-swift`

### Android (Kotlin)

```kotlin
import livekit.LiveKit

val room = LiveKit.create(applicationContext)
room.connect(
    url = ack.livekitUrl,
    token = ack.livekitToken
)
```

依赖：`https://github.com/livekit/client-sdk-android`

### Web (JavaScript)

```javascript
import { Room } from 'livekit-client'

const room = new Room()
await room.connect(
    ack.livekit_url,
    ack.livekit_token
)
```

依赖：`npm install livekit-client`

### Flutter

```dart
import 'package:livekit_client/livekit_client.dart';

final room = Room();
await room.connect(
    ack.livekitUrl,
    ack.livekitToken,
);
```

依赖：`https://pub.dev/packages/livekit_client`

---

## 错误处理

### 通话冲突

如果用户已在其他通话中，发起新通话会失败：

```json
{ "error": "caller already in call" }
```

**客户端处理**：提示"您正在通话中，请先挂断当前通话"。

### 对方已在通话中

```json
{ "error": "callee already in call" }
```

**客户端处理**：提示"对方正在通话中"。

### VDA 服务不可用

IM 会返回 gRPC 错误：

```json
{ "error": "vda voice event: Unavailable" }
```

**客户端处理**：提示"语音服务暂不可用，请稍后重试"。

### LiveKit 连接失败

```javascript
room.on(RoomEvent.Disconnected, (reason) => {
    // reason 可能是网络问题或 token 过期
    // 重新获取 token 后重连
})
```

---

## 断线处理

### IM 长连接断开

- IM 断开后，VDA 自动清理语音状态：
  - 主叫在振铃阶段断线 → 通话被自动取消
  - 通话中一方断线 → 清除该用户的在线状态，对方仍可挂断
- **客户端**应在 IM 重连后重新加入 LiveKit 房间（如果通话仍在进行中）

### WebRTC 连接断开

- LiveKit SDK 会自动重连（ice restart）
- 如果重连失败，SDK 会触发 `Disconnected` 事件
- 客户端应调用 `call_end` 或 `room_leave` 清理状态

---

## 注意事项

- LiveKit token 有时效性（默认 24 小时），长时间通话中可能需要续期（当前 VDA 不支持 token refresh，需挂断重拨）
- 语音房间当前没有鉴权（任何已登录用户可加入任意房间），生产环境需增加房间成员校验
- 静音状态存储在 VDA Redis，不与 LiveKit server 同步 — 重连后静音状态可能丢失
