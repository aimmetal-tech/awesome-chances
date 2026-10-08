# awesome-chances 项目主记录

## 1. 项目目标

面向校园学生和计算机自学初学者，建设通过桌面浏览器使用的 AI 导学、成长规划与真实任务推荐平台。核心是“能力—兴趣”双维度匹配，统一闭环为 `用户画像 → 任务画像 → 推荐/学习路线 → 执行反馈 → 画像更新`。

规划包含导学 Agent、竞赛规划、项目导学和开源贡献匹配。规划不等于已实现能力，描述项目时应区分真实功能与演示数据。

## 2. 技术栈与当前状态

- 前端：Next.js、React、TypeScript、App Router、Tailwind CSS、ESLint、pnpm，版本以 package.json / pnpm-lock.yaml 为准。`src/` 和前端配置已恢复初始化框架；当前仓库没有此前七栏目演示页面或认证 BFF。
- 后端：Go 1.25 + Gin 1.12 模块化单体，pgx 接入 PostgreSQL，Argon2id 密码哈希、Cookie 会话；版本以 backend/go.mod 为准。
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

<!-- BEGIN:nextjs-agent-rules -->

## This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->
