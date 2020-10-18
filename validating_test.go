package validating

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"testing"
	"time"
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

func makeErrsMap(errs Errors) map[string]Error {
	if errs == nil {
		return nil
	}

	formatted := make(map[string]Error, len(errs))
	for _, err := range errs {
		formatted[err.Field()] = err
	}
	return formatted
}

func negateErrs(errs Errors, validatorName, fieldName, msg string) Errors {
	if len(errs) == 0 {
		return NewErrors(fieldName, ErrInvalid, msg)
	}
	switch errs[0].Kind() {
	case ErrUnsupported:
		return NewErrors(fieldName, ErrUnsupported, "cannot use validator `"+validatorName+"`")
	case ErrUnrecognized:
		return NewErrors(fieldName, ErrUnrecognized, "of an unrecognized type")
	default:
		return nil
	}
}

func TestAll(t *testing.T) {
	cases := []struct {
		schemaMaker func() Schema
		errs        Errors
	}{
		{
			func() Schema {
				value := ""
				return Schema{
					F("value", &value): All(),
				}
			},
			nil,
		},
		{
			func() Schema {
				value := ""
				return Schema{
					F("value", &value): All(Nonzero()),
				}
			},
			NewErrors("value", ErrInvalid, "is zero valued"),
		},
		{
			func() Schema {
				value := "a"
				return Schema{
					F("value", &value): All(Nonzero(), Len(2, 5)),
				}
			},
			NewErrors("value", ErrInvalid, "with an invalid length"),
		},
		{
			func() Schema {
				value := "abc"
				return Schema{
					F("value", &value): All(Nonzero(), Len(2, 5), In("a", "ab")),
				}
			},
			NewErrors("value", ErrInvalid, "is not one of given values"),
		},
		{
			func() Schema {
				value := "abc"
				return Schema{
					F("value", &value): All(Nonzero(), Len(2, 5), In("a", "ab", "abc")),
				}
			},
			nil,
		},
	}
	for _, c := range cases {
		errs := Validate(c.schemaMaker())
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestAny(t *testing.T) {
	cases := []struct {
		schemaMaker func() Schema
		errs        Errors
	}{
		{
			func() Schema {
				value := ""
				return Schema{
					F("value", &value): Any(),
				}
			},
			nil,
		},
		{
			func() Schema {
				value := ""
				return Schema{
					F("value", &value): Any(Nonzero()),
				}
			},
			NewErrors("value", ErrInvalid, "is zero valued"),
		},
		{
			func() Schema {
				value := "a"
				return Schema{
					F("value", &value): Any(Nonzero(), Len(2, 5)),
				}
			},
			nil,
		},
		{
			func() Schema {
				value := "abc"
				return Schema{
					F("value", &value): Any(Len(1, 2), In("a", "ab")),
				}
			},
			Errors{
				NewError("value", ErrInvalid, "with an invalid length"),
				NewError("value", ErrInvalid, "is not one of given values"),
			},
		},
	}
	for _, c := range cases {
		errs := Validate(c.schemaMaker())
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestNot(t *testing.T) {
	cases := []struct {
		schemaMaker func() Schema
		errs        Errors
	}{
		{
			func() Schema {
				value := []string{"foo"}
				return Schema{
					F("value", &value): Not(Any(Len(2, 3), Eq("foo"))),
				}
			},
			NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			func() Schema {
				value := ""
				return Schema{
					F("value", &value): Not(Nonzero()),
				}
			},
			nil,
		},
		{
			func() Schema {
				value := "a"
				return Schema{
					F("value", &value): Not(All(Nonzero(), Len(2, 5))),
				}
			},
			nil,
		},
		{
			func() Schema {
				value := "a"
				return Schema{
					F("value", &value): Not(Any(Nonzero(), Len(2, 5))),
				}
			},
			NewErrors("value", ErrInvalid, "is invalid"),
		},
		{
			func() Schema {
				value := "a"
				return Schema{
					F("value", &value): Not(Any(Nonzero(), Len(2, 5))).Msg("is not ok"),
				}
			},
			NewErrors("value", ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		errs := Validate(c.schemaMaker())
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestNested(t *testing.T) {
	cases := []struct {
		schemaMaker func() Schema
		errs        Errors
	}{
		{
			func() Schema {
				post := Post{}
				return Schema{
					F("author", &post.Author): Nested(Schema{}),
				}
			},
			nil,
		},
		{
			func() Schema {
				post := Post{}
				return Schema{
					F("author", &post.Author): Nested(Schema{
						F(".name", &post.Author.Name): Nonzero(),
						F(".age", &post.Author.Age):   Nonzero(),
					}),
				}
			},
			Errors{
				NewError("author.name", ErrInvalid, "is zero valued"),
				NewError("author.age", ErrInvalid, "is zero valued"),
			},
		},
		{
			func() Schema {
				post := Post{Author: Author{"russell", 10}}
				return Schema{
					F("author", &post.Author): Nested(Schema{
						F(".name", &post.Author.Name): Nonzero(),
						F(".age", &post.Author.Age):   Nonzero(),
					}),
				}
			},
			nil,
		},
	}
	for _, c := range cases {
		errs := Validate(c.schemaMaker())
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestNestedMulti(t *testing.T) {
	cases := []struct {
		schemaMaker func() Schema
		errs        Errors
	}{
		{
			func() Schema {
				post := Post{}
				return Schema{
					F("comments", &post.Comments): NestedMulti(func() []Schema {
						return nil
					}),
				}
			},
			nil,
		},
		{
			func() Schema {
				post := Post{Comments: []Comment{
					{"", time.Time{}},
				}}
				return Schema{
					F("comments", &post.Comments): NestedMulti(func() (schemas []Schema) {
						for i := range post.Comments {
							schemas = append(schemas, Schema{
								F(fmt.Sprintf("[%d].content", i), &post.Comments[i].Content):      Nonzero(),
								F(fmt.Sprintf("[%d].created_at", i), &post.Comments[i].CreatedAt): Nonzero(),
							})
						}
						return
					}),
				}
			},
			Errors{
				NewError("comments[0].content", ErrInvalid, "is zero valued"),
				NewError("comments[0].created_at", ErrInvalid, "is zero valued"),
			},
		},
		{
			func() Schema {
				post := Post{Comments: []Comment{
					{Content: "thanks", CreatedAt: time.Now()},
				}}
				return Schema{
					F("comments", &post.Comments): NestedMulti(func() (schemas []Schema) {
						for i := range post.Comments {
							schemas = append(schemas, Schema{
								F(fmt.Sprintf("[%d].content", i), &post.Comments[i].Content):      Nonzero(),
								F(fmt.Sprintf("[%d].created_at", i), &post.Comments[i].CreatedAt): Nonzero(),
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
		errs := Validate(c.schemaMaker())
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestLazy(t *testing.T) {
	cases := []struct {
		schemaMaker func(*bool) Schema
		gotFlag     bool
		wantFlag    bool
	}{
		{
			schemaMaker: func(flag *bool) Schema {
				post := Post{}
				return Schema{
					F("title", &post.Title): Lazy(func() Validator {
						*flag = true
						return Len(2, 5)
					}),
				}
			},
			wantFlag: true,
		},
		{
			schemaMaker: func(flag *bool) Schema {
				post := Post{}
				return Schema{
					F("title", &post.Title): All(
						Nonzero(),
						Lazy(func() Validator {
							*flag = true
							return Len(2, 5)
						}),
					),
				}
			},
			wantFlag: false,
		},
	}
	for _, c := range cases {
		Validate(c.schemaMaker(&c.gotFlag)) // nolint:errcheck
		if !reflect.DeepEqual(c.gotFlag, c.wantFlag) {
			t.Errorf("Got (%+v) != Want (%+v)", c.gotFlag, c.wantFlag)
		}
	}
}

func TestAssert(t *testing.T) {
	cases := []struct {
		schemaMaker func() Schema
		errs        Errors
	}{
		{
			func() Schema {
				post := Post{}
				return Schema{
					F("comments", &post.Comments): Assert(true),
				}
			},
			nil,
		},
		{
			func() Schema {
				post := Post{}
				return Schema{
					F("comments", &post.Comments): Assert(false),
				}
			},
			NewErrors("comments", ErrInvalid, "is invalid"),
		},
		{
			func() Schema {
				post := Post{}
				return Schema{
					F("comments", &post.Comments): Assert(false).Msg("is not ok"),
				}
			},
			NewErrors("comments", ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		errs := Validate(c.schemaMaker())
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestNonzero(t *testing.T) {
	cases := []struct {
		valuePtrMaker func() interface{}
		msg           string
		errs          Errors
	}{
		{
			valuePtrMaker: func() interface{} {
				value := uint8(0)
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []uint8
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []uint16
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []uint32
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []uint64
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []int8
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []int16
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []int32
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []int64
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []float32
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []float64
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []uint
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []int
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []bool
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []string
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []time.Time
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
				var value []time.Duration
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "is zero valued"),
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
			errs: NewErrors("value", ErrUnrecognized, "of an unrecognized type"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := int(0)
				return &value
			},
			msg:  "is not ok",
			errs: NewErrors("value", ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		errs := Validate(Schema{
			F("value", c.valuePtrMaker()): Nonzero().Msg(c.msg),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestLen(t *testing.T) {
	cases := []struct {
		valuePtrMaker func() interface{}
		msg           string
		errs          Errors
	}{
		{
			valuePtrMaker: func() interface{} {
				value := uint8(0)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint8)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []uint8
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint16)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []uint16
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint32)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []uint32
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint64)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []uint64
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int8)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []int8
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int16)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []int16
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int32)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []int32
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int64)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []int64
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*float32)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []float32
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*float64)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []float64
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*uint)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []uint
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*int)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []int
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*bool)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []bool
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []string
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := (*time.Time)(nil)
				return &value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []time.Time
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Len`"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []time.Duration
				return &value
			},
			errs: NewErrors("value", ErrInvalid, "with an invalid length"),
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
			errs: NewErrors("value", ErrUnrecognized, "of an unrecognized type"),
		},
		{
			valuePtrMaker: func() interface{} {
				var value []int
				return &value
			},
			msg:  "is not ok",
			errs: NewErrors("value", ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		errs := Validate(Schema{
			F("value", c.valuePtrMaker()): Len(1, math.MaxInt64).Msg(c.msg),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestRuneCount(t *testing.T) {
	cases := []struct {
		valuePtrMaker func() (interface{}, int, int)
		msg           string
		errs          Errors
	}{
		{
			valuePtrMaker: func() (interface{}, int, int) {
				value := 0
				return &value, 1, 2
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `RuneCount`"),
		},
		{
			valuePtrMaker: func() (interface{}, int, int) {
				value := ""
				return &value, 1, 2
			},
			errs: NewErrors("value", ErrInvalid, "the number of runes is not between the given range"),
		},
		{
			valuePtrMaker: func() (interface{}, int, int) {
				value := "a"
				return &value, 1, 2
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, int, int) {
				value := "你好"
				return &value, 1, 2
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, int, int) {
				value := "你好吗"
				return &value, 1, 2
			},
			errs: NewErrors("value", ErrInvalid, "the number of runes is not between the given range"),
		},
		{
			valuePtrMaker: func() (interface{}, int, int) {
				value := "你好吗"
				return &value, 1, 2
			},
			msg:  "is not ok",
			errs: NewErrors("value", ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		valuePtr, min, max := c.valuePtrMaker()
		errs := Validate(Schema{
			F("value", valuePtr): RuneCount(min, max).Msg(c.msg),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestEq_Ne(t *testing.T) {
	cases := []struct {
		valuePtrMaker func() (interface{}, interface{})
		msg           string
		errs          Errors
	}{
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint8(0)
				other := uint8(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint8(2)
				other := uint8(2)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint8)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint8
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint16(0)
				other := uint16(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint16(2)
				other := uint16(2)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint16)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint16
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint32(0)
				other := uint32(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint32(2)
				other := uint32(2)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint32)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint32
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint64(0)
				other := uint64(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint64(2)
				other := uint64(2)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint64)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint64
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int8(0)
				other := int8(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int8(2)
				other := int8(2)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int8)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int8
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int16(0)
				other := int16(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int16(2)
				other := int16(2)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int16)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int16
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int32(0)
				other := int32(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int32(2)
				other := int32(2)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int32)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int32
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int64(0)
				other := int64(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int64(2)
				other := int64(2)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int64)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int64
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float32(0)
				other := float32(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float32(2)
				other := float32(2)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*float32)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []float32
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float64(0)
				other := float64(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float64(2)
				other := float64(2)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*float64)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []float64
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint(0)
				other := uint(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint(2)
				other := uint(2)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*uint)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(0)
				other := int(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(2)
				other := int(2)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*int)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := false
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*bool)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []bool
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := ""
				other := "a"
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []string
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value, _ := time.Parse(time.RFC3339, "2020-10-14T00:00:00Z")
				other, _ := time.Parse(time.RFC3339, "2020-10-13T00:00:00Z")
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value, _ := time.Parse(time.RFC3339, "2020-10-14T00:00:00Z")
				other, _ := time.Parse(time.RFC3339, "2020-10-14T00:00:00Z")
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*time.Time)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []time.Time
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Duration(0)
				other := time.Duration(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "does not equal the given value"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Duration(2)
				other := time.Duration(2)
				return &value, other
			},
			errs: nil,
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*time.Duration)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []time.Duration
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Eq`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := func() {}
				return &value, value
			},
			errs: NewErrors("value", ErrUnrecognized, "of an unrecognized type"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(0)
				other := int(1)
				return &value, other
			},
			msg:  "is not ok",
			errs: NewErrors("value", ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			valuePtr, other := c.valuePtrMaker()

			// Test Eq
			errs := Validate(Schema{
				F("value", valuePtr): Eq(other).Msg(c.msg),
			})
			if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
				t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
			}

			// Test Ne
			negativeWantErrs := negateErrs(c.errs, "Ne", "value", "equals the given value")
			errs = Validate(Schema{
				F("value", valuePtr): Ne(other).Msg(c.msg),
			})
			if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(negativeWantErrs)) {
				t.Errorf("Got (%+v) != Want (%+v)", errs, negativeWantErrs)
			}
		})
	}
}

func TestGt_Lte(t *testing.T) {
	cases := []struct {
		valuePtrMaker func() (interface{}, interface{})
		msg           string
		errs          Errors
	}{
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint8(0)
				other := uint8(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint8
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint16(0)
				other := uint16(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint16
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint32(0)
				other := uint32(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint32
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint64(0)
				other := uint64(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint64
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int8(0)
				other := int8(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int8
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int16(0)
				other := int16(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int16
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int32(0)
				other := int32(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int32
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int64(0)
				other := int64(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int64
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float32(0)
				other := float32(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []float32
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float64(0)
				other := float64(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []float64
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint(0)
				other := uint(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(0)
				other := int(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := false
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*bool)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []bool
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := ""
				other := "a"
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []string
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Time{}
				other := time.Now()
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []time.Time
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Duration(0)
				other := time.Duration(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than or equal to given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []time.Duration
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gt`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := func() {}
				return &value, value
			},
			errs: NewErrors("value", ErrUnrecognized, "of an unrecognized type"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(0)
				other := int(1)
				return &value, other
			},
			msg:  "is not ok",
			errs: NewErrors("value", ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		valuePtr, other := c.valuePtrMaker()

		// Test Gt
		errs := Validate(Schema{
			F("value", valuePtr): Gt(other).Msg(c.msg),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}

		// Test Lte
		negativeWantErrs := negateErrs(c.errs, "Lte", "value", "is greater than given value")
		errs = Validate(Schema{
			F("value", valuePtr): Lte(other).Msg(c.msg),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(negativeWantErrs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, negativeWantErrs)
		}
	}
}

func TestGte_Lt(t *testing.T) {
	cases := []struct {
		valuePtrMaker func() (interface{}, interface{})
		msg           string
		errs          Errors
	}{
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint8(0)
				other := uint8(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint8
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint16(0)
				other := uint16(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint16
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint32(0)
				other := uint32(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint32
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint64(0)
				other := uint64(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint64
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int8(0)
				other := int8(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int8
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int16(0)
				other := int16(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int16
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int32(0)
				other := int32(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int32
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int64(0)
				other := int64(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int64
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float32(0)
				other := float32(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []float32
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := float64(0)
				other := float64(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []float64
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := uint(0)
				other := uint(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []uint
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(0)
				other := int(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []int
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := false
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := (*bool)(nil)
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []bool
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := ""
				other := "a"
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []string
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Time{}
				other := time.Now()
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []time.Time
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := time.Duration(0)
				other := time.Duration(1)
				return &value, other
			},
			errs: NewErrors("value", ErrInvalid, "is lower than given value"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				var value []time.Duration
				return &value, value
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `Gte`"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := func() {}
				return &value, value
			},
			errs: NewErrors("value", ErrUnrecognized, "of an unrecognized type"),
		},
		{
			valuePtrMaker: func() (interface{}, interface{}) {
				value := int(0)
				other := int(1)
				return &value, other
			},
			msg:  "is not ok",
			errs: NewErrors("value", ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		valuePtr, other := c.valuePtrMaker()

		// Test Gte
		errs := Validate(Schema{
			F("value", valuePtr): Gte(other).Msg(c.msg),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}

		// Test Lt
		negativeWantErrs := negateErrs(c.errs, "Lt", "value", "is greater than or equal to given value")
		errs = Validate(Schema{
			F("value", valuePtr): Lt(other).Msg(c.msg),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(negativeWantErrs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, negativeWantErrs)
		}
	}
}

func TestIn_Nin(t *testing.T) {
	cases := []struct {
		valuePtrMaker func() (interface{}, []interface{})
		msg           string
		errs          Errors
	}{
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint8(0)
				return &value, []interface{}{uint8(1), uint8(2)}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []uint8
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint16(0)
				return &value, []interface{}{uint16(1), uint16(2)}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []uint16
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint32(0)
				return &value, []interface{}{uint32(1), uint32(2)}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []uint32
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint64(0)
				return &value, []interface{}{uint64(1), uint64(2)}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []uint64
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int8(0)
				return &value, []interface{}{int8(1), int8(2)}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []int8
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int16(0)
				return &value, []interface{}{int16(1), int16(2)}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []int16
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int32(0)
				return &value, []interface{}{int32(1), int32(2)}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []int32
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int64(0)
				return &value, []interface{}{int64(1), int64(2)}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []int64
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := float32(0)
				return &value, []interface{}{float32(1), float32(2)}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []float32
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := float64(0)
				return &value, []interface{}{float64(1), float64(2)}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []float64
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := uint(0)
				return &value, []interface{}{uint(1), uint(2)}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []uint
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int(0)
				return &value, []interface{}{int(1), int(2)}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []int
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := false
				return &value, []interface{}{true}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []bool
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := ""
				return &value, []interface{}{"a", "ab"}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []string
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := time.Time{}
				return &value, []interface{}{time.Now()}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []time.Time
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := time.Duration(0)
				return &value, []interface{}{time.Duration(1), time.Duration(2)}
			},
			errs: NewErrors("value", ErrInvalid, "is not one of given values"),
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
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				var value []time.Duration
				return &value, nil
			},
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `In`"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := func() {}
				return &value, nil
			},
			errs: NewErrors("value", ErrUnrecognized, "of an unrecognized type"),
		},
		{
			valuePtrMaker: func() (interface{}, []interface{}) {
				value := int(0)
				return &value, []interface{}{int(1), int(2)}
			},
			msg:  "is not ok",
			errs: NewErrors("value", ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			valuePtr, others := c.valuePtrMaker()

			// Test In
			errs := Validate(Schema{
				F("value", valuePtr): In(others...).Msg(c.msg),
			})
			if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
				t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
			}

			// Test Nin
			negativeWantErrs := negateErrs(c.errs, "Nin", "value", "is one of given values")
			errs = Validate(Schema{
				F("value", valuePtr): Nin(others...).Msg(c.msg),
			})
			if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(negativeWantErrs)) {
				t.Errorf("Got (%+v) != Want (%+v)", errs, negativeWantErrs)
			}
		})
	}
}

func TestRegexpMatch(t *testing.T) {
	cases := []struct {
		valuePtrMaker func() interface{}
		re            *regexp.Regexp
		msg           string
		errs          Errors
	}{
		{
			valuePtrMaker: func() interface{} {
				value := 0
				return &value
			},
			re:   regexp.MustCompile(``),
			errs: NewErrors("value", ErrUnsupported, "cannot use validator `RegexpMatch`"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := "x13012345678"
				return &value
			},
			re:   regexp.MustCompile(`^(86)?1\d{10}$`), // cellphone
			errs: NewErrors("value", ErrInvalid, "does not match the given regular expression"),
		},
		{
			valuePtrMaker: func() interface{} {
				value := "x13012345678"
				return &value
			},
			re:   regexp.MustCompile(`^(86)?1\d{10}$`), // cellphone
			msg:  "invalid cellphone",
			errs: NewErrors("value", ErrInvalid, "invalid cellphone"),
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
		errs := Validate(Schema{
			F("value", c.valuePtrMaker()): RegexpMatch(c.re).Msg(c.msg),
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}
