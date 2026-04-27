# 产品文档索引

**项目名称**：Game Backend Cloud  
**更新时间**：2026-04-27  
**文档目标**：作为产品文档总入口，帮助快速理解整个平台的文档结构、阅读顺序和各文档职责。  

---

## 1. 阅读建议

如果第一次了解这个项目，建议按下面顺序阅读：

1. `ProductVision.md`
2. `SaaSTenantModel.md`
3. `PlatformArchitecture.md`
4. `PermissionRoleModel.md`
5. `ConfigEnvironmentModel.md`
6. `v0.2/PRD.md`
7. `v0.2/TechDesign.md`
8. `v0.2/IterationTaskList.md`

这个顺序对应：

- 先看产品愿景
- 再看 SaaS 与平台底座模型
- 再看当前版本的产品与执行文档

---

## 2. 文档分层

当前文档分成 4 层。

## 2.1 产品总纲层

- [产品总纲](./ProductVision.md)

作用：

- 定义整个平台是什么
- 定义目标客户、模块树、产品路线

## 2.2 SaaS 平台模型层

- [SaaS 租户与项目模型](./SaaSTenantModel.md)
- [平台模块架构说明](./PlatformArchitecture.md)
- [权限与角色模型](./PermissionRoleModel.md)
- [配置中心与环境配置模型](./ConfigEnvironmentModel.md)

作用：

- 定义多租户、项目、环境、服务实例
- 定义平台分层与服务关系
- 定义权限、配置、发布等平台级基础规则

## 2.3 版本文档层

- [v0.2 产品需求文档](./v0.2/PRD.md)
- [v0.2 技术设计说明](./v0.2/TechDesign.md)
- [v0.2 迭代开发任务清单](./v0.2/IterationTaskList.md)

作用：

- 描述当前版本要交付什么
- 描述当前版本的技术状态
- 描述当前版本的执行任务

## 2.4 专项计划层

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

- 当前版本交付哪些产品能力
- 当前版本不做什么

### v0.2/TechDesign.md

负责回答：

- 当前仓库的真实技术现状是什么
- 当前服务已经做到哪一步

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

- 平台愿景
- SaaS 模型
- 模块架构
- 权限模型
- 配置模型
- 当前版本 PRD
- 当前版本技术设计
- 当前版本任务清单

这意味着后续可以从“继续写文档”切换到“按任务推进实现”。

