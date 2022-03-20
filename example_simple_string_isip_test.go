package validating_test

import (
	"fmt"
	"net"

	v "github.com/RussellLuo/validating/v3"
)

func Example_simpleStringIsIP() {
	isIP := func(value string) bool {
		return net.ParseIP(value) != nil
	}

	value := "192.168.0."
	err := v.Validate(v.Value(value, v.Is(isIP).Msg("invalid IP")))
	fmt.Printf("%+v\n", err)

	// Output:
	// INVALID(invalid IP)
}
