package handler

import (
	"net/http"

	"github.com/Rakesh2908/taskflow/api/response"
	"github.com/Rakesh2908/taskflow/dto"
	"github.com/Rakesh2908/taskflow/service"
	"github.com/Rakesh2908/taskflow/utils"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	svc       service.AuthService
	validator *validator.Validate
}

func NewAuthHandler(svc service.AuthService, v *validator.Validate) *AuthHandler {
	if v == nil {
		v = validator.New()
	}
	return &AuthHandler{svc: svc, validator: v}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
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

	res, err := h.svc.Register(r.Context(), req)
	if err != nil {
		if err == utils.ErrConflict {
			response.WriteError(w, http.StatusConflict, "conflict", nil)
			return
		}
		writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusCreated, res)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
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

	res, err := h.svc.Login(r.Context(), req)
	if err != nil {
		if err == utils.ErrInvalidCredentials {
			response.WriteError(w, http.StatusUnauthorized, "invalid credentials", nil)
			return
		}
		writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, res)
}

