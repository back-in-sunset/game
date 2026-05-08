# VDA gRPC API 参考

## 服务概述

VDA (Voice Data Audio) 是语音业务信令服务，负责 1v1 通话和语音房间的信令管理，签发 LiveKit WebRTC 访问 token。

**职责边界**：
- VDA 管理通话/房间的**业务信令**（发起、接听、拒接、挂断、静音）
- VDA 签发 LiveKit JWT token，客户端用该 token 直连 LiveKit 进行**媒体传输**（Opus 音频 over UDP）
- VDA **不处理**音频编码/解码/转发 — 这些由 LiveKit 负责

**调用方**：IM 服务通过 `HandleVoiceEvent` 统一入口转发客户端语音信令；其他服务可直接调用具体 RPC。

## Proto 定义

```protobuf
syntax = "proto3";
package vdaclient;
option go_package = "/vda";

service VDA {
  // 1v1 通话
  rpc InitiateCall(InitiateCallRequest) returns (InitiateCallResponse);
  rpc AcceptCall(AcceptCallRequest) returns (AcceptCallResponse);
  rpc RejectCall(RejectCallRequest) returns (ActionResponse);
  rpc EndCall(EndCallRequest) returns (ActionResponse);
  rpc CancelCall(CancelCallRequest) returns (ActionResponse);

  // 语音房间
  rpc JoinVoiceRoom(JoinVoiceRoomRequest) returns (JoinVoiceRoomResponse);
  rpc LeaveVoiceRoom(LeaveVoiceRoomRequest) returns (ActionResponse);
  rpc MuteToggle(MuteToggleRequest) returns (MuteToggleResponse);

  // IM 事件转发（统一入口）
  rpc HandleVoiceEvent(VoiceEventRequest) returns (VoiceEventResponse);

  // 状态查询
  rpc GetCallState(GetCallStateRequest) returns (CallStateResponse);
  rpc GetRoomState(GetRoomStateRequest) returns (RoomStateResponse);
}
```

---

## 公共类型

### CallerInfo

调用方身份信息，domain 为 `"platform"` 或 `"tenant"`。

```json
{
  "UserID": 1001,
  "Domain": "platform",
  "TenantID": "",
  "ProjectID": "",
  "Environment": ""
}
```

### ActionResponse

通用操作响应。

```json
{
  "Success": true,
  "Message": ""
}
```

---

## 1v1 通话 API

通话状态机：

```
ringing ──accept──▶ connected ──end──▶ ended
   │                    │
   ├──reject──▶ rejected│
   └──cancel──▶ cancelled
```

### InitiateCall

发起 1v1 通话。主叫获得 LiveKit token，被叫等待推送通知。

**请求**：

```json
{
  "Caller": {
    "UserID": 1001,
    "Domain": "platform"
  },
  "Callee": 2002
}
```

**响应**：

```json
{
  "CallID": "call_a1b2c3d4",
  "State": "ringing",
  "LiveKitRoom": "call_platform_1001_2002_a1b2c3d4",
  "LiveKitToken": "eyJhbGciOi...",
  "LiveKitUrl": "http://livekit:7880"
}
```

**错误**：
- 主叫已在通话中 → `caller already in call`
- 被叫已在通话中 → `callee already in call`

### AcceptCall

被叫接听通话，获得 LiveKit token 进入媒体房间。

**请求**：

```json
{
  "CallID": "call_a1b2c3d4",
  "UserID": 2002
}
```

**响应**：

```json
{
  "CallID": "call_a1b2c3d4",
  "State": "connected",
  "LiveKitToken": "eyJhbGciOi...",
  "LiveKitUrl": "http://livekit:7880"
}
```

**错误**：
- 通话不存在 → `call not found`
- 不是被叫 → `user is not the callee`
- 状态不是 ringing → `expected ringing`

### RejectCall

被叫拒接通话。

**请求**：

```json
{
  "CallID": "call_a1b2c3d4",
  "UserID": 2002
}
```

**响应**：

```json
{ "Success": true }
```

**错误**：
- 不是被叫 → `user is not the callee`

### EndCall

通话中任意一方挂断。

**请求**：

```json
{
  "CallID": "call_a1b2c3d4",
  "UserID": 1001
}
```

**响应**：

```json
{ "Success": true }
```

**错误**：
- 不是通话参与者 → `user is not a participant`
- 状态不允许挂断 → `cannot end in state rejected`

### CancelCall

主叫在对方接听前取消通话。

**请求**：

```json
{
  "CallID": "call_a1b2c3d4",
  "UserID": 1001
}
```

**响应**：

```json
{ "Success": true }
```

**错误**：
- 不是主叫 → `user is not the caller`

---

## 语音房间 API

### JoinVoiceRoom

加入语音房间，获得 LiveKit token。

**请求**：

```json
{
  "RoomID": "room-lobby",
  "User": {
    "UserID": 1001,
    "Domain": "platform"
  }
}
```

**响应**：

```json
{
  "RoomID": "room-lobby",
  "LiveKitToken": "eyJhbGciOi...",
  "Participants": [1001, 2002],
  "LiveKitUrl": "http://livekit:7880"
}
```

### LeaveVoiceRoom

离开语音房间。

**请求**：

```json
{
  "RoomID": "room-lobby",
  "UserID": 1001
}
```

**响应**：

```json
{ "Success": true }
```

### MuteToggle

切换麦克风静音状态。

**请求**：

```json
{
  "RoomID": "room-lobby",
  "UserID": 1001,
  "Muted": true
}
```

**响应**：

```json
{
  "RoomID": "room-lobby",
  "UserID": 1001,
  "Muted": true
}
```

---

## HandleVoiceEvent（IM 统一入口）

IM 将客户端语音信令统一转发到此 RPC，VDA 根据 `Action` 字段路由到对应的 call/room manager。

**请求**：

```json
{
  "Action": "call_invite",
  "Caller": {
    "UserID": 1001,
    "Domain": "platform"
  },
  "CallID": "",
  "RoomID": "",
  "ReceiverID": 2002,
  "PayloadJson": ""
}
```

**支持的 Action**：

| Action | 对应逻辑 | 触发条件 |
|--------|---------|---------|
| `call_invite` | `s.calls.Initiate()` | 发起通话 |
| `call_accept` | `s.calls.Accept()` | 接听通话 |
| `call_reject` | `s.calls.Reject()` | 拒接通话 |
| `call_end` | `s.calls.End()` | 挂断通话 |
| `call_cancel` | `s.calls.Cancel()` | 取消通话 |
| `room_join` | `s.rooms.Join()` | 加入房间 |
| `room_leave` | `s.rooms.Leave()` | 离开房间 |
| `mute_toggle` | `s.rooms.ToggleMute()` | 静音切换 |
| `disconnect` | `s.calls.OnDisconnect()` | IM 断线清理 |

**响应**：

```json
{
  "CallID": "call_a1b2c3d4",
  "State": "ringing",
  "LiveKitToken": "eyJhbGciOi...",
  "LiveKitRoom": "call_platform_1001_2002_a1b2c3d4",
  "LiveKitUrl": "http://livekit:7880",
  "PushJson": "",
  "PushTargets": [2002, 1001]
}
```

`PushTargets` 指示 IM 应向哪些用户推送通知。例如 `call_invite` 返回 `[callee, caller]`，IM 应向被叫推送来电通知。

---

## 状态查询 API

### GetCallState

查询通话当前状态。

**请求**：

```json
{ "CallID": "call_a1b2c3d4" }
```

**响应**：

```json
{
  "CallID": "call_a1b2c3d4",
  "State": "connected",
  "Caller": 1001,
  "Callee": 2002,
  "LiveKitRoom": "call_platform_1001_2002_a1b2c3d4"
}
```

### GetRoomState

查询房间参与者与静音状态。

**请求**：

```json
{ "RoomID": "room-lobby" }
```

**响应**：

```json
{
  "RoomID": "room-lobby",
  "Participants": [1001, 2002, 3003],
  "Muted": [2002]
}
```

---

## 错误码

所有 RPC 通过 gRPC status code 返回错误：

| gRPC Code | 场景 |
|-----------|------|
| `InvalidArgument` | 请求参数无效 |
| `NotFound` | 通话/房间不存在 |
| `FailedPrecondition` | 状态不允许当前操作（如非 ringing 状态下 Accept） |
| `PermissionDenied` | 用户无权操作（如非被叫 Accept） |
| `AlreadyExists` | 用户已在其他通话中 |
| `Internal` | Redis/LiveKit 内部错误 |
