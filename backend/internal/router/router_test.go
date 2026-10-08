package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"awesome-chances/backend/internal/adapters/demo"
	"awesome-chances/backend/internal/adapters/memory"
	"awesome-chances/backend/internal/handler"
	"awesome-chances/backend/internal/service"
)

func TestBootstrapAndProfileValidation(t *testing.T) {
	catalog, err := demo.Load()
	if err != nil {
		t.Fatal(err)
	}
	handler := New(handler.New(service.New(catalog, memory.NewFeedbackStore())))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/bootstrap", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"mode":"demo"`) {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, test := range []struct {
		name   string
		body   string
		status int
	}{
		{"empty profile", `{"profile":{}}`, 400},
		{"malformed JSON", `{`, 400},
		{"trailing JSON", `{} {}`, 400},
		{"unknown field", `{"unexpected":true}`, 400},
		{"oversized body", strings.Repeat("x", 70<<10), 400},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest("POST", "/api/v1/recommendations", strings.NewReader(test.body)))
			if w.Code != test.status {
				t.Fatal(w.Code, w.Body.String())
			}
		})
	}
	profile := catalog.Profile
	for _, invalid := range []bool{false, true} {
		if invalid {
			skill := profile.Skills["git"]
			skill.Level = 1.5
			profile.Skills["git"] = skill
		}
		body, _ := json.Marshal(map[string]any{"profile": profile})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("POST", "/api/v1/recommendations", bytes.NewReader(body)))
		want := 200
		if invalid {
			want = 400
		}
		if w.Code != want {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}

func TestFeedbackValidationAndIdempotency(t *testing.T) {
	catalog, _ := demo.Load()
	handler := New(handler.New(service.New(catalog, memory.NewFeedbackStore())))
	for _, test := range []struct {
		body   string
		status int
	}{
		{`{"id":"test-1","taskId":"api-project","type":"start"}`, 201},
		{`{"id":"test-1","taskId":"api-project","type":"start"}`, 200},
		{`{"id":"test-1","taskId":"api-project","type":"complete"}`, 409},
		{`{"id":"test-2","taskId":"nonexistent","type":"start"}`, 400},
		{`{"id":"test-3","taskId":"api-project","type":"raise-skill"}`, 400},
	} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/feedback", strings.NewReader(test.body)))
		if w.Code != test.status {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}
