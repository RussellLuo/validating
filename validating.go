package validating

// Field represents a (Name, Value) pair that needs to be validated.
type Field struct {
	Name  string
	Value interface{}
}

// F is a shortcut for creating a pointer to Field.
func F(name string, value interface{}) *Field {
	return &Field{Name: name, Value: value}
}

// Validator is an interface for representing a validating's validator.
type Validator interface {
	Validate(field *Field) Errors
}

// Validate invokes v.Validate with an empty field.
func Validate(v Validator) (errs Errors) {
	return v.Validate(&Field{})
}
