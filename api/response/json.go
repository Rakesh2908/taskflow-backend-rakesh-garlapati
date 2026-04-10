package response

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string, fields map[string]string) {
	type errBody struct {
		Error  string            `json:"error"`
		Fields map[string]string `json:"fields,omitempty"`
	}

	body := errBody{Error: msg}
	if fields != nil {
		body.Fields = fields
	}
	WriteJSON(w, status, body)
}

