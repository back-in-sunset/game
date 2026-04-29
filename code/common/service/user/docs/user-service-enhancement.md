# User 服务功能增强文档

## 1. 背景与目标

当前 `user` 服务已具备注册、登录、用户信息、资料更新、手机号/邮箱绑定、验证码发送等基础能力。  
为满足生产可用性、风控可扩展性与运维治理要求，需要补齐账号安全、会话治理、审计与管理能力。

本文定义后续增强范围、接口建议、数据模型建议与分期实施策略。

---

## 2. 增强范围总览

1. 账号找回与重置密码  
2. Token 生命周期管理  
3. 验证码消费审计  
4. 统一错误码体系  
5. 邮箱/手机号格式与规则校验  
6. 登录与敏感操作风控  
7. 用户状态管理（禁用/冻结/注销）  
8. 资料与敏感操作审计  
9. 事务与唯一约束回归测试  
10. 管理端能力（封禁、解封、重置）  
11. RPC/API 协议能力收敛  
12. 迁移脚本与发布回滚方案  

---

## 3. 详细功能说明

### 3.1 账号找回与重置密码

#### 目标
- 支持“忘记密码”场景，无需旧密码可重置。
- 保证验证码链路、风控链路、审计链路完整。

#### 建议接口
- `POST /api/user/password/reset/send-code`
- `POST /api/user/password/reset/confirm`

#### 核心流程
1. 提交手机号/邮箱发送验证码。  
2. 校验验证码与风控策略。  
3. 重置密码（加密存储）。  
4. 使历史 token 失效（见 3.2）。

---

### 3.2 Token 生命周期管理

#### 目标
- 支持会话延续、主动登出与强制下线。

#### 建议能力
- Refresh Token：短 access token + 长 refresh token。  
- Token 版本号机制：用户级 `token_version`，修改密码/重置密码后递增。  
- 黑名单机制：高风险 token 即时失效。  
- 设备会话：按设备维度维护会话并支持踢设备。

#### 建议接口
- `POST /api/user/token/refresh`
- `POST /api/user/logout`
- `POST /api/user/session/kick`

---

### 3.3 验证码消费审计

#### 目标
- 可追踪验证码发送、验证、失败原因、风控拒绝。

#### 建议数据表
- `user_verify_audit`
  - `id`
  - `scene`（bind_mobile/bind_email/reset_password/login_2fa）
  - `target`
  - `channel`（sms/email）
  - `action`（send/verify）
  - `result`（success/failed/rejected）
  - `reason`
  - `ip`
  - `device_id`
  - `created_at`

---

### 3.4 统一错误码体系

#### 目标
- API/RPC 错误语义统一，便于前端处理与监控聚合。

#### 建议
- 统一错误结构：`code + message + details`。
- 预定义错误码段：
  - `U1xxx` 参数错误
  - `U2xxx` 认证鉴权错误
  - `U3xxx` 业务冲突（手机号/邮箱已占用）
  - `U4xxx` 风控拒绝
  - `U5xxx` 服务内部异常

#### 当前已实现错误码（API）
- `U1001` mobile is required
- `U1002` password is required
- `U1003` name is required
- `U1004` uid invalid
- `U1005` oldPassword and newPassword are required
- `U1006` old and new password must differ
- `U1007` new password length must be >= 6
- `U1008` newMobile is required
- `U1009` verifyCode is required
- `U1010` email and verifyCode are required
- `U1011` email is required
- `U1012` avatar length must be <= 1024
- `U1013` bio length must be <= 1024
- `U1014` location length must be <= 255
- `U1015` birthday must use YYYY-MM-DD
- `U2003` forbidden
- `U3002` old password is incorrect
- `U3003` verification code invalid or expired
- `U3004` mobile already in use
- `U3005` email already in use
- `U4001` rate limited
- `U5001` login failed
- `U5002` user query failed
- `U5003` user update failed
- `U5004` sms verifier unavailable
- `U5005` mobile index query failed
- `U5006` email verifier unavailable
- `U5007` email index query failed
- `U5008` profile model unavailable
- `U5009` profile upsert failed

#### 错误响应示例
```json
{
  "code": "U1004",
  "message": "uid invalid"
}
```

---

### 3.5 邮箱/手机号格式与规则校验

#### 目标
- 防止无效目标进入验证码和绑定链路。

#### 校验建议
- 手机号：国家码、长度、号段合法性。  
- 邮箱：RFC 基础格式 + 域名策略（黑名单、临时邮箱域）。  
- 统一归一化：
  - 手机号标准化（E.164）
  - 邮箱小写化与去空白

---

### 3.6 登录与敏感操作风控

#### 目标
- 可插拔风控平台对接，降低撞库与盗号风险。

#### 当前基础
- 已将验证码频控抽象为 `RateLimiter` 接口。

#### 扩展建议
- 增加 `RiskEvaluator` 接口：
  - 输入：场景、uid、ip、device、geo、target
  - 输出：通过/拒绝/需二次验证
- 场景：
  - 登录
  - 改绑手机号
  - 绑定邮箱
  - 重置密码

---

### 3.7 用户状态管理

#### 目标
- 支持账户封禁与安全冻结。

#### 建议字段
- `status`：active / frozen / banned / deleted
- `status_reason`
- `status_expire_at`

#### 生效点
- 登录前置校验
- 修改资料、改绑、改密前置校验

---

### 3.8 资料与敏感操作审计

#### 目标
- 满足合规与追责需求。

#### 建议记录
- 用户资料修改（前后值摘要）  
- 密码修改/重置  
- 手机号变更  
- 邮箱绑定/解绑  
- 管理员操作（操作者ID、工单ID）

---

### 3.9 事务与唯一约束回归测试

#### 目标
- 保证 `user`、`user_mobile_index`、`user_profile` 一致性。

#### 必测场景
1. 注册成功：`user` 与 `user_mobile_index` 同时写入  
2. 手机号修改：`user` + 索引更新原子性  
3. 删除用户：`user` + 索引删除原子性  
4. 邮箱绑定并发冲突：唯一约束与业务错误码映射  
5. 回滚路径：中途失败无脏数据

---

### 3.10 管理端能力

#### 目标
- 支持运营与安全团队日常处置。

#### 建议接口
- `POST /admin/user/freeze`
- `POST /admin/user/unfreeze`
- `POST /admin/user/reset-password`
- `GET /admin/user/audit/list`

---

### 3.11 RPC/API 协议能力收敛

#### 目标
- 避免仅 API 层有能力，RPC 层缺失导致横向复用困难。

#### 建议
- 在 `user.proto` 增加：
  - `UpdateUserProfile`
  - `ChangePassword`
  - `ChangeMobile`
  - `BindEmail`
- `UserInfoResponse` 补齐 profile 字段。

---

### 3.12 迁移与发布回滚

#### 目标
- 降低线上变更风险。

#### 建议步骤
1. 先发 DDL（`user_mobile_index`、`user_profile`、邮箱唯一索引）。  
2. 再发读兼容代码（可无表运行）。  
3. 再打开写路径。  
4. 观察指标稳定后移除兼容分支。

#### 回滚策略
- 回滚应用时保留新表不删。  
- 索引冲突需提前清洗脏数据。

---

## 4. 分期落地建议

### Phase 1（1-2 周）
- 忘记密码链路
- 统一错误码
- 邮箱/手机号严格校验
- 审计日志基础表

### Phase 2（1-2 周）
- Token refresh / logout / token version
- 用户状态管理
- 管理端封禁/解禁

### Phase 3（持续）
- RiskEvaluator 平台化接入
- 验证码审计大盘与异常告警
- 完整集成回归与压测

---

## 5. 验收标准

1. 安全功能：改密/重置/改绑可用且可审计。  
2. 一致性：索引与主数据无不一致。  
3. 可观测性：关键操作有指标和日志。  
4. 可扩展性：频控/风控可替换实现，无业务侵入。  

---

## 6. API 联调示例（当前已实现）

### 6.0 `POST /api/user/login`

请求示例：
```json
{
  "mobile": "13800000000",
  "password": "123456"
}
```

成功响应：
```json
{
  "accessToken": "jwt-token",
  "accessExpire": 1714377600
}
```

常见失败码：`U1001` `U1002` `U5001`

### 6.0.1 `POST /api/user/register`

请求示例：
```json
{
  "name": "tom",
  "gender": 1,
  "mobile": "13800000000",
  "password": "123456"
}
```

成功响应：
```json
{
  "id": 10001,
  "name": "tom",
  "gender": 1,
  "mobile": "13800000000"
}
```

常见失败码：`U1001` `U1002` `U1003`

### 6.0.2 `POST /api/user/userinfo`

请求示例：
```json
{}
```

成功响应：
```json
{
  "id": 10001,
  "name": "tom",
  "gender": 1,
  "mobile": "13800000000",
  "avatar": "https://cdn.example.com/a.png",
  "bio": "hello",
  "birthday": "1998-08-08",
  "location": "Shanghai",
  "extra": "{\"lang\":\"zh\"}"
}
```

常见失败码：`U1004`

### 6.1 `POST /api/user/profile/update`

请求示例：
```json
{
  "avatar": "https://cdn.example.com/a.png",
  "bio": "hello",
  "birthday": "1998-08-08",
  "location": "Shanghai",
  "extra": "{\"lang\":\"zh\"}"
}
```

成功响应：
```json
{
  "success": true,
  "message": "ok"
}
```

常见失败码：`U1004` `U1012` `U1013` `U1014` `U1015` `U5008` `U5009`

### 6.2 `POST /api/user/password/change`

请求示例：
```json
{
  "oldPassword": "old123456",
  "newPassword": "new123456"
}
```

成功响应：
```json
{
  "success": true,
  "message": "ok"
}
```

常见失败码：`U1004` `U1005` `U1006` `U1007` `U3002` `U5002` `U5003`

### 6.3 `POST /api/user/mobile/change`

请求示例：
```json
{
  "newMobile": "13800000000",
  "verifyCode": "123456"
}
```

成功响应：
```json
{
  "success": true,
  "message": "ok"
}
```

常见失败码：`U1004` `U1008` `U1009` `U3003` `U3004` `U5002` `U5003` `U5004` `U5005`

### 6.4 `POST /api/user/code/send/mobile`

请求示例：
```json
{
  "mobile": "13800000000"
}
```

成功响应：
```json
{
  "success": true,
  "message": "ok"
}
```

常见失败码：`U1001` `U4001` `U5004`

### 6.5 `POST /api/user/code/send/email`

请求示例：
```json
{
  "email": "test@example.com"
}
```

成功响应：
```json
{
  "success": true,
  "message": "ok"
}
```

常见失败码：`U1011` `U4001` `U5006`

### 6.6 `POST /api/user/email/bind`

请求示例：
```json
{
  "email": "test@example.com",
  "verifyCode": "123456"
}
```

成功响应：
```json
{
  "success": true,
  "message": "ok"
}
```

常见失败码：`U1004` `U1010` `U3003` `U3005` `U5002` `U5003` `U5006` `U5007`
