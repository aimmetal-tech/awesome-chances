package router

import (
	"awesome-chances/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// 【路由组 6：统一学习任务】
// 主要目的：提供跨类型任务查询和通用计划预览，供各功能复用同一任务模型。
// 对应实现：handler/catalog.go（6.1 查询）、handler/planning.go（6.2 计划）；路由编号与 Handler 注释保持一致。
func registerTasks(api *gin.RouterGroup, h *handler.Handler) {
	group := api.Group("/tasks") // 路由组 6：统一前缀 /api/v1/tasks。
	// 【聚合 6.1：任务查询】
	group.Match(readMethods, "", h.Tasks)    // [6.1.1] 按任务类型、兴趣分类、关键词和分页参数查询统一任务列表。
	group.Match(readMethods, "/:id", h.Task) // [6.1.2] 按任务 ID 查询任意类型的任务详情。
	// 【聚合 6.2：通用计划预览】
	group.POST("/:id/plan", h.TaskPlan) // [6.2.1] 校验画像，预览任务步骤、技能缺口和预计天数，不保存学习计划。
}
