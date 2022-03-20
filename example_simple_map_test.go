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
	err := v.Validate(v.Map(func() map[string]v.Schema {
		schemas := make(map[string]v.Schema)
		for name, age := range ages {
			schemas[name] = v.Value(age, v.Nonzero[int]())
		}
		return schemas
	}))
	fmt.Printf("%+v\n", err)

	// Output:
	// [foo]: INVALID(is zero valued)
}

func Example_simpleMapEachMapValue() {
	ages := map[string]int{
		"foo": 0,
		"bar": 1,
	}
	err := v.Validate(v.Value(ages, v.EachMapValue[map[string]int](v.Nonzero[int]())))
	fmt.Printf("%+v\n", err)

	// Output:
	// [foo]: INVALID(is zero valued)
}
