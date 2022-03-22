package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v3"
)

func Example_simpleSlice() {
	names := []string{"", "foo"}
	err := v.Validate(v.Value(names, v.Slice(func(s []string) (schemas []v.Schema) {
		for _, name := range s {
			schemas = append(schemas, v.Value(name, v.Nonzero[string]()))
		}
		return schemas
	})))
	fmt.Printf("%+v\n", err)

	// Output:
	// [0]: INVALID(is zero valued)
}
