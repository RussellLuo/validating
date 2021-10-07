package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v2"
)

func Example_simpleValue() {
	value := 0
	// See https://github.com/golang/go/issues/41176.
	err := v.Validate(v.Validator[int](v.Value(value, v.Validator[int](v.Eq(2)))))
	fmt.Printf("%+v\n", err)

	// Output:
	// INVALID(is not between given range)
}
