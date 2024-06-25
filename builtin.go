package validating

import (
	"fmt"
	"regexp"
	"strconv"
	"unicode/utf8"

	"golang.org/x/exp/constraints"
)

// Func is an adapter to allow the use of ordinary functions as
// validators. If f is a function with the appropriate signature,
// Func(f) is a Validator that calls f.
type Func func(field *Field) Errors

// Validate calls f(field).
func (f Func) Validate(field *Field) Errors {
	return f(field)
}

// Schema is a field mapping, which defines
// the corresponding validator for each field.
type Schema map[*Field]Validator

// Validate validates fields per the given according to the schema.
func (s Schema) Validate(field *Field) (errs Errors) {
	return validateSchema(s, field, func(name string) string { return name })
}

// Value is a shortcut function used to create a schema for a simple value.
func Value(value any, validator Validator) Schema {
	return Schema{
		F("", value): validator,
	}
}

// Nested is a composite validator factory used to create a validator, which will
// delegate the actual validation to the validator returned by f.
func Nested[T any](f func(T) Validator) Validator {
	return Func(func(field *Field) Errors {
		v, ok := field.Value.(T)
		if !ok {
			return NewUnsupportedErrors(field, "Nested")
		}

		return f(v).Validate(field)
	})
}

// Map is a composite validator factory used to create a validator, which will
// do the validation per the schemas associated with a map.
func Map[T map[K]V, K comparable, V any](f func(T) map[K]Validator) Validator {
	return Func(func(field *Field) (errs Errors) {
		v, ok := field.Value.(T)
		if !ok {
			return NewUnsupportedErrors(field, "Map")
		}

		validators := f(v)
		for k, validator := range validators {
			s := toSchema(v[k], validator)
			err := validateSchema(s, field, func(name string) string {
				return name + fmt.Sprintf("[%v]", k)
			})
			if err != nil {
				errs.Append(err...)
			}
		}
		return
	})
}

// Slice is a composite validator factory used to create a validator, which will
// do the validation per the schemas associated with a slice.
func Slice[T ~[]E, E any](f func(T) []Validator) Validator {
	return Func(func(field *Field) (errs Errors) {
		v, ok := field.Value.(T)
		if !ok {
			return NewUnsupportedErrors(field, "Slice")
		}

		validators := f(v)
		for i, validator := range validators {
			s := toSchema(v[i], validator)
			err := validateSchema(s, field, func(name string) string {
				return name + "[" + strconv.Itoa(i) + "]"
			})
			if err != nil {
				errs.Append(err...)
			}
		}
		return
	})
}

// Array is an alias of Slice.
func Array[T ~[]E, E any](f func(T) []Validator) Validator {
	return Slice[T](f)
}

// MessageValidator is a validator that allows users to customize the INVALID
// error message by calling Msg().
type MessageValidator struct {
	Message   string
	Validator Validator
}

// Msg sets the INVALID error message.
func (mv *MessageValidator) Msg(msg string) *MessageValidator {
	if msg != "" {
		mv.Message = msg
	}
	return mv
}

// Validate delegates the actual validation to its inner validator.
func (mv *MessageValidator) Validate(field *Field) Errors {
	return mv.Validator.Validate(field)
}

// All is a composite validator factory used to create a validator, which will
// succeed only when all sub-validators succeed.
func All(validators ...Validator) Validator {
	return Func(func(field *Field) Errors {
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

// AnyValidator is a validator that allows users to change the returned errors
// by calling LastError().
type AnyValidator struct {
	returnLastError bool // Whether to return the last error if all validators fail.
	validators      []Validator
}

// Any is a composite validator factory used to create a validator, which will
// succeed as long as any sub-validator succeeds.
func Any(validators ...Validator) *AnyValidator {
	return &AnyValidator{validators: validators}
}

// LastError makes AnyValidator return the error from the last validator
// if all inner validators fail.
func (av *AnyValidator) LastError() *AnyValidator {
	av.returnLastError = true
	return av
}

// Validate delegates the actual validation to its inner validators.
func (av *AnyValidator) Validate(field *Field) Errors {
	var errs Errors
	var lastErr Errors

	for _, v := range av.validators {
		lastErr = v.Validate(field)
		if lastErr == nil {
			return nil
		}
		errs.Append(lastErr...)
	}

	if av.returnLastError {
		return lastErr
	}
	return errs
}

// Or is an alias of Any.
var Or = Any

// Not is a composite validator factory used to create a validator, which will
// succeed when the given validator fails.
func Not(validator Validator) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "is invalid",
		Validator: Func(func(field *Field) Errors {
			errs := validator.Validate(field)
			if len(errs) == 0 {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}

			var newErrs Errors
			for _, err := range errs {
				// Unsupported errors should be retained.
				if err.Kind() == ErrUnsupported {
					newErrs.Append(err)
				}
			}
			return newErrs
		}),
	}
	return
}

// Is is a leaf validator factory used to create a validator, which will
// succeed when the predicate function f returns true for the field's value.
func Is[T any](f func(T) bool) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "is invalid",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(T)
			if !ok {
				return NewUnsupportedErrors(field, "Is")
			}

			if !f(v) {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// Nonzero is a leaf validator factory used to create a validator, which will
// succeed when the field's value is nonzero.
func Nonzero[T comparable]() (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "is zero valued",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(T)
			if !ok {
				return NewUnsupportedErrors(field, "Nonzero")
			}

			var zero T
			if v == zero {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// Zero is a leaf validator factory used to create a validator, which will
// succeed when the field's value is zero.
func Zero[T comparable]() (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "is nonzero",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(T)
			if !ok {
				return NewUnsupportedErrors(field, "Nonzero")
			}

			var zero T
			if v != zero {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// ZeroOr is a composite validator factory used to create a validator, which will
// succeed if the field's value is zero, or if the given validator succeeds.
//
// ZeroOr will return the error from the given validator if it fails.
func ZeroOr[T comparable](validator Validator) Validator {
	return Any(Zero[T](), validator).LastError()
}

// LenString is a leaf validator factory used to create a validator, which will
// succeed when the length of the string field is between min and max.
func LenString(min, max int) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "has an invalid length",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(string)
			if !ok {
				return NewUnsupportedErrors(field, "LenString")
			}

			l := len(v)
			if l < min || l > max {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// LenSlice is a leaf validator factory used to create a validator, which will
// succeed when the length of the slice field is between min and max.
func LenSlice[T ~[]E, E any](min, max int) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "has an invalid length",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(T)
			if !ok {
				return NewUnsupportedErrors(field, "LenSlice")
			}

			l := len(v)
			if l < min || l > max {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// RuneCount is a leaf validator factory used to create a validator, which will
// succeed when the number of runes in the field's value is between min and max.
func RuneCount(min, max int) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "the number of runes is not between the given range",
		Validator: Func(func(field *Field) Errors {
			valid := false

			switch v := field.Value.(type) {
			case string:
				l := utf8.RuneCountInString(v)
				valid = l >= min && l <= max
			case []byte:
				l := utf8.RuneCount(v)
				valid = l >= min && l <= max
			default:
				return NewUnsupportedErrors(field, "RuneCount")
			}

			if !valid {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// Eq is a leaf validator factory used to create a validator, which will
// succeed when the field's value equals the given value.
func Eq[T comparable](value T) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "does not equal the given value",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(T)
			if !ok {
				return NewUnsupportedErrors(field, "Eq")
			}

			if v != value {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// Ne is a leaf validator factory used to create a validator, which will
// succeed when the field's value does not equal the given value.
func Ne[T comparable](value T) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "equals the given value",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(T)
			if !ok {
				return NewUnsupportedErrors(field, "Ne")
			}

			if v == value {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// Gt is a leaf validator factory used to create a validator, which will
// succeed when the field's value is greater than the given value.
func Gt[T constraints.Ordered](value T) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "is lower than or equal to the given value",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(T)
			if !ok {
				return NewUnsupportedErrors(field, "Gt")
			}

			if v <= value {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// Gte is a leaf validator factory used to create a validator, which will
// succeed when the field's value is greater than or equal to the given value.
func Gte[T constraints.Ordered](value T) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "is lower than the given value",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(T)
			if !ok {
				return NewUnsupportedErrors(field, "Gte")
			}

			if v < value {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// Lt is a leaf validator factory used to create a validator, which will
// succeed when the field's value is lower than the given value.
func Lt[T constraints.Ordered](value T) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "is greater than or equal to the given value",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(T)
			if !ok {
				return NewUnsupportedErrors(field, "Lt")
			}

			if v >= value {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// Lte is a leaf validator factory used to create a validator, which will
// succeed when the field's value is lower than or equal to the given value.
func Lte[T constraints.Ordered](value T) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "is greater than the given value",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(T)
			if !ok {
				return NewUnsupportedErrors(field, "Lte")
			}

			if v > value {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// Range is a leaf validator factory used to create a validator, which will
// succeed when the field's value is between min and max.
func Range[T constraints.Ordered](min, max T) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "is not between the given range",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(T)
			if !ok {
				return NewUnsupportedErrors(field, "Range")
			}

			if v < min || v > max {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// In is a leaf validator factory used to create a validator, which will
// succeed when the field's value is equal to one of the given values.
func In[T comparable](values ...T) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "is not one of the given values",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(T)
			if !ok {
				return NewUnsupportedErrors(field, "In")
			}

			valid := false
			for _, value := range values {
				if v == value {
					valid = true
					break
				}
			}

			if !valid {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// Nin is a leaf validator factory used to create a validator, which will
// succeed when the field's value is not equal to any of the given values.
func Nin[T comparable](values ...T) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "is one of the given values",
		Validator: Func(func(field *Field) Errors {
			v, ok := field.Value.(T)
			if !ok {
				return NewUnsupportedErrors(field, "Nin")
			}

			valid := true
			for _, value := range values {
				if v == value {
					valid = false
					break
				}
			}

			if !valid {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// Match is a leaf validator factory used to create a validator, which will
// succeed when the field's value matches the given regular expression.
func Match(re *regexp.Regexp) (mv *MessageValidator) {
	mv = &MessageValidator{
		Message: "does not match the given regular expression",
		Validator: Func(func(field *Field) Errors {
			valid := false

			switch v := field.Value.(type) {
			case string:
				valid = re.MatchString(v)
			case []byte:
				valid = re.Match(v)
			default:
				return NewUnsupportedErrors(field, "Match")
			}

			if !valid {
				return NewErrors(field.Name, ErrInvalid, mv.Message)
			}
			return nil
		}),
	}
	return
}

// toSchema converts the given validator to a Schema if it's not already.
func toSchema(value any, validator Validator) Schema {
	s, ok := validator.(Schema)
	if !ok {
		s = Value(value, validator)
	}
	return s
}

// validateSchema do the validation per the given schema, which is associated
// with the given field.
func validateSchema(schema Schema, field *Field, prefixFunc func(string) string) (errs Errors) {
	prefix := prefixFunc(field.Name)

	for f, v := range schema {
		if prefix != "" {
			name := prefix
			if f.Name != "" {
				name = name + "." + f.Name
			}
			f = F(name, f.Value)
		}
		if err := v.Validate(f); err != nil {
			errs.Append(err...)
		}
	}
	return
}
