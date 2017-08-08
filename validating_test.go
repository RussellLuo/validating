package validating_test

import (
	"reflect"
	"testing"
	"time"

	"fmt"
	v "github.com/RussellLuo/validating"
)

type Author struct {
	Name string
	Age  int
}

type Comment struct {
	Content   string
	CreatedAt time.Time
}

type Post struct {
	Author    Author
	Title     string
	CreatedAt time.Time
	Likes     int
	Comments  []Comment
}

func makeErrsMap(errs v.Errors) map[string]v.Error {
	if errs == nil {
		return nil
	}

	formatted := make(map[string]v.Error, len(errs))
	for _, err := range errs {
		formatted[err.Field()] = err
	}
	return formatted
}

func TestAll(t *testing.T) {
	cases := []struct {
		schemaMaker func() v.Schema
		errs        v.Errors
	}{
		{
			func() v.Schema {
				value := ""
				return v.Schema{
					v.F("value", &value): v.All(),
				}
			},
			nil,
		},
		{
			func() v.Schema {
				value := ""
				return v.Schema{
					v.F("value", &value): v.All(v.Nonzero()),
				}
			},
			v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			func() v.Schema {
				value := "a"
				return v.Schema{
					v.F("value", &value): v.All(v.Nonzero(), v.Len(2, 5)),
				}
			},
			v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
		},
		{
			func() v.Schema {
				value := "abc"
				return v.Schema{
					v.F("value", &value): v.All(v.Nonzero(), v.Len(2, 5), v.In("a", "ab")),
				}
			},
			v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			func() v.Schema {
				value := "abc"
				return v.Schema{
					v.F("value", &value): v.All(v.Nonzero(), v.Len(2, 5), v.In("a", "ab", "abc")),
				}
			},
			nil,
		},
	}
	for _, c := range cases {
		errs := v.Validate(c.schemaMaker())
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestAny(t *testing.T) {
	cases := []struct {
		schemaMaker func() v.Schema
		errs        v.Errors
	}{
		{
			func() v.Schema {
				value := ""
				return v.Schema{
					v.F("value", &value): v.Any(),
				}
			},
			nil,
		},
		{
			func() v.Schema {
				value := ""
				return v.Schema{
					v.F("value", &value): v.Any(v.Nonzero()),
				}
			},
			v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			func() v.Schema {
				value := "a"
				return v.Schema{
					v.F("value", &value): v.Any(v.Nonzero(), v.Len(2, 5)),
				}
			},
			nil,
		},
		{
			func() v.Schema {
				value := "abc"
				return v.Schema{
					v.F("value", &value): v.Any(v.Len(1, 2), v.In("a", "ab")),
				}
			},
			v.Errors{
				v.NewError("value", v.ErrInvalid, "with an invalid length"),
				v.NewError("value", v.ErrInvalid, "is not one of given values"),
			},
		},
	}
	for _, c := range cases {
		errs := v.Validate(c.schemaMaker())
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestNot(t *testing.T) {
	cases := []struct {
		schemaMaker func() v.Schema
		errs        v.Errors
	}{
		{
			func() v.Schema {
				value := ""
				return v.Schema{
					v.F("value", &value): v.Not(v.Nonzero()),
				}
			},
			nil,
		},
		{
			func() v.Schema {
				value := "a"
				return v.Schema{
					v.F("value", &value): v.Not(v.All(v.Nonzero(), v.Len(2, 5))),
				}
			},
			nil,
		},
		{
			func() v.Schema {
				value := "a"
				return v.Schema{
					v.F("value", &value): v.Not(v.Any(v.Nonzero(), v.Len(2, 5))),
				}
			},
			v.NewErrors("value", v.ErrInvalid, "is invalid"),
		},
		{
			func() v.Schema {
				value := "a"
				return v.Schema{
					v.F("value", &value): v.Not(v.Any(v.Nonzero(), v.Len(2, 5)), "is not ok"),
				}
			},
			v.NewErrors("value", v.ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		errs := v.Validate(c.schemaMaker())
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestNested(t *testing.T) {
	cases := []struct {
		schemaMaker func() v.Schema
		errs        v.Errors
	}{
		{
			func() v.Schema {
				post := Post{}
				return v.Schema{
					v.F("author", &post.Author): v.Nested(v.Schema{}),
				}
			},
			nil,
		},
		{
			func() v.Schema {
				post := Post{}
				return v.Schema{
					v.F("author", &post.Author): v.Nested(v.Schema{
						v.F("name", &post.Author.Name): v.Nonzero(),
						v.F("age", &post.Author.Age):   v.Nonzero(),
					}),
				}
			},
			v.Errors{
				v.NewError("author.name", v.ErrInvalid, "is zero valued"),
				v.NewError("author.age", v.ErrInvalid, "is zero valued"),
			},
		},
		{
			func() v.Schema {
				post := Post{Author: Author{"russell", 10}}
				return v.Schema{
					v.F("author", &post.Author): v.Nested(v.Schema{
						v.F("name", &post.Author.Name): v.Nonzero(),
						v.F("age", &post.Author.Age):   v.Nonzero(),
					}),
				}
			},
			nil,
		},
	}
	for _, c := range cases {
		errs := v.Validate(c.schemaMaker())
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestNestedMulti(t *testing.T) {
	cases := []struct {
		schemaMaker func() v.Schema
		errs        v.Errors
	}{
		{
			func() v.Schema {
				post := Post{}
				return v.Schema{
					v.F("comments", &post.Comments): v.NestedMulti(func() []v.Schema {
						return nil
					}),
				}
			},
			nil,
		},
		{
			func() v.Schema {
				post := Post{Comments: []Comment{
					{"", time.Time{}},
				}}
				return v.Schema{
					v.F("comments", &post.Comments): v.NestedMulti(func() (schemas []v.Schema) {
						for i := range post.Comments {
							schemas = append(schemas, v.Schema{
								v.F(fmt.Sprintf("[%d].content", i), &post.Comments[i].Content):      v.Nonzero(),
								v.F(fmt.Sprintf("[%d].created_at", i), &post.Comments[i].CreatedAt): v.Nonzero(),
							})
						}
						return
					}),
				}
			},
			v.Errors{
				v.NewError("comments.[0].content", v.ErrInvalid, "is zero valued"),
				v.NewError("comments.[0].created_at", v.ErrInvalid, "is zero valued"),
			},
		},
		{
			func() v.Schema {
				post := Post{Comments: []Comment{
					{Content: "thanks", CreatedAt: time.Now()},
				}}
				return v.Schema{
					v.F("comments", &post.Comments): v.NestedMulti(func() (schemas []v.Schema) {
						for i := range post.Comments {
							schemas = append(schemas, v.Schema{
								v.F(fmt.Sprintf("[%d].content", i), &post.Comments[i].Content):      v.Nonzero(),
								v.F(fmt.Sprintf("[%d].created_at", i), &post.Comments[i].CreatedAt): v.Nonzero(),
							})
						}
						return
					}),
				}
			},
			nil,
		},
	}
	for _, c := range cases {
		errs := v.Validate(c.schemaMaker())
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestLazy(t *testing.T) {
	cases := []struct {
		schemaMaker func(*bool) v.Schema
		gotFlag     bool
		wantFlag    bool
	}{
		{
			schemaMaker: func(flag *bool) v.Schema {
				post := Post{}
				return v.Schema{
					v.F("title", &post.Title): v.Lazy(func() v.Validator {
						*flag = true
						return v.Len(2, 5)
					}),
				}
			},
			wantFlag: true,
		},
		{
			schemaMaker: func(flag *bool) v.Schema {
				post := Post{}
				return v.Schema{
					v.F("title", &post.Title): v.All(
						v.Nonzero(),
						v.Lazy(func() v.Validator {
							*flag = true
							return v.Len(2, 5)
						}),
					),
				}
			},
			wantFlag: false,
		},
	}
	for _, c := range cases {
		v.Validate(c.schemaMaker(&c.gotFlag))
		if !reflect.DeepEqual(c.gotFlag, c.wantFlag) {
			t.Errorf("Got (%+v) != Want (%+v)", c.gotFlag, c.wantFlag)
		}
	}
}

func TestAssert(t *testing.T) {
	cases := []struct {
		schemaMaker func() v.Schema
		errs        v.Errors
	}{
		{
			func() v.Schema {
				post := Post{}
				return v.Schema{
					v.F("comments", &post.Comments): v.Assert(true),
				}
			},
			nil,
		},
		{
			func() v.Schema {
				post := Post{}
				return v.Schema{
					v.F("comments", &post.Comments): v.Assert(false),
				}
			},
			v.NewErrors("comments", v.ErrInvalid, "is invalid"),
		},
		{
			func() v.Schema {
				post := Post{}
				return v.Schema{
					v.F("comments", &post.Comments): v.Assert(false, "is not ok"),
				}
			},
			v.NewErrors("comments", v.ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		errs := v.Validate(c.schemaMaker())
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestNonzero(t *testing.T) {
	cases := []struct {
		valuePtrMaker func() interface{}
		msgs          []string
		errs          v.Errors
	}{
		{
			valuePtrMaker: func() interface{} {
				value := uint8(0)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := uint8(1)
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint8)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := uint8(1)
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint8{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint8{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := uint16(0)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := uint16(1)
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint16)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := uint16(1)
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint16{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint16{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := uint32(0)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := uint32(1)
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint32)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := uint32(1)
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint32{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint32{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := uint64(0)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := uint64(1)
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint64)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := uint64(1)
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint64{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint64{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := int8(0)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := int8(1)
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int8)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := int8(1)
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int8{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int8{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := int16(0)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := int16(1)
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int16)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := int16(1)
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int16{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int16{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := int32(0)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := int32(1)
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int32)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := int32(1)
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int32{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int32{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := int64(0)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := int64(1)
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int64)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := int64(1)
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int64{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int64{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := float32(0)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := float32(1)
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*float32)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := float32(1)
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []float32{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []float32{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := float64(0)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := float64(1)
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*float64)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := float64(1)
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []float64{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []float64{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := uint(0)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := uint(1)
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := uint(1)
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := int(0)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := int(1)
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := int(1)
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := false
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := true
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*bool)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := true
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []bool{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []bool{true}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := ""
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := "a"
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*string)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := "a"
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []string{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []string{"a"}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := time.Time{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := time.Now()
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*time.Time)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				inner := time.Now()
				value := &inner
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []time.Time{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []time.Time{time.Now()}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := time.Duration(0)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := time.Duration(1)
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := func() {}
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnrecognized, "of unrecognized type"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := int(0)
				return &value
			},
			msgs: []string{"is not ok"},
			errs: v.NewErrors("value", v.ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		errs := v.Validate(v.Schema{
			v.F("value", c.valuePtrMaker()): v.Nonzero(c.msgs...),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}
