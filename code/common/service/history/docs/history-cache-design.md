# History 缓存技术设计

## 1. 背景

`history` 服务同时具备高读和高写特征：

- 写路径高频：用户浏览、播放进度更新、重复刷新都会产生写入。
- 读路径高频：历史列表是用户侧常驻页面，分页访问频繁。
- 数据要求偏弱一致：允许短时间不一致，也允许小窗口数据丢失。

基于上述约束，当前实现采用：

- Redis 作为主读写层
- MySQL 作为最终落库存储
- 异步刷盘保证最终一致
- Redis 不可用时回退到 MySQL 直写直读

该方案优先优化吞吐和响应时间，不追求强一致。

## 2. 总体架构

核心入口位于 `history/internal/historycache/manager.go`，API 本地模式与 RPC 模式共用同一套缓存逻辑。

### 2.1 运行模式

1. `RecordHistory`
   - 优先写 Redis
   - 标记脏用户和脏条目
   - 后台异步写回 MySQL

2. `ListHistory`
   - 优先查 Redis
   - Redis miss 时可回退 MySQL

3. `DeleteHistoryItem` / `ClearHistoryByType` / `ClearHistoryAll`
   - 先删除 Redis 可见数据
   - 再通过 tombstone + dirty 标记异步同步到 MySQL

4. 后台 flush worker
   - 周期扫描脏用户
   - 按用户维度执行 upsert 或 soft delete
   - 成功后清理 dirty 标记

### 2.2 生命周期

- RPC 模式：`rpc/internal/svc.ServiceContext.Start/Stop`
- API 本地直连模式：`api/internal/svc.ServiceContext.Start/Stop`

当 `CacheRedis` 未配置时，自动退化为仅 MySQL 模式。

## 3. Redis 键设计

### 3.1 列表与详情

- `history:list:{userID}`
  - `ZSET`
  - member：`{sortID}:{mediaType}:{mediaID}`
  - score：`last_seen_at unix`
  - 用于全量历史分页

- `history:list:{userID}:{mediaType}`
  - `ZSET`
  - member：`{sortID}:{mediaType}:{mediaID}`
  - score：`last_seen_at unix`
  - 用于按类型分页

- `history:item:{userID}:{mediaType}:{mediaID}`
  - `HASH`
  - 保存完整记录字段：
    - `id`
    - `user_id`
    - `media_type`
    - `media_id`
    - `title`
    - `cover`
    - `author_id`
    - `progress_ms`
    - `duration_ms`
    - `finished`
    - `source`
    - `device`
    - `meta`
    - `first_seen_at`
    - `last_seen_at`
    - `sort_id`

### 3.2 脏数据与删除标记

- `history:dirty:users`
  - `ZSET`
  - member：`userID`
  - score：最近一次脏写时间

- `history:dirty:items:{userID}`
  - `SET`
  - member：`{mediaType}:{mediaID}`
  - 表示该用户哪些历史项待刷盘

- `history:deleted:{userID}`
  - `SET`
  - member 类型：
    - `all`
    - `type:{mediaType}`
    - `item:{mediaType}:{mediaID}`
  - 用于删除 tombstone，避免 Redis 删除后又从 MySQL 回源旧数据

- `history:seq:{userID}`
  - `STRING`
  - 自增序列，用于生成 Redis 列表 member 的排序主键 `sortID`

- `history:flush:lock:{userID}`
  - `STRING`
  - 用户级 flush 互斥锁，防止并发刷同一用户

## 4. 读写流程

### 4.1 记录历史

`Manager.Record()` 的处理顺序：

1. 读取 Redis 当前 item，继承已有 `id` 与 `first_seen_at`
2. 通过 `history:seq:{userID}` 生成新的 `sortID`
3. 写入 `history:item:*`
4. 更新全量和类型两个 `ZSET`
5. 标记：
   - `history:dirty:users`
   - `history:dirty:items:{userID}`
6. 删除与该条记录相关的 tombstone

如果 Redis 写失败：

- 记录日志
- 回退到 `model.UpsertRecord()`

### 4.2 读取历史

`Manager.List()` 的处理顺序：

1. 优先检查目标列表 key 是否存在
2. 从 Redis `ZSET` 分页读取 member
3. 根据 member 解析出 `mediaType/mediaID`
4. 批量逐条读取 `history:item:*`
5. 若发现列表中存在脏 member，但详情已丢失：
   - 直接从 ZSET 清理该 member
   - 跳过该记录
6. 当 Redis miss 且允许回退时，转走 MySQL

当前分页游标语义：

- `cursor`：上一页最后一条的 `last_seen_at`
- `lastID`：上一页最后一条的 Redis `sortID`

这保证 Redis 列表分页在同一用户下有稳定顺序，即使 MySQL `id` 尚未回写也不影响翻页。

### 4.3 删除与清空

#### 删除单条

`Manager.DeleteItem()`：

1. 删除 `history:item:*`
2. 从两个 ZSET 删除对应 member
3. 写入 `item:{mediaType}:{mediaID}` tombstone
4. 标记 dirty user + dirty item

#### 按类型清空

`Manager.ClearByType()`：

1. 枚举 `history:list:{userID}:{mediaType}` 中的 member
2. 删除全量列表和类型列表中的 member
3. 删除对应 item key
4. 删除 dirty item 集合中的对应 identity
5. 写入 `type:{mediaType}` tombstone

#### 全量清空

`Manager.ClearAll()`：

1. 枚举 `history:list:{userID}` 全量 member
2. 删除全部 item key
3. 删除该用户相关列表 key 与 dirty item key
4. 写入 `all` tombstone

## 5. 异步刷盘

### 5.1 触发方式

后台 goroutine 周期执行 `FlushOnce()`：

- 默认周期：`FlushIntervalSeconds`
- 每轮最多处理 `FlushBatchUsers`
- 每个用户最多处理 `FlushBatchItems`

### 5.2 刷盘规则

1. 读取 `history:dirty:users`
2. 对每个用户加 `history:flush:lock:{userID}` 互斥锁
3. 优先处理删除标记：
   - `all` -> `SoftDeleteAll`
   - `type:{mediaType}` -> `SoftDeleteByType`
4. 再处理普通 dirty item：
   - Redis item 仍存在 -> `UpsertRecord`
   - Redis item 不存在且 tombstone 存在 -> `SoftDeleteItem`
5. 成功后清理对应 dirty/tombstone
6. 如果用户已无 dirty item 且无 tombstone，从 `history:dirty:users` 移除

### 5.3 失败策略

- 单用户 flush 失败，不影响其他用户
- 失败后 dirty 标记保留，下次继续重试
- 不做死信队列；当前仅依赖重试和日志定位

## 6. MySQL 侧配合

### 6.1 Upsert 优化

`model.UpsertRecord()` 已从：

- `select for update`
- `insert/update`
- `findOne`

优化为：

- 单条 `INSERT ... ON DUPLICATE KEY UPDATE`
- 再按 `LastInsertId` 回读一次

这样降低了写路径的数据库往返次数和锁竞争。

### 6.2 索引

历史表保留以下核心索引：

- `uk_user_media_active (user_id, media_type, media_id, deleted)`
- `idx_user_last_seen (user_id, deleted, last_seen_at, id)`
- `idx_user_type_last_seen (user_id, media_type, deleted, last_seen_at, id)`
- `idx_last_seen (last_seen_at)`

其中：

- 前三个服务在线查询与 upsert 唯一约束
- `idx_last_seen` 用于 `PurgeExpired`

## 7. 配置项

当前 `api/etc/history.yaml` 与 `rpc/etc/history.yaml` 中新增：

### 7.1 CacheRedis

示例：

```yaml
CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node
```

### 7.2 HistoryCache

```yaml
HistoryCache:
  ListTTLSeconds: 604800
  DetailTTLSeconds: 604800
  FlushIntervalSeconds: 60
  FlushBatchUsers: 100
  FlushBatchItems: 200
  DirtyTTLSeconds: 3600
  DeleteMarkerTTL: 3600
  ReadFallbackToDB: true
  WriteBackEnabled: true
```

说明：

- `ReadFallbackToDB`
  - Redis miss 或 Redis 错误时是否允许回退 MySQL

- `WriteBackEnabled`
  - 是否启用异步刷盘 worker
  - 为 `false` 时可作为“纯缓存试运行关闭刷盘”的开关

## 8. 已知限制

1. 当前实现只显式维护两类 `mediaType`
   - `post`
   - `video`
   - 如果后续新增类型，需要同步调整 `ClearAll()` 中的列表 key 删除逻辑

2. `ClearByType()` / `ClearAll()` 仍然是用户维度的 O(n) 删除
   - 适合单用户历史规模有限的场景
   - 若后续用户历史显著增大，需要进一步改造为惰性清理或版本号隔离

3. 列表 Redis 分页使用的是 Redis `sortID`
   - 不等于 MySQL `id`
   - 但对外接口兼容，且不影响分页稳定性

4. 当前无 MQ
   - 刷盘依赖服务进程内 worker
   - 若未来实例数增多、刷盘量变大，可考虑改成 Redis Stream 或 MQ 消费模型

## 9. 测试覆盖

当前已覆盖：

- 单元测试
  - `history/internal/historycache/manager_test.go`
  - 验证 record/list/flush/delete/clear 基本行为

- 集成测试
  - `history/rpc/integration/history_integration_test.go`
  - 验证 MySQL 读写链路与业务流程

建议后续补充：

1. Redis 可用、MySQL 不可用时的短期容错测试
2. Redis 不可用时的回退路径测试
3. flush 重试与 tombstone 冲突场景测试
4. 高并发同用户多次进度更新压测

