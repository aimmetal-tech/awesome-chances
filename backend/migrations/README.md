# PostgreSQL 版本迁移

`001_auth.up.sql` 创建 users 和 auth_sessions，包含邮箱唯一约束、密码哈希标记、外键与会话有效期检查/索引。`embed.go` 将 SQL 编入服务，不依赖启动目录。

按 backend/.env.example 在 backend/.env 填写 PG 连接参数并设置 DATABASE_ENABLED=true（兼容已有根 .env），然后在 backend/ 执行：

```bash
go run ./cmd/migrate
go run ./cmd/api
```

也可设置 DATABASE_AUTO_MIGRATE=true 由 API 启动执行。默认 false，需要先迁移。Database.Migrate 在事务与 advisory lock 中执行，schema_migrations 记录版本、SHA-256 校验和与执行时间；重复运行不重复建表，修改已执行 SQL 会失败，应新增版本。

数据库本身和 PostgreSQL 账号需由用户准备，本工具不创建数据库/角色，不自动执行删除或回滚。迁移失败时事务回滚；启动会检查版本。新业务表继续新增版本 SQL，画像/任务/反馈 Record 当前尚未持久化。
