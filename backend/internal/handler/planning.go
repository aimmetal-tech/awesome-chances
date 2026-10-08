package handler

import (
	"github.com/gin-gonic/gin"
)

// 计划实现聚合：承接竞赛备赛、项目导学及通用任务计划，预览均不伪装成 AI 生成。



// 【路由组 1：竞赛规划】
// 主要目的：查询示例竞赛任务，并根据画像预览备赛步骤、技能缺口和预计天数。
// 所属 Router：router/competitions.go；本段实现聚合 1.2，按编号定位对应接口。


// 【聚合 1.2：备赛计划预览】
// CompetitionPlan [1.2.1] POST /api/v1/competitions/{id}/plan
// 功能：校验画像，返回预置备赛步骤、技能缺口和预计天数，不生成真实赛事日期。
func (h *Handler) CompetitionPlan(c *gin.Context) {
	h.previewPlan(c, "competition")
}



// 【路由组 2：项目导学】
// 主要目的：查询示例项目，并结合画像提供预置实践步骤和技能缺口。
// 所属 Router：router/projects.go；本段实现聚合 2.2，按编号定位对应接口。


// 【聚合 2.2：项目导学预览】
// ProjectGuide [2.2.1] POST /api/v1/projects/{id}/guide
// 功能：校验画像，返回预置实践步骤、技能缺口和预计天数，尚不分析真实源码。
func (h *Handler) ProjectGuide(c *gin.Context) {
	h.previewPlan(c, "project")
}



// 【路由组 6：统一学习任务】
// 主要目的：提供跨类型任务查询和通用计划预览，供各功能复用同一任务模型。
// 所属 Router：router/tasks.go；本段实现聚合 6.2，按编号定位对应接口。


// 【聚合 6.2：通用计划预览】
// TaskPlan [6.2.1] POST /api/v1/tasks/{id}/plan
// 功能：校验画像，预览任务步骤、技能缺口和预计天数，不保存学习计划。
func (h *Handler) TaskPlan(c *gin.Context) { h.previewPlan(c, "") }
