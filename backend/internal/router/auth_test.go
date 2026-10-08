package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"awesome-chances/backend/internal/adapters/memory"
	"awesome-chances/backend/internal/handler"
	"awesome-chances/backend/internal/model"
	"awesome-chances/backend/internal/security"
	"awesome-chances/backend/internal/service"
)

// Test-only repository; runtime authentication is always backed by PostgreSQL.
type authFixture struct {
	mu       sync.Mutex
	users    map[string]model.UserRecord
	sessions map[string]model.SessionRecord
}

func (r *authFixture) CreateUser(_ context.Context, user model.UserRecord) (model.UserRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[user.Email]; ok {
		return model.UserRecord{}, model.ErrEmailExists
	}
	r.users[user.Email] = user
	return user, nil
}
func (r *authFixture) UserByEmail(_ context.Context, email string) (model.UserRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[email]
	if !ok {
		return u, model.ErrUserNotFound
	}
	return u, nil
}
func (r *authFixture) CreateSession(_ context.Context, s model.SessionRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.TokenHash] = s
	return nil
}
func (r *authFixture) SessionUser(_ context.Context, hash string, now time.Time) (model.UserRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[hash]
	if ok && now.Before(s.ExpiresAt) {
		for _, u := range r.users {
			if u.ID == s.UserID {
				return u, nil
			}
		}
	}
	return model.UserRecord{}, model.ErrUnauthenticated
}
func (r *authFixture) DeleteSession(_ context.Context, hash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, hash)
	return nil
}

func authAPI(t *testing.T, limit int) (http.Handler, *authFixture) {
	t.Helper()
	_, catalog := testAPI(t)
	repo := &authFixture{users: map[string]model.UserRecord{}, sessions: map[string]model.SessionRecord{}}
	cfg := model.AuthConfig{CookieName: "test_session", CookieSecure: true, SessionTTL: time.Hour, MinPasswordLength: 15, MaxHashJobs: 2, RateLimitPerMinute: limit, AllowedOrigins: []string{"http://localhost:3000"}}
	auth, err := service.NewAuth(repo, cfg)
	if err != nil {
		t.Fatal(err)
	}
	return New(handler.New(service.New(catalog, memory.NewFeedbackStore())).WithAuth(auth, cfg)), repo
}

func authRequest(api http.Handler, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	return w
}

func TestRegistrationLoginSessionExpiryAndLogout(t *testing.T) {
	api, repo := authAPI(t, 100)
	password := "correct horse battery staple"
	register := `{"email":" Learner@Example.com ","displayName":" 学习者 ","password":"` + password + `"}`
	w := authRequest(api, "POST", "/api/v1/auth/register", register, nil)
	if w.Code != 201 || strings.Contains(w.Body.String(), password) || strings.Contains(w.Body.String(), "passwordHash") || strings.Contains(w.Body.String(), "argon2id") {
		t.Fatal(w.Code, w.Body.String())
	}
	var response model.Response[model.UserView]
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Meta.Mode != "persistent" || response.Data.Email != "learner@example.com" || response.Data.DisplayName != "学习者" {
		t.Fatal(response)
	}
	stored := repo.users["learner@example.com"]
	if stored.PasswordHash == password || !strings.HasPrefix(stored.PasswordHash, "$argon2id$") {
		t.Fatal("plaintext password storage")
	}
	if authRequest(api, "POST", "/api/v1/auth/register", register, nil).Code != 409 {
		t.Fatal("duplicate registration")
	}
	wrong := authRequest(api, "POST", "/api/v1/auth/login", `{"email":"learner@example.com","password":"wrong password"}`, nil)
	missing := authRequest(api, "POST", "/api/v1/auth/login", `{"email":"missing@example.com","password":"wrong password"}`, nil)
	if wrong.Code != 401 || missing.Code != 401 || wrong.Body.String() != missing.Body.String() {
		t.Fatal("credential error leaked account existence")
	}
	login := `{"email":"LEARNER@example.com","password":"` + password + `"}`
	w = authRequest(api, "POST", "/api/v1/auth/login", login, nil)
	if w.Code != 200 || strings.Contains(w.Body.String(), "token") {
		t.Fatal(w.Code, w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatal("session cookie missing")
	}
	cookie := cookies[0]
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" || cookie.MaxAge <= 0 {
		t.Fatal("cookie attributes")
	}
	if _, exists := repo.sessions[cookie.Value]; exists {
		t.Fatal("raw session token persisted")
	}
	key := security.TokenHash(cookie.Value)
	if _, exists := repo.sessions[key]; !exists {
		t.Fatal("session hash missing")
	}
	if authRequest(api, "GET", "/api/v1/auth/me", "", nil).Code != 401 {
		t.Fatal("anonymous user accepted")
	}
	if authRequest(api, "GET", "/api/v1/auth/me", "", cookie).Code != 200 {
		t.Fatal("valid session rejected")
	}
	w = authRequest(api, "POST", "/api/v1/auth/logout", "", cookie)
	if w.Code != 200 || len(repo.sessions) != 0 || w.Result().Cookies()[0].MaxAge != -1 {
		t.Fatal("logout did not revoke session")
	}
	if authRequest(api, "GET", "/api/v1/auth/me", "", cookie).Code != 401 {
		t.Fatal("revoked session accepted")
	}
	if authRequest(api, "POST", "/api/v1/auth/logout", "", cookie).Code != 200 {
		t.Fatal("logout not idempotent")
	}
	w = authRequest(api, "POST", "/api/v1/auth/login", login, nil)
	cookie = w.Result().Cookies()[0]
	key = security.TokenHash(cookie.Value)
	session := repo.sessions[key]
	session.ExpiresAt = time.Now().Add(-time.Second)
	repo.sessions[key] = session
	if authRequest(api, "GET", "/api/v1/auth/me", "", cookie).Code != 401 {
		t.Fatal("expired session accepted")
	}
}

func TestAuthInputOriginRateLimitAndDisabledDatabase(t *testing.T) {
	api, _ := authAPI(t, 100)
	for _, body := range []string{`{}`, `{"email":"bad","displayName":"test","password":"correct horse battery staple"}`, `{"email":"test@example.com","displayName":"test","password":"short"}`, `{"unexpected":true}`, `{} {}`} {
		if authRequest(api, "POST", "/api/v1/auth/register", body, nil).Code != 400 {
			t.Fatal("invalid registration accepted")
		}
	}
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://attacker.example")
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatal("cross-origin auth request accepted")
	}
	w = request(api, "POST", "/api/v1/auth/login", []byte(`{}`))
	if w.Code != 415 {
		t.Fatal("form/non-JSON login accepted")
	}
	limited, _ := authAPI(t, 1)
	if authRequest(limited, "HEAD", "/api/v1/auth/me", "", nil).Code != 401 || authRequest(limited, "GET", "/api/v1/auth/me", "", nil).Code != 429 {
		t.Fatal("auth limiter")
	}
	disabled, _ := testAPI(t)
	for _, item := range []struct{ method, path string }{{"POST", "register"}, {"POST", "login"}, {"GET", "me"}, {"HEAD", "me"}, {"POST", "logout"}} {
		w := authRequest(disabled, item.method, "/api/v1/auth/"+item.path, `{}`, nil)
		if w.Code != 503 || !strings.Contains(w.Body.String(), "auth_unavailable") {
			t.Fatal("database disabled status", w.Code, w.Body.String())
		}
	}
}
