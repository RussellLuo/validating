package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v3"
)

func Example_simpleSlice() {
	names := []string{"", "foo"}
	err := v.Validate(v.Value(names, v.Slice(func(s []string) (schemas []v.Validator) {
		for range s {
			schemas = append(schemas, v.Nonzero[string]())
		}
		return schemas
	})))
	fmt.Printf("%+v\n", err)

	// Output:
	// [0]: INVALID(is zero valued)
}
