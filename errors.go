package validating

import (
	"fmt"
	"strings"
)

const (
	ErrUnsupported  = "UNSUPPORTED"
	ErrUnrecognized = "UNRECOGNIZED"
	ErrInvalid      = "INVALID"
)

type Error interface {
	error
	Field() string
	Kind() string
	Message() string
}

type Errors []Error

func NewErrors(field, kind, message string) (errs Errors) {
	errs.Append(NewError(field, kind, message))
	return errs
}

func (e *Errors) Append(err Error) {
	*e = append(*e, err)
}

func (e *Errors) Extend(errs Errors) {
	*e = append(*e, errs...)
}

// HasField will return if errors contains field
func (e Errors) HasField(name string) bool {
	for _, err := range e {
		if err.Field() == name {
			return true
		}
	}
	return false
}

// JSONErrors transform errors to json usable format
func (e Errors) JSONErrors() *map[string][]string {
	jsons := make(map[string][]string)
	for _, err := range e {
		jsons[err.Field()] = append(jsons[err.Field()], err.Message())
	}
	return &jsons
}

func (e Errors) Error() string {
	strs := make([]string, len(e))
	for i, err := range e {
		strs[i] = err.Error()
	}
	return strings.Join(strs, ", ")
}

type errorImpl struct {
	field   string
	kind    string
	message string
}

func NewError(field, kind, message string) Error {
	return &errorImpl{field, kind, message}
}

func (e *errorImpl) Field() string {
	return e.field
}

func (e *errorImpl) Kind() string {
	return e.kind
}

func (e *errorImpl) Message() string {
	return e.message
}

func (e *errorImpl) Error() string {
	return fmt.Sprintf("%s: %s(%s)", e.field, e.kind, e.message)
}
