# Go 后端

Go 1.25 + Gin 模块化单体；pgx 访问 PostgreSQL，密码使用 Argon2id，会话通过 HttpOnly Cookie 传递，数据库仅保存令牌哈希。后端独立提供 API，不托管 Web 页面。

## 本地运行

在本目录执行（PowerShell）：

```powershell
Copy-Item .env.example .env
go run ./cmd/api
```

已有 `.env` 时不要覆盖。配置加载顺序为显式 `ENV_FILE`，否则当前目录 `.env`，否则上层 `.env`；已有进程环境优先。端口由 `BACKEND_HOST` / `BACKEND_PORT` 设置，示例默认地址为 `http://127.0.0.1:8080`。

- 健康检查：`GET /healthz`。
- 业务接口：`/api/v1/*`；8 个路由组与编号见 `internal/router/`。
- 注册/登录/身份/退出：`POST /api/v1/auth/register`、`POST /api/v1/auth/login`、`GET /api/v1/auth/me`、`POST /api/v1/auth/logout`。
- 完整请求、响应及状态码见 [OpenAPI 中文契约](internal/openapi/openapi.yaml)，目前没有 Swagger UI 路由。

## 数据库

模板中的 PG 连接值全部留空。数据库及角色需预先创建，填写 PG 参数并设 `DATABASE_ENABLED=true` 后执行：

```powershell
go run ./cmd/migrate
go run ./cmd/api
```

也可配置 `DATABASE_AUTO_MIGRATE=true`，由 API 启动时迁移；迁移规则见 [migrations/README.md](migrations/README.md)。启用数据库后连接/迁移失败会停止启动，不使用内存账号替代。

`DATABASE_ENABLED=false` 时可测试示例查询和推荐，认证返回 503。方向、任务和计划目前为 demo，反馈保存在进程内存；画像/任务/反馈的数据库记录仍为设计。真实 PostgreSQL 尚未联调，AI/GitHub/实时赛事未接入。

## 模块与检查

`cmd/` 放启动入口，`internal/router/ → handler/ → service/` 处理请求；`model/` 放结构体，`recommendation/` 放规则，`provider/` 放外部接口，`adapters/` 放实现，`config/` 和 `security/` 放配置及认证辅助，`migrations/` 放版本 SQL。

```powershell
go test ./...
go vet ./...
```

普通测试不依赖真实服务；数据库集成测试默认跳过，只在设置专用测试库的 `TEST_DATABASE_URL` 时执行。密钥、环境文件、构建输出和测试报告均不提交。
