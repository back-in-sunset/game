以下是根据我们此前全部讨论内容（游戏后端、社区、IM、语音消息、自建信令服务器、Godot对接）及您最终指定的技术栈（APISIX Ingress、Golang+go-zero、ScyllaDB、Redis、Kafka）形成的 **完整 PRD v0.1 文档**。本版本专注于可交付的 MVP，包含所有必要细节，可直接用于开发团队实施。

---

# 产品需求文档（PRD）v0.1 完整版

**产品名称**：星域·共鸣（Stellar Resonance）  
**版本**：v0.1（MVP - 最小可行产品）  
**发布日期**：2026-04-23  
**负责人**：产品委员会  

---

## 1. 产品定位与目标

### 1.1 产品定位
一款 **内置强社交能力的多人对战游戏**，玩家无需切换应用即可完成组队、聊天、分享动态、收发语音消息，实现“游戏即社区”的沉浸式体验。

### 1.2 MVP 核心目标
- ✅ 实现 **1v1 快速匹配与帧同步战斗**（基础对战闭环）  
- ✅ 提供 **好友系统 + 文字私聊**（含离线消息）  
- ✅ 提供 **社区动态广场**（发帖、点赞、评论）  
- ✅ 支持 **语音消息**（录制发送、云端存储、离线播放）  
- ✅ 搭建 **自建语音信令服务器**（为后续实时语音房间奠定基础）  

### 1.3 非目标（v0.1 不做）
- ❌ 5v5 团战、战队系统、排行榜  
- ❌ 实时语音房间（开黑，留待 v0.2）  
- ❌ 直播、视频分享、商业化（付费/广告）  
- ❌ 复杂赛事系统、UGC 地图编辑器  

---

## 2. 用户角色与核心流程

| 角色 | 说明 | 核心权限 |
|------|------|----------|
| 游客 | 未登录，可浏览公开动态 | 只读社区（无法点赞评论），不能游戏、聊天 |
| 注册用户 | 完成注册/登录 | 全部核心功能：对战、好友、聊天、发动态、语音消息 |
| 管理员 | 后台运营 | 动态审核、用户禁言、敏感词管理 |

**核心用户旅程**（MVP）：
1. 注册/登录 → 完善昵称/头像  
2. 开始匹配 → 进入1v1战斗 → 结算并显示结果  
3. 添加好友 → 发送文字/语音消息  
4. 发布动态 → 浏览广场 → 点赞/评论他人动态  

---

## 3. 功能详述（MVP）

### 3.1 用户与账号模块
| 功能点 | 描述 | 验收标准 |
|--------|------|----------|
| 注册 | 手机号/邮箱 + 密码 + 昵称 | 必填项校验；验证码可选（MVP 可简化） |
| 登录 | 账号/密码 → 返回 JWT token | token 有效期 7 天；支持刷新 token |
| 个人主页 | 展示昵称、头像、段位（固定“新秀”）、我的动态列表 | 头像支持默认图片；段位文案显示正确 |
| 修改资料 | 修改昵称、上传头像 | 头像上传至 MinIO，更新 ScyllaDB |

### 3.2 游戏核心模块

#### 3.2.1 匹配系统
- **匹配模式**：1v1，基于简单 ELO 分（初始 1200）  
- **匹配算法**：优先分差 ≤ 100 的对手；等待超过 30 秒，逐步放宽至 200 分差  
- **匹配队列**：使用 Redis 存储（`match:queue:1v1`），先进先出  
- **超时处理**：匹配超过 60 秒，弹出提示让玩家选择继续或取消  

#### 3.2.2 战斗房间
- **同步方式**：帧同步（帧率 15 fps），服务器仅转发操作指令，客户端各自演算  
- **房间生命周期**：匹配成功创建房间（等待 5 秒）→ 对局进行（最长 2 分钟）→ 结算 → 销毁房间  
- **断线重连**：60 秒内可重连，期间 AI 托管（简单移动/攻击）；超过 60 秒判负  
- **操作指令**：客户端每帧上报输入（方向、攻击键），服务器收集后广播给双方  

#### 3.2.3 战斗结算
- **结算数据**：胜负、评分（击杀/死亡/伤害简版）  
- **ELO 变化**：胜方 +20，负方 -20（平局不变化）  
- **奖励**：经验值（后续用于升级，MVP 仅展示）  

### 3.3 社区模块

#### 3.3.1 动态广场
- **发布动态**：文字（≤ 200 字）+ 最多 4 张图片（每个 ≤ 2MB）  
- **互动**：点赞、评论（一级评论，不支持回复）  
- **展示顺序**：按发布时间倒序（暂不做热度排序）  
- **敏感词过滤**：发布时同步检测（调用第三方 API 或本地词库），命中则拒绝发布并提示  

#### 3.3.2 个人动态管理
- 用户可删除自己发布的动态  
- 管理员可删除任何动态（后台预留接口）  

### 3.4 IM 系统（文字 + 语音消息）

#### 3.4.1 文字私聊
- **发送**：文本 + emoji（不限制长度，但 UI 建议≤500 字）  
- **接收**：在线用户通过 WebSocket 实时推送；离线用户存入 ScyllaDB 离线消息表  
- **已读状态**：消息送达即标记为已读（简化版，不实现已读回执）  
- **历史消息**：拉取最近 100 条（分页）  

#### 3.4.2 语音消息
- **录制**：长按麦克风按钮录制（最长 30 秒），松手自动发送  
- **格式处理**：客户端录音（AAC/M4A）→ 上传至 `media-service` → 转码为 opus（降低带宽）→ 存储至 MinIO → 返回 `file_id`  
- **消息发送**：IM 服务收到 `{type:"voice", file_id, duration}` 后，转发给接收方，同时存储元数据到 ScyllaDB  
- **播放**：接收方点击消息 → `media-service` 生成预签名 URL → 下载并播放  
- **离线语音**：离线用户上线后拉取消息列表，按需下载播放  

### 3.5 自建语音信令服务器（基础版）

**目标**：为下一版本的实时语音房间（P2P/SFU）提供信令能力，MVP 阶段仅完成**房间管理、成员列表、SDP/ICE 转发**框架，不实际承载音频流。

| 功能 | 描述 |
|------|------|
| 房间绑定 | 每个战斗房间自动创建对应的信令房间（通过 Kafka 事件 `RoomCreated` 触发） |
| 成员加入/离开 | 玩家进入战斗房间时自动加入信令房间，离开时退出 |
| 信令转发 | 转发 WebRTC 的 offer/answer/ice-candidate 消息（为后续 P2P 直连预留） |
| 接口协议 | WebSocket（JSON 格式） |
| 存储 | Redis（`signaling:room:{roomId}` 存储成员列表） |

**注意**：MVP 不开启实际的 WebRTC 连接，信令服务器只做消息透传和房间管理，便于后续无缝升级。

---

## 4. 非功能需求

| 项目 | 目标指标 | 说明 |
|------|----------|------|
| 匹配耗时 | P99 ≤ 10 秒 | 从点击匹配到进入房间 |
| 战斗操作延迟 | 端到端 ≤ 100 ms（国内同区域） | 帧同步指令转发 |
| IM 文字消息 | 在线延迟 ≤ 300 ms | 从发送到接收显示 |
| 语音消息下载 | 首字节 ≤ 1 秒 | 预签名 URL + CDN 加速（可选） |
| 系统可用性 | 99.5% | 每月允许约 3.6 小时不可用 |
| 并发支持 | 500 CCU（同时在线对战） | 单集群支撑 |
| 数据一致性 | 最终一致性（允许短暂延迟） | 动态点赞计数、ELO 更新可容忍几秒延迟 |

---

## 5. 版本规划（Roadmap）

| 版本 | 功能增量 | 预计工期 |
|------|----------|----------|
| **v0.1 (MVP)** | 匹配+1v1战斗+好友+文字私聊+语音消息+动态广场+信令服务器基础 | 8 周 |
| v0.2 | 实时语音房间（SFU，集成 LiveKit）、战队系统、排行榜、动态热度排序 | 6 周 |
| v0.3 | 5v5 团战、观战模式、语音转文字、群聊 | 6 周 |
| v1.0 | 赛事系统、UGC 地图编辑器、商业化（通行证/会员） | 8 周 |

---

## 6. 技术选型（完整版）

### 6.1 总体架构
- **微服务架构**：所有业务服务用 **Golang + go-zero** 框架构建。  
- **API 网关**：**APISIX Ingress**（部署于 K8s），负责路由、JWT 认证、限流、可观测性。  
- **服务发现与配置**：**etcd**（APISIX 与 go-zero 共用集群）。  
- **异步消息**：**Kafka**，解耦战斗结算 → 社区动态、IM 离线消息、数据统计。  
- **部署环境**：Kubernetes（v1.26+），公有云（国内可用区）。

### 6.2 数据存储
| 组件 | 用途 | 部署方式 |
|------|------|----------|
| **ScyllaDB** | 主力持久化：用户、好友、动态、评论、聊天记录、语音消息元数据 | 3 节点集群，一致性级别 QUORUM |
| **Redis** | 缓存（用户信息、未读数）、匹配队列、房间临时状态、WebSocket 路由 | 哨兵模式（3 副本）或集群 |
| **MinIO** | 对象存储：语音文件、动态图片 | 单机（开发）或分布式（生产） |
| **Kafka** | 消息队列：`battle_finished`, `im_offline`, `post_like` 等 | 3 节点集群，保留 7 天 |

#### ScyllaDB 核心表设计（CQL）

```cql
-- 用户表
CREATE TABLE users (
    user_id uuid PRIMARY KEY,
    nickname text,
    avatar_url text,
    elo int,
    created_at timestamp,
    updated_at timestamp
);

-- 好友关系（双向，user_id1 < user_id2）
CREATE TABLE friendships (
    user_id1 uuid,
    user_id2 uuid,
    status text,  -- 'pending', 'accepted', 'blocked'
    created_at timestamp,
    PRIMARY KEY ((user_id1), user_id2)
);

-- 动态表（posts）
CREATE TABLE posts (
    post_id uuid,
    user_id uuid,
    content text,
    images list<text>,
    like_count int,
    comment_count int,
    created_at timestamp,
    PRIMARY KEY (post_id, created_at)
) WITH CLUSTERING ORDER BY (created_at DESC);

-- 评论表
CREATE TABLE comments (
    post_id uuid,
    comment_id uuid,
    user_id uuid,
    content text,
    created_at timestamp,
    PRIMARY KEY ((post_id), created_at, comment_id)
) WITH CLUSTERING ORDER BY (created_at DESC);

-- 聊天记录（按发送者+接收者分桶，避免大分区）
CREATE TABLE im_messages (
    conversation_id text,  -- "userA:userB" 按字典序拼接
    msg_id timeuuid,
    from_id uuid,
    to_id uuid,
    type text,  -- 'text', 'voice'
    content text,
    voice_file_id text,
    duration int,
    status text, -- 'sent', 'delivered'
    created_at timestamp,
    PRIMARY KEY ((conversation_id), msg_id)
) WITH CLUSTERING ORDER BY (msg_id DESC);
```

### 6.3 微服务拆分（共 7 个服务）

| 服务 | 职责 | 对外协议 | 依赖存储 |
|------|------|----------|----------|
| `user-svc` | 用户注册/登录/资料管理 | HTTP | ScyllaDB, Redis |
| `match-svc` | 匹配队列管理、ELO 计算 | gRPC | Redis |
| `game-svc` | 战斗房间管理、帧同步广播 | WebSocket + gRPC (内部) | Redis, ScyllaDB, Kafka |
| `community-svc` | 动态发布/点赞/评论 | HTTP | ScyllaDB, Kafka |
| `im-svc` | 文字消息收发、离线存储 | WebSocket | ScyllaDB, Redis, Kafka |
| `media-svc` | 语音上传/转码/下载 | HTTP | MinIO, ScyllaDB (元数据) |
| `signaling-svc` | WebRTC 信令转发 | WebSocket | Redis |

**额外基础设施**：APISIX Gateway、etcd、Kafka、Redis、ScyllaDB、MinIO、Prometheus+Grafana、Jaeger。

### 6.4 通信协议

- **HTTP/HTTPS**：用户服务、社区服务、媒体服务（上传下载）、内部服务健康检查  
- **WebSocket/WSS**：IM 服务（文字聊天）、信令服务（通过 APISIX 代理）  
- **gRPC**：服务间高频调用（如 match → game 创建房间），go-zero 原生支持  
- **Kafka**：异步事件（战斗结束 → 更新 ELO & 生成动态；离线消息 → 推送）  

### 6.5 go-zero 服务示例（用户登录）

```go
// user-svc/internal/logic/loginLogic.go
func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginResp, error) {
    // 1. 从 ScyllaDB 查询用户
    var user User
    err := l.svcCtx.DB.Query("SELECT user_id, nickname, elo, password_hash FROM users WHERE email = ?", req.Email).Scan(&user)
    // 2. 验证密码 bcrypt
    // 3. 生成 JWT（密钥与 APISIX 共享）
    token, _ := jwt.NewClient().GenerateToken(jwt.AuthParams{UserId: user.UserId})
    // 4. 存入 Redis 会话
    l.svcCtx.Redis.Set(fmt.Sprintf("session:%s", user.UserId), token, 7*24*time.Hour)
    return &types.LoginResp{Token: token, UserId: user.UserId}, nil
}
```

### 6.6 APISIX Ingress 配置示例（关键路由）

```yaml
apiVersion: apisix.apache.org/v2
kind: ApisixRoute
metadata:
  name: stellar-route
spec:
  http:
    - name: user-api
      match:
        paths: ["/api/v1/user/*", "/api/v1/community/*"]
      backends:
        - serviceName: user-svc
          servicePort: 8080
        - serviceName: community-svc
          servicePort: 8080
      plugins:
        - name: jwt-auth
          enable: true
          config:
            secret: "shared-jwt-secret"
        - name: limit-req
          config:
            rate: 20
            burst: 40
    - name: media-upload
      match:
        paths: ["/api/v1/media/upload"]
      backends:
        - serviceName: media-svc
          servicePort: 8080
      plugins:
        - name: proxy-rewrite
          config:
            headers:
              X-Forwarded-Proto: "https"
    - name: im-websocket
      match:
        paths: ["/ws/im"]
      backends:
        - serviceName: im-svc
          servicePort: 8081
    - name: signaling-websocket
      match:
        paths: ["/ws/signaling"]
      backends:
        - serviceName: signaling-svc
          servicePort: 8082
```

### 6.7 Kafka 主题设计

| Topic | 生产者 | 消费者 | 用途 |
|-------|--------|--------|------|
| `battle_finished` | game-svc | user-svc（更新ELO）, community-svc（自动发战报动态） | 战斗结算事件 |
| `im_offline` | im-svc | push-svc（预留） | 离线消息触发推送（v0.1 可暂不实现推送） |
| `post_like` | community-svc | community-svc（异步更新点赞计数） | 削峰处理点赞 |

### 6.8 自建信令服务器（go-zero 实现要点）

- 使用 `go-zero` 的 `ws` 模块（基于 gorilla/websocket）  
- 房间信息存储在 Redis `signaling:room:{roomId}` (Set 结构存 userId)  
- 每个连接维护在本地（`sync.Map`），节点间信令转发通过 **Redis Pub/Sub**（MVP 单节点可省略）  
- 消息格式 JSON：

```json
// 客户端 → 信令服务器
{
  "type": "join",
  "roomId": "room_123",
  "userId": "user_456"
}

{
  "type": "offer",
  "targetUserId": "user_789",
  "sdp": { ... }
}
// 服务端广播
{
  "type": "user_joined",
  "userId": "user_456",
  "members": ["user_456", "user_789"]
}
```

### 6.9 Godot 客户端对接（总结）

- **HTTP 请求**：使用 `HTTPRequest` 节点调用 APISIX 暴露的 REST API（登录、社区、上传等）  
- **WebSocket 连接**：使用 `WebSocketPeer` 封装，连接 `/ws/im` 和 `/ws/signaling`  
- **战斗帧同步**：`game-svc` 使用独立 TCP 或 WebSocket 二进制协议（protobuf），客户端每帧发送操作，接收服务器广播的输入合集  
- **语音录制**：调用 Godot 的 `AudioEffectCapture` 获取麦克风数据，编码为 AAC 后上传  
- **语音播放**：下载 opus 文件后，使用 `AudioStreamPlayer` + 解码库（或 Godot 4 内置 opus 支持）播放  

---

## 7. 可运维性设计

### 7.1 可观测性
- **日志**：服务输出 JSON 格式，由 **Fluentd** 收集到 **Elasticsearch**，Kibana 查看  
- **指标**：**Prometheus** 采集 go-zero 的 `/metrics` 以及 APISIX stats，**Grafana** 展示仪表板（QPS、延迟、错误率）  
- **链路追踪**：**Jaeger**（go-zero 集成 OpenTelemetry SDK）  

### 7.2 弹性伸缩
- **K8s HPA**：每个服务基于 CPU/内存 + 自定义指标（如 WebSocket 连接数）自动扩缩容  
- **无状态设计**：所有服务均无状态（game-svc 房间状态存于 Redis，支持水平扩展）  

### 7.3 安全性
- **JWT 认证**：APISIX 统一验证，服务间调用通过 gRPC 拦截器传递 token  
- **敏感词过滤**：动态发布/IM 文本调用第三方 API（如阿里绿网）或本地 DFA 库  
- **限流**：APISIX 按用户/IP 限流（每个接口 20rps）  
- **防作弊**：帧同步指令加入时间戳和签名，服务端校验合法性  

---

## 8. MVP 交付清单（Checklist）

| 类别 | 交付内容 |
|------|----------|
| **后端服务** | user-svc, match-svc, game-svc, community-svc, im-svc, media-svc, signaling-svc 共 7 个服务的代码（go-zero） |
| **网关** | APISIX Ingress 配置文件（路由、插件） |
| **数据层** | ScyllaDB 集群（3 节点）建表 CQL 脚本；Redis 哨兵配置；Kafka 集群 topic 创建脚本；MinIO 部署 |
| **客户端** | Godot 项目源码（实现所有对接接口） |
| **部署** | Kubernetes yaml（Deployment, Service, ConfigMap, HPA）；Dockerfile 每个服务 |
| **监控** | Prometheus 规则、Grafana 面板（JSON 模板）、Jaeger 配置 |
| **文档** | API 文档（Swagger/go-zero 自动生成）、运维手册、测试报告 |

---

## 9. 附录：术语表

| 术语 | 解释 |
|------|------|
| 帧同步 | 服务器只转发玩家操作指令，各客户端独立运算游戏逻辑，保证确定性 |
| ELO | 匹配评分系统，根据胜负增减分数 |
| CCU | 同时在线用户数（Concurrent Users） |
| SDP | Session Description Protocol，WebRTC 连接描述 |
| ICE | Interactive Connectivity Establishment，用于 NAT 穿透 |
| SFU | Selective Forwarding Unit，媒体流选择性转发单元 |
| go-zero | 一款高性能的 Golang 微服务框架，内置工具链和治理能力 |

---

**文档状态**：定稿，可进入开发阶段。  
**审批**：产品委员会 | 技术委员会  

---

如需进一步细化某个模块的接口定义（如 game-svc 的 protobuf 协议、信令服务器的详细 API），请告知，我将补充至附录。
