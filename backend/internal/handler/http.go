package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"awesome-chances/backend/internal/model"
	"awesome-chances/backend/internal/service"
)

// 【公共 HTTP 辅助：供全部路由组使用】
// 主要目的：组装 Handler、统一 JSON 解析/响应及错误映射；业务接口按职责聚合，按组标题定位。
type Handler struct {
	app        *service.App
	auth       *service.Auth
	authConfig model.AuthConfig
}

func New(app *service.App) *Handler { return &Handler{app: app} }

func (h *Handler) WithAuth(auth *service.Auth, cfg model.AuthConfig) *Handler {
	h.auth = auth
	h.authConfig = cfg
	return h
}

// RouteError 为 Gin 的 404/405 与中间件提供统一错误结构。
func RouteError(code, message string) model.ErrorResponse {
	return model.ErrorResponse{Error: model.ErrorDetail{Code: code, Message: message}}
}

func write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func success[T any](w http.ResponseWriter, status int, data T) {
	write(w, status, model.Response[T]{Data: data, Meta: model.Meta{Mode: "demo"}})
}

func failure(w http.ResponseWriter, status int, code, message string) {
	write(w, status, model.ErrorResponse{Error: model.ErrorDetail{Code: code, Message: message}})
}

func serviceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrEmailExists):
		failure(w, 409, "email_exists", err.Error())
	case errors.Is(err, model.ErrCredentials):
		failure(w, 401, "invalid_credentials", err.Error())
	case errors.Is(err, model.ErrUnauthenticated):
		failure(w, 401, "unauthenticated", err.Error())
	case errors.Is(err, model.ErrAuthUnavailable):
		failure(w, 503, "auth_unavailable", err.Error())
	case errors.Is(err, model.ErrAuthBusy):
		failure(w, 429, "auth_busy", err.Error())
	case errors.Is(err, service.ErrInvalidInput):
		failure(w, 400, "invalid_input", err.Error())
	case errors.Is(err, service.ErrNotFound):
		failure(w, 404, "not_found", "请求的资源不存在")
	case errors.Is(err, model.ErrFeedbackConflict):
		failure(w, 409, "feedback_conflict", err.Error())
	default:
		failure(w, 500, "internal_error", "服务暂时无法处理请求")
	}
}

func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		failure(w, 400, "invalid_json", "请求需为有效 JSON，大小不超过 64 KiB，且不能包含未知字段")
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		failure(w, 400, "invalid_json", "请求只能包含一个 JSON 对象")
		return false
	}
	return true
}
