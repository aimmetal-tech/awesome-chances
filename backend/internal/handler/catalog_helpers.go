package handler

import (
	"awesome-chances/backend/internal/model"
	"github.com/gin-gonic/gin"
	"strconv"
)

// 查询公共辅助：供路由组 1、2、4、6 复用，不对应独立路由编号。
func (h *Handler) listTasks(c *gin.Context, taskType string) {
	w := c.Writer
	r := c.Request

	params := r.URL.Query()
	if taskType != "" && params.Has("type") && params.Get("type") != taskType {
		failure(w, 400, "invalid_input", "该栏目不支持查询其他任务类型")
		return
	}
	if taskType == "" {
		taskType = params.Get("type")
	}
	page, pageSize := 1, 20
	var err error
	if params.Has("page") {
		page, err = strconv.Atoi(params.Get("page"))
		if err != nil {
			failure(w, 400, "invalid_input", "page 必须为整数")
			return
		}
	}
	if params.Has("pageSize") {
		pageSize, err = strconv.Atoi(params.Get("pageSize"))
		if err != nil {
			failure(w, 400, "invalid_input", "pageSize 必须为整数")
			return
		}
	}
	data, err := h.app.Tasks(model.TaskQuery{Type: taskType, Category: params.Get("category"), Query: params.Get("q"), Page: page, PageSize: pageSize})
	if err != nil {
		serviceError(w, err)
		return
	}
	success(w, 200, data)
}

func (h *Handler) getTask(c *gin.Context, taskType string) {
	w := c.Writer

	data, err := h.app.Task(c.Param("id"), taskType)
	if err != nil {
		serviceError(w, err)
		return
	}
	success(w, 200, data)
}
