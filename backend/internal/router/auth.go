package router

import (
	"awesome-chances/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// 【路由组 8：用户认证】
// 主要目的：持久化注册/登录、会话身份查询及退出，账号与会话只存 PostgreSQL。
// 对应实现：handler/auth.go；与 Handler 使用相同编号。
func registerAuth(api *gin.RouterGroup, h *handler.Handler) {
	auth := api.Group("/auth") // 路由组 8：共享 /api/v1/auth，统一检查来源及认证流量。
	auth.Use(h.AuthBoundary())
	// 【聚合 8.1：账号注册与登录】
	auth.POST("/register", h.Register) // [8.1.1] 校验邮箱/昵称/密码，保存 Argon2id 哈希，返回用户。
	auth.POST("/login", h.Login)       // [8.1.2] 验证邮箱密码，持久化会话哈希并设置 HttpOnly Cookie。
	// 【聚合 8.2：会话身份与退出】
	auth.Match(readMethods, "/me", h.CurrentUser) // [8.2.1] 根据有效会话返回当前用户，缺失/过期会话返回 401。
	auth.POST("/logout", h.Logout)                // [8.2.2] 删除当前数据库会话并清理 Cookie，重复退出仍成功。
}
