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

// Schema is a field mapping, which defines
// the corresponding validator for each field.
type Schema map[Field]Validator

// Validate validates fields according to the pre-defined schema.
func Validate(schema Schema) (errs Errors) {
	for f, v := range schema {
		if err := v.Validate(f); err != nil {
			errs.Extend(err)
		}
	}
	return
}
