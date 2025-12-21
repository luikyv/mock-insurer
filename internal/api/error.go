package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/luikyv/mock-insurer/internal/timeutil"
)

type Error struct {
	code        string
	statusCode  int
	description string
}

func (err Error) Error() string {
	return fmt.Sprintf("%s %s", err.code, err.description)
}

func NewError(code string, status int, description string) Error {
	err := Error{
		code:        code,
		statusCode:  status,
		description: description,
	}

	return err
}

// WriteError writes an API error response to the provided http.ResponseWriter.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr Error
	if !errors.As(err, &apiErr) {
		slog.ErrorContext(r.Context(), "unknown error", "error", err)
		WriteError(w, r, Error{"INTERNAL_ERROR", http.StatusInternalServerError, "internal error"})
		return
	}

	slog.InfoContext(r.Context(), "returning error", "error", err, "status_code", apiErr.statusCode)
	description := apiErr.description
	if len(description) > 2048 {
		description = description[:2048]
	}

	if apiErr.statusCode == http.StatusUnprocessableEntity {
		WriteJSON(w, map[string]any{
			"errors": map[string]any{
				"code":   apiErr.code,
				"title":  apiErr.code,
				"detail": description,
			},
		}, apiErr.statusCode)
		return
	}

	WriteJSON(w, map[string]any{
		"errors": []map[string]any{{
			"code":            apiErr.code,
			"title":           apiErr.code,
			"detail":          description,
			"requestDateTime": timeutil.DateTimeNow(),
		}},
	}, apiErr.statusCode)
}
