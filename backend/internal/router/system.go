package router

import (
	"awesome-chances/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// 【路由组 7：系统与初始化】
// 主要目的：检查后端存活，并提供 Web 首次加载所需的演示数据和接入状态。
// 对应实现：handler/system.go；路由编号与 Handler 注释保持一致。
func registerSystem(root *gin.Engine, api *gin.RouterGroup, h *handler.Handler) {
	// 【聚合 7.1：服务健康检查】
	root.Match(readMethods, "/healthz", h.Health) // [7.1.1] 返回服务存活状态，不检查数据库或外部 Provider。
	// 【聚合 7.2：Web 初始化】
	api.Match(readMethods, "/bootstrap", h.Bootstrap) // [7.2.1] 聚合默认画像、方向、任务、默认推荐和能力接入状态。
}
