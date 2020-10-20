package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v2"
)

func Example_simpleValue() {
	value := 0
	err := v.Validate(v.Value(&value, v.Range(1, 5)))
	fmt.Printf("%+v\n", err)

	// Output:
	// INVALID(is not between given range)
}
