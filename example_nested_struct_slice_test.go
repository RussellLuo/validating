package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v2"
)

type Phone struct {
	Number, Remark string
}

type Person4 struct {
	Name   string
	Age    int
	Phones []*Phone
}

func makeSchema4(p *Person4) v.Schema {
	return v.Schema{
		v.F("name", &p.Name): v.Len(1, 5),
		v.F("age", &p.Age):   v.Nonzero(),
		v.F("phones", &p.Phones): v.ZeroOr(
			v.Slice(func() (schemas []v.Schema) {
				for _, phone := range p.Phones {
					schemas = append(schemas, v.Schema{
						v.F("number", &phone.Number): v.Nonzero(),
						v.F("remark", &phone.Remark): v.Len(5, 7),
					})
				}
				return
			}),
		),
	}
}

func Example_nestedStructSlice() {
	p1 := Person4{}
	err := v.Validate(makeSchema4(&p1))
	fmt.Printf("err of p1: %+v\n", err)

	p2 := Person4{Phones: []*Phone{
		{"13011112222", "private"},
		{"13033334444", "business"},
	}}
	err = v.Validate(makeSchema4(&p2))
	fmt.Printf("err of p2: %+v\n", err)
}
