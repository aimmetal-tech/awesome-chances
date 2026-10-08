# awesome-chances 项目主记录

## 1. 项目目标

面向校园学生和计算机自学初学者，建设通过桌面浏览器使用的 AI 导学、成长规划与真实任务推荐平台。核心是“能力—兴趣”双维度匹配，统一闭环为 `用户画像 → 任务画像 → 推荐/学习路线 → 执行反馈 → 画像更新`。

规划包含导学 Agent、竞赛规划、项目导学和开源贡献匹配。规划不等于已实现能力，描述项目时应区分真实功能与演示数据。

## 2. 技术栈与当前状态

- 前端：Next.js、React、TypeScript、App Router、Tailwind CSS、ESLint、pnpm，版本以 package.json / pnpm-lock.yaml 为准。当前仍为基础框架，已采用简体中文及思源黑体优先的本机字体栈；没有此前七栏目演示页面或认证 BFF。
- 后端：Go 1.25 + Gin 1.12 模块化单体，GORM + PostgreSQL Driver 接入 PostgreSQL，database/sql 管理连接池（驱动底层 pgx），保留版本 SQL 迁移；Argon2id 密码哈希、Cookie 会话，版本以 backend/go.mod 为准。
- 已有后端路由聚合、领域/HTTP/数据库模型、示例查询、规则推荐、计划预览、内存反馈、注册/登录/当前用户/退出，以及 PostgreSQL 连接和用户/会话迁移代码。
- PostgreSQL 连接值待填写，真实数据库尚未联调；数据库关闭时认证返回 503。画像、任务和反馈尚未实现数据库持久化。
- Agent、MCP、真实 LLM、GitHub 搜索与实时赛事尚未接入；插件入口返回明确的未接入状态。外部能力通过 Provider + Adapter 隔离。

## 3. 目录与接口

- `src/app/`：前端页面；`src/app/api/` 是未来 Next.js BFF 的约定位置，目前未实现。BFF 仅负责转发、聚合和页面数据适配，核心业务在 backend。
- `backend/cmd/api/`：API 入口；`backend/cmd/migrate/`：迁移入口。
- `backend/internal/router/`：Gin 路由注册与聚合；`handler/`：请求/响应；`service/`：业务；`model/`：领域、HTTP、数据库记录及配置结构。
- `backend/internal/recommendation/`：推荐规则；`provider/`：外部能力契约；`adapters/`：demo、内存与 PostgreSQL 实现；`config/`：环境配置；`security/`：密码与令牌辅助。
- `backend/migrations/`：版本 SQL；`backend/internal/openapi/openapi.yaml`：OpenAPI 3.1 中文契约，修改接口时同步维护。
- 路由组：1 竞赛、2 项目、3 导学、4 开源、5 画像与成长、6 任务、7 系统、8 认证。业务接口前缀 `/api/v1`，健康检查 `/healthz`；后端不托管前端页面。
- 示例数据唯一来源为 `backend/internal/adapters/demo/`。调用链为 `router → handler → service → 推荐/Repository`，数据库记录不直接作为 HTTP 响应。

## 4. 运行与记录维护

前端按原 README 使用 pnpm；后端独立运行方式见 `backend/README.md`。后端配置模板为 `backend/.env.example`，在 backend 目录运行时优先读取该目录的 `.env`，也兼容上层 `.env`；进程环境优先。真实凭据不提交。

本文件是项目的主记录，仅保留项目目标、实际架构和关键状态。结构性变化按影响范围同步本文件及对应模块 README；日常目标、工作日志、个人约束和接力流水账不写入本文件。后端专属工作指引位于仓库外的 `../AGENTS.md`，是本地协作辅助，不是克隆项目运行的必要文件。

此前演示前端备份位于仓库外 `../frontend/`，不参与主项目构建或提交。`docs/`、`data/` 当前为忽略的本地记录/规划；缓存、报告、私人环境配置和其他与后端功能无直接关系的新内容须加入忽略规则，不提交。必要的后端源码、测试、迁移、契约与无秘密配置模板正常维护。

- AGENTS.md 应反映最新的项目理解、实现状态和协作约定。
- README.md 应反映最新的项目介绍、目录说明、开发命令和部署方式。
- 完成变更前，核对两份文档与实际代码是否一致。

## 5. 多级 AGENTS.md 约定

- **项目中可能存在多个子 AGENTS.md**，例如 `backend/AGENTS.md`、`src/AGENTS.md` 或更深层目录中的 AGENTS.md。不得假设根目录文件是唯一的协作指引。
- 根 AGENTS.md 适用于整个项目；子 AGENTS.md 适用于其所在目录及下级目录，不适用于同级目录或其他目录。
- 修改任何文件前，必须沿项目根目录到目标文件所在目录的路径，查找并读取所有适用的 AGENTS.md，按从根到子目录的顺序理解约定。
- 子 AGENTS.md 可以补充或细化上级约定；约定冲突时，以更深层目录中的 AGENTS.md 为准，未覆盖的上级约定继续生效。用户明确指令优先于这些文档约定。
- 修改涉及多个目录时，分别检查各目录适用的指引；结构性变更还须同步更新受影响的子 AGENTS.md。

## 6. 语言与字体约定

- **简体中文是项目的首选语言**。页面文案、提示信息和项目文档默认使用简体中文；技术名词、代码标识符和必要的专有名称可保留原文。
- 当前阶段不考虑 i18n，不引入多语言依赖、语言切换或按语言划分的路由。页面根元素使用 `lang="zh-CN"`。
- **思源黑体是首选中文字体**。全局字体及组件字体应先使用思源黑体，再回退到系统自带字体，最后使用通用无衬线字体。
- 当前使用本机字体栈：`"Source Han Sans SC", "Source Han Sans CN", "思源黑体", "PingFang SC", "Microsoft YaHei", system-ui, sans-serif`。未安装思源黑体时使用系统字体；当前不依赖在线字体服务。
- 全局样式与 Tailwind 的 `font-sans` 使用同一字体栈，避免组件覆盖中文字体优先级。

## 7. 贡献流程

- 贡献流程与填写示例见 [CONTRIBUTING.md](./CONTRIBUTING.md)。
- 开始编写新功能前，必须创建对应 Issue，说明功能目标、实现范围和验收标准；已有相同功能 Issue 时复用，避免重复。
- 创建 PR 后，必须在对应 Issue 正文中用 `Related` 附上实际 PR 链接，并在 PR 描述中用 `Closes #编号` 引用已全部完成的 Issue，或用 `Related #编号` 引用仅关联或部分完成的 Issue。同一 Issue 对应多个 PR 时，列出全部链接；尚未创建 PR 时标记“待创建”。
- `Closes` 在 PR 合并到仓库默认分支后自动关闭对应 Issue；`Related` 不自动关闭 Issue。子任务 PR 使用 `Related` 引用尚未全部完成的父 Issue。
- 允许通过一个大 Issue 和多个子 Issue 拆分功能。父 Issue 记录整体目标、总体验收标准、子 Issue 清单及 PR 汇总；子 Issue 记录具体任务、验收标准、父 Issue 链接和对应 PR 链接。
- 提交前读取各级适用的 AGENTS.md，运行与改动相关的检查，在 PR 中如实记录验证结果，并按文档维护要求同步受影响的文档。

<!-- BEGIN:nextjs-agent-rules -->

## This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->
