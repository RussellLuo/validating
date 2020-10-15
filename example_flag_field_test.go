package validating_test

import (
	"flag"
	"fmt"

	v "github.com/RussellLuo/validating/v2"
)

func Example_flagField() {
	// import "flag"
	// import "fmt"
	// import v "github.com/RussellLuo/validating/v2"

	value := flag.String("value", "", "Value argument")
	flag.Parse()

	err := v.Validate(v.Schema{
		v.F("value", value): v.All(v.Nonzero(), v.Len(2, 5)),
	})
	fmt.Printf("err: %+v\n", err)

	// Output:
	// err: value: INVALID(is zero valued)
}
