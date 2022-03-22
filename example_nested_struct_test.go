package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v3"
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
		v.F("name", p.Name): v.LenString(1, 5),
		v.F("age", p.Age):   v.Gte(10),
		v.F("address", p.Address): v.Nested(func(addr Address) v.Validator {
			return v.Schema{
				v.F("country", addr.Country):   v.Nonzero[string](),
				v.F("province", addr.Province): v.Nonzero[string](),
				v.F("city", addr.City):         v.Nonzero[string](),
			}
		}),
	})
	fmt.Printf("err: %+v\n", err)

	//err: name: INVALID(has an invalid length), age: INVALID(is lower than the given value), address.country: INVALID(is zero valued), address.province: INVALID(is zero valued), address.city: INVALID(is zero valued)
}
