package httpx

import (
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/zagvozdeen/again/internal/valid"
)

type Problem struct {
	Title  string       `json:"title"`
	Status int          `json:"status"`
	Code   string       `json:"code"`
	Errors []FieldError `json:"errors,omitempty"`
}

type FieldError struct {
	Pointer string `json:"pointer"`
	Code    string `json:"code"`
}

type Error struct {
	Status int
	Code   string
}

func (e *Error) Error() string { return e.Code }

var (
	ErrBadJSON      = &Error{http.StatusBadRequest, "bad_json"}
	ErrUnauthorized = &Error{http.StatusUnauthorized, "unauthorized"}
	ErrForbidden    = &Error{http.StatusForbidden, "forbidden"}
	ErrNotFound     = &Error{http.StatusNotFound, "not_found"}
	ErrTooLarge     = &Error{http.StatusRequestEntityTooLarge, "too_large"}
)

func WriteError(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	var p Problem
	if fields, ok := errors.AsType[valid.Errors](err); ok {
		p = Problem{Status: http.StatusUnprocessableEntity, Code: "validation"}
		for field, code := range fields {
			p.Errors = append(p.Errors, FieldError{Pointer: "#/" + field, Code: code})
		}
		slices.SortFunc(p.Errors, func(a, b FieldError) int { return strings.Compare(a.Pointer, b.Pointer) })
	} else if known, ok := errors.AsType[*Error](err); ok {
		p = Problem{Status: known.Status, Code: known.Code}
	} else {
		log.ErrorContext(r.Context(), "Request failed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		p = Problem{Status: http.StatusInternalServerError, Code: "internal"}
	}
	p.Title = http.StatusText(p.Status)

	b, err := json.Marshal(p)
	if err != nil {
		log.ErrorContext(r.Context(), "Failed to encode problem", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_, _ = w.Write(b)
}
