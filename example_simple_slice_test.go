package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v3"
)

func Example_simpleSlice() {
	names := []string{"", "foo"}
	err := v.Validate(v.Slice(func() (schemas []v.Schema) {
		for _, name := range names {
			schemas = append(schemas, v.Value(name, v.Nonzero[string]()))
		}
		return schemas
	}))
	fmt.Printf("%+v\n", err)

	// Output:
	// [0]: INVALID(is zero valued)
}

func Example_simpleSliceEach() {
	names := []string{"", "foo"}
	err := v.Validate(v.Value(names, v.Each[[]string](v.Nonzero[string]())))
	fmt.Printf("%+v\n", err)

	// Output:
	// [0]: INVALID(is zero valued)
}
