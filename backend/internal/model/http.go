package model

// Request/response contracts are separate from future database records.
type ProfileRequest struct {
	Profile Profile `json:"profile"`
}

type ChatRequest struct {
	Profile Profile `json:"profile"`
	Message string  `json:"message"`
}

type IssueSearchRequest struct {
	Profile Profile `json:"profile"`
	Query   string  `json:"query"`
}

type TaskQuery struct {
	Type     string
	Category string
	Query    string
	Page     int
	PageSize int
}

type Page[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type Meta struct {
	Mode string `json:"mode"`
}
type Response[T any] struct {
	Data T    `json:"data"`
	Meta Meta `json:"meta"`
}
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type Catalog struct {
	Profile    Profile     `json:"profile"`
	Directions []Direction `json:"directions"`
	Tasks      []Task      `json:"tasks"`
}
type BootstrapResponse struct {
	Catalog
	Recommendations []Recommendation  `json:"recommendations"`
	Capabilities    map[string]string `json:"capabilities"`
}
type ProfileValidationResponse struct {
	Profile   Profile `json:"profile"`
	Valid     bool    `json:"valid"`
	Persisted bool    `json:"persisted"`
}
type FeedbackResponse struct {
	Accepted       bool   `json:"accepted"`
	Created        bool   `json:"created"`
	Persistence    string `json:"persistence"`
	AbilityUpdated bool   `json:"abilityUpdated"`
}

// TaskPlan is a preview of existing demo steps, not AI output or a saved plan.
type TaskPlan struct {
	TaskID        string   `json:"taskId"`
	Title         string   `json:"title"`
	Steps         []string `json:"steps"`
	Minutes       int      `json:"minutes"`
	EstimatedDays int      `json:"estimatedDays"`
	Gaps          []Gap    `json:"gaps"`
	Readiness     string   `json:"readiness"`
	Source        string   `json:"source"`
	Generated     bool     `json:"generated"`
}
