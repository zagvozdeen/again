package httpx

import (
	"encoding/json/v2"
	"errors"
	"net/http"

	"github.com/zagvozdeen/again/internal/valid"
)

const MaxBody = 1 << 20

type Handler func(w http.ResponseWriter, r *http.Request) error

func JSON(w http.ResponseWriter, status int, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(b)
	return nil
}

type Input interface {
	Validate(v *valid.Validator)
}

func Decode[T any, PT interface {
	*T
	Input
}](w http.ResponseWriter, r *http.Request) (T, error) {
	var in T

	body := http.MaxBytesReader(w, r.Body, MaxBody)
	defer func() { _ = body.Close() }()
	if err := json.UnmarshalRead(body, &in, json.RejectUnknownMembers(true)); err != nil {
		var zero T
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return zero, ErrTooLarge
		}
		return zero, ErrBadJSON
	}

	var v valid.Validator
	PT(&in).Validate(&v)
	if err := v.Err(); err != nil {
		var zero T
		return zero, err
	}
	return in, nil
}
