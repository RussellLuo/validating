package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v3"
)

func Example_simplePointer() {
	value := 0
	ptr := &value

	err := v.Validate(v.Value(ptr, v.All(
		v.Nonzero[*int](),
		v.Nested(func(ptr *int) v.Validator { return v.Value(*ptr, v.Range(1, 5)) }),
	)))
	fmt.Printf("%+v\n", err)

	// Output:
	// INVALID(is not between the given range)
}
