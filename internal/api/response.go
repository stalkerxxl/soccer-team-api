package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/stalkerxxl/soccer-team-api/internal/i18n"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// writeJSON writes a JSON response with the provided HTTP status code.
func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError writes a localized API error response.
func writeError(w http.ResponseWriter, r *http.Request, localizer *i18n.Localizer, statusCode int, code string, messageKey string) {
	writeJSON(w, statusCode, errorResponse{
		Code:    code,
		Message: localizer.Message(r.Header.Get("Accept-Language"), messageKey),
	})
}

// decodeJSON decodes a single JSON object and rejects unknown fields.
func decodeJSON(r *http.Request, target any) (err error) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	if decoder.More() {
		return errors.New("request body must contain a single json object")
	}

	return nil
}
