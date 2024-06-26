package validating_test

import (
	"fmt"
	"time"

	v "github.com/RussellLuo/validating/v3"
)

func mapNonzero(field *v.Field) v.Errors {
	value, ok := field.Value.(map[string]time.Time)
	if !ok {
		var want map[string]time.Time
		return v.NewUnsupportedErrors("mapNonzero", field, want)
	}
	if len(value) == 0 {
		return v.NewErrors(field.Name, v.ErrInvalid, "is zero valued")
	}
	return nil
}

type MyValidator struct{}

func (mv MyValidator) Validate(field *v.Field) v.Errors {
	return mapNonzero(field)
}

func Example_customizations() {
	var value map[string]time.Time

	errs := v.Validate(v.Schema{
		v.F("value", value): v.Func(mapNonzero),
	})
	fmt.Printf("errs from the func-validator: %+v\n", errs)

	errs = v.Validate(v.Schema{
		v.F("value", value): MyValidator{},
	})
	fmt.Printf("errs from the struct-validator: %+v\n", errs)

	// Output:
	// errs from the func-validator: value: INVALID(is zero valued)
	// errs from the struct-validator: value: INVALID(is zero valued)
}
