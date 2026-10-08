package handler

import (
	"github.com/gin-gonic/gin"
)



// 【路由组 7：系统与初始化】
// 主要目的：检查后端存活，并提供 Web 首次加载所需的演示数据和接入状态。
// 对应注册：router/system.go；本文件只实现这一组的业务接口。


// 【聚合 7.1：服务健康检查】
// Health [7.1.1] GET /healthz
// 功能：返回服务存活状态，不检查数据库或外部 Provider。
func (h *Handler) Health(c *gin.Context) {
	w := c.Writer

	write(w, 200, struct {
		Status string `json:"status"`
		Mode   string `json:"mode"`
	}{"ok", "demo"})
}


// 【聚合 7.2：Web 初始化】
// Bootstrap [7.2.1] GET /api/v1/bootstrap
// 功能：聚合默认画像、方向、任务、默认推荐和能力接入状态。
func (h *Handler) Bootstrap(c *gin.Context) {
	w := c.Writer

	success(w, 200, h.app.Bootstrap())
}
