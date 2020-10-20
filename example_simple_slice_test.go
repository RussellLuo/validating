package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v2"
)

func Example_simpleSlice() {
	names := []string{"", "foo"}
	err := v.Validate(v.Slice(func() (schemas []v.Schema) {
		for _, name := range names {
			name := name
			schemas = append(schemas, v.Schema{
				v.F("", &name): v.Nonzero(),
			})
		}
		return schemas
	}))
	fmt.Printf("%+v\n", err)

	// Output:
	// [0]: INVALID(is zero valued)
}
