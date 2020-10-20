package validating

import (
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"
)

// FromFunc is an adapter to allow the use of ordinary functions as
// validators. If f is a function with the appropriate signature,
// FromFunc(f) is a Validator that calls f.
type FromFunc func(field Field) Errors

// Validate calls f(field).
func (f FromFunc) Validate(field Field) Errors {
	return f(field)
}

// validateSchema do the validation per the given schema, which is associated
// with the given field.
func validateSchema(schema Schema, field Field, prefixFunc func(string) string) (errs Errors) {
	prefix := prefixFunc(field.Name)

	for f, v := range schema {
		if prefix != "" {
			name := prefix
			if f.Name != "" {
				name = name + "." + f.Name
			}
			f = F(name, f.ValuePtr)
		}
		if err := v.Validate(f); err != nil {
			errs.Extend(err)
		}
	}
	return
}

// Schema is a field mapping, which defines
// the corresponding validator for each field.
type Schema map[Field]Validator

// Validate validates fields per the given according to the schema.
func (s Schema) Validate(field Field) (errs Errors) {
	return validateSchema(s, field, func(name string) string {
		return name
	})
}

// Var is a shortcut for creating a schema for a variable.
func Var(valuePtr interface{}, validator Validator) Schema {
	return Schema{
		F("", valuePtr): validator,
	}
}

// Map is a composite validator factory to create a validator, which will
// do the validation per the schemas associated with a map.
func Map(f func() map[string]Schema) Validator {
	schemas := f()
	return FromFunc(func(field Field) (errs Errors) {
		for k, s := range schemas {
			err := validateSchema(s, field, func(name string) string {
				return name + "[" + k + "]"
			})
			if err != nil {
				errs.Extend(err)
			}
		}
		return
	})
}

// Slice is a composite validator factory to create a validator, which will
// do the validation per the schemas associated with a slice.
func Slice(f func() []Schema) Validator {
	schemas := f()
	return FromFunc(func(field Field) (errs Errors) {
		for i, s := range schemas {
			err := validateSchema(s, field, func(name string) string {
				return name + "[" + strconv.Itoa(i) + "]"
			})
			if err != nil {
				errs.Extend(err)
			}
		}
		return
	})
}

// Array is an alias of Slice.
var Array = Slice

// MessageValidator is a validator that allows users to customize the INVALID
// error message by calling Msg().
type MessageValidator struct {
	message   string
	validator Validator
}

// Msg sets the INVALID error message.
func (mv *MessageValidator) Msg(msg string) *MessageValidator {
	if msg != "" {
		mv.message = msg
	}
	return mv
}

// Validate delegates the actual validation to its inner validator.
func (mv *MessageValidator) Validate(field Field) Errors {
	return mv.validator.Validate(field)
}

// All is a composite validator factory to create a validator, which will
// succeed only when all sub-validators succeed.
func All(validators ...Validator) Validator {
	return FromFunc(func(field Field) Errors {
		for _, v := range validators {
			if errs := v.Validate(field); errs != nil {
				return errs
			}
		}
		return nil
	})
}

// And is an alias of All.
var And = All

// Any is a composite validator factory to create a validator, which will
// succeed as long as any sub-validator succeeds.
func Any(validators ...Validator) Validator {
	return FromFunc(func(field Field) Errors {
		var errs Errors
		for _, v := range validators {
			err := v.Validate(field)
			if err == nil {
				return nil
			}
			errs.Extend(err)
		}
		return errs
	})
}

// Or is an alias of Any.
var Or = Any

// not is a helper function to negate the given validator.
func not(validatorName string, validator Validator, field Field, msg string) Errors {
	errs := validator.Validate(field)
	if len(errs) == 0 {
		return NewErrors(field.Name, ErrInvalid, msg)
	}
	switch errs[0].Kind() {
	case ErrUnsupported:
		return NewErrors(field.Name, ErrUnsupported, "cannot use validator `"+validatorName+"`")
	case ErrUnrecognized:
		return NewErrors(field.Name, ErrUnrecognized, "of an unrecognized type")
	default:
		return nil
	}
}

// merge merges multiple errors, which occur from the composite validator, into one error.
func merge(validatorName string, validator Validator, field Field, msg string) Errors {
	errs := validator.Validate(field)
	if len(errs) == 0 {
		return nil
	}
	switch errs[0].Kind() {
	case ErrUnsupported:
		return NewErrors(field.Name, ErrUnsupported, "cannot use validator `"+validatorName+"`")
	case ErrUnrecognized:
		return NewErrors(field.Name, ErrUnrecognized, "of an unrecognized type")
	default:
		return NewErrors(field.Name, ErrInvalid, msg)
	}
}

// Not is a composite validator factory to create a validator, which will
// succeed when the given validator fails.
func Not(validator Validator) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "is invalid",
		validator: FromFunc(func(field Field) Errors {
			errs := validator.Validate(field)
			if len(errs) == 0 {
				return NewErrors(field.Name, ErrInvalid, mv.message)
			}
			for _, err := range errs {
				switch err.Kind() {
				case ErrUnsupported, ErrUnrecognized:
					return []Error{err}
				}
			}
			return nil
		}),
	}
	return
}

// Nested is a composite validator factory to create a validator, which will
// delegate the actual validation to its inner schema.
//
// This composite validator is DEPRECATED.
func Nested(schema Schema) Validator {
	return FromFunc(func(field Field) Errors {
		nestedSchema := make(Schema, len(schema))
		for f, v := range schema {
			nestedSchema[F(field.Name+f.Name, f.ValuePtr)] = v
		}
		return Validate(nestedSchema)
	})
}

// NestedMulti is a composite validator factory to create a validator, which will
// delegate the actual validation to its inner multiple schemas, which are
// returned by calling f.
//
// This composite validator is DEPRECATED.
func NestedMulti(f func() []Schema) Validator {
	schemas := f()
	validators := make([]Validator, len(schemas))
	for i, schema := range schemas {
		validators[i] = Nested(schema)
	}
	return FromFunc(func(field Field) (errs Errors) {
		for _, v := range validators {
			if err := v.Validate(field); err != nil {
				errs.Extend(err)
			}
		}
		return
	})
}

// Lazy is a composite validator factory to create a validator, which will
// call f only as needed, to delegate the actual validation to
// the validator returned by f.
func Lazy(f func() Validator) Validator {
	return FromFunc(func(field Field) Errors {
		return f().Validate(field)
	})
}

// Assert is a leaf validator factory to create a validator, which will
// succeed only when the boolean expression evaluates to true.
func Assert(b bool) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "is invalid",
		validator: FromFunc(func(field Field) Errors {
			if !b {
				return NewErrors(field.Name, ErrInvalid, mv.message)
			}
			return nil
		}),
	}
	return
}

// Nonzero is a leaf validator factory to create a validator, which will
// succeed when the field's value is nonzero.
func Nonzero() (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "is zero valued",
		validator: FromFunc(func(field Field) Errors {
			valid := false

			switch t := field.ValuePtr.(type) {
			case *uint8:
				valid = *t != 0
			case **uint8:
				valid = *t != nil
			case *[]uint8:
				valid = len(*t) != 0
			case *uint16:
				valid = *t != 0
			case **uint16:
				valid = *t != nil
			case *[]uint16:
				valid = len(*t) != 0
			case *uint32:
				valid = *t != 0
			case **uint32:
				valid = *t != nil
			case *[]uint32:
				valid = len(*t) != 0
			case *uint64:
				valid = *t != 0
			case **uint64:
				valid = *t != nil
			case *[]uint64:
				valid = len(*t) != 0
			case *int8:
				valid = *t != 0
			case **int8:
				valid = *t != nil
			case *[]int8:
				valid = len(*t) != 0
			case *int16:
				valid = *t != 0
			case **int16:
				valid = *t != nil
			case *[]int16:
				valid = len(*t) != 0
			case *int32:
				valid = *t != 0
			case **int32:
				valid = *t != nil
			case *[]int32:
				valid = len(*t) != 0
			case *int64:
				valid = *t != 0
			case **int64:
				valid = *t != nil
			case *[]int64:
				valid = len(*t) != 0
			case *float32:
				valid = *t != 0
			case **float32:
				valid = *t != nil
			case *[]float32:
				valid = len(*t) != 0
			case *float64:
				valid = *t != 0
			case **float64:
				valid = *t != nil
			case *[]float64:
				valid = len(*t) != 0
			case *uint:
				valid = *t != 0
			case **uint:
				valid = *t != nil
			case *[]uint:
				valid = len(*t) != 0
			case *int:
				valid = *t != 0
			case **int:
				valid = *t != nil
			case *[]int:
				valid = len(*t) != 0
			case *bool:
				valid = *t
			case **bool:
				valid = *t != nil
			case *[]bool:
				valid = len(*t) != 0
			case *string:
				valid = *t != ""
			case **string:
				valid = *t != nil
			case *[]string:
				valid = len(*t) != 0
			case *time.Time:
				valid = !t.IsZero()
			case **time.Time:
				valid = *t != nil
			case *[]time.Time:
				valid = len(*t) != 0
			case *time.Duration:
				valid = *t != 0
			case **time.Duration:
				valid = *t != nil
			case *[]time.Duration:
				valid = len(*t) != 0
			default:
				return NewErrors(field.Name, ErrUnrecognized, "of an unrecognized type")
			}

			if !valid {
				return NewErrors(field.Name, ErrInvalid, mv.message)
			}
			return nil
		}),
	}
	return
}

// Len is a leaf validator factory to create a validator, which will
// succeed when the field's length is between min and max.
func Len(min, max int) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "with an invalid length",
		validator: FromFunc(func(field Field) Errors {
			valid := false

			switch t := field.ValuePtr.(type) {
			case *uint8, **uint8, *uint16, **uint16,
				*uint32, **uint32, *uint64, **uint64,
				*int8, **int8, *int16, **int16,
				*int32, **int32, *int64, **int64,
				*float32, **float32, *float64, **float64,
				*uint, **uint, *int, **int,
				*bool, **bool,
				**string,
				*time.Time, **time.Time,
				**time.Duration:
				return NewErrors(field.Name, ErrUnsupported, "cannot use validator `Len`")
			case *[]uint8:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]uint16:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]uint32:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]uint64:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]int8:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]int16:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]int32:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]int64:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]float32:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]float64:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]uint:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]int:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]bool:
				l := len(*t)
				valid = l >= min && l <= max
			case *string:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]string:
				l := len(*t)
				valid = l >= min && l <= max
			case *[]time.Time:
				l := len(*t)
				valid = l >= min && l <= max
			case *time.Duration:
				valid = *t >= time.Duration(min) && *t <= time.Duration(max)
			case *[]time.Duration:
				l := len(*t)
				valid = l >= min && l <= max
			default:
				return NewErrors(field.Name, ErrUnrecognized, "of an unrecognized type")
			}

			if !valid {
				return NewErrors(field.Name, ErrInvalid, mv.message)
			}
			return nil
		}),
	}
	return
}

// RuneCount is a leaf validator factory to create a validator, which will
// succeed when the number of runes in the field's value is between min and max.
func RuneCount(min, max int) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "the number of runes is not between the given range",
		validator: FromFunc(func(field Field) Errors {
			valid := false

			switch t := field.ValuePtr.(type) {
			case *string:
				l := utf8.RuneCountInString(*t)
				valid = l >= min && l <= max
			case *[]byte:
				l := utf8.RuneCount(*t)
				valid = l >= min && l <= max
			default:
				return NewErrors(field.Name, ErrUnsupported, "cannot use validator `RuneCount`")
			}

			if !valid {
				return NewErrors(field.Name, ErrInvalid, mv.message)
			}
			return nil
		}),
	}
	return
}

// Eq is a leaf validator factory to create a validator, which will
// succeed when the field's value equals the given value.
func Eq(value interface{}) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "does not equal the given value",
		validator: FromFunc(func(field Field) Errors {
			valid := false

			switch t := field.ValuePtr.(type) {
			case **uint8, *[]uint8, **uint16, *[]uint16,
				**uint32, *[]uint32, **uint64, *[]uint64,
				**int8, *[]int8, **int16, *[]int16,
				**int32, *[]int32, **int64, *[]int64,
				**float32, *[]float32, **float64, *[]float64,
				**uint, *[]uint, **int, *[]int,
				*bool, **bool, *[]bool,
				**string, *[]string,
				**time.Time, *[]time.Time,
				**time.Duration, *[]time.Duration:
				return NewErrors(field.Name, ErrUnsupported, "cannot use validator `Eq`")
			case *uint8:
				valid = *t == value.(uint8)
			case *uint16:
				valid = *t == value.(uint16)
			case *uint32:
				valid = *t == value.(uint32)
			case *uint64:
				valid = *t == value.(uint64)
			case *int8:
				valid = *t == value.(int8)
			case *int16:
				valid = *t == value.(int16)
			case *int32:
				valid = *t == value.(int32)
			case *int64:
				valid = *t == value.(int64)
			case *float32:
				valid = *t == value.(float32)
			case *float64:
				valid = *t == value.(float64)
			case *uint:
				valid = *t == value.(uint)
			case *int:
				valid = *t == value.(int)
			case *string:
				valid = *t == value.(string)
			case *time.Time:
				valid = (*t).Equal(value.(time.Time))
			case *time.Duration:
				valid = *t == value.(time.Duration)
			default:
				return NewErrors(field.Name, ErrUnrecognized, "of an unrecognized type")
			}

			if !valid {
				return NewErrors(field.Name, ErrInvalid, mv.message)
			}
			return nil
		}),
	}
	return
}

// Ne is a leaf validator factory to create a validator, which will
// succeed when the field's value does not equal the given value.
func Ne(value interface{}) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "equals the given value",
		validator: FromFunc(func(field Field) Errors {
			return not("Ne", Eq(value), field, mv.message)
		}),
	}
	return
}

// Gt is a leaf validator factory to create a validator, which will
// succeed when the field's value is greater than the given value.
func Gt(value interface{}) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "is lower than or equal to given value",
		validator: FromFunc(func(field Field) Errors {
			valid := false

			switch t := field.ValuePtr.(type) {
			case **uint8, *[]uint8, **uint16, *[]uint16,
				**uint32, *[]uint32, **uint64, *[]uint64,
				**int8, *[]int8, **int16, *[]int16,
				**int32, *[]int32, **int64, *[]int64,
				**float32, *[]float32, **float64, *[]float64,
				**uint, *[]uint, **int, *[]int,
				*bool, **bool, *[]bool,
				**string, *[]string,
				**time.Time, *[]time.Time,
				**time.Duration, *[]time.Duration:
				return NewErrors(field.Name, ErrUnsupported, "cannot use validator `Gt`")
			case *uint8:
				valid = *t > value.(uint8)
			case *uint16:
				valid = *t > value.(uint16)
			case *uint32:
				valid = *t > value.(uint32)
			case *uint64:
				valid = *t > value.(uint64)
			case *int8:
				valid = *t > value.(int8)
			case *int16:
				valid = *t > value.(int16)
			case *int32:
				valid = *t > value.(int32)
			case *int64:
				valid = *t > value.(int64)
			case *float32:
				valid = *t > value.(float32)
			case *float64:
				valid = *t > value.(float64)
			case *uint:
				valid = *t > value.(uint)
			case *int:
				valid = *t > value.(int)
			case *string:
				valid = *t > value.(string)
			case *time.Time:
				valid = (*t).After(value.(time.Time))
			case *time.Duration:
				valid = *t > value.(time.Duration)
			default:
				return NewErrors(field.Name, ErrUnrecognized, "of an unrecognized type")
			}

			if !valid {
				return NewErrors(field.Name, ErrInvalid, mv.message)
			}
			return nil
		}),
	}
	return
}

// Gte is a leaf validator factory to create a validator, which will
// succeed when the field's value is greater than or equal to the given value.
func Gte(value interface{}) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "is lower than given value",
		validator: FromFunc(func(field Field) Errors {
			return merge("Gte", Any(Gt(value), Eq(value)), field, mv.message)
		}),
	}
	return
}

// Lt is a leaf validator factory to create a validator, which will
// succeed when the field's value is lower than the given value.
func Lt(value interface{}) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "is greater than or equal to given value",
		validator: FromFunc(func(field Field) Errors {
			return not("Lt", Gte(value), field, mv.message)
		}),
	}
	return
}

// Lte is a leaf validator factory to create a validator, which will
// succeed when the field's value is lower than or equal to the given value.
func Lte(value interface{}) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "is greater than given value",
		validator: FromFunc(func(field Field) Errors {
			return not("Lte", Gt(value), field, mv.message)
		}),
	}
	return
}

// Range is a shortcut of `All(Gte(min), Lte(max))`.
func Range(min, max interface{}) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "is not between given range",
		validator: FromFunc(func(field Field) Errors {
			return merge("Range", All(Gte(min), Lte(max)), field, mv.message)
		}),
	}
	return
}

// In is a leaf validator factory to create a validator, which will
// succeed when the field's value is equal to one of the given values.
func In(values ...interface{}) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "is not one of given values",
		validator: FromFunc(func(field Field) Errors {
			valid := false

			switch t := field.ValuePtr.(type) {
			case **uint8, *[]uint8, **uint16, *[]uint16,
				**uint32, *[]uint32, **uint64, *[]uint64,
				**int8, *[]int8, **int16, *[]int16,
				**int32, *[]int32, **int64, *[]int64,
				**float32, *[]float32, **float64, *[]float64,
				**uint, *[]uint, **int, *[]int,
				**bool, *[]bool,
				**string, *[]string,
				**time.Time, *[]time.Time,
				**time.Duration, *[]time.Duration:
				return NewErrors(field.Name, ErrUnsupported, "cannot use validator `In`")
			case *uint8:
				for _, value := range values {
					if *t == value.(uint8) {
						valid = true
						break
					}
				}
			case *uint16:
				for _, value := range values {
					if *t == value.(uint16) {
						valid = true
						break
					}
				}
			case *uint32:
				for _, value := range values {
					if *t == value.(uint32) {
						valid = true
						break
					}
				}
			case *uint64:
				for _, value := range values {
					if *t == value.(uint64) {
						valid = true
						break
					}
				}
			case *int8:
				for _, value := range values {
					if *t == value.(int8) {
						valid = true
						break
					}
				}
			case *int16:
				for _, value := range values {
					if *t == value.(int16) {
						valid = true
						break
					}
				}
			case *int32:
				for _, value := range values {
					if *t == value.(int32) {
						valid = true
						break
					}
				}
			case *int64:
				for _, value := range values {
					if *t == value.(int64) {
						valid = true
						break
					}
				}
			case *float32:
				for _, value := range values {
					if *t == value.(float32) {
						valid = true
						break
					}
				}
			case *float64:
				for _, value := range values {
					if *t == value.(float64) {
						valid = true
						break
					}
				}
			case *uint:
				for _, value := range values {
					if *t == value.(uint) {
						valid = true
						break
					}
				}
			case *int:
				for _, value := range values {
					if *t == value.(int) {
						valid = true
						break
					}
				}
			case *bool:
				for _, value := range values {
					if *t == value.(bool) {
						valid = true
						break
					}
				}
			case *string:
				for _, value := range values {
					if *t == value.(string) {
						valid = true
						break
					}
				}
			case *time.Time:
				for _, value := range values {
					if (*t).Equal(value.(time.Time)) {
						valid = true
						break
					}
				}
			case *time.Duration:
				for _, value := range values {
					if *t == value.(time.Duration) {
						valid = true
						break
					}
				}
			default:
				return NewErrors(field.Name, ErrUnrecognized, "of an unrecognized type")
			}

			if !valid {
				return NewErrors(field.Name, ErrInvalid, mv.message)
			}
			return nil
		}),
	}
	return
}

// Nin is a leaf validator factory to create a validator, which will
// succeed when the field's value is not equal to any of the given values.
func Nin(values ...interface{}) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "is one of given values",
		validator: FromFunc(func(field Field) Errors {
			return not("Nin", In(values...), field, mv.message)
		}),
	}
	return
}

// RegexpMatch is a leaf validator factory to create a validator, which will
// succeed when the field's value matches the given regular expression.
func RegexpMatch(re *regexp.Regexp) (mv *MessageValidator) {
	mv = &MessageValidator{
		message: "does not match the given regular expression",
		validator: FromFunc(func(field Field) Errors {
			valid := false

			switch t := field.ValuePtr.(type) {
			case *string:
				valid = re.MatchString(*t)
			case *[]byte:
				valid = re.Match(*t)
			default:
				return NewErrors(field.Name, ErrUnsupported, "cannot use validator `RegexpMatch`")
			}

			if !valid {
				return NewErrors(field.Name, ErrInvalid, mv.message)
			}
			return nil
		}),
	}
	return
}
