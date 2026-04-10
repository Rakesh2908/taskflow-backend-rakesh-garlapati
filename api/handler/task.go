package handler

import (
	"net/http"

	"github.com/Rakesh2908/taskflow/api/response"
	"github.com/Rakesh2908/taskflow/dto"
	"github.com/Rakesh2908/taskflow/middleware"
	repo "github.com/Rakesh2908/taskflow/repository"
	"github.com/Rakesh2908/taskflow/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type TaskHandler struct {
	tasks     service.TaskService
	projects  service.ProjectService
	validator *validator.Validate
}

func NewTaskHandler(tasks service.TaskService, projects service.ProjectService, v *validator.Validate) *TaskHandler {
	if v == nil {
		v = validator.New()
	}
	return &TaskHandler{tasks: tasks, projects: projects, validator: v}
}

func (h *TaskHandler) ListForProject(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}

	var filters repo.TaskFilters

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = &status
	}

	if assignee := r.URL.Query().Get("assignee"); assignee != "" {
		aid, err := uuid.Parse(assignee)
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, "invalid assignee", nil)
			return
		}
		filters.AssigneeID = &aid
	}

	page, err := parsePage(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	tasks, err := h.tasks.List(r.Context(), projectID, filters, page)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) CreateForProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromCtx(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}

	var req dto.CreateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json", nil)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		if fields, ok := validationFields(req, err); ok {
			response.WriteError(w, http.StatusBadRequest, "validation failed", fields)
			return
		}
		response.WriteError(w, http.StatusBadRequest, "validation failed", nil)
		return
	}

	t, err := h.tasks.Create(r.Context(), userID, projectID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, t)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromCtx(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	taskIDStr := chi.URLParam(r, "id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}

	var req dto.UpdateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json", nil)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		if fields, ok := validationFields(req, err); ok {
			response.WriteError(w, http.StatusBadRequest, "validation failed", fields)
			return
		}
		response.WriteError(w, http.StatusBadRequest, "validation failed", nil)
		return
	}

	t, err := h.tasks.Update(r.Context(), userID, taskID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, t)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromCtx(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	taskIDStr := chi.URLParam(r, "id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}

	if err := h.tasks.Delete(r.Context(), userID, taskID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

