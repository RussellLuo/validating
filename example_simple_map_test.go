package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v2"
)

func Example_simpleMap() {
	ages := map[string]int{
		"foo": 0,
		"bar": 1,
	}
	err := v.Validate(v.Map(func() map[string]v.Schema {
		schemas := make(map[string]v.Schema)
		for name, age := range ages {
			age := age
			schemas[name] = v.Schema{
				v.F("", &age): v.Nonzero(),
			}
		}
		return schemas
	}))
	fmt.Printf("%+v\n", err)

	// Output:
	// [foo]: INVALID(is zero valued)
}
