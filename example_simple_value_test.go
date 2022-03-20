package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v3"
)

func Example_simpleValue() {
	err := v.Validate(v.Value(0, v.Range(1, 5)))
	fmt.Printf("%+v\n", err)

	// Output:
	// INVALID(is not between the given range)
}
