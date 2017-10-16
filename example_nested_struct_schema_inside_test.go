package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating"
)

type Address struct {
	Country, Province, City string
}

func (a *Address) Schema() v.Schema {
	return v.Schema{
		v.F("country", &a.Country):  v.Nonzero(),
		v.F("province", &a.Country): v.Nonzero(),
		v.F("city", &a.City):        v.Nonzero(),
	}
}

type Person struct {
	Name    string
	Age     int
	Address Address
}

func (p *Person) Schema() v.Schema {
	return v.Schema{
		v.F("name", &p.Name):       v.Len(1, 5),
		v.F("age", &p.Age):         v.Gte(10),
		v.F("address", &p.Address): v.Nested(p.Address.Schema()),
	}
}

func Example_nestedStruct_schemaInside() {
	p := Person{}
	err := v.Validate(p.Schema())
	fmt.Printf("err: %+v\n", err)
}
