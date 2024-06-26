package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v3"
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
		v.F("name", p.Name): v.LenString(1, 5),
		v.F("age", p.Age):   v.Nonzero[int](),
		v.F("phones", p.Phones): v.EachSlice[[]*Phone](v.Nested(func(phone *Phone) v.Validator {
			return v.Schema{
				v.F("number", phone.Number): v.Nonzero[string](),
				v.F("remark", phone.Remark): v.LenString(5, 7),
			}
		})),
	}
}

// The equivalent implementation using Slice.
//nolint:golint,unused
func makeSchema4_Slice(p *Person4) v.Schema {
	return v.Schema{
		v.F("name", p.Name): v.LenString(1, 5),
		v.F("age", p.Age):   v.Nonzero[int](),
		v.F("phones", p.Phones): v.Slice(func(s []*Phone) (schemas []v.Validator) {
			for _, phone := range s {
				schemas = append(schemas, v.Schema{
					v.F("number", phone.Number): v.Nonzero[string](),
					v.F("remark", phone.Remark): v.LenString(5, 7),
				})
			}
			return
		}),
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

	//err of p1: name: INVALID(has an invalid length), age: INVALID(is zero valued)
	//err of p2: name: INVALID(has an invalid length), age: INVALID(is zero valued), phones[1].remark: INVALID(has an invalid length)
}
