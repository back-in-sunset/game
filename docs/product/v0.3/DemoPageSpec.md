# Demo 页面规格说明书

**版本**：v0.3
**更新时间**：2026-05-11
**基于**：`docs/product/ProductThreeHorizons.md` — 近期开发者工具形态
**状态**：已实施

---

## 1. 背景

ProductThreeHorizons 定义近期（0-6 月）产品形态为"开发者工具"。第一个关键交付物是公开 Demo 页面，用于发到 V2EX/知乎/Reddit 展示产品能力。

当前前端是一个内部调试工具（Dashboard 带硬编码帖子，ConversationRow 显示 debug 信息），不适合对外展示。

## 2. 功能范围

### 包含

1. **IM 聊天面板** — 打开即显示会话列表 + 消息面板，自动连接 IM WebSocket
2. **语音通话** — 点击按钮发起 WebRTC 语音通话，两个 tab 之间可互通
3. **Powered by 标识** — 底部 "Powered by Game Backend Cloud" 链接
4. **简洁 UI** — 不暴露内部调试面板、状态徽章、频道导航

### 不包含

- 登录/注册流程（使用内嵌只读 token）
- 好友管理、历史搜索、设置页（内部工具功能）
- 多人语音房间（只做 1v1）
- 移动端适配（桌面浏览器优先）

## 3. 页面布局

```
┌──────────────────────────────────────────────────┐
│  Game Backend Cloud — 实时通信能力演示            │
│                                                  │
│  ┌─会话列表────┐  ┌─消息面板──────────────────┐   │
│  │ Alice     ● │  │ Alice: 嗨！想讨论产品...  │   │
│  │ Support   ● │  │ You:  好的，我看看       │   │
│  │            │  │                           │   │
│  │            │  │ [输入消息...] [发送]       │   │
│  └────────────┘  └───────────────────────────┘   │
│                                                  │
│  [发起语音通话]  [挂断]  [静音]  通话时长: 00:00  │
│                                                  │
│  Powered by Game Backend Cloud                   │
└──────────────────────────────────────────────────┘
```

## 4. 技术方案

### 4.1 架构

- **不使用 WebShell** — 独立布局，不暴露频道导航、状态徽章、内部链接
- 自动连接 IM（`useIMBootstrap` 逻辑内联到 DemoPage）
- 左侧会话列表 + 右侧消息面板 + 底部通话控制栏
- 使用 hardcoded dev token，无需登录

### 4.2 IM 连接

复用现有 `imStore`：
- 页面 mount 时调用 `imStore.getState().connect(imWsUrl, { token, domain, scope })`
- 复用现有 `ConversationRow` + `MessageBubble` + `Composer` 组件
- 会话数据从 `@game/shared` 的硬编码 conversations 加载

### 4.3 语音通话

修复 VoicePage "发起呼叫" 按钮（当前是 no-op 桩）：
- Demo 场景下硬编码 LiveKit token（从 VDA 服务获取一次测试 token）
- 两个 tab 使用同一 room，各自连接 LiveKit
- 通话控制：发起、挂断、静音、时长显示

### 4.4 路由

- `/demo` — 公开 Demo 入口（IM 聊天 + 1v1 通话 + 创建多人房间）
- `/demo/room/:roomId` — 多人语音房间（参与者面板 + 房间消息 + 房间顶栏）
- `/` — 内部 Dashboard（保留不变）

## 5. 文件变更

### 5.1 Demo 页面（已实施）

| 文件 | 操作 | 说明 |
|------|------|------|
| `frontend/apps/web/src/pages/DemoPage.tsx` | 新建 | 公开 Demo 页面主体 |
| `frontend/apps/web/src/pages/DemoRoomPage.tsx` | 新建 | 多人语音房间页面 |
| `frontend/apps/web/src/store/roomStore.ts` | 新建 | 房间状态管理 |
| `frontend/packages/api/src/livekit/token.ts` | 新建 | LiveKit JWT 签发工具 |
| `frontend/apps/web/src/services/livekit.ts` | 修改 | 扩展参与者感知 |
| `frontend/apps/web/src/pages/VoicePage.tsx` | 修改 | 修复"发起呼叫"按钮 |
| `frontend/apps/web/src/app/App.tsx` | 修改 | 添加 `/demo` 和 `/demo/room/:roomId` 路由 |

## 6. 不做什么

- 不做新 UI 组件库 — 复用 `@game/ui` 现有组件
- 不做 VDA token API 集成 — demo 场景下硬编码 LiveKit token
- 不做响应式/移动端 — 桌面浏览器优先
- 不做 i18n — 中文界面
- 不做 analytics/埋点
- 不部署到公网（本 spec 只覆盖前端改动，部署另议）

## 7. 验证标准

1. `pnpm dev` → 访问 `http://localhost:5173/demo`
2. 看到会话列表（Alice / Support），对话可切换
3. 消息面板显示预置消息，可发送新消息
4. 开两个浏览器 tab，互相收发消息
5. 点击"发起语音通话" → 两个 tab 进入通话状态 → 可听到音频
6. 挂断 → 通话结束，时长显示正确
7. 点击"创建房间" → 选择 5/10 人 → 跳转 `/demo/room/{id}`
8. 复制房间链接 → 新 tab 打开 → 两人在同一房间，参与者面板显示
9. 房间顶栏显示人数、时长，静音/取消静音生效
10. 离开房间 → 返回 /demo

## 8. 后续步骤

Demo 页面完成后：

1. **Web SDK npm 包** — 从 `packages/api` 提取，发布为 `@gamecloud/web-sdk`
2. **接入文档** — 从零到首条消息的 5 分钟教程（中英双语 README）
3. **3 篇技术文章** — IM 协议设计 / LiveKit 实践 / 微服务拆分
4. **GitHub 仓库** — 开源 frontend SDK + demo 项目

## 9. 关联文档

- 产品三态规划：`docs/product/ProductThreeHorizons.md`
- 商业化分析：`docs/product/attention.md`
- v0.2 PRD：`docs/product/v0.2/PRD.md`
