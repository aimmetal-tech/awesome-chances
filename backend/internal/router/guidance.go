package router

import (
	"awesome-chances/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// 【路由组 3：方向导学】
// 主要目的：介绍学习方向、岗位工作及技术栈，并预留 AI 导学对话入口。
// 对应实现：handler/catalog.go（3.1 方向）、handler/integrations.go（3.2 对话）；路由编号与 Handler 注释保持一致。
func registerGuidance(api *gin.RouterGroup, h *handler.Handler) {
	group := api.Group("/guidance") // 路由组 3：统一前缀 /api/v1/guidance。
	// 【聚合 3.1：学习方向查询】
	directions := group.Group("/directions")           // 聚合 3.1：共享 /directions 前缀。
	directions.Match(readMethods, "", h.Directions)    // [3.1.1] 返回全部预置学习方向。
	directions.Match(readMethods, "/:id", h.Direction) // [3.1.2] 返回指定方向的工作内容、技术栈和预置学习顺序。
	// 【聚合 3.2：AI 导学对话】
	group.POST("/chat", h.GuidanceChat) // [3.2.1] 校验画像和消息，合法请求返回 501，说明生成式 AI 尚未接入。
}
