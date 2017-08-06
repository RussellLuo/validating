package validating

type Error interface {
	error
	Field() string
	Message() string
}

type Errors []Error

func NewErrors(field, message string) (errs Errors) {
	errs.Append(NewError(field, message))
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
	message string
}

func NewError(field, message string) Error {
	return &errorImpl{field, message}
}

func (e *errorImpl) Field() string {
	return e.field
}

func (e *errorImpl) Message() string {
	return e.message
}

func (e *errorImpl) Error() string {
	return e.field + ": " + e.message
}
