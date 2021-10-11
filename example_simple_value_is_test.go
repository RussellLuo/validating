package validating_test

import (
	"fmt"
	v "github.com/RussellLuo/validating/v2"
	"net"
)

func Example_simpleValueIsIP() {
	isIP := v.Is[string](func(value string) bool {
		return net.ParseIP(value) != nil
	})

	value := "192.168.0."
	// See https://github.com/golang/go/issues/41176.
	//err := v.Validate(v.Value(value, isIP))
	err := v.Validate(v.Validator[string](v.Value(value, v.Validator[string](isIP))))
	fmt.Printf("%+v\n", err)

	// Output:
	// INVALID(is invalid)
}
