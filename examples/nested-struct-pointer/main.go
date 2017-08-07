package main

import (
	"fmt"

	v "github.com/RussellLuo/validating"
)

type Address struct {
	Country, Province, City string
}

type Person struct {
	Name    string
	Age     int
	Address *Address
}

func makeSchema(p *Person) v.Schema {
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

func main() {
	p1 := Person{}
	err := v.Validate(makeSchema(&p1))
	fmt.Printf("err of p1: %+v\n", err)

	p2 := Person{Age: 60, Address: &Address{}}
	err = v.Validate(makeSchema(&p2))
	fmt.Printf("err of p2: %+v\n", err)
}
