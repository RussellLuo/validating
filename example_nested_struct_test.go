package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating"
)

type Address1 struct {
	Country, Province, City string
}

type Person1 struct {
	Name    string
	Age     int
	Address Address1
}

func Example_nestedStruct() {
	p := Person1{}
	err := v.Validate(v.Schema{
		v.F("name", &p.Name): v.Len(1, 5),
		v.F("age", &p.Age):   v.Gte(10),
		v.F("address", &p.Address): v.Nested(v.Schema{
			v.F("country", &p.Address.Country):  v.Nonzero(),
			v.F("province", &p.Address.Country): v.Nonzero(),
			v.F("city", &p.Address.City):        v.Nonzero(),
		}),
	})
	fmt.Printf("err: %+v\n", err)
}
