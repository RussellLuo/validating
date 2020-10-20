package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v2"
)

func Example_singleVariable() {
	value := 0
	err := v.Validate(v.Var(&value, v.Range(1, 5)))
	fmt.Printf("%+v\n", err)

	// Output:
	// INVALID(is not between given range)
}
