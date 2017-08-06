package main

import (
	"fmt"

	v "github.com/RussellLuo/validating"
)

func main() {
	value := 0
	err := v.Validate(v.Schema{
		v.F("value", &value): v.Range(1, 5),
	})
	fmt.Printf("err: %+v\n", err)
}
