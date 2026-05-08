# Friend gRPC API 参考

## 服务概述

Friend 服务管理社交关系链，支持好友申请、好友管理、黑名单功能。采用写扩散存储模型，MySQL 起步，预留 ScyllaDB 迁移路径。

**调用方**：IM 服务（用于权限校验）、客户端 API 网关。

## Proto 定义

```protobuf
syntax = "proto3";
package friendclient;
option go_package = "/friend";

service Friend {
  rpc SendRequest(SendRequestReq) returns (SendRequestResp);
  rpc AcceptRequest(AcceptRequestReq) returns (AcceptRequestResp);
  rpc RejectRequest(RejectRequestReq) returns (RejectRequestResp);
  rpc ListRequests(ListRequestsReq) returns (ListRequestsResp);
  rpc RemoveFriend(RemoveFriendReq) returns (RemoveFriendResp);
  rpc ListFriends(ListFriendsReq) returns (ListFriendsResp);
  rpc CheckFriendship(CheckFriendshipReq) returns (CheckFriendshipResp);
  rpc BlockUser(BlockUserReq) returns (BlockUserResp);
  rpc UnblockUser(UnblockUserReq) returns (UnblockUserResp);
  rpc ListBlockedUsers(ListBlockedUsersReq) returns (ListBlockedUsersResp);
}
```

---

## 好友申请

### SendRequest

发送好友申请。检查是否已是好友、是否已有待处理申请、是否被对方拉黑。

**请求**：

```json
{
  "from_user_id": 1001,
  "to_user_id": 2002,
  "domain": "platform",
  "tenant_id": "",
  "message": "hello"
}
```

**响应**：

```json
{ "request_id": 123456789 }
```

**错误**：
- `F1002` — 不能对自己操作
- `F1003` — 已经是好友
- `F1007` — 已被对方拉黑
- `F1009` — 已有待处理的申请

### AcceptRequest

同意好友申请。创建双向好友关系，更新申请状态。

**请求**：

```json
{
  "request_id": 123456789,
  "user_id": 2002
}
```

**响应**：`{}`

**错误**：
- `F1004` — 申请不存在
- `F1005` — 申请已处理

### RejectRequest

拒绝好友申请。

**请求**：

```json
{
  "request_id": 123456789,
  "user_id": 2002
}
```

**响应**：`{}`

### ListRequests

查询收到的申请列表。

**请求**：

```json
{
  "user_id": 2002,
  "status": 0,
  "offset": 0,
  "limit": 20
}
```

status: `0`=pending, `1`=accepted, `2`=rejected

**响应**：

```json
{
  "requests": [
    {
      "request_id": 123456789,
      "from_user_id": 1001,
      "to_user_id": 2002,
      "message": "hello",
      "status": 0,
      "created_at": 1745000000000
    }
  ],
  "total": 1
}
```

---

## 好友管理

### RemoveFriend

解除好友关系。删除双向记录，清缓存。

**请求**：

```json
{
  "user_id": 1001,
  "friend_id": 2002,
  "domain": "platform",
  "tenant_id": ""
}
```

**响应**：`{}`

**错误**：
- `F1006` — 不是好友

### ListFriends

查询好友列表。优先从 Redis ZSET 读取，miss 时加载 DB 并回填缓存。

**请求**：

```json
{
  "user_id": 1001,
  "domain": "platform",
  "tenant_id": "",
  "offset": 0,
  "limit": 50
}
```

**响应**：

```json
{
  "friend_ids": [2002, 3003],
  "total": 2
}
```

### CheckFriendship

检查是否为好友。优先查 Redis 标记位。

**请求**：

```json
{
  "user_id": 1001,
  "target_id": 2002,
  "domain": "platform",
  "tenant_id": ""
}
```

**响应**：

```json
{ "is_friend": true }
```

---

## 黑名单

### BlockUser

拉黑用户。如果是好友则先解除好友关系。

**请求**：

```json
{
  "user_id": 1001,
  "blocked_user_id": 2002,
  "domain": "platform",
  "tenant_id": ""
}
```

**响应**：`{}`

**错误**：
- `F1002` — 不能对自己操作
- `F1008` — 已在黑名单中

### UnblockUser

取消拉黑。

**请求**：

```json
{
  "user_id": 1001,
  "blocked_user_id": 2002,
  "domain": "platform",
  "tenant_id": ""
}
```

**响应**：`{}`

**错误**：
- `F1010` — 不在黑名单中

### ListBlockedUsers

查询黑名单列表。

**请求**：

```json
{
  "user_id": 1001,
  "domain": "platform",
  "tenant_id": "",
  "offset": 0,
  "limit": 20
}
```

**响应**：

```json
{
  "blocked_user_ids": [2002],
  "total": 1
}
```

---

## 错误码

| Code | 说明 |
|------|------|
| F1001 | 用户不存在 |
| F1002 | 不能对自己操作 |
| F1003 | 已经是好友 |
| F1004 | 申请不存在 |
| F1005 | 申请已处理 |
| F1006 | 不是好友 |
| F1007 | 已被对方拉黑 |
| F1008 | 已在黑名单中 |
| F1009 | 已有待处理的申请 |
| F1010 | 不在黑名单中 |

所有错误通过 gRPC status code 返回：
- `AlreadyExists` — F1003, F1008, F1009
- `NotFound` — F1001, F1004, F1006, F1010
- `PermissionDenied` — F1007
- `FailedPrecondition` — F1005
- `InvalidArgument` — F1002
