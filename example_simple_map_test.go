package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v3"
)

func Example_simpleMap() {
	ages := map[string]int{
		"foo": 0,
		"bar": 1,
	}
	err := v.Validate(v.Value(ages, v.Map(func(m map[string]int) map[string]v.Schema {
		schemas := make(map[string]v.Schema)
		for name, age := range m {
			schemas[name] = v.Value(age, v.Nonzero[int]())
		}
		return schemas
	})))
	fmt.Printf("%+v\n", err)

	// Output:
	// [foo]: INVALID(is zero valued)
}
