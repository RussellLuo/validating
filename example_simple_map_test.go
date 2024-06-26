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
	err := v.Validate(v.Value(ages, v.EachMap[map[string]int](v.Nonzero[int]())))
	fmt.Printf("%+v\n", err)

	// Output:
	// [foo]: INVALID(is zero valued)
}
