package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating"
)

type Address2 struct {
	Country, Province, City string
}

func (a *Address2) Schema() v.Schema {
	return v.Schema{
		v.F("country", &a.Country):  v.Nonzero(),
		v.F("province", &a.Country): v.Nonzero(),
		v.F("city", &a.City):        v.Nonzero(),
	}
}

type Person2 struct {
	Name    string
	Age     int
	Address Address2
}

func (p *Person2) Schema() v.Schema {
	return v.Schema{
		v.F("name", &p.Name):       v.Len(1, 5),
		v.F("age", &p.Age):         v.Gte(10),
		v.F("address", &p.Address): v.Nested(p.Address.Schema()),
	}
}

func Example_nestedStructSchemaInside() {
	p := Person2{}
	err := v.Validate(p.Schema())
	fmt.Printf("err: %+v\n", err)
}
