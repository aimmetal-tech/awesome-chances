package router

import (
	"awesome-chances/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// 【路由组 5：画像与成长】
// 主要目的：提供默认画像、画像校验、统一推荐和行为反馈，连接成长闭环的首版演示。
// 对应实现：handler/profile.go；路由编号与 Handler 注释保持一致。
func registerProfile(api *gin.RouterGroup, h *handler.Handler) {
	// 【聚合 5.1：画像读取与校验】
	profile := api.Group("/profile")                 // 聚合 5.1：共享 /profile 前缀。
	profile.Match(readMethods, "", h.DefaultProfile) // [5.1.1] 返回默认演示画像，不读取登录用户或浏览器本地画像。
	profile.POST("/validate", h.ValidateProfile)     // [5.1.2] 校验画像字段与取值范围，返回校验结果，不保存画像。
	// 【聚合 5.2：统一任务推荐】
	api.POST("/recommendations", h.Recommendations) // [5.2.1] 根据提交的画像运行统一规则排序，返回推荐理由与技能缺口。
	// 【聚合 5.3：行为反馈接收】
	api.POST("/feedback", h.Feedback) // [5.3.1] 校验并记录内存反馈，首次提交 201、重复 200、冲突 409，不提升能力。
}
