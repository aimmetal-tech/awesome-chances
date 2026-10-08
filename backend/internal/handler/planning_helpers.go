package handler

import (
	"awesome-chances/backend/internal/model"
	"github.com/gin-gonic/gin"
)

// 计划公共辅助：供路由组 1、2、6 复用，不对应独立路由编号。
func (h *Handler) previewPlan(c *gin.Context, taskType string) {
	w := c.Writer
	r := c.Request

	var input model.ProfileRequest
	if !decode(w, r, &input) {
		return
	}
	data, err := h.app.PreviewPlan(c.Param("id"), taskType, input.Profile)
	if err != nil {
		serviceError(w, err)
		return
	}
	success(w, 200, data)
}
