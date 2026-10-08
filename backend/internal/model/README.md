# 第一版 model

统一放置后端数据结构，不耦合 Gin/pgx。UserRecord 和 SessionRecord 已由 PostgreSQL Repository 和用户/会话迁移使用；其他 Record 仍为后续设计。

| 文件 | 内容 | 用途 |
| --- | --- | --- |
| `domain.go` | Profile、SkillEstimate、Task、Direction、Scores、Gap、Recommendation、Feedback | 推荐和业务使用的领域数据；Profile.Validate 做范围/词表校验 |
| `http.go` | ProfileRequest、ChatRequest、IssueSearchRequest、TaskQuery、Page、Response、ErrorResponse、BootstrapResponse、TaskPlan 等 | handler 的请求/响应及查询数据；兼容现有 Web JSON |
| `records.go` | UserRecord、ProfileRecord、SkillEvidenceRecord、TaskRecord、RecommendationRecord、FeedbackRecord、CompetitionRecord、ProjectRecord、IssueRecord | 后续 PostgreSQL 存储设计，不直接作为 HTTP 响应 |
| `auth.go` | RegisterRequest、LoginRequest、UserView、LoginResponse、SessionRecord 与认证错误 | 密码仅在入参中，UserView 永不带哈希，会话记录只保存令牌哈希 |
| `config.go` | AppConfig、DatabaseConfig、AuthConfig | 后端 .env / 环境变量配置模型，兼容根 .env；凭据不返回 HTTP |
| `errors.go` | ErrFeedbackConflict | 临时反馈适配器与业务/HTTP 共用的冲突语义 |

## 拟定表关系

| 表 | 主键 / 关系 | 主要字段 |
| --- | --- | --- |
| users | id | email（规范化唯一）、display_name、password_hash、created_at、updated_at |
| auth_sessions | token_hash；user_id → users.id | created_at、expires_at；只存会话 SHA-256 哈希 |
| user_profiles | user_id → users.id；一人一份当前画像 | version、profile(JSONB)、updated_at |
| skill_evidence | id；user_id → users.id | skill_id、来源类型/ID、level、confidence、observed_at、verified_at |
| tasks | id | version、task(JSONB)、前置任务 IDs、来源 URL/置信度、可用状态、时间戳 |
| recommendations | id；user_id / task_id | 画像/任务版本、画像快照、推荐结果快照、规则配置、created_at |
| feedback_events | (user_id, event_id) | 目标类型/ID、事件类型、可选数值、发生/接收时间 |
| competitions | task_id → tasks.id | 官网、报名截止/开赛时间、参赛限制、组队规则、核验时间 |
| projects | task_id → tasks.id | 仓库 URL、许可证、技术栈、分析的 commit/revision |
| issues | task_id → tasks.id；(repository, number) 唯一 | owner/repo、Issue 编号/URL、标签、开关状态、认领状态、抓取时间 |

首版合并能力、兴趣、目标和偏好到 Profile JSONB；先保留技能证据独立表。技能词表与验证当前使用已有四项，真实扩展时一起调整模型校验、迁移和前端契约。用户 ID 以后从认证上下文取得，不能信任客户端自行声明的用户归属。

`db` 标签只定义拟定列名，嵌套结构/切片需要 PostgreSQL Adapter 显式序列化，记录需要 Repository 校验。实际 migration 必须补齐外键/唯一索引、分数 [0,1] 检查、目标/事件/来源枚举与版本约束；JSONB 中的 task.id 应与 tasks.id 一致。前置 ID 和多态反馈目标由业务验证，不能把 JSONB 标签当作外键约束。

未知网址、赛事日期或核验时间使用指针表示 NULL，不能补造事实。推荐保存画像、任务及规则输入，便于回溯；证据保留来源和核验状态，自报完成不等于验证成功。用户/会话迁移代码已提供，是否执行由数据库配置决定；真实 PostgreSQL 尚未联调，其他模型仍为设计，demo Catalog 和反馈仍走原演示路径。
