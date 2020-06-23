package validating_test

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"testing"
	"time"

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

func negateErrs(errs v.Errors, validatorName, fieldName, msg string) v.Errors {
	if errs == nil {
		return v.NewErrors(fieldName, v.ErrInvalid, msg)
	}
	switch errs[0].Kind() {
	case v.ErrUnrecognized:
		return errs
	case v.ErrUnsupported:
		return v.NewErrors(fieldName, v.ErrUnsupported, "cannot use validator `"+validatorName+"`")
	default:
		return nil
	}
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
		v.Validate(c.schemaMaker(&c.gotFlag)) // nolint:errcheck
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
				temp := uint8(1)
				value := &temp
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
				temp := uint16(1)
				value := &temp
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
				temp := uint32(1)
				value := &temp
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
				temp := uint64(1)
				value := &temp
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
				temp := int8(1)
				value := &temp
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
				temp := int16(1)
				value := &temp
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
				temp := int32(1)
				value := &temp
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
				temp := int64(1)
				value := &temp
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
				temp := float32(1)
				value := &temp
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
				temp := float64(1)
				value := &temp
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
				temp := uint(1)
				value := &temp
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
				temp := int(1)
				value := &temp
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
				temp := true
				value := &temp
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
				temp := "a"
				value := &temp
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
				temp := time.Now()
				value := &temp
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
				value := (*time.Duration)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				temp := time.Duration(1)
				value := &temp
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []time.Duration{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []time.Duration{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := func() {}
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnrecognized, "of an unrecognized type"),
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

func TestLen(t *testing.T) {
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint8)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint8{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint16)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint16{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint32)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint32{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint64)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint64{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int8)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int8{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int16)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int16{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int32)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int32{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int64)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int64{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*float32)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []float32{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*float64)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []float64{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []uint{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*bool)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []bool{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []string{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*time.Time)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []time.Time{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
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
				value := (*time.Duration)(nil)
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []time.Duration{}
				return &value
			},
			errs: v.NewErrors("value", v.ErrInvalid, "with an invalid length"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []time.Duration{1}
				return &value
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := func() {}
				return &value
			},
			errs: v.NewErrors("value", v.ErrUnrecognized, "of an unrecognized type"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := []int{}
				return &value
			},
			msgs: []string{"is not ok"},
			errs: v.NewErrors("value", v.ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		errs := v.Validate(v.Schema{
			v.F("value", c.valuePtrMaker()): v.Len(1, math.MaxInt64, c.msgs...),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestGt_Lte(t *testing.T) {
	cases := []struct {
		valuePtrMaker func() (interface{}, interface{})
		msgs          []string
		errs          v.Errors
	}{
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint8(0)
				other := uint8(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint8(2)
				other := uint8(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint8)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []uint8{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint16(0)
				other := uint16(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint16(2)
				other := uint16(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint16)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []uint16{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint32(0)
				other := uint32(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint32(2)
				other := uint32(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint32)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []uint32{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint64(0)
				other := uint64(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint64(2)
				other := uint64(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint64)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []uint64{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int8(0)
				other := int8(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int8(2)
				other := int8(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int8)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []int8{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int16(0)
				other := int16(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int16(2)
				other := int16(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int16)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []int16{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int32(0)
				other := int32(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int32(2)
				other := int32(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int32)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []int32{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int64(0)
				other := int64(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int64(2)
				other := int64(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int64)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []int64{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float32(0)
				other := float32(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float32(2)
				other := float32(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*float32)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []float32{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float64(0)
				other := float64(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float64(2)
				other := float64(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*float64)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []float64{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint(0)
				other := uint(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint(2)
				other := uint(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []uint{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(0)
				other := int(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(2)
				other := int(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []int{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := false
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*bool)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []bool{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := ""
				other := "a"
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := "a"
				other := ""
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*string)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []string{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Time{}
				other := time.Now()
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Now()
				other := time.Time{}
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*time.Time)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []time.Time{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Duration(0)
				other := time.Duration(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Duration(2)
				other := time.Duration(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*time.Duration)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []time.Duration{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := func() {}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnrecognized, "of an unrecognized type"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(0)
				other := int(1)
				return &value, other
			},
			msgs: []string{"is not ok"},
			errs: v.NewErrors("value", v.ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		valuePtr, other := c.valuePtrMaker()

		// Test Gt
		errs := v.Validate(v.Schema{
			v.F("value", valuePtr): v.Gt(other, c.msgs...),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}

		// Test Lte
		negativeWantErrs := negateErrs(c.errs, "Lte", "value", "is greater than given value")
		errs = v.Validate(v.Schema{
			v.F("value", valuePtr): v.Lte(other, c.msgs...),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(negativeWantErrs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestGte_Lt(t *testing.T) {
	cases := []struct {
		valuePtrMaker func() (interface{}, interface{})
		msgs          []string
		errs          v.Errors
	}{
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint8(0)
				other := uint8(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint8(1)
				other := uint8(0)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint8(1)
				other := uint8(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint8)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []uint8{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint16(0)
				other := uint16(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint16(1)
				other := uint16(0)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint16(1)
				other := uint16(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint16)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []uint16{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint32(0)
				other := uint32(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint32(1)
				other := uint32(0)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint32(1)
				other := uint32(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint32)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []uint32{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint64(0)
				other := uint64(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint64(1)
				other := uint64(0)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint64(1)
				other := uint64(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint64)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []uint64{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int8(0)
				other := int8(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int8(1)
				other := int8(0)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int8(1)
				other := int8(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int8)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []int8{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int16(0)
				other := int16(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int16(1)
				other := int16(0)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int16(1)
				other := int16(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int16)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []int16{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int32(0)
				other := int32(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int32(1)
				other := int32(0)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int32(1)
				other := int32(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int32)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []int32{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int64(0)
				other := int64(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int64(1)
				other := int64(0)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int64(1)
				other := int64(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int64)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []int64{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float32(0)
				other := float32(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float32(1)
				other := float32(0)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float32(1)
				other := float32(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*float32)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []float32{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float64(0)
				other := float64(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float64(1)
				other := float64(0)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float64(1)
				other := float64(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*float64)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []float64{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint(0)
				other := uint(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint(1)
				other := uint(0)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint(1)
				other := uint(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []uint{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(0)
				other := int(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(1)
				other := int(0)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(1)
				other := int(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []int{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := false
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*bool)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []bool{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := ""
				other := "a"
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := "a"
				other := ""
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := "a"
				other := "a"
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*string)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []string{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Time{}
				other := time.Now()
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Now()
				other := time.Time{}
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Time{}
				other := time.Time{}
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*time.Time)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []time.Time{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Duration(0)
				other := time.Duration(1)
				return &value, other
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Duration(1)
				other := time.Duration(0)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Duration(1)
				other := time.Duration(1)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*time.Duration)(nil)
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := []time.Duration{}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := func() {}
				return &value, value
			},
			errs: v.NewErrors("value", v.ErrUnrecognized, "of an unrecognized type"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(0)
				other := int(1)
				return &value, other
			},
			msgs: []string{"is not ok"},
			errs: v.NewErrors("value", v.ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		valuePtr, other := c.valuePtrMaker()

		// Test Gte
		errs := v.Validate(v.Schema{
			v.F("value", valuePtr): v.Gte(other, c.msgs...),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}

		// Test Lt
		negativeWantErrs := negateErrs(c.errs, "Lt", "value", "is greater than or equal to given value")
		errs = v.Validate(v.Schema{
			v.F("value", valuePtr): v.Lt(other, c.msgs...),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(negativeWantErrs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestIn(t *testing.T) {
	cases := []struct {
		valuePtrMaker func() (interface{}, []interface{})
		msgs          []string
		errs          v.Errors
	}{
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint8(0)
				return &value, []interface{}{uint8(1), uint8(2)}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint8(1)
				return &value, []interface{}{uint8(1), uint8(2)}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*uint8)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []uint8{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint16(0)
				return &value, []interface{}{uint16(1), uint16(2)}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint16(1)
				return &value, []interface{}{uint16(1), uint16(2)}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*uint16)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []uint16{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint32(0)
				return &value, []interface{}{uint32(1), uint32(2)}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint32(1)
				return &value, []interface{}{uint32(1), uint32(2)}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*uint32)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []uint32{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint64(0)
				return &value, []interface{}{uint64(1), uint64(2)}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint64(1)
				return &value, []interface{}{uint64(1), uint64(2)}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*uint64)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []uint64{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int8(0)
				return &value, []interface{}{int8(1), int8(2)}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int8(1)
				return &value, []interface{}{int8(1), int8(2)}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*int8)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []int8{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int16(0)
				return &value, []interface{}{int16(1), int16(2)}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int16(1)
				return &value, []interface{}{int16(1), int16(2)}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*int16)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []int16{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int32(0)
				return &value, []interface{}{int32(1), int32(2)}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int32(1)
				return &value, []interface{}{int32(1), int32(2)}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*int32)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []int32{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int64(0)
				return &value, []interface{}{int64(1), int64(2)}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int64(1)
				return &value, []interface{}{int64(1), int64(2)}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*int64)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []int64{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := float32(0)
				return &value, []interface{}{float32(1), float32(2)}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := float32(1)
				return &value, []interface{}{float32(1), float32(2)}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*float32)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []float32{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := float64(0)
				return &value, []interface{}{float64(1), float64(2)}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := float64(1)
				return &value, []interface{}{float64(1), float64(2)}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*float64)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []float64{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint(0)
				return &value, []interface{}{uint(1), uint(2)}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint(1)
				return &value, []interface{}{uint(1), uint(2)}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*uint)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []uint{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int(0)
				return &value, []interface{}{int(1), int(2)}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int(1)
				return &value, []interface{}{int(1), int(2)}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*int)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []int{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := false
				return &value, []interface{}{true}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := true
				return &value, []interface{}{true}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*bool)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []bool{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := ""
				return &value, []interface{}{"a", "ab"}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := "a"
				return &value, []interface{}{"a", "ab"}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*string)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []string{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := time.Time{}
				return &value, []interface{}{time.Now()}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := time.Time{}
				return &value, []interface{}{value}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*time.Time)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []time.Time{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := time.Duration(0)
				return &value, []interface{}{time.Duration(1), time.Duration(2)}
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of given values"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := time.Duration(1)
				return &value, []interface{}{time.Duration(1), time.Duration(2)}
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := (*time.Duration)(nil)
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := []time.Duration{}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := func() {}
				return &value, nil
			},
			errs: v.NewErrors("value", v.ErrUnrecognized, "of an unrecognized type"),
		},
	}
	for _, c := range cases {
		valuePtr, others := c.valuePtrMaker()
		errs := v.Validate(v.Schema{
			v.F("value", valuePtr): v.In(others...),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestRegexpMatch(t *testing.T) {
	cases := []struct {
		valuePtrMaker func() interface{}
		re            *regexp.Regexp
		msgs          []string
		errs          v.Errors
	}{
		{
			valuePtrMaker: func() interface{} {
				value := 0
				return &value
			},
			re:   regexp.MustCompile(``),
			errs: v.NewErrors("value", v.ErrUnsupported, "cannot use validator `RegexpMatch`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := "x13012345678"
				return &value
			},
			re:   regexp.MustCompile(`^(86)?1\d{10}$`), // cellphone
			errs: v.NewErrors("value", v.ErrInvalid, "does not match the given regular expression"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := "x13012345678"
				return &value
			},
			re:   regexp.MustCompile(`^(86)?1\d{10}$`), // cellphone
			msgs: []string{"invalid cellphone"},
			errs: v.NewErrors("value", v.ErrInvalid, "invalid cellphone"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := "13012345678"
				return &value
			},
			re:   regexp.MustCompile(`^(86)?1\d{10}$`), // cellphone
			errs: nil,
		},
		{
			valuePtrMaker: func() interface{} {
				value := []byte("13012345678")
				return &value
			},
			re:   regexp.MustCompile(`^(86)?1\d{10}$`), // cellphone
			errs: nil,
		},
	}
	for _, c := range cases {
		errs := v.Validate(v.Schema{
			v.F("value", c.valuePtrMaker()): v.RegexpMatch(c.re, c.msgs...),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}
