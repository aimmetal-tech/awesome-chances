package handler

import (
	"github.com/gin-gonic/gin"
)

// 查询实现聚合：承接竞赛、项目、导学、开源和任务 Router；不同组按三行分开。



// 【路由组 1：竞赛规划】
// 主要目的：查询示例竞赛任务，并根据画像预览备赛步骤、技能缺口和预计天数。
// 所属 Router：router/competitions.go；本段实现聚合 1.1，按编号定位对应接口。


// 【聚合 1.1：竞赛查询】
// Competitions [1.1.1] GET /api/v1/competitions
// 功能：查询示例竞赛任务列表，支持分类、搜索和分页。
func (h *Handler) Competitions(c *gin.Context) {
	h.listTasks(c, "competition")
}

// Competition [1.1.2] GET /api/v1/competitions/{id}
// 功能：按任务 ID 查询竞赛详情，其他类型或不存在的任务返回 404。
func (h *Handler) Competition(c *gin.Context) { h.getTask(c, "competition") }



// 【路由组 2：项目导学】
// 主要目的：查询示例项目，并结合画像提供预置实践步骤和技能缺口。
// 所属 Router：router/projects.go；本段实现聚合 2.1，按编号定位对应接口。


// 【聚合 2.1：项目查询】
// Projects [2.1.1] GET /api/v1/projects
// 功能：查询示例项目列表，支持分类、搜索和分页。
func (h *Handler) Projects(c *gin.Context) { h.listTasks(c, "project") }

// Project [2.1.2] GET /api/v1/projects/{id}
// 功能：按任务 ID 查询项目详情，其他类型或不存在的任务返回 404。
func (h *Handler) Project(c *gin.Context) { h.getTask(c, "project") }



// 【路由组 3：方向导学】
// 主要目的：介绍学习方向、岗位工作及技术栈，并预留 AI 导学对话入口。
// 所属 Router：router/guidance.go；本段实现聚合 3.1，按编号定位对应接口。


// 【聚合 3.1：学习方向查询】
// Directions [3.1.1] GET /api/v1/guidance/directions
// 功能：返回全部预置学习方向。
func (h *Handler) Directions(c *gin.Context) {
	w := c.Writer

	success(w, 200, h.app.Directions())
}

// Direction [3.1.2] GET /api/v1/guidance/directions/{id}
// 功能：返回指定方向的工作内容、技术栈和预置学习顺序。
func (h *Handler) Direction(c *gin.Context) {
	w := c.Writer

	data, err := h.app.Direction(c.Param("id"))
	if err != nil {
		serviceError(w, err)
		return
	}
	success(w, 200, data)
}



// 【路由组 4：开源贡献】
// 主要目的：查询示例 Issue，并预留根据画像实时检索开源贡献任务的入口。
// 所属 Router：router/open_source.go；本段实现聚合 4.1，按编号定位对应接口。


// 【聚合 4.1：示例 Issue 查询】
// Issues [4.1.1] GET /api/v1/open-source/issues
// 功能：查询示例 Issue 列表，支持分类、搜索和分页。
func (h *Handler) Issues(c *gin.Context) { h.listTasks(c, "issue") }

// Issue [4.1.2] GET /api/v1/open-source/issues/{id}
// 功能：按任务 ID 查询示例 Issue，其他类型或不存在的任务返回 404。
func (h *Handler) Issue(c *gin.Context) { h.getTask(c, "issue") }



// 【路由组 6：统一学习任务】
// 主要目的：提供跨类型任务查询和通用计划预览，供各功能复用同一任务模型。
// 所属 Router：router/tasks.go；本段实现聚合 6.1，按编号定位对应接口。


// 【聚合 6.1：任务查询】
// Tasks [6.1.1] GET /api/v1/tasks
// 功能：按任务类型、兴趣分类、关键词和分页参数查询统一任务列表。
func (h *Handler) Tasks(c *gin.Context) { h.listTasks(c, "") }

// Task [6.1.2] GET /api/v1/tasks/{id}
// 功能：按任务 ID 查询任意类型的任务详情。
func (h *Handler) Task(c *gin.Context) { h.getTask(c, "") }
