package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"awesome-chances/backend/internal/model"
	"awesome-chances/backend/internal/recommendation"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
)

// FeedbackRepository is the only data-write boundary currently needed.
type FeedbackRepository interface {
	Add(model.Feedback) (bool, error)
}

// Catalog is loaded once at startup and treated as immutable demo data.
type App struct {
	catalog  model.Catalog
	feedback FeedbackRepository
	rules    recommendation.Config
}

func New(catalog model.Catalog, feedback FeedbackRepository) *App {
	return &App{catalog: catalog, feedback: feedback, rules: recommendation.DefaultConfig()}
}

func invalid(message string) error { return fmt.Errorf("%w: %s", ErrInvalidInput, message) }

func (a *App) DefaultProfile() model.Profile { return a.catalog.Profile }
func (a *App) Directions() []model.Direction { return a.catalog.Directions }
func (a *App) Direction(id string) (model.Direction, error) {
	for _, direction := range a.catalog.Directions {
		if direction.ID == id {
			return direction, nil
		}
	}
	return model.Direction{}, ErrNotFound
}

func (a *App) Bootstrap() model.BootstrapResponse {
	return model.BootstrapResponse{
		Catalog:         a.catalog,
		Recommendations: recommendation.Rank(a.catalog.Profile, a.catalog.Tasks, a.rules),
		Capabilities:    map[string]string{"recommendation": "demo", "github": "planned", "ai": "planned", "competitions": "demo"},
	}
}

func (a *App) ValidateProfile(profile model.Profile) error {
	if err := profile.Validate(); err != nil {
		return invalid(err.Error())
	}
	return nil
}

func (a *App) Recommend(profile model.Profile) ([]model.Recommendation, error) {
	if err := a.ValidateProfile(profile); err != nil {
		return nil, err
	}
	return recommendation.Rank(profile, a.catalog.Tasks, a.rules), nil
}

func (a *App) Tasks(query model.TaskQuery) (model.Page[model.Task], error) {
	if query.Type != "" && query.Type != "resource" && query.Type != "project" && query.Type != "competition" && query.Type != "issue" {
		return model.Page[model.Task]{}, invalid("任务类型需为 resource、project、competition 或 issue")
	}
	if query.Category != "" {
		if _, ok := model.InterestLabels[query.Category]; !ok {
			return model.Page[model.Task]{}, invalid("未知兴趣分类")
		}
	}
	if query.Page < 1 || query.Page > 1000000 || query.PageSize < 1 || query.PageSize > 100 || utf8.RuneCountInString(query.Query) > 200 {
		return model.Page[model.Task]{}, invalid("page 需为 1–1000000，pageSize 需为 1–100，q 不超过 200 字符")
	}
	matched := make([]model.Task, 0)
	needle := strings.ToLower(strings.TrimSpace(query.Query))
	for _, task := range a.catalog.Tasks {
		if query.Type != "" && query.Type != task.Type {
			continue
		}
		if query.Category != "" && query.Category != task.Category {
			continue
		}
		haystack := strings.ToLower(task.Title + " " + task.Description + " " + strings.Join(task.Tags, " "))
		if needle != "" && !strings.Contains(haystack, needle) {
			continue
		}
		matched = append(matched, task)
	}
	start := min((query.Page-1)*query.PageSize, len(matched))
	end := min(start+query.PageSize, len(matched))
	return model.Page[model.Task]{Items: matched[start:end], Total: len(matched), Page: query.Page, PageSize: query.PageSize}, nil
}

// expectedType prevents a project ID being used as a competition/issue ID.
func (a *App) Task(id, expectedType string) (model.Task, error) {
	for _, task := range a.catalog.Tasks {
		if task.ID == id && (expectedType == "" || task.Type == expectedType) {
			return task, nil
		}
	}
	return model.Task{}, ErrNotFound
}

func (a *App) PreviewPlan(id, expectedType string, profile model.Profile) (model.TaskPlan, error) {
	if err := a.ValidateProfile(profile); err != nil {
		return model.TaskPlan{}, err
	}
	task, err := a.Task(id, expectedType)
	if err != nil {
		return model.TaskPlan{}, err
	}
	result := recommendation.Rank(profile, []model.Task{task}, a.rules)[0]
	return model.TaskPlan{
		TaskID: task.ID, Title: task.Title, Steps: task.Steps, Minutes: task.Minutes,
		EstimatedDays: int(math.Ceil(float64(task.Minutes) / float64(profile.DailyMinutes))),
		Gaps:          result.Gaps, Readiness: result.Readiness, Source: task.Source, Generated: false,
	}, nil
}

func (a *App) SubmitFeedback(event model.Feedback) (model.FeedbackResponse, error) {
	if strings.TrimSpace(event.ID) == "" || len(event.ID) > 80 {
		return model.FeedbackResponse{}, invalid("事件 ID 非空且最多 80 字节")
	}
	if _, err := a.Task(event.TaskID, ""); err != nil {
		return model.FeedbackResponse{}, invalid("任务不存在")
	}
	switch event.Type {
	case "save", "unsave", "start", "complete", "too-hard", "suitable":
	default:
		return model.FeedbackResponse{}, invalid("未知反馈类型")
	}
	created, err := a.feedback.Add(event)
	if err != nil {
		return model.FeedbackResponse{}, err
	}
	return model.FeedbackResponse{Accepted: true, Created: created, Persistence: "process-memory", AbilityUpdated: false}, nil
}
