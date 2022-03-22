package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v3"
)

type Address2 struct {
	Country, Province, City string
}

type Person2 struct {
	Name    string
	Age     int
	Address *Address2
}

func makeSchema2(p *Person2) v.Schema {
	return v.Schema{
		v.F("name", p.Name): v.LenString(1, 5),
		v.F("age", p.Age):   v.Lte(50),
		v.F("address", p.Address): v.All(
			v.Is(func(addr *Address2) bool { return addr != nil }).Msg("is nil"),
			v.Nested(func(addr *Address2) v.Validator {
				return v.Schema{
					v.F("country", addr.Country):   v.Nonzero[string](),
					v.F("province", addr.Province): v.Nonzero[string](),
					v.F("city", addr.City):         v.Nonzero[string](),
				}
			}),
		),
	}
}

func Example_nestedStructPointer() {
	p1 := Person2{}
	err := v.Validate(makeSchema2(&p1))
	fmt.Printf("err of p1: %+v\n", err)

	p2 := Person2{Age: 60, Address: &Address2{}}
	err = v.Validate(makeSchema2(&p2))
	fmt.Printf("err of p2: %+v\n", err)
}
