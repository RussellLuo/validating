package validating

// Field represents a (Name, ValuePtr) pair need to be validated.
type Field struct {
	Name     string
	ValuePtr interface{}
}

// F is a shortcut for creating an instance of Field.
func F(name string, valuePtr interface{}) Field {
	return Field{name, valuePtr}
}

// Validator is an interface for representing a validating's validator.
type Validator interface {
	Validate(field Field) Errors
}

// Validate invokes v.Validate with an empty field.
func Validate(v Validator) (errs Errors) {
	return v.Validate(Field{})
}
