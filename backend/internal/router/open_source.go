package router

import (
	"awesome-chances/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// 【路由组 4：开源贡献】
// 主要目的：查询示例 Issue，并预留根据画像实时检索开源贡献任务的入口。
// 对应实现：handler/catalog.go（4.1 Issue）、handler/integrations.go（4.2 搜索）；路由编号与 Handler 注释保持一致。
func registerOpenSource(api *gin.RouterGroup, h *handler.Handler) {
	group := api.Group("/open-source") // 路由组 4：统一前缀 /api/v1/open-source。
	// 【聚合 4.1：示例 Issue 查询】
	issues := group.Group("/issues")           // 聚合 4.1：共享 /issues 前缀。
	issues.Match(readMethods, "", h.Issues)    // [4.1.1] 查询示例 Issue 列表，支持分类、搜索和分页。
	issues.Match(readMethods, "/:id", h.Issue) // [4.1.2] 按任务 ID 查询示例 Issue，其他类型或不存在的任务返回 404。
	// 【聚合 4.2：实时开源搜索】
	group.POST("/search", h.SearchIssues) // [4.2.1] 校验画像和搜索词，合法请求返回 501，说明 GitHub Provider 尚未接入。
}
