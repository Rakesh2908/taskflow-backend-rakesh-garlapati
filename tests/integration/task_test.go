package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateTask(t *testing.T) {
	srv := newTestServer(t)

	token := mustRegisterAndLogin(t, srv.URL, "Owner", "owner@example.com", "password123")

	// Create project
	res := makeRequest(t, http.MethodPost, srv.URL+"/projects", map[string]any{
		"name":        "P1",
		"description": "D1",
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
	if proj.ID == "" {
		t.Fatalf("empty project id")
	}

	// Create task
	res = makeRequest(t, http.MethodPost, srv.URL+"/projects/"+proj.ID+"/tasks", map[string]any{
		"title":       "T1",
		"description": "TD",
		"priority":    "low",
	}, token)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create task status = %d", res.StatusCode)
	}
}

func TestForbiddenDelete(t *testing.T) {
	srv := newTestServer(t)

	tokenA := mustRegisterAndLogin(t, srv.URL, "UserA", "a@example.com", "password123")
	tokenB := mustRegisterAndLogin(t, srv.URL, "UserB", "b@example.com", "password123")

	// A creates project
	res := makeRequest(t, http.MethodPost, srv.URL+"/projects", map[string]any{
		"name":        "P1",
		"description": "D1",
	}, tokenA)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create project status = %d", res.StatusCode)
	}
	var proj struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&proj); err != nil {
		t.Fatalf("decode project: %v", err)
	}

	// A creates task
	res = makeRequest(t, http.MethodPost, srv.URL+"/projects/"+proj.ID+"/tasks", map[string]any{
		"title":    "T1",
		"priority": "low",
	}, tokenA)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create task status = %d", res.StatusCode)
	}
	var task struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&task); err != nil {
		t.Fatalf("decode task: %v", err)
	}

	// B tries to delete A's task
	res = makeRequest(t, http.MethodDelete, srv.URL+"/tasks/"+task.ID, nil, tokenB)
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("delete status = %d", res.StatusCode)
	}
}

func TestDeleteTaskSetsJSONContentType(t *testing.T) {
	srv := newTestServer(t)

	token := mustRegisterAndLogin(t, srv.URL, "Owner", "ct-delete@example.com", "password123")

	// Create project
	res := makeRequest(t, http.MethodPost, srv.URL+"/projects", map[string]any{"name": "P1"}, token)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create project status = %d", res.StatusCode)
	}
	var proj struct{ ID string `json:"id"` }
	_ = json.NewDecoder(res.Body).Decode(&proj)

	// Create task
	res = makeRequest(t, http.MethodPost, srv.URL+"/projects/"+proj.ID+"/tasks", map[string]any{
		"title":    "T1",
		"priority": "low",
	}, token)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create task status = %d", res.StatusCode)
	}
	var task struct{ ID string `json:"id"` }
	_ = json.NewDecoder(res.Body).Decode(&task)

	// Delete task should still set JSON content-type
	res = makeRequest(t, http.MethodDelete, srv.URL+"/tasks/"+task.ID, nil, token)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct == "" {
		t.Fatalf("expected Content-Type to be set")
	}
}

func TestFilterByStatus(t *testing.T) {
	srv := newTestServer(t)

	token := mustRegisterAndLogin(t, srv.URL, "Owner", "filter@example.com", "password123")

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

	// Create 3 tasks (all start as todo)
	taskIDs := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		res = makeRequest(t, http.MethodPost, srv.URL+"/projects/"+proj.ID+"/tasks", map[string]any{
			"title":    "T",
			"priority": "low",
		}, token)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("create task status = %d", res.StatusCode)
		}
		var task struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(res.Body).Decode(&task); err != nil {
			t.Fatalf("decode task: %v", err)
		}
		taskIDs = append(taskIDs, task.ID)
	}

	// Patch statuses: in_progress, done, todo
	patch := func(id, status string) {
		res = makeRequest(t, http.MethodPatch, srv.URL+"/tasks/"+id, map[string]any{
			"status": status,
		}, token)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("patch status=%s statusCode=%d", status, res.StatusCode)
		}
	}
	patch(taskIDs[0], "in_progress")
	patch(taskIDs[1], "done")
	patch(taskIDs[2], "todo")

	// Filter by done
	res = makeRequest(t, http.MethodGet, srv.URL+"/projects/"+proj.ID+"/tasks?status=done", nil, token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d", res.StatusCode)
	}
	var tasks []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&tasks); err != nil {
		t.Fatalf("decode tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0]["status"] != "done" {
		t.Fatalf("expected status done, got %v", tasks[0]["status"])
	}
}

