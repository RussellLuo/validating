package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v3"
)

func Example_simpleSlice() {
	names := []string{"", "foo"}
	err := v.Validate(v.Value(names, v.EachSlice[[]string](v.Nonzero[string]())))
	fmt.Printf("%+v\n", err)

	// Output:
	// [0]: INVALID(is zero valued)
}
