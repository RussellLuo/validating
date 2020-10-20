package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v2"
)

type Person5 struct {
	Name string
	Age  int
}

func Example_simpleStruct() {
	p := Person5{Age: 1}
	err := v.Validate(v.Schema{
		v.F("name", &p.Name): v.Len(1, 5).Msg("length is not between 1 and 5"),
		v.F("age", &p.Age):   v.Nonzero(),
	})
	fmt.Printf("%+v\n", err)

	// Output:
	// name: INVALID(length is not between 1 and 5)
}
