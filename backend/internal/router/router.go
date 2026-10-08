package router

import (
	"io"
	"log/slog"

	"awesome-chances/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// readMethods 在原 Gin 分组中同时注册 GET/HEAD，保留组中间件和读取接口兼容性。
var readMethods = []string{"GET", "HEAD"}

// New 聚合所有 Gin 路由组；编号与对应 Handler 注释一致。
// Router 紧凑排列，1/2/3 行阅读间隔只应用于 Handler 的实现函数。
func New(h *handler.Handler, loggers ...*slog.Logger) *gin.Engine {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if len(loggers) > 0 && loggers[0] != nil {
		logger = loggers[0]
	}
	gin.SetMode(gin.ReleaseMode)
	root := gin.New()
	root.HandleMethodNotAllowed = true
	root.RedirectTrailingSlash = false
	_ = root.SetTrustedProxies(nil)
	root.Use(requestLog(logger))
	root.NoRoute(func(c *gin.Context) { c.JSON(404, handler.RouteError("not_found", "路由不存在")) })
	root.NoMethod(func(c *gin.Context) {
		c.JSON(405, handler.RouteError("method_not_allowed", "该路由不支持此请求方法"))
	})
	api := root.Group("/api/v1") // 版本聚合：业务接口共享 /api/v1。
	// 【路由组 1：竞赛规划】查询竞赛及预览备赛步骤。
	registerCompetitions(api, h)
	// 【路由组 2：项目导学】查询项目及预览实践步骤。
	registerProjects(api, h)
	// 【路由组 3：方向导学】方向查询与 AI 对话入口。
	registerGuidance(api, h)
	// 【路由组 4：开源贡献】示例 Issue 与实时搜索入口。
	registerOpenSource(api, h)
	// 【路由组 5：画像与成长】画像、推荐、反馈。
	registerProfile(api, h)
	// 【路由组 6：统一学习任务】任务查询和计划预览。
	registerTasks(api, h)
	// 【路由组 7：系统与初始化】健康检查和 Web 初始化。
	registerSystem(root, api, h)
	// 【路由组 8：用户认证】注册、登录、当前用户和退出。
	registerAuth(api, h)
	// 读取接口通过 Gin Match 同时注册 GET/HEAD，保留分组中间件。
	return root
}
