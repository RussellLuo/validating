package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating"
)

type Person struct {
	Name string
	Age  int
}

func Example_simpleStruct() {
	p := Person{}
	err := v.Validate(v.Schema{
		v.F("name", &p.Name): v.Len(1, 5, "length is not between 1 and 5"),
		v.F("age", &p.Age):   v.Nonzero(),
	})
	fmt.Printf("err: %+v\n", err)

	// Output:
	// err: name: INVALID(length is not between 1 and 5), age: INVALID(is zero valued)
}
