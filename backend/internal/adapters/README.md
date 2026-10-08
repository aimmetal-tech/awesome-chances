# 适配器

- demo/：唯一预置方向与示例任务，明确标记 demo。
- memory/：容量限制的临时反馈，服务重启清空。
- postgres/：GORM + PostgreSQL Driver 管理连接，database/sql 连接池、Ping、Close、版本 migration；AuthRepository 使用 ORM 参数绑定执行用户/会话 CRUD，处理唯一冲突、过期身份及撤销。版本 SQL 通过 GORM 事务执行，保留 advisory lock 和校验和。运行时认证不使用内存替身。

配置模型在 model/config.go，读取在 config/；密码/令牌辅助在 security/。GitHub、Jev、LLM、Embedding、竞赛数据源等仍按需要新增适配器。
