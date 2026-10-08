package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"awesome-chances/backend/internal/adapters/demo"
	"awesome-chances/backend/internal/adapters/memory"
	"awesome-chances/backend/internal/handler"
	"awesome-chances/backend/internal/model"
	"awesome-chances/backend/internal/service"
)

func testAPI(t *testing.T) (http.Handler, model.Catalog) {
	t.Helper()
	catalog, err := demo.Load()
	if err != nil {
		t.Fatal(err)
	}
	return New(handler.New(service.New(catalog, memory.NewFeedbackStore()))), catalog
}

func request(api http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	api.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewReader(body)))
	return w
}

func readData[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var response model.Response[T]
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err, w.Body.String())
	}
	if response.Meta.Mode != "demo" || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("missing demo metadata or cache policy")
	}
	return response.Data
}

func TestAggregatedRoutesAndTypeIsolation(t *testing.T) {
	api, _ := testAPI(t)
	for _, path := range []string{
		"/healthz", "/api/v1/bootstrap", "/api/v1/profile",
		"/api/v1/guidance/directions", "/api/v1/guidance/directions/backend",
		"/api/v1/tasks", "/api/v1/tasks/api-project",
		"/api/v1/competitions", "/api/v1/competitions/algorithm-prep",
		"/api/v1/projects", "/api/v1/projects/api-project",
		"/api/v1/open-source/issues", "/api/v1/open-source/issues/docs-issue",
	} {
		t.Run(path, func(t *testing.T) {
			w := request(api, "GET", path, nil)
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
		})
	}
	for path, taskType := range map[string]string{
		"/api/v1/competitions": "competition", "/api/v1/projects": "project", "/api/v1/open-source/issues": "issue",
	} {
		page := readData[model.Page[model.Task]](t, request(api, "GET", path, nil))
		if len(page.Items) == 0 {
			t.Fatal("empty module", path)
		}
		for _, task := range page.Items {
			if task.Type != taskType {
				t.Fatal("cross-module task", path, task.ID)
			}
		}
	}
	for _, path := range []string{"/api/v1/tasks/missing", "/api/v1/competitions/api-project", "/api/v1/projects/docs-issue", "/api/v1/open-source/issues/api-project", "/api/v1/guidance/directions/missing"} {
		w := request(api, "GET", path, nil)
		if w.Code != 404 || !strings.Contains(w.Body.String(), `"code":"not_found"`) {
			t.Fatal(path, w.Code, w.Body.String())
		}
	}
	w := request(api, "DELETE", "/api/v1/tasks", nil)
	if w.Code != 405 || !strings.Contains(w.Header().Get("Allow"), "GET") {
		t.Fatal("method restriction", w.Code, w.Header())
	}
	if request(api, "GET", "/api/v1/missing", nil).Code != 404 {
		t.Fatal("unknown route")
	}
	bootstrap := readData[model.BootstrapResponse](t, request(api, "GET", "/api/v1/bootstrap", nil))
	if len(bootstrap.Tasks) != 9 || len(bootstrap.Directions) != 4 || len(bootstrap.Recommendations) != 9 || bootstrap.Profile.Name == "" || bootstrap.Capabilities["ai"] != "planned" {
		t.Fatal("Web bootstrap contract changed")
	}
}

func TestTaskFilteringPaginationAndBoundaries(t *testing.T) {
	api, _ := testAPI(t)
	page := readData[model.Page[model.Task]](t, request(api, "GET", "/api/v1/tasks?type=project&category=backend&q=API&pageSize=1", nil))
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != "api-project" {
		t.Fatal(page)
	}
	page = readData[model.Page[model.Task]](t, request(api, "GET", "/api/v1/tasks?page=2&pageSize=2", nil))
	if page.Total != 9 || len(page.Items) != 2 || page.Items[0].ID != "data-project" {
		t.Fatal(page)
	}
	for _, path := range []string{"/api/v1/tasks?q=nonexistent", "/api/v1/tasks?page=1000000"} {
		w := request(api, "GET", path, nil)
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"items":[]`) {
			t.Fatal("empty array contract", w.Code, w.Body.String())
		}
	}
	for _, suffix := range []string{"page=0", "page=-1", "page=1000001", "page=x", "page=", "pageSize=0", "pageSize=101", "type=unknown", "category=unknown", "q=" + url.QueryEscape(strings.Repeat("字", 201))} {
		w := request(api, "GET", "/api/v1/tasks?"+suffix, nil)
		if w.Code != 400 {
			t.Fatal(suffix, w.Code, w.Body.String())
		}
	}
	if request(api, "GET", "/api/v1/projects?type=issue", nil).Code != 400 {
		t.Fatal("module type override")
	}
}

func TestProfileValidationAndPlanPreviewAreNotPersistence(t *testing.T) {
	api, catalog := testAPI(t)
	profile := catalog.Profile
	profile.Name = "新的测试者"
	profile.DailyMinutes = 15
	profile.Skills = map[string]model.SkillEstimate{}
	for key := range model.SkillLabels {
		profile.Skills[key] = model.SkillEstimate{Level: 0, Confidence: .25}
	}
	body, _ := json.Marshal(model.ProfileRequest{Profile: profile})
	w := request(api, "POST", "/api/v1/profile/validate", body)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	validated := readData[model.ProfileValidationResponse](t, w)
	if !validated.Valid || validated.Persisted || validated.Profile.Name != profile.Name {
		t.Fatal(validated)
	}
	for _, path := range []string{"/api/v1/tasks/api-project/plan", "/api/v1/projects/api-project/guide", "/api/v1/competitions/algorithm-prep/plan"} {
		w := request(api, "POST", path, body)
		if w.Code != 200 {
			t.Fatal(path, w.Code, w.Body.String())
		}
		plan := readData[model.TaskPlan](t, w)
		if len(plan.Steps) == 0 || len(plan.Gaps) == 0 || plan.Readiness != "prepare" || plan.Source != "demo" || plan.Generated {
			t.Fatal(plan)
		}
		if plan.TaskID == "api-project" && plan.EstimatedDays != 12 {
			t.Fatal("daily time ignored", plan)
		}
	}
	if request(api, "POST", "/api/v1/competitions/api-project/plan", body).Code != 404 {
		t.Fatal("cross-module plan")
	}
	if request(api, "POST", "/api/v1/projects/api-project/guide", []byte(`{"profile":{}}`)).Code != 400 {
		t.Fatal("invalid plan profile")
	}
	current := readData[model.Profile](t, request(api, "GET", "/api/v1/profile", nil))
	if current.Name != catalog.Profile.Name || current.DailyMinutes != 45 || current.Skills["python"].Level != .4 {
		t.Fatal("preview changed server profile", current)
	}
}

func TestUnconnectedIntegrationsHaveHonestStatus(t *testing.T) {
	api, catalog := testAPI(t)
	chat, _ := json.Marshal(model.ChatRequest{Profile: catalog.Profile, Message: "我该学什么？"})
	search, _ := json.Marshal(model.IssueSearchRequest{Profile: catalog.Profile, Query: "Go"})
	for _, test := range []struct {
		path string
		body []byte
	}{
		{"/api/v1/guidance/chat", chat}, {"/api/v1/open-source/search", search},
	} {
		w := request(api, "POST", test.path, test.body)
		if w.Code != 501 || !strings.Contains(w.Body.String(), `"code":"feature_not_connected"`) {
			t.Fatal(w.Code, w.Body.String())
		}
		if request(api, "POST", test.path, []byte(`{}`)).Code != 400 {
			t.Fatal("invalid integration input accepted")
		}
	}
	for _, message := range []string{" ", strings.Repeat("字", 2001)} {
		body, _ := json.Marshal(model.ChatRequest{Profile: catalog.Profile, Message: message})
		if request(api, "POST", "/api/v1/guidance/chat", body).Code != 400 {
			t.Fatal("invalid chat length")
		}
	}
	for _, query := range []string{" ", strings.Repeat("字", 201)} {
		body, _ := json.Marshal(model.IssueSearchRequest{Profile: catalog.Profile, Query: query})
		if request(api, "POST", "/api/v1/open-source/search", body).Code != 400 {
			t.Fatal("invalid search length")
		}
	}
}
