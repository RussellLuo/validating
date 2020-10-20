package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v2"
)

type Address struct {
	Country, Province, City string
}

type Person struct {
	Name    string
	Age     int
	Address Address
}

func Example_nestedStruct() {
	p := Person{}
	err := v.Validate(v.Schema{
		v.F("name", &p.Name): v.Len(1, 5),
		v.F("age", &p.Age):   v.Gte(10),
		v.F("address", &p.Address): v.Schema{
			v.F("country", &p.Address.Country):  v.Nonzero(),
			v.F("province", &p.Address.Country): v.Nonzero(),
			v.F("city", &p.Address.City):        v.Nonzero(),
		},
	})
	fmt.Printf("err: %+v\n", err)
}
