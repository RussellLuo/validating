package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating"
)

func Example_singleField() {
	// Package "validating" imported as "v"

	value := 0
	err := v.Validate(v.Schema{
		v.F("value", &value): v.Range(1, 5),
	})
	fmt.Printf("err: %+v\n", err)
	// Output:
	// err: value: INVALID(is not between given range)
}
