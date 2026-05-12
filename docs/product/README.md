# 产品文档索引

**项目名称**：Game Backend Cloud  
**更新时间**：2026-05-09  
**文档目标**：作为产品文档总入口，帮助快速理解整个平台的文档结构、阅读顺序和各文档职责。  

---

## 1. 阅读建议

如果第一次了解这个项目，建议按下面顺序阅读：

1. `ProductVision.md` — 产品是什么、目标客户、模块树
2. `attention.md` — **当前实际状态总览**（技术现状 + 商业化分析 + 差距分析）
3. `ProductThreeHorizons.md` — **产品三态规划**（近期/中期/远期）
4. `v0.2/PRD.md` — 当前版本已交付和计划交付的功能
5. `v0.2/TechDesign.md` — 当前仓库真实技术状态
6. `v0.2/IterationTaskList.md` — 当前开发任务和优先级
7. `v0.3/DemoPageSpec.md` — **公开 Demo 页面规格（已实施）**
8. `SaaSTenantModel.md` — 未来 SaaS 多租户模型（蓝图）
9. `PlatformArchitecture.md` — 平台模块分层架构（蓝图）
10. `PermissionRoleModel.md` — 未来权限与角色模型（蓝图）
11. `ConfigEnvironmentModel.md` — 未来配置中心模型（蓝图）

这个顺序对应：

- 先看产品愿景和当前实际状态
- 再看当前版本的执行文档
- 最后看 SaaS 平台蓝图（尚未实施）

---

## 2. 文档分层

当前文档分成 5 层。

## 2.1 产品总纲层

- [产品总纲](./ProductVision.md)

作用：

- 定义整个平台是什么
- 定义目标客户、模块树、产品路线

## 2.2 当前状态总览层

- [商业化分析与现状评估](./attention.md)

作用：

- 描述当前所有服务的实际成熟度和商业化就绪度
- 提供商业模式、定价、推广、变现节奏的具体指导
- 对比 PRD 承诺 vs 实际交付的差距分析
- 是了解"现在到底有什么"的最快入口

## 2.3 SaaS 平台模型层（蓝图，尚未实施）

- [SaaS 租户与项目模型](./SaaSTenantModel.md)
- [平台模块架构说明](./PlatformArchitecture.md)
- [权限与角色模型](./PermissionRoleModel.md)
- [配置中心与环境配置模型](./ConfigEnvironmentModel.md)

作用：

- 定义多租户、项目、环境、服务实例
- 定义平台分层与服务关系
- 定义权限、配置、发布等平台级基础规则

## 2.4 版本文档层

- [v0.2 产品需求文档](./v0.2/PRD.md)
- [v0.2 技术设计说明](./v0.2/TechDesign.md)
- [v0.2 迭代开发任务清单](./v0.2/IterationTaskList.md)

作用：

- 描述当前版本已交付和计划交付的功能
- 描述当前版本的真实技术状态
- 描述当前版本的执行任务和优先级

## 2.5 专项计划层

- [用户与评论专项计划](../../plan/user-comment-service-plan.md)
- [Docker 依赖说明](../../deploy/docker/README.md)

作用：

- 记录具体服务或具体领域的执行计划
- 承接版本任务的落地细节

---

## 3. 各文档职责

### ProductVision.md

负责回答：

- 这个平台最终要做成什么
- 为什么是 SaaS 化游戏后端平台
- 模块边界和阶段路线是什么

### attention.md

负责回答：

- 现在到底有什么、哪些能卖、哪些还缺
- 怎么定价、怎么推广、什么时候开始收费
- PRD 承诺和实际代码之间差了什么
- 当前最该做的 3 件事是什么

### SaaSTenantModel.md

负责回答：

- Tenant / Project / Environment 是什么
- 多租户如何隔离
- 配额和计费应该挂在哪一层

### PlatformArchitecture.md

负责回答：

- 平台分几层
- 各层职责是什么
- 模块之间如何调用

### PermissionRoleModel.md

负责回答：

- 平台管理员、租户管理员、项目管理员、运营、开发分别能做什么
- 高风险操作如何控制

### ConfigEnvironmentModel.md

负责回答：

- 配置在哪一层生效
- 配置如何继承、覆盖、发布和回滚

### v0.2/PRD.md

负责回答：

- 当前版本已交付哪些产品能力
- 当前版本不做什么

### v0.2/TechDesign.md

负责回答：

- 当前仓库的真实技术现状是什么
- 当前每个服务已经做到哪一步

### v0.2/IterationTaskList.md

负责回答：

- 当前开发应该先做什么
- P0 / P1 / P2 如何排序

---

## 4. 后续维护建议

### 4.1 新版本文档

后续新增版本时，建议继续按目录组织：

```text
docs/product/v0.3/
docs/product/v0.4/
```

每个版本目录至少包含：

- `PRD.md`
- `TechDesign.md`
- `IterationTaskList.md`

### 4.2 上位模型文档

以下文档建议长期稳定，不要频繁改文件名：

- `ProductVision.md`
- `SaaSTenantModel.md`
- `PlatformArchitecture.md`
- `PermissionRoleModel.md`
- `ConfigEnvironmentModel.md`

这些文档更适合改内容版本，而不是改文件名。

### 4.3 任务落地文档

当某个模块进入落地阶段时，建议在：

- `plan/`

下面增加专项计划，用来承接更细粒度执行任务。

---

## 5. 当前文档状态

当前已形成完整文档骨架：

- 平台愿景（ProductVision）
- 当前状态总览（attention — **2026-05-09 更新，反映实际代码状态**）
- 产品三态规划（ProductThreeHorizons — **2026-05-09 创建**）
- SaaS 模型蓝图（SaaSTenant / PlatformArchitecture / PermissionRole / ConfigEnvironment）
- 当前版本 PRD（v0.2 — **2026-05-11 修订，IM/好友/VDA 前端已超出原计划**）
- 当前版本技术设计（v0.2 — **2026-05-11 修订，LiveKit 前端集成已落地**）
- 当前版本任务清单（v0.2 — **2026-05-11 修订，LiveKit 集成和好友 UI 已标记完成**）
- Demo 页面规格（v0.3 — **2026-05-11 修订，DemoPage + DemoRoomPage 已实施**）

v0.2 版本文档已从”保守描述”更新为”基于代码实际状态的准确描述”。v0.3 Demo 页面已实施，包括多人语音房间。SaaS 蓝图文档保持为设计参考，尚未进入实施阶段。

