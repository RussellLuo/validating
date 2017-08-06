package main

import (
	"fmt"

	v "github.com/RussellLuo/validating"
)

func main() {
	value := 0
	err := v.Validate(v.Schema{
		v.F("value", &value): v.Nonzero(),
	})
	fmt.Printf("err: %+v\n", err)
}
