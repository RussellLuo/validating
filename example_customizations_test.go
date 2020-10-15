package validating_test

import (
	"fmt"
	"time"

	v "github.com/RussellLuo/validating/v2"
)

func validate(field v.Field) v.Errors {
	switch t := field.ValuePtr.(type) {
	case *map[string]time.Time:
		if *t == nil {
			return v.NewErrors(field.Name, v.ErrInvalid, "is empty")
		}
		return nil
	default:
		return v.NewErrors(field.Name, v.ErrUnsupported, "is unsupported")
	}
}

type MyValidator struct{}

func (mv *MyValidator) Validate(field v.Field) v.Errors {
	return validate(field)
}

func Example_customizations() {
	var value map[string]time.Time

	// do validation by funcValidator
	funcValidator := v.FromFunc(func(field v.Field) v.Errors {
		return validate(field)
	})
	errs := v.Validate(v.Schema{
		v.F("value", &value): funcValidator,
	})
	fmt.Printf("errs from funcValidator: %+v\n", errs)

	// do validation by structValidator
	structValidator := &MyValidator{}
	errs = v.Validate(v.Schema{
		v.F("value", &value): structValidator,
	})
	fmt.Printf("errs from structValidator: %+v\n", errs)
}
