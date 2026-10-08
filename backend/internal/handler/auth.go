package handler

import (
	"time"

	"awesome-chances/backend/internal/model"
	"github.com/gin-gonic/gin"
)



// 【路由组 8：用户认证】
// 主要目的：注册、登录、会话身份和退出；所有账号/会话操作通过 Service 和 PostgreSQL。
// 对应注册：router/auth.go；不直接执行 SQL，不返回密码哈希或会话令牌。


// 【聚合 8.1：账号注册与登录】
// Register [8.1.1] POST /api/v1/auth/register
// 功能：校验并持久化用户，密码保存为带随机盐的 Argon2id 哈希，注册不自动登录。
func (h *Handler) Register(c *gin.Context) {
	if !h.authReady(c) {
		return
	}
	var input model.RegisterRequest
	if !decode(c.Writer, c.Request, &input) {
		return
	}
	data, err := h.auth.Register(c.Request.Context(), input)
	if err != nil {
		serviceError(c.Writer, err)
		return
	}
	write(c.Writer, 201, model.Response[model.UserView]{Data: data, Meta: model.Meta{Mode: "persistent"}})
}

// Login [8.1.2] POST /api/v1/auth/login
// 功能：校验凭据，写入会话哈希，向浏览器设置 HttpOnly/SameSite Cookie。
func (h *Handler) Login(c *gin.Context) {
	if !h.authReady(c) {
		return
	}
	var input model.LoginRequest
	if !decode(c.Writer, c.Request, &input) {
		return
	}
	data, token, err := h.auth.Login(c.Request.Context(), input)
	if err != nil {
		serviceError(c.Writer, err)
		return
	}
	h.setSessionCookie(c, token, data.ExpiresAt)
	write(c.Writer, 200, model.Response[model.LoginResponse]{Data: data, Meta: model.Meta{Mode: "persistent"}})
}


// 【聚合 8.2：会话身份与退出】
// CurrentUser [8.2.1] GET /api/v1/auth/me
// 功能：读取数据库中的有效会话身份，过期/缺失会话返回 401。
func (h *Handler) CurrentUser(c *gin.Context) {
	if !h.authReady(c) {
		return
	}
	token, _ := c.Cookie(h.authConfig.CookieName)
	data, err := h.auth.CurrentUser(c.Request.Context(), token)
	if err != nil {
		serviceError(c.Writer, err)
		return
	}
	write(c.Writer, 200, model.Response[model.UserView]{Data: data, Meta: model.Meta{Mode: "persistent"}})
}

// Logout [8.2.2] POST /api/v1/auth/logout
// 功能：删除当前会话并清除 Cookie，不影响其他设备会话，重复退出仍成功。
func (h *Handler) Logout(c *gin.Context) {
	if !h.authReady(c) {
		return
	}
	token, _ := c.Cookie(h.authConfig.CookieName)
	if err := h.auth.Logout(c.Request.Context(), token); err != nil {
		serviceError(c.Writer, err)
		return
	}
	h.setSessionCookie(c, "", time.Time{})
	write(c.Writer, 200, model.Response[struct {
		LoggedOut bool `json:"loggedOut"`
	}]{Data: struct {
		LoggedOut bool `json:"loggedOut"`
	}{true}, Meta: model.Meta{Mode: "persistent"}})
}
