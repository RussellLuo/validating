package validating

const (
	ErrUnsupported  = "unsupported"
	ErrUnrecognized = "unrecognized"
	ErrInvalid      = "invalid"
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
	return e.field + ": " + e.message
}
