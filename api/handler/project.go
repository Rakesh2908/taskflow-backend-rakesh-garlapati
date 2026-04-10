package handler

import (
	"net/http"

	"github.com/Rakesh2908/taskflow/api/response"
	"github.com/Rakesh2908/taskflow/dto"
	"github.com/Rakesh2908/taskflow/middleware"
	"github.com/Rakesh2908/taskflow/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	svc       service.ProjectService
	validator *validator.Validate
}

func NewProjectHandler(svc service.ProjectService, v *validator.Validate) *ProjectHandler {
	if v == nil {
		v = validator.New()
	}
	return &ProjectHandler{svc: svc, validator: v}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromCtx(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	page, err := parsePage(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	projects, err := h.svc.List(r.Context(), userID, page)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromCtx(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req dto.CreateProjectRequest
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

	p, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, p)
}

func (h *ProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}

	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, p)
}

func (h *ProjectHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromCtx(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}

	stats, err := h.svc.Stats(r.Context(), userID, projectID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, stats)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromCtx(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}

	var req dto.UpdateProjectRequest
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

	p, err := h.svc.Update(r.Context(), userID, projectID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, p)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromCtx(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}

	if err := h.svc.Delete(r.Context(), userID, projectID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

