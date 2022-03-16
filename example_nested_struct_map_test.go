package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating/v2"
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
		v.F("name", &p.Name): v.Len(1, 5),
		v.F("age", &p.Age):   v.Nonzero(),
		v.F("family", &p.Family): v.ZeroOr(
			v.Map(func() map[string]v.Schema {
				schemas := make(map[string]v.Schema)
				for relation, member := range p.Family {
					schemas[relation] = v.Schema{
						v.F("name", &member.Name): v.Len(10, 15).Msg("is too long"),
					}
				}
				return schemas
			}),
		),
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
}
