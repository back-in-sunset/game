# 迭代开发任务清单

**项目名称**：Game 社交化游戏后端  
**文档版本**：v0.2  
**更新时间**：2026-04-27  
**文档目标**：提供直接可执行的开发任务拆分和优先级。  

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

---

## 2. P1

- [ ] 为 `user` 增加刷新 token
- [ ] 为 `user` 增加资料修改能力
- [ ] 为 `user/comment` 统一错误码
- [ ] 为 `user/comment` 增加最小集成测试
- [ ] 清理 `user/utils/cachex` 已知失败测试

---

## 3. P2

- [ ] 明确 IM 服务协议模型
- [ ] 明确 IM 服务鉴权方式
- [ ] 设计单聊文本消息最小闭环
- [ ] 确定 TCP 与 WebSocket 的接入策略

---

## 4. 建议开发顺序

1. 先落 `tenant/project/environment` 平台最小模型
2. 继续收口 `comment` 服务高级能力
3. 补 `user/comment` 的测试和联调
4. 明确 IM 设计，再决定是否落地消息闭环
5. 在基础服务稳定后，再扩展动态、好友、消息体系

---

## 5. 当前不建议优先推进

- APISIX 接入
- Kafka 引入
- ScyllaDB 替换 MySQL
- 复杂游戏对战模块

---

## 6. 关联文档

- 产品需求文档：`docs/product/v0.2/PRD.md`
- 技术设计说明：`docs/product/v0.2/TechDesign.md`
- 平台模型 TDD 计划：`plan/platform-tenant-project-environment-tdd-plan.md`
- 用户与评论专项计划：`plan/user-comment-service-plan.md`
