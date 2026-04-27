# User And Comment Service Plan

## 背景

本轮目标是先把 `user` 服务和 `comment` 服务补到可用状态，优先修复空实现、主键返回不一致、基础参数缺失校验，以及 API/RPC 主链路未打通的问题。

## 任务拆分

### 1. user 服务完善

- [x] 审查 `api/rpc/model` 链路，定位注册、登录、用户信息接口的关键缺口
- [x] 修复注册时 `user_id` 写入与返回不一致的问题
- [x] 为注册接口补充 `name/mobile/password` 必填校验
- [x] 为登录接口补充 `mobile/password` 必填校验
- [x] 为 `userinfo` 补充 `x-uid` 与请求用户 ID 一致性校验

### 2. comment 服务完善

- [x] 审查 `api/rpc/model` 链路，定位评论新增、删除、查询接口的空实现
- [x] 打通 API `Add` 到 RPC `AddComment` 的请求透传与返回映射
- [x] 打通 API `Get` 到 RPC `GetComment` 的请求透传与返回映射
- [x] 打通 API `Delete` 到 RPC `DeleteComment` 的请求透传与返回映射
- [x] 为评论新增、删除、查询、列表补充基础参数校验
- [x] 增加统一评论响应映射，明确 `ID/CommentID/CreatedAt` 字段转换
- [x] 在 model 层补充评论逻辑删除能力
- [x] 为 `/api/comments/:id` 增加 `path:"id"` 绑定到 `CommentID`

### 3. 验证

- [x] 执行 `gofmt`
- [x] 尝试执行 `go test`
- [ ] 获取 `user` 与 `comment` 服务的干净全量通过结果

## 已完成改动

### user 服务

- 修正 `UserModel.Insert`，显式写入 `user_id`
- 注册 RPC 改为返回雪花 ID，而不是依赖数据库自增返回值
- 登录/注册 API 与 RPC 都补了基础参数校验
- `UserInfo` RPC 补充鉴权一致性检查，避免跨用户读取

### comment 服务

- API 层补齐 `Add/Get/Delete/List` 的参数校验和 RPC 调用
- RPC 层补齐 `AddComment/GetComment/DeleteComment`
- 增加统一映射函数，避免评论响应字段丢失
- model 层新增逻辑删除方法，当前通过更新评论索引状态实现删除

## 当前风险与后续事项

- 当前 `DeleteComment` 是逻辑删除，只更新索引状态，没有同步扣减主题统计字段
- `comment` 服务仍有点赞、取消点赞、屏蔽、置顶等能力未继续完善
- 全量 `go test ./...` 受本机模块缓存、仓库现有依赖和已有失败测试影响，暂未拿到干净通过结果
- `user/utils/cachex` 存在仓库原有失败测试，需要单独修复

## 建议下一步

1. 继续补 `comment` 的点赞、取消点赞、屏蔽、置顶链路
2. 为 `user/comment` 增加最小可运行的接口联调用例
3. 清理并修复仓库现有失败测试，再做一次全量回归
