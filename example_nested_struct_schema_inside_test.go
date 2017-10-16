package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating"
)

type Address3 struct {
	Country, Province, City string
}

func (a *Address3) Schema() v.Schema {
	return v.Schema{
		v.F("country", &a.Country):  v.Nonzero(),
		v.F("province", &a.Country): v.Nonzero(),
		v.F("city", &a.City):        v.Nonzero(),
	}
}

type Person3 struct {
	Name    string
	Age     int
	Address Address3
}

func (p *Person3) Schema() v.Schema {
	return v.Schema{
		v.F("name", &p.Name):       v.Len(1, 5),
		v.F("age", &p.Age):         v.Gte(10),
		v.F("address", &p.Address): v.Nested(p.Address.Schema()),
	}
}

func Example_nestedStructSchemaInside() {
	p := Person3{}
	err := v.Validate(p.Schema())
	fmt.Printf("err: %+v\n", err)
}
