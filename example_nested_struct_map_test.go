package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v3"
)

type Member struct {
	Name string
}

type Person1 struct {
	Name   string
	Age    int
	Family map[string]*Member
}

func makeSchema1(p *Person1) v.Schema {
	return v.Schema{
		v.F("name", p.Name): v.LenString(1, 5),
		v.F("age", p.Age):   v.Nonzero[int](),
		v.F("family", p.Family): v.EachMap[map[string]*Member](v.Nested(func(member *Member) v.Validator {
			return v.Schema{
				v.F("name", member.Name): v.LenString(10, 15).Msg("is too long"),
			}
		})),
	}
}

// The equivalent implementation using Map.
//
//nolint:golint,unused
func makeSchema1_Map(p *Person1) v.Schema {
	return v.Schema{
		v.F("name", p.Name): v.LenString(1, 5),
		v.F("age", p.Age):   v.Nonzero[int](),
		v.F("family", p.Family): v.Map(func(m map[string]*Member) map[string]v.Validator {
			schemas := make(map[string]v.Validator)
			for relation, member := range m {
				schemas[relation] = v.Schema{
					v.F("name", member.Name): v.LenString(10, 15).Msg("is too long"),
				}
			}
			return schemas
		}),
	}
}

func Example_nestedStructMap() {
	p1 := Person1{}
	err := v.Validate(makeSchema1(&p1))
	fmt.Printf("err of p1: %+v\n", err)

	p2 := Person1{Family: map[string]*Member{
		"father": {"father's name"},
		"mother": {"mother's name is long"},
	}}
	err = v.Validate(makeSchema1(&p2))
	fmt.Printf("err of p2: %+v\n", err)

	//err of p1: name: INVALID(has an invalid length), age: INVALID(is zero valued)
	//err of p2: name: INVALID(has an invalid length), age: INVALID(is zero valued), family[mother].name: INVALID(is too long)
}
