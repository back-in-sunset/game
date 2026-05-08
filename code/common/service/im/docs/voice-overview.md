# 语音功能产品概览

## 功能简介

IM 语音功能支持两类场景：

| 场景 | 说明 | 参与者 |
|------|------|--------|
| **1v1 语音通话** | 类似微信语音通话，一对一实时音频 | 2 人 |
| **语音房间** | 类似 Discord 语音频道，多人实时音频 | 最多 20 人 |

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                        客户端 App                            │
│  ┌─────────────────┐          ┌──────────────────┐         │
│  │   IM SDK         │          │  LiveKit SDK      │         │
│  │  (TCP/WebSocket) │          │  (WebRTC/UDP)     │         │
│  └───────┬─────────┘          └────────┬──────────┘         │
└──────────┼──────────────────────────────┼───────────────────┘
           │                              │
           │  业务信令（信令）               │  媒体信令（音频）
           │  call_invite / accept / end   │  Opus over UDP
           │  room_join / leave / mute     │  WebRTC SDP/ICE
           │                              │
     ┌─────▼──────┐  gRPC   ┌───────┐    │
     │  IM Server │◄───────►│  VDA  │    │
     │  :8091     │         │ :9101 │────┼── LiveKit API ──┐
     └────────────┘         └───┬───┘    │                  │
                                │        │         ┌───────▼────────┐
                                │        │         │  LiveKit Server │
                                │        └────────►│  :7880 (API)    │
                                │                  │  :7881 (WS)     │
                                │                  │  :50000+ (UDP)  │
                                │                  └────────────────┘
                                │
                          ┌─────▼──────┐
                          │   Redis    │
                          │   :6379    │
                          └────────────┘
```

**双层信令**：

- **业务信令** — 通过 IM 现有的 WebSocket/TCP 长连接，复用 `OpSendMsg`，新增 `call_*` / `room_*` / `mute_toggle` action
- **媒体信令** — 通过 LiveKit SDK 直连 LiveKit 服务器，WebRTC + Opus 音频编码，UDP 传输

这种设计的优势：客户端只需维护一条 IM 长连接即可完成所有业务操作；音频数据走 LiveKit 专有通道，不经过 IM/VDA，延迟更低。

## 功能列表

### 1v1 通话

| 功能 | Action | 说明 |
|------|--------|------|
| 发起通话 | `call_invite` | 主叫发起，获得 LiveKit token 立刻进入媒体房间 |
| 接听通话 | `call_accept` | 被叫接听，获得 LiveKit token 进入媒体房间 |
| 拒接来电 | `call_reject` | 被叫拒接，主叫收到拒接通知 |
| 取消通话 | `call_cancel` | 主叫在对方接听前取消 |
| 挂断通话 | `call_end` | 任意一方主动挂断 |

**状态机**：

```
ringing (振铃中)
   ├── accept ──▶ connected (通话中) ── end ──▶ ended (已挂断)
   ├── reject ──▶ rejected (已拒接)
   └── cancel ──▶ cancelled (已取消)
```

**约束**：
- 一个用户同时只能处于一通 1v1 通话中
- 主叫/被叫已在通话中时，新的通话请求会被拒绝
- 主叫在振铃期间断线 → 通话自动取消
- 被叫在通话中断线 → 对方可继续挂断

### 语音房间

| 功能 | Action | 说明 |
|------|--------|------|
| 加入房间 | `room_join` | 加入指定房间，获得 LiveKit token |
| 离开房间 | `room_leave` | 离开房间，从参与者列表移除 |
| 静音切换 | `mute_toggle` | 切换本人麦克风静音/取消静音 |

**约束**：
- 房间无需提前创建，第一个加入者自动创建
- 最后一个人离开后房间自动销毁
- 参与者上限 20 人（当前无硬限制，生产环境需增加）
- 房间没有鉴权，任何已登录用户可加入任意房间

## SDK 依赖

客户端需要同时集成两个 SDK：

| SDK | 用途 | 链接 |
|-----|------|------|
| IM SDK | TCP/WebSocket 长连接，文本消息 + 语音信令 | 自研（参照 client-protocol.md 实现） |
| LiveKit SDK | WebRTC 音频采集/播放/传输 | [livekit.io/sdks](https://docs.livekit.io/home/client-sdks/) |

LiveKit 官方支持 iOS / Android / Web / Flutter / React Native / Unity。

## 与文本消息的关系

- 语音信令和文本消息共享**同一条 IM 长连接**
- 语音信令使用与 `send` 消息相同的 `OpSendMsg` 操作码
- 通过 `action` 字段区分：`send` = 发文本，`call_invite` = 语音通话
- 语音通话期间，文本消息照常收发，互不影响
- 用户在线状态（IM presence）与语音状态（VDA voice presence）独立管理

## 当前限制

- 1v1 通话不支持升级为多人通话
- 语音房间不支持密码/白名单/管理员等权限控制
- LiveKit token 不可续期，超长通话需挂断重拨（当前 token 有效期 24h）
- 静音状态仅存储在 VDA，不与 LiveKit server 同步
- 不支持通话中的屏幕共享 / 视频 / 文字聊天
- 房间参与者信息通过 IM push 通知，暂不支持主动查询房间成员列表（需调用 `GetRoomState` RPC）
