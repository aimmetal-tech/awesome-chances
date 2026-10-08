package router

import (
	"awesome-chances/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// 【路由组 1：竞赛规划】
// 主要目的：查询示例竞赛任务，并根据画像预览备赛步骤、技能缺口和预计天数。
// 对应实现：handler/catalog.go（1.1 查询）、handler/planning.go（1.2 备赛）；路由编号与 Handler 注释保持一致。
func registerCompetitions(api *gin.RouterGroup, h *handler.Handler) {
	group := api.Group("/competitions") // 路由组 1：统一前缀 /api/v1/competitions。
	// 【聚合 1.1：竞赛查询】
	group.Match(readMethods, "", h.Competitions)    // [1.1.1] 查询示例竞赛任务列表，支持分类、搜索和分页。
	group.Match(readMethods, "/:id", h.Competition) // [1.1.2] 按任务 ID 查询竞赛详情，其他类型或不存在的任务返回 404。
	// 【聚合 1.2：备赛计划预览】
	group.POST("/:id/plan", h.CompetitionPlan) // [1.2.1] 校验画像，返回预置备赛步骤、技能缺口和预计天数，不生成真实赛事日期。
}
