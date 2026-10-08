package handler

import (
	"awesome-chances/backend/internal/model"
	"github.com/gin-gonic/gin"
)



// 【路由组 5：画像与成长】
// 主要目的：提供默认画像、画像校验、统一推荐和行为反馈，连接成长闭环的首版演示。
// 对应注册：router/profile.go；本文件只实现这一组的业务接口。


// 【聚合 5.1：画像读取与校验】
// DefaultProfile [5.1.1] GET /api/v1/profile
// 功能：返回默认演示画像，不读取登录用户或浏览器本地画像。
func (h *Handler) DefaultProfile(c *gin.Context) {
	w := c.Writer

	success(w, 200, h.app.DefaultProfile())
}

// ValidateProfile [5.1.2] POST /api/v1/profile/validate
// 功能：校验画像字段与取值范围，返回校验结果，不保存画像。
func (h *Handler) ValidateProfile(c *gin.Context) {
	w := c.Writer
	r := c.Request

	var input model.ProfileRequest
	if !decode(w, r, &input) {
		return
	}
	if err := h.app.ValidateProfile(input.Profile); err != nil {
		serviceError(w, err)
		return
	}
	success(w, 200, model.ProfileValidationResponse{Profile: input.Profile, Valid: true, Persisted: false})
}


// 【聚合 5.2：统一任务推荐】
// Recommendations [5.2.1] POST /api/v1/recommendations
// 功能：根据提交的画像运行统一规则排序，返回推荐理由与技能缺口。
func (h *Handler) Recommendations(c *gin.Context) {
	w := c.Writer
	r := c.Request

	var input model.ProfileRequest
	if !decode(w, r, &input) {
		return
	}
	data, err := h.app.Recommend(input.Profile)
	if err != nil {
		serviceError(w, err)
		return
	}
	success(w, 200, data)
}


// 【聚合 5.3：行为反馈接收】
// Feedback [5.3.1] POST /api/v1/feedback
// 功能：校验并记录内存反馈，首次提交 201、重复 200、冲突 409，不提升能力。
func (h *Handler) Feedback(c *gin.Context) {
	w := c.Writer
	r := c.Request

	var input model.Feedback
	if !decode(w, r, &input) {
		return
	}
	data, err := h.app.SubmitFeedback(input)
	if err != nil {
		serviceError(w, err)
		return
	}
	status := 200
	if data.Created {
		status = 201
	}
	success(w, status, data)
}
