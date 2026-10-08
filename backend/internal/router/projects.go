package router

import (
	"awesome-chances/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// 【路由组 2：项目导学】
// 主要目的：查询示例项目，并结合画像提供预置实践步骤和技能缺口。
// 对应实现：handler/catalog.go（2.1 查询）、handler/planning.go（2.2 导学）；路由编号与 Handler 注释保持一致。
func registerProjects(api *gin.RouterGroup, h *handler.Handler) {
	group := api.Group("/projects") // 路由组 2：统一前缀 /api/v1/projects。
	// 【聚合 2.1：项目查询】
	group.Match(readMethods, "", h.Projects)    // [2.1.1] 查询示例项目列表，支持分类、搜索和分页。
	group.Match(readMethods, "/:id", h.Project) // [2.1.2] 按任务 ID 查询项目详情，其他类型或不存在的任务返回 404。
	// 【聚合 2.2：项目导学预览】
	group.POST("/:id/guide", h.ProjectGuide) // [2.2.1] 校验画像，返回预置实践步骤、技能缺口和预计天数，尚不分析真实源码。
}
