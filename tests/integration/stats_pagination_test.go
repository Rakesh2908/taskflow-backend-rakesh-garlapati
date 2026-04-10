package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestProjectStats(t *testing.T) {
	srv := newTestServer(t)

	token := mustRegisterAndLogin(t, srv.URL, "Owner", "stats-owner@example.com", "password123")

	// Create project
	res := makeRequest(t, http.MethodPost, srv.URL+"/projects", map[string]any{
		"name": "P1",
	}, token)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create project status = %d", res.StatusCode)
	}
	var proj struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&proj); err != nil {
		t.Fatalf("decode project: %v", err)
	}

	// Create 3 tasks
	for i := 0; i < 3; i++ {
		res = makeRequest(t, http.MethodPost, srv.URL+"/projects/"+proj.ID+"/tasks", map[string]any{
			"title":    "T",
			"priority": "low",
		}, token)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("create task status = %d", res.StatusCode)
		}
	}

	// Fetch stats: should show at least todo=3
	res = makeRequest(t, http.MethodGet, srv.URL+"/projects/"+proj.ID+"/stats", nil, token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("stats status = %d", res.StatusCode)
	}

	var payload struct {
		ByStatus map[string]int `json:"by_status"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode stats: %v", err)
	}
	if payload.ByStatus["todo"] != 3 {
		t.Fatalf("expected todo=3, got %d", payload.ByStatus["todo"])
	}
}

func TestPaginationProjectsAndTasks(t *testing.T) {
	srv := newTestServer(t)

	token := mustRegisterAndLogin(t, srv.URL, "Owner", "paging@example.com", "password123")

	// Create 2 projects
	var p1, p2 struct{ ID string }
	res := makeRequest(t, http.MethodPost, srv.URL+"/projects", map[string]any{"name": "P1"}, token)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create project1 status = %d", res.StatusCode)
	}
	_ = json.NewDecoder(res.Body).Decode(&p1)

	res = makeRequest(t, http.MethodPost, srv.URL+"/projects", map[string]any{"name": "P2"}, token)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create project2 status = %d", res.StatusCode)
	}
	_ = json.NewDecoder(res.Body).Decode(&p2)

	// List projects with limit=1 => should return exactly 1
	res = makeRequest(t, http.MethodGet, srv.URL+"/projects?page=1&limit=1", nil, token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("list projects status = %d", res.StatusCode)
	}
	var projects []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&projects); err != nil {
		t.Fatalf("decode projects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	// Create 3 tasks in p1
	for i := 0; i < 3; i++ {
		res = makeRequest(t, http.MethodPost, srv.URL+"/projects/"+p1.ID+"/tasks", map[string]any{
			"title":    "T",
			"priority": "low",
		}, token)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("create task status = %d", res.StatusCode)
		}
	}

	// List tasks with limit=2 => should return exactly 2
	res = makeRequest(t, http.MethodGet, srv.URL+"/projects/"+p1.ID+"/tasks?page=1&limit=2", nil, token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("list tasks status = %d", res.StatusCode)
	}
	var tasks []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&tasks); err != nil {
		t.Fatalf("decode tasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

