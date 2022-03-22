package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v3"
)

type Address3 struct {
	Country, Province, City string
}

func (a *Address3) Schema() v.Schema {
	return v.Schema{
		v.F("country", a.Country):   v.Nonzero[string](),
		v.F("province", a.Province): v.Nonzero[string](),
		v.F("city", a.City):         v.Nonzero[string](),
	}
}

type Person3 struct {
	Name    string
	Age     int
	Address Address3
}

func (p *Person3) Schema() v.Schema {
	return v.Schema{
		v.F("name", p.Name):       v.LenString(1, 5),
		v.F("age", p.Age):         v.Gte(10),
		v.F("address", p.Address): p.Address.Schema(),
	}
}

func Example_nestedStructSchemaInside() {
	p := Person3{}
	err := v.Validate(p.Schema())
	fmt.Printf("err: %+v\n", err)

	//err: name: INVALID(has an invalid length), age: INVALID(is lower than the given value), address.country: INVALID(is zero valued), address.province: INVALID(is zero valued), address.city: INVALID(is zero valued)
}
