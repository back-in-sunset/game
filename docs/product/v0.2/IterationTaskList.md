# 迭代开发任务清单

**项目名称**：Game 社交化游戏后端  
**文档版本**：v0.2  
**更新时间**：2026-05-09（基于实际代码状态修订）  
**文档目标**：提供直接可执行的开发任务拆分和优先级。已完成项标记为 [x]。  

---

## 1. P0

- [ ] 为平台落地 `tenant/project/environment` 最小模型
- [ ] 以 TDD 方式完成平台模型的领域层测试
- [ ] 以 TDD 方式完成平台模型的仓储层测试
- [ ] 以 TDD 方式完成平台模型的应用服务层测试
- [ ] 为 `comment` 完善点赞能力
- [ ] 为 `comment` 完善取消点赞能力
- [ ] 为 `comment` 完善屏蔽能力
- [ ] 为 `comment` 完善取消屏蔽能力
- [ ] 为 `comment` 完善置顶能力
- [ ] 为 `comment` 完善取消置顶能力
- [ ] 修复评论删除后的统计一致性问题
- [ ] 为 `user` 和 `comment` 增加最小联调脚本
- [x] ~~明确 IM 服务协议模型~~ → 已落地：16 字节大端 header 二进制帧协议
- [x] ~~明确 IM 服务鉴权方式~~ → 已落地：JWT RS256 + platform API 公钥验证
- [x] ~~设计单聊文本消息最小闭环~~ → 已落地：发送/路由/在线投递/离线存储
- [x] ~~确定 TCP 与 WebSocket 的接入策略~~ → 已落地：WS :8082 + TCP :8091 双接入
- [x] ~~好友服务基础 CRUD~~ → 已落地：go-zero RPC + HTTP API
- [x] ~~VDA 语音信令服务~~ → 已落地：LiveKit 集成 + gRPC token 签发
- [x] ~~前端 monorepo 基础架构~~ → 已落地：pnpm + Vite + Expo + 4 共享包
- [x] 前端 LiveKit Web SDK 集成，端到端语音通话
- [x] 前端好友页面完整 UI
- [ ] 前端 Mobile 端 IM 功能实现

---

## 2. P1

- [ ] 为 `user` 增加刷新 token
- [ ] 为 `user` 增加资料修改能力
- [ ] 为 `user/comment` 统一错误码
- [ ] 为 `user/comment` 增加最小集成测试
- [ ] 清理 `user/utils/cachex` 已知失败测试
- [ ] IM 已读回执
- [ ] IM 消息搜索
- [ ] 好友黑名单功能
- [ ] 前端 E2E 测试（Playwright）

---

## 3. P2

- [ ] IM 群聊/频道
- [ ] IM 多媒体消息（图片/语音片段）
- [ ] 好友分组/标签
- [ ] VDA 语音分钟数计量
- [ ] 前端组件测试（Vitest + React Testing Library）
- [ ] 前端 Storybook 文档
- [ ] Mobile 端语音通话

---

## 4. 建议开发顺序（2026-05-11 更新）

1. ~~前端 LiveKit 集成 → 端到端语音通话打通~~ ✅ 已完成
2. ~~前端好友页面完整 UI~~ ✅ 已完成
3. 写 3 篇技术文章 + 整理 SDK npm 包（公开推广准备）
4. 补齐 E2E 测试（Playwright）
5. 落 `tenant/project/environment` 平台最小模型
6. 继续收口 `comment` 服务高级能力（点赞/屏蔽/置顶）
7. 补 `user/comment` 的测试和联调
8. IM 群聊/频道
9. 前端 Mobile 端追赶 Web 端功能
10. 在基础服务稳定后，再扩展动态、攻略等社区内容

---

## 5. 当前不建议优先推进

- APISIX 接入
- Kafka 引入
- ScyllaDB 替换 MySQL
- 复杂游戏对战模块
- 完整计费系统（先做语音分钟数计量即可）

---

## 6. 关联文档

- 产品需求文档：`docs/product/v0.2/PRD.md`
- 技术设计说明：`docs/product/v0.2/TechDesign.md`
- 商业化分析：`docs/product/attention.md`
- 平台模型 TDD 计划：`plan/platform-tenant-project-environment-tdd-plan.md`
- 用户与评论专项计划：`plan/user-comment-service-plan.md`
