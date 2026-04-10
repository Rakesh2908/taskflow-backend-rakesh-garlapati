package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Rakesh2908/taskflow/api/handler"
	"github.com/Rakesh2908/taskflow/config"
	"github.com/Rakesh2908/taskflow/repository"
	"github.com/Rakesh2908/taskflow/routes"
	"github.com/Rakesh2908/taskflow/service"
	"github.com/Rakesh2908/taskflow/tests/testhelper"
	"github.com/go-playground/validator/v10"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	pool := testhelper.SetupTestDB(t)

	cfg := config.Config{
		DBUrl:      "",
		JWTSecret:  "test-jwt-secret",
		Port:       "8080",
		BcryptCost: 12,
	}

	usersRepo := repository.NewUserRepository(pool)
	projectsRepo := repository.NewProjectRepository(pool)
	tasksRepo := repository.NewTaskRepository(pool)

	authSvc := service.NewAuthService(cfg, usersRepo)
	projectSvc := service.NewProjectService(projectsRepo, tasksRepo)
	taskSvc := service.NewTaskService(tasksRepo, projectsRepo)

	v := validator.New()
	authH := handler.NewAuthHandler(authSvc, v)
	projectH := handler.NewProjectHandler(projectSvc, v)
	taskH := handler.NewTaskHandler(taskSvc, projectSvc, v)

	r := routes.NewRouter(&cfg, authH, projectH, taskH)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

func makeRequest(t *testing.T, method, url string, body any, token string) *http.Response {
	t.Helper()

	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		rdr = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	return res
}

func mustRegisterAndLogin(t *testing.T, serverURL, name, email, pass string) string {
	t.Helper()

	res := makeRequest(t, http.MethodPost, serverURL+"/auth/register", map[string]any{
		"name":     name,
		"email":    email,
		"password": pass,
	}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("register status = %d", res.StatusCode)
	}

	res = makeRequest(t, http.MethodPost, serverURL+"/auth/login", map[string]any{
		"email":    email,
		"password": pass,
	}, "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d", res.StatusCode)
	}

	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if payload.Token == "" || strings.Count(payload.Token, ".") != 2 {
		t.Fatalf("expected jwt-like token, got %q", payload.Token)
	}
	return payload.Token
}

func TestRegisterSuccess(t *testing.T) {
	srv := newTestServer(t)

	res := makeRequest(t, http.MethodPost, srv.URL+"/auth/register", map[string]any{
		"name":     "Alice",
		"email":    "alice@example.com",
		"password": "password123",
	}, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d", res.StatusCode)
	}

	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Token == "" || strings.Count(payload.Token, ".") != 2 {
		t.Fatalf("expected jwt-like token, got %q", payload.Token)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	srv := newTestServer(t)

	body := map[string]any{
		"name":     "Alice",
		"email":    "dup@example.com",
		"password": "password123",
	}
	res := makeRequest(t, http.MethodPost, srv.URL+"/auth/register", body, "")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("first register status = %d", res.StatusCode)
	}

	res = makeRequest(t, http.MethodPost, srv.URL+"/auth/register", body, "")
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("second register status = %d", res.StatusCode)
	}
}

func TestLoginSuccess(t *testing.T) {
	srv := newTestServer(t)

	token := mustRegisterAndLogin(t, srv.URL, "Alice", "login@example.com", "password123")
	if token == "" {
		t.Fatalf("empty token")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	srv := newTestServer(t)

	_ = mustRegisterAndLogin(t, srv.URL, "Alice", "wrongpass@example.com", "password123")

	res := makeRequest(t, http.MethodPost, srv.URL+"/auth/login", map[string]any{
		"email":    "wrongpass@example.com",
		"password": "not-the-password",
	}, "")
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", res.StatusCode)
	}
}

