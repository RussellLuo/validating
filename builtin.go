package validating

import (
	"time"
)

// validateFunc represents a prototype of the validator's Validate function.
type validateFunc func(field Field) Errors

// funcValidator is a validator which is made from a function.
type funcValidator struct {
	f validateFunc
}

func (v *funcValidator) Validate(field Field) Errors {
	return v.f(field)
}

// FromFunc creates a leaf validator from a function.
func FromFunc(f validateFunc) Validator {
	return &funcValidator{f}
}

// All creates a composite validator, which will succeed
// only when all sub-validators succeed.
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

// Any creates a composite validator, which will succeed
// as long as any sub-validator succeeds.
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

// Nested creates a composite validator, which will delegate
// the validation work to its inner schema.
func Nested(schema Schema) Validator {
	return FromFunc(func(field Field) Errors {
		nestedSchema := make(Schema, len(schema))
		for f, v := range schema {
			nestedSchema[F(field.Name+"."+f.Name, f.ValuePtr)] = v
		}
		return Validate(nestedSchema)
	})
}

// NestedMulti creates a composite validator, which will delegate
// the validation work to its inner multiple schemas, which are
// returned by calling f.
func NestedMulti(f func() []Schema) Validator {
	schemas := f()
	validators := make([]Validator, len(schemas))
	for i, schema := range schemas {
		validators[i] = Nested(schema)
	}
	return FromFunc(func(field Field) Errors {
		var errs Errors
		for _, v := range validators {
			err := v.Validate(field)
			if err != nil {
				errs.Extend(err)
			}
		}
		return errs
	})
}

func getMsg(validatorName, defaultMsg string, msgs ...string) string {
	switch len(msgs) {
	case 0:
		return defaultMsg
	case 1:
		return msgs[0]
	default:
		panic(validatorName + " only accepts at most one `msg`!")
	}
}

// Assert creates a leaf validator, which will succeed
// only when the boolean expression evaluates to true.
func Assert(b bool, msgs ...string) Validator {
	msg := getMsg("Assert", "is invalid", msgs...)
	return FromFunc(func(field Field) Errors {
		if !b {
			return NewErrors(field.Name, ErrInvalid, msg)
		}
		return nil
	})
}

// Lazy creates a leaf validator, which will call f only as needed,
// to delegate the validation work to the validator returned by f.
func Lazy(f func() Validator) Validator {
	return FromFunc(func(field Field) Errors {
		return f().Validate(field)
	})
}

// Nonzero creates a leaf validator, which will succeed
// when the field's value is nonzero.
func Nonzero(msgs ...string) Validator {
	msg := getMsg("Nonzero", "is zero valued", msgs...)
	return FromFunc(func(field Field) Errors {
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
			valid = *t != false
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
		default:
			return NewErrors(field.Name, ErrUnrecognized, "of unrecognized type")
		}

		if !valid {
			return NewErrors(field.Name, ErrInvalid, msg)
		}
		return nil
	})
}

// Len creates a leaf validator, which will succeed
// when the field's value is between min and max.
func Len(min, max int, msgs ...string) Validator {
	msg := getMsg("Len", "with an invalid length", msgs...)
	return FromFunc(func(field Field) Errors {
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
			*time.Time, **time.Time:
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
		default:
			return NewErrors(field.Name, ErrUnrecognized, "of an unrecognized type")
		}

		if !valid {
			return NewErrors(field.Name, ErrInvalid, msg)
		}
		return nil
	})
}

// Gt creates a leaf validator, which will succeed
// when the field's value is greater than the given value.
func Gt(value interface{}, msgs ...string) Validator {
	msg := getMsg("Gt", "is lower than or equal to given value", msgs...)
	return FromFunc(func(field Field) Errors {
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
			**time.Time, *[]time.Time:
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
			return NewErrors(field.Name, ErrInvalid, msg)
		}
		return nil
	})
}

// Gte creates a leaf validator, which will succeed
// when the field's value is greater than or equal to the given value.
func Gte(value interface{}, msgs ...string) Validator {
	msg := getMsg("Gte", "is lower than give value", msgs...)
	return FromFunc(func(field Field) Errors {
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
			**time.Time, *[]time.Time:
			return NewErrors(field.Name, ErrUnsupported, "cannot use validator `Gte`")
		case *uint8:
			valid = *t >= value.(uint8)
		case *uint16:
			valid = *t >= value.(uint16)
		case *uint32:
			valid = *t >= value.(uint32)
		case *uint64:
			valid = *t >= value.(uint64)
		case *int8:
			valid = *t >= value.(int8)
		case *int16:
			valid = *t >= value.(int16)
		case *int32:
			valid = *t >= value.(int32)
		case *int64:
			valid = *t >= value.(int64)
		case *float32:
			valid = *t >= value.(float32)
		case *float64:
			valid = *t >= value.(float64)
		case *uint:
			valid = *t >= value.(uint)
		case *int:
			valid = *t >= value.(int)
		case *string:
			valid = *t >= value.(string)
		case *time.Time:
			valid = !(*t).Before(value.(time.Time))
		case *time.Duration:
			valid = *t >= value.(time.Duration)
		default:
			return NewErrors(field.Name, ErrUnrecognized, "of an unrecognized type")
		}

		if !valid {
			return NewErrors(field.Name, ErrInvalid, msg)
		}
		return nil
	})
}

// Lt creates a leaf validator, which will succeed
// when the field's value is lower than the given value.
func Lt(value interface{}, msgs ...string) Validator {
	msg := getMsg("Lt", "is greater than or equal to give value", msgs...)
	validator := Gte(value, msg)
	return FromFunc(func(field Field) Errors {
		errs := validator.Validate(field)
		if errs == nil {
			return NewErrors(field.Name, ErrInvalid, msg)
		}
		switch errs[0].Kind() {
		case ErrUnsupported:
			return errs
		case ErrUnrecognized:
			return NewErrors(field.Name, ErrUnrecognized, "cannot use validator `Lt`")
		default:
			return nil
		}
	})
}

// Lte creates a leaf validator, which will succeed
// when the field's value is lower than or equal to the given value.
func Lte(value interface{}, msgs ...string) Validator {
	msg := getMsg("Lte", "is greater than give value", msgs...)
	validator := Gt(value, msg)
	return FromFunc(func(field Field) Errors {
		errs := validator.Validate(field)
		if errs == nil {
			return NewErrors(field.Name, ErrInvalid, msg)
		}
		switch errs[0].Kind() {
		case ErrUnsupported:
			return errs
		case ErrUnrecognized:
			return NewErrors(field.Name, ErrUnrecognized, "cannot use validator `Lte`")
		default:
			return nil
		}
	})
}

// Range is a shortcut of `All(Gte(min), Lte(max))`.
func Range(min, max interface{}, msgs ...string) Validator {
	msg := getMsg("Range", "is not between given range", msgs...)
	return All(Gte(min, msg), Lte(max, msg))
}

// In creates a leaf validator, which will succeed
// when the field's value is equal to one of the given values.
func In(values ...interface{}) Validator {
	msg := "is not one of given values"
	return FromFunc(func(field Field) Errors {
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
			**time.Time, *[]time.Time:
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
			return NewErrors(field.Name, ErrInvalid, msg)
		}
		return nil
	})
}
