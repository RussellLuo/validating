package validating

import (
	"fmt"
	"strings"
)

const (
	ErrUnsupported = "UNSUPPORTED"
	ErrInvalid     = "INVALID"
)

type Error interface {
	error
	Field() string
	Kind() string
	Message() string
}

type Errors []Error

func NewErrors(field, kind, message string) Errors {
	return []Error{NewError(field, kind, message)}
}

func NewUnsupportedErrors(field *Field, validatorName string) Errors {
	return NewErrors(field.Name, ErrUnsupported, fmt.Sprintf("cannot use validator `%s` on type %T", validatorName, field.Value))
}

func (e *Errors) Append(errs ...Error) {
	*e = append(*e, errs...)
}

func (e Errors) Error() string {
	strs := make([]string, len(e))
	for i, err := range e {
		strs[i] = err.Error()
	}
	return strings.Join(strs, ", ")
}

// Map converts the given errors to a map[string]Error, where the keys
// of the map are the field names.
func (e Errors) Map() map[string]Error {
	if len(e) == 0 {
		return nil
	}
	m := make(map[string]Error, len(e))
	for _, err := range e {
		m[err.Field()] = err
	}
	return m
}

type errorImpl struct {
	field   string
	kind    string
	message string
}

func NewError(field, kind, message string) Error {
	return errorImpl{field, kind, message}
}

func (e errorImpl) Field() string {
	return e.field
}

func (e errorImpl) Kind() string {
	return e.kind
}

func (e errorImpl) Message() string {
	return e.message
}

func (e errorImpl) Error() string {
	s := fmt.Sprintf("%s(%s)", e.kind, e.message)
	if e.field == "" {
		return s
	}
	return fmt.Sprintf("%s: %s", e.field, s)
}
