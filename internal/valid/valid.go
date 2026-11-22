package valid

import (
	"cmp"
	"slices"
	"unicode/utf8"
)

type Errors map[string]string

func (e Errors) Error() string {
	return "validation failed"
}

type Rule[T any] func(T) string

type Validator struct {
	errs Errors
}

func (v *Validator) Check[T any](field string, val T, rules ...Rule[T]) T {
	for _, rule := range rules {
		if code := rule(val); code != "" {
			v.Fail(field, code)
			break
		}
	}
	return val
}

func (v *Validator) Fail(field, code string) {
	if v.errs == nil {
		v.errs = Errors{}
	}
	if _, ok := v.errs[field]; !ok {
		v.errs[field] = code
	}
}

func (v *Validator) Err() error {
	if len(v.errs) == 0 {
		return nil
	}
	return v.errs
}

func Required[T comparable](val T) string {
	var zero T
	if val == zero {
		return "required"
	}
	return ""
}

func MinLen(n int) Rule[string] {
	return func(s string) string {
		if utf8.RuneCountInString(s) < n {
			return "too_short"
		}
		return ""
	}
}

func MaxLen(n int) Rule[string] {
	return func(s string) string {
		if utf8.RuneCountInString(s) > n {
			return "too_long"
		}
		return ""
	}
}

func OneOf[T comparable](allowed ...T) Rule[T] {
	return func(val T) string {
		if !slices.Contains(allowed, val) {
			return "invalid"
		}
		return ""
	}
}

func Between[T cmp.Ordered](lo, hi T) Rule[T] {
	return func(val T) string {
		if val < lo || val > hi {
			return "out_of_range"
		}
		return ""
	}
}
