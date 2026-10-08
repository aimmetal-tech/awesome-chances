package handler

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"awesome-chances/backend/internal/model"
	"github.com/gin-gonic/gin"
)

// 【认证公共辅助：供路由组 8 使用】不对应独立 HTTP 路由。
func (h *Handler) authReady(c *gin.Context) bool {
	if h.auth != nil {
		return true
	}
	serviceError(c.Writer, model.ErrAuthUnavailable)
	c.Abort()
	return false
}

func (h *Handler) setSessionCookie(c *gin.Context, token string, expires time.Time) {
	age := int(time.Until(expires).Seconds())
	if token == "" {
		age = -1
		expires = time.Unix(1, 0)
	}
	http.SetCookie(c.Writer, &http.Cookie{Name: h.authConfig.CookieName, Value: token, Path: "/", HttpOnly: true, Secure: h.authConfig.CookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: age, Expires: expires})
}

// AuthBoundary 使用明确的 origin 和有界 IP 窗口保护认证入口，禁止跨站提交。
// 不信任 X-Forwarded-For；未来部署代理时需显式配置可信代理。
func (h *Handler) AuthBoundary() gin.HandlerFunc {
	type window struct {
		until time.Time
		count int
	}
	entries := make(map[string]window)
	var mu sync.Mutex
	return func(c *gin.Context) {
		if !h.authReady(c) {
			return
		}
		if c.Request.Method == http.MethodPost {
			origin := c.GetHeader("Origin")
			if origin != "" {
				u, err := url.Parse(origin)
				allowed := err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.User == nil && u.Host == c.Request.Host && u.Path == "" && u.RawQuery == "" && u.Fragment == ""
				for _, expected := range h.authConfig.AllowedOrigins {
					allowed = allowed || origin == expected
				}
				if !allowed {
					failure(c.Writer, 403, "invalid_origin", "不允许跨站认证请求")
					c.Abort()
					return
				}
			}
			if !strings.HasPrefix(strings.ToLower(c.GetHeader("Content-Type")), "application/json") && c.Request.URL.Path != "/api/v1/auth/logout" {
				failure(c.Writer, 415, "unsupported_media_type", "认证请求需为 application/json")
				c.Abort()
				return
			}
		}
		limit := h.authConfig.RateLimitPerMinute
		if limit < 1 {
			limit = 30
		}
		ip, now := c.ClientIP(), time.Now()
		mu.Lock()
		entry := entries[ip]
		if !now.Before(entry.until) {
			if len(entries) >= 1024 {
				for key, item := range entries {
					if !now.Before(item.until) {
						delete(entries, key)
					}
				}
			}
			if len(entries) >= 1024 {
				mu.Unlock()
				failure(c.Writer, 429, "rate_limited", "认证请求较多，请稍后重试")
				c.Abort()
				return
			}
			entry = window{until: now.Add(time.Minute)}
		}
		entry.count++
		entries[ip] = entry
		mu.Unlock()
		if entry.count > limit {
			c.Header("Retry-After", "60")
			failure(c.Writer, 429, "rate_limited", "认证请求过于频繁，请稍后重试")
			c.Abort()
			return
		}
		c.Next()
	}
}
