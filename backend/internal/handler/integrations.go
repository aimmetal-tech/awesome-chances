package handler

import (
	"awesome-chances/backend/internal/model"
	"github.com/gin-gonic/gin"
	"strings"
	"unicode/utf8"
)

// 插件入口聚合：承接导学对话与实时开源搜索；Provider 未接入时明确返回 501。



// 【路由组 3：方向导学】
// 主要目的：介绍学习方向、岗位工作及技术栈，并预留 AI 导学对话入口。
// 所属 Router：router/guidance.go；本段实现聚合 3.2，按编号定位对应接口。


// 【聚合 3.2：AI 导学对话】
// GuidanceChat [3.2.1] POST /api/v1/guidance/chat
// 功能：校验画像和消息，合法请求返回 501，说明生成式 AI 尚未接入。
func (h *Handler) GuidanceChat(c *gin.Context) {
	w := c.Writer
	r := c.Request

	var input model.ChatRequest
	if !decode(w, r, &input) {
		return
	}
	if err := h.app.ValidateProfile(input.Profile); err != nil {
		serviceError(w, err)
		return
	}
	if strings.TrimSpace(input.Message) == "" || utf8.RuneCountInString(input.Message) > 2000 {
		failure(w, 400, "invalid_input", "message 需为 1–2000 字符")
		return
	}
	failure(w, 501, "feature_not_connected", "导学对话尚未接入 GenerationProvider，请先使用预置方向介绍")
}



// 【路由组 4：开源贡献】
// 主要目的：查询示例 Issue，并预留根据画像实时检索开源贡献任务的入口。
// 所属 Router：router/open_source.go；本段实现聚合 4.2，按编号定位对应接口。


// 【聚合 4.2：实时开源搜索】
// SearchIssues [4.2.1] POST /api/v1/open-source/search
// 功能：校验画像和搜索词，合法请求返回 501，说明 GitHub Provider 尚未接入。
func (h *Handler) SearchIssues(c *gin.Context) {
	w := c.Writer
	r := c.Request

	var input model.IssueSearchRequest
	if !decode(w, r, &input) {
		return
	}
	if err := h.app.ValidateProfile(input.Profile); err != nil {
		serviceError(w, err)
		return
	}
	if strings.TrimSpace(input.Query) == "" || utf8.RuneCountInString(input.Query) > 200 {
		failure(w, 400, "invalid_input", "query 需为 1–200 字符")
		return
	}
	failure(w, 501, "feature_not_connected", "实时开源搜索尚未接入 RepositoryProvider，请先使用示例 Issue 列表")
}
