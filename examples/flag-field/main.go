package main

import (
	"flag"
	"fmt"

	v "github.com/RussellLuo/validating"
)

func main() {
	value := flag.String("value", "", "Value argument")
	flag.Parse()

	err := v.Validate(v.Schema{
		v.F("value", value): v.All(v.Nonzero(), v.Len(2, 5)),
	})
	fmt.Printf("err: %+v\n", err)
}
