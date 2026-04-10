package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"reflect"
	"strings"

	"github.com/Rakesh2908/taskflow/api/response"
	repo "github.com/Rakesh2908/taskflow/repository"
	"github.com/Rakesh2908/taskflow/utils"
	"github.com/go-playground/validator/v10"
)

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func validationFields(req any, err error) (map[string]string, bool) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil, false
	}

	t := reflect.TypeOf(req)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	fields := make(map[string]string, len(ve))
	for _, fe := range ve {
		fields[jsonFieldName(t, fe.StructField(), fe.Field())] = validationMessage(fe.Tag())
	}
	return fields, true
}

func validationMessage(tag string) string {
	switch tag {
	case "required":
		return "is required"
	default:
		return tag
	}
}

func jsonFieldName(t reflect.Type, structFieldName, fallback string) string {
	if t.Kind() != reflect.Struct {
		return fallback
	}
	sf, ok := t.FieldByName(structFieldName)
	if !ok {
		return fallback
	}
	tag := sf.Tag.Get("json")
	if tag == "" {
		return fallback
	}
	name := strings.Split(tag, ",")[0]
	if name == "" || name == "-" {
		return fallback
	}
	return name
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch err {
	case utils.ErrNotFound:
		response.WriteError(w, http.StatusNotFound, "not found", nil)
	case utils.ErrForbidden:
		response.WriteError(w, http.StatusForbidden, "forbidden", nil)
	case utils.ErrConflict:
		response.WriteError(w, http.StatusConflict, "conflict", nil)
	default:
		response.WriteError(w, http.StatusInternalServerError, "internal server error", nil)
	}
}

func parsePage(r *http.Request) (*repo.Page, error) {
	q := r.URL.Query()
	pageStr := q.Get("page")
	limitStr := q.Get("limit")
	if pageStr == "" && limitStr == "" {
		return nil, nil
	}

	page := 1
	limit := 20
	var err error

	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return nil, errors.New("invalid page")
		}
	}
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			return nil, errors.New("invalid limit")
		}
	}
	if limit > 100 {
		limit = 100
	}

	return &repo.Page{
		Limit:  limit,
		Offset: (page - 1) * limit,
	}, nil
}

