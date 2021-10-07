package validating

// Field represents a (Name, Value) pair that needs to be validated.
type Field[T any] struct {
	Name  string
	Value T
}

// F is a shortcut for creating an instance of Field.
func F[T any](name string, value T) *Field[T] {
	return &Field[T]{name, value}
}

// Validator is an interface for representing a validating's validator.
type Validator[T any] interface {
	Validate(field *Field[T]) Errors
}

// Validate invokes v.Validate with an empty field.
func Validate[T any](v Validator[T]) (errs Errors) {
	return v.Validate(&Field[T]{})
}
