package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating"
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
		v.F("name", &p.Name): v.Len(1, 5),
		v.F("age", &p.Age):   v.Lte(50),
		v.F("address", &p.Address): v.All(
			v.Assert(p.Address != nil, "is nil"),
			v.Lazy(func() v.Validator {
				return v.Nested(v.Schema{
					v.F("country", &p.Address.Country):  v.Nonzero(),
					v.F("province", &p.Address.Country): v.Nonzero(),
					v.F("city", &p.Address.City):        v.Nonzero(),
				})
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
