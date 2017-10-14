package validating_test

import (
	"flag"
	"fmt"

	v "github.com/RussellLuo/validating"
)

func Example_flagField() {
	// import "flag"
	// import v "github.com/RussellLuo/validating

	value := flag.String("value", "", "Value argument")
	flag.Parse()

	err := v.Validate(v.Schema{
		v.F("value", value): v.All(v.Nonzero(), v.Len(2, 5)),
	})
	fmt.Printf("err: %+v\n", err)

	// Output:
	// err: value: INVALID(is zero valued)
}
