package validating_test

import (
	"fmt"

	v "github.com/RussellLuo/validating"
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
		v.F("family", &p.Family): v.All(
			v.Assert(p.Family != nil, "is empty"),
			v.NestedMulti(func() (schemas []v.Schema) {
				for relation, member := range p.Family {
					schemas = append(schemas, v.Schema{
						v.F(fmt.Sprintf("[%s].name", relation), &member.Name): v.Len(10, 15, "is too long"),
					})
				}
				return
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
