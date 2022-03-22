package validating_test

import (
	"reflect"
	"regexp"
	"testing"
	"time"

	v "github.com/RussellLuo/validating/v3"
)

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

func TestNested(t *testing.T) {
	cases := []struct {
		name      string
		value     interface{}
		validator v.Validator
		errs      v.Errors
	}{
		{
			name:  "invalid",
			value: struct{ Foo int }{Foo: 0},
			validator: v.Nested(func(s struct{ Foo int }) v.Validator {
				return v.Schema{
					v.F("foo", s.Foo): v.Nonzero[int](),
				}
			}),
			errs: v.NewErrors("foo", v.ErrInvalid, "is zero valued"),
		},
		{
			name:  "valid",
			value: struct{ Foo int }{Foo: 1},
			validator: v.Nested(func(s struct{ Foo int }) v.Validator {
				return v.Schema{
					v.F("foo", s.Foo): v.Nonzero[int](),
				}
			}),
			errs: nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := v.Validate(v.Value(c.value, c.validator))
			if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
				t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
			}
		})
	}
}

func TestMap(t *testing.T) {
	type Stat struct {
		Count int
	}

	cases := []struct {
		name      string
		value     interface{}
		validator v.Validator
		errs      v.Errors
	}{
		{
			name:  "nil map",
			value: map[string]Stat(nil),
			validator: v.Map(func(m map[string]Stat) map[string]v.Schema {
				return nil
			}),
			errs: nil,
		},
		{
			name: "invalid",
			value: map[string]Stat{
				"visitor": {Count: 0},
				"visit":   {Count: 0},
			},
			validator: v.Map(func(m map[string]Stat) map[string]v.Schema {
				schemas := make(map[string]v.Schema)
				for k, s := range m {
					schemas[k] = v.Schema{
						v.F("count", s.Count): v.Nonzero[int](),
					}
				}
				return schemas
			}),
			errs: v.Errors{
				v.NewError("stats[visitor].count", v.ErrInvalid, "is zero valued"),
				v.NewError("stats[visit].count", v.ErrInvalid, "is zero valued"),
			},
		},
		{
			name: "valid",
			value: map[string]Stat{
				"visitor": {Count: 1},
				"visit":   {Count: 2},
			},
			validator: v.Map(func(m map[string]Stat) map[string]v.Schema {
				schemas := make(map[string]v.Schema)
				for k, s := range m {
					schemas[k] = v.Schema{
						v.F("count", s.Count): v.Nonzero[int](),
					}
				}
				return schemas
			}),
			errs: nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := v.Validate(v.Schema{
				v.F("stats", c.value): c.validator,
			})
			if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
				t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
			}
		})
	}
}

func TestSlice(t *testing.T) {
	type Comment struct {
		Content   string
		CreatedAt time.Time
	}

	cases := []struct {
		name      string
		value     interface{}
		validator v.Validator
		errs      v.Errors
	}{
		{
			name:  "nil slice",
			value: []Comment(nil),
			validator: v.Slice(func(s []Comment) (schemas []v.Schema) {
				return nil
			}),
			errs: nil,
		},
		{
			name: "invalid",
			value: []Comment{
				{Content: "", CreatedAt: time.Time{}},
			},

			validator: v.Slice(func(s []Comment) (schemas []v.Schema) {
				for _, c := range s {
					schemas = append(schemas, v.Schema{
						v.F("content", c.Content):      v.Nonzero[string](),
						v.F("created_at", c.CreatedAt): v.Nonzero[time.Time](),
					})
				}
				return
			}),
			errs: v.Errors{
				v.NewError("comments[0].content", v.ErrInvalid, "is zero valued"),
				v.NewError("comments[0].created_at", v.ErrInvalid, "is zero valued"),
			},
		},
		{
			name:  "nil slice",
			value: []Comment(nil),
			validator: v.Slice(func(s []Comment) (schemas []v.Schema) {
				for _, c := range s {
					schemas = append(schemas, v.Schema{
						v.F("content", c.Content):      v.Nonzero[string](),
						v.F("created_at", c.CreatedAt): v.Nonzero[time.Time](),
					})
				}
				return
			}),
			errs: nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := v.Validate(v.Schema{
				v.F("comments", c.value): c.validator,
			})
			if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
				t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
			}
		})
	}
}

func TestAll(t *testing.T) {
	cases := []struct {
		schema v.Schema
		errs   v.Errors
	}{
		{
			v.Schema{
				v.F("value", ""): v.All(),
			},
			nil,
		},
		{
			v.Schema{
				v.F("value", ""): v.All(v.Nonzero[string]()),
			},
			v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			v.Schema{
				v.F("value", "a"): v.All(v.Nonzero[string](), v.LenString(2, 5)),
			},
			v.NewErrors("value", v.ErrInvalid, "has an invalid length"),
		},
		{
			v.Schema{
				v.F("value", "abc"): v.All(v.Nonzero[string](), v.LenString(2, 5), v.In("a", "ab")),
			},
			v.NewErrors("value", v.ErrInvalid, "is not one of the given values"),
		},
		{
			v.Schema{
				v.F("value", "abc"): v.All(v.Nonzero[string](), v.LenString(2, 5), v.In("a", "ab", "abc")),
			},
			nil,
		},
	}
	for _, c := range cases {
		errs := v.Validate(c.schema)
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestAny(t *testing.T) {
	cases := []struct {
		schema v.Schema
		errs   v.Errors
	}{
		{
			v.Schema{
				v.F("value", ""): v.Any(),
			},
			nil,
		},
		{
			v.Schema{
				v.F("value", ""): v.Any(v.Nonzero[string]()),
			},
			v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			v.Schema{
				v.F("value", "a"): v.Any(v.Nonzero[string](), v.LenString(2, 5)),
			},
			nil,
		},
		{
			v.Schema{
				v.F("value", "abc"): v.Any(v.LenString(1, 2), v.In("a", "ab")),
			},
			v.Errors{
				v.NewError("value", v.ErrInvalid, "has an invalid length"),
				v.NewError("value", v.ErrInvalid, "is not one of the given values"),
			},
		},
		{
			v.Schema{
				v.F("value", "abc"): v.Any(v.LenString(1, 2), v.In("a", "ab")).LastError(),
			},
			v.Errors{
				v.NewError("value", v.ErrInvalid, "is not one of the given values"),
			},
		},
	}
	for _, c := range cases {
		errs := v.Validate(c.schema)
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestNot(t *testing.T) {
	cases := []struct {
		schema v.Schema
		errs   v.Errors
	}{
		{
			v.Schema{
				v.F("value", []string{"foo"}): v.Not(v.Eq("foo")),
			},
			v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Eq` on type []string"),
		},
		{
			v.Schema{
				v.F("value", ""): v.Not(v.Nonzero[string]()),
			},
			nil,
		},
		{
			v.Schema{
				v.F("value", "a"): v.Not(v.All(v.Nonzero[string](), v.LenString(2, 5))),
			},
			nil,
		},
		{
			v.Schema{
				v.F("value", "a"): v.Not(v.Any(v.Nonzero[string](), v.LenString(2, 5))),
			},
			v.NewErrors("value", v.ErrInvalid, "is invalid"),
		},
		{
			v.Schema{
				v.F("value", "a"): v.Not(v.Any(v.Nonzero[string](), v.LenString(2, 5))).Msg("is not ok"),
			},
			v.NewErrors("value", v.ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		errs := v.Validate(c.schema)
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestIs(t *testing.T) {
	cases := []struct {
		name      string
		value     interface{}
		validator v.Validator
		errs      v.Errors
	}{
		{
			name:      "int invalid",
			value:     0,
			validator: v.Is(func(i int) bool { return i == 1 }),
			errs:      v.NewErrors("value", v.ErrInvalid, "is invalid"),
		},
		{
			name:      "int valid",
			value:     1,
			validator: v.Is(func(i int) bool { return i == 1 }),
			errs:      nil,
		},
		{
			name:      "string invalid",
			value:     "",
			validator: v.Is(func(s string) bool { return s == "a" }),
			errs:      v.NewErrors("value", v.ErrInvalid, "is invalid"),
		},
		{
			name:      "string valid",
			value:     "a",
			validator: v.Is(func(s string) bool { return s == "a" }),
			errs:      nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := v.Validate(v.Schema{
				v.F("value", c.value): c.validator,
			})
			if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
				t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
			}
		})
	}
}

func TestNonzero(t *testing.T) {
	cases := []struct {
		value     interface{}
		validator v.Validator
		errs      v.Errors
	}{
		{
			value:     0,
			validator: v.Nonzero[int](),
			errs:      v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			value:     1,
			validator: v.Nonzero[int](),
			errs:      nil,
		},
		{
			value:     false,
			validator: v.Nonzero[bool](),
			errs:      v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			value:     true,
			validator: v.Nonzero[bool](),
			errs:      nil,
		},
		{
			value:     "",
			validator: v.Nonzero[string](),
			errs:      v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			value:     "a",
			validator: v.Nonzero[string](),
			errs:      nil,
		},
		{
			value:     time.Time{},
			validator: v.Nonzero[time.Time](),
			errs:      v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			value:     time.Now(),
			validator: v.Nonzero[time.Time](),
			errs:      nil,
		},
		{
			value:     struct{ Foo int }{},
			validator: v.Nonzero[struct{ Foo int }](),
			errs:      v.NewErrors("value", v.ErrInvalid, "is zero valued"),
		},
		{
			value:     struct{ Foo int }{Foo: 1},
			validator: v.Nonzero[struct{ Foo int }](),
			errs:      nil,
		},
	}
	for _, c := range cases {
		errs := v.Validate(v.Schema{
			v.F("value", c.value): c.validator,
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestZero(t *testing.T) {
	cases := []struct {
		value     interface{}
		validator v.Validator
		errs      v.Errors
	}{
		{
			value:     0,
			validator: v.Zero[int](),
			errs:      nil,
		},
		{
			value:     1,
			validator: v.Zero[int](),
			errs:      v.NewErrors("value", v.ErrInvalid, "is nonzero"),
		},
		{
			value:     false,
			validator: v.Zero[bool](),
			errs:      nil,
		},
		{
			value:     true,
			validator: v.Zero[bool](),
			errs:      v.NewErrors("value", v.ErrInvalid, "is nonzero"),
		},
		{
			value:     "",
			validator: v.Zero[string](),
			errs:      nil,
		},
		{
			value:     "a",
			validator: v.Zero[string](),
			errs:      v.NewErrors("value", v.ErrInvalid, "is nonzero"),
		},
		{
			value:     time.Time{},
			validator: v.Zero[time.Time](),
			errs:      nil,
		},
		{
			value:     time.Now(),
			validator: v.Zero[time.Time](),
			errs:      v.NewErrors("value", v.ErrInvalid, "is nonzero"),
		},
		{
			value:     struct{ Foo int }{},
			validator: v.Zero[struct{ Foo int }](),
			errs:      nil,
		},
		{
			value:     struct{ Foo int }{Foo: 1},
			validator: v.Zero[struct{ Foo int }](),
			errs:      v.NewErrors("value", v.ErrInvalid, "is nonzero"),
		},
	}
	for _, c := range cases {
		errs := v.Validate(v.Schema{
			v.F("value", c.value): c.validator,
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestLenString(t *testing.T) {
	cases := []struct {
		value     interface{}
		validator v.Validator
		errs      v.Errors
	}{
		{
			value:     0,
			validator: v.LenString(1, 2),
			errs:      v.NewErrors("value", v.ErrUnsupported, "cannot use validator `LenString` on type int"),
		},
		{
			value:     "",
			validator: v.LenString(1, 2),
			errs:      v.NewErrors("value", v.ErrInvalid, "has an invalid length"),
		},
		{
			value:     "",
			validator: v.LenString(1, 2).Msg("bad length"),
			errs:      v.NewErrors("value", v.ErrInvalid, "bad length"),
		},
		{
			value:     "a",
			validator: v.LenString(1, 2).Msg("bad length"),
			errs:      nil,
		},
	}
	for _, c := range cases {
		errs := v.Validate(v.Schema{
			v.F("value", c.value): c.validator,
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestLenSlice(t *testing.T) {
	cases := []struct {
		value     interface{}
		validator v.Validator
		errs      v.Errors
	}{
		{
			value:     "",
			validator: v.LenSlice[[]string](1, 2),
			errs:      v.NewErrors("value", v.ErrUnsupported, "cannot use validator `LenSlice` on type string"),
		},
		{
			value:     []int(nil),
			validator: v.LenSlice[[]int](1, 2),
			errs:      v.NewErrors("value", v.ErrInvalid, "has an invalid length"),
		},
		{
			value:     []int{0},
			validator: v.LenSlice[[]int](1, 2),
			errs:      nil,
		},
		{
			value:     []string(nil),
			validator: v.LenSlice[[]string](1, 2),
			errs:      v.NewErrors("value", v.ErrInvalid, "has an invalid length"),
		},
		{
			value:     []string{""},
			validator: v.LenSlice[[]string](1, 2),
			errs:      nil,
		},
		{
			value:     []bool(nil),
			validator: v.LenSlice[[]bool](1, 2),
			errs:      v.NewErrors("value", v.ErrInvalid, "has an invalid length"),
		},
		{
			value:     []bool{true},
			validator: v.LenSlice[[]bool](1, 2),
			errs:      nil,
		},
	}
	for _, c := range cases {
		errs := v.Validate(v.Schema{
			v.F("value", c.value): c.validator,
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestRuneCount(t *testing.T) {
	cases := []struct {
		value     interface{}
		validator v.Validator
		errs      v.Errors
	}{
		{
			value:     0,
			validator: v.RuneCount(1, 2),
			errs:      v.NewErrors("value", v.ErrUnsupported, "cannot use validator `RuneCount` on type int"),
		},
		{
			value:     "",
			validator: v.RuneCount(1, 2),
			errs:      v.NewErrors("value", v.ErrInvalid, "the number of runes is not between the given range"),
		},
		{
			value:     "a",
			validator: v.RuneCount(1, 2),
			errs:      nil,
		},
		{
			value:     "你好",
			validator: v.RuneCount(1, 2),
			errs:      nil,
		},
		{
			value:     "你好吗",
			validator: v.RuneCount(1, 2).Msg("is not ok"),
			errs:      v.NewErrors("value", v.ErrInvalid, "is not ok"),
		},
	}
	for _, c := range cases {
		errs := v.Validate(v.Schema{
			v.F("value", c.value): c.validator,
		})
		if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
			t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
		}
	}
}

func TestEq_Ne_Gt_Gte_Lt_Lte(t *testing.T) {
	cases := []struct {
		name   string
		schema v.Schema
		errs   v.Errors
	}{
		// Eq
		{
			name: "Eq int err",
			schema: v.Schema{
				v.F("value", 1): v.Eq(0),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "does not equal the given value"),
		},
		{
			name: "Eq int ok",
			schema: v.Schema{
				v.F("value", 1): v.Eq(1),
			},
			errs: nil,
		},
		{
			name: "Eq bool err",
			schema: v.Schema{
				v.F("value", true): v.Eq(false),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "does not equal the given value"),
		},
		{
			name: "Eq bool ok",
			schema: v.Schema{
				v.F("value", true): v.Eq(true),
			},
			errs: nil,
		},
		{
			name: "Eq string err",
			schema: v.Schema{
				v.F("value", "a"): v.Eq(""),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "does not equal the given value"),
		},
		{
			name: "Eq string ok",
			schema: v.Schema{
				v.F("value", "a"): v.Eq("a"),
			},
			errs: nil,
		},
		{
			name: "Eq struct err",
			schema: v.Schema{
				v.F("value", struct{ Foo int }{Foo: 1}): v.Eq(struct{ Foo int }{}),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "does not equal the given value"),
		},
		{
			name: "Eq struct err",
			schema: v.Schema{
				v.F("value", struct{ Foo int }{Foo: 1}): v.Eq(struct{ Foo int }{Foo: 1}),
			},
			errs: nil,
		},
		// Ne
		{
			name: "Ne int err",
			schema: v.Schema{
				v.F("value", 1): v.Ne(1),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "equals the given value"),
		},
		{
			name: "Ne int ok",
			schema: v.Schema{
				v.F("value", 1): v.Ne(0),
			},
			errs: nil,
		},
		{
			name: "Ne bool err",
			schema: v.Schema{
				v.F("value", true): v.Ne(true),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "equals the given value"),
		},
		{
			name: "Ne bool ok",
			schema: v.Schema{
				v.F("value", true): v.Ne(false),
			},
			errs: nil,
		},
		{
			name: "Ne string err",
			schema: v.Schema{
				v.F("value", "a"): v.Ne("a"),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "equals the given value"),
		},
		{
			name: "Ne string ok",
			schema: v.Schema{
				v.F("value", "a"): v.Ne(""),
			},
			errs: nil,
		},
		{
			name: "Ne struct err",
			schema: v.Schema{
				v.F("value", struct{ Foo int }{Foo: 1}): v.Ne(struct{ Foo int }{Foo: 1}),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "equals the given value"),
		},
		{
			name: "Ne struct ok",
			schema: v.Schema{
				v.F("value", struct{ Foo int }{Foo: 1}): v.Ne(struct{ Foo int }{}),
			},
			errs: nil,
		},
		// Gt
		{
			name: "Gt int err",
			schema: v.Schema{
				v.F("value", 1): v.Gt(1),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to the given value"),
		},
		{
			name: "Gt int ok",
			schema: v.Schema{
				v.F("value", 1): v.Gt(0),
			},
			errs: nil,
		},
		{
			name: "Gt float err",
			schema: v.Schema{
				v.F("value", 1.0): v.Gt(1.0),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to the given value"),
		},
		{
			name: "Gt float ok",
			schema: v.Schema{
				v.F("value", 1.0): v.Gt(0.0),
			},
			errs: nil,
		},
		{
			name: "Gt string err",
			schema: v.Schema{
				v.F("value", "a"): v.Gt("a"),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than or equal to the given value"),
		},
		{
			name: "Gt string ok",
			schema: v.Schema{
				v.F("value", "a"): v.Gt(""),
			},
			errs: nil,
		},
		// Gte
		{
			name: "Gte int err",
			schema: v.Schema{
				v.F("value", 1): v.Gte(2),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than the given value"),
		},
		{
			name: "Gte int ok0",
			schema: v.Schema{
				v.F("value", 1): v.Gte(1),
			},
			errs: nil,
		},
		{
			name: "Gte int ok1",
			schema: v.Schema{
				v.F("value", 1): v.Gte(0),
			},
			errs: nil,
		},
		{
			name: "Gte float err",
			schema: v.Schema{
				v.F("value", 1.0): v.Gte(2.0),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than the given value"),
		},
		{
			name: "Gte float ok0",
			schema: v.Schema{
				v.F("value", 1.0): v.Gte(1.0),
			},
		},
		{
			name: "Gte float ok1",
			schema: v.Schema{
				v.F("value", 1.0): v.Gte(0.0),
			},
			errs: nil,
		},
		{
			name: "Gte string err",
			schema: v.Schema{
				v.F("value", "a"): v.Gte("ab"),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is lower than the given value"),
		},
		{
			name: "Gte string ok0",
			schema: v.Schema{
				v.F("value", "a"): v.Gte("a"),
			},
			errs: nil,
		},
		{
			name: "Gte string ok1",
			schema: v.Schema{
				v.F("value", "a"): v.Gte(""),
			},
			errs: nil,
		},
		// Lt
		{
			name: "Lt int err",
			schema: v.Schema{
				v.F("value", 1): v.Lt(1),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is greater than or equal to the given value"),
		},
		{
			name: "Lt int ok",
			schema: v.Schema{
				v.F("value", 1): v.Lt(2),
			},
			errs: nil,
		},
		{
			name: "Lt float err",
			schema: v.Schema{
				v.F("value", 1.0): v.Lt(1.0),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is greater than or equal to the given value"),
		},
		{
			name: "Lt float ok",
			schema: v.Schema{
				v.F("value", 1.0): v.Lt(2.0),
			},
			errs: nil,
		},
		{
			name: "Lt string err",
			schema: v.Schema{
				v.F("value", "a"): v.Lt("a"),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is greater than or equal to the given value"),
		},
		{
			name: "Lt string ok",
			schema: v.Schema{
				v.F("value", "a"): v.Lt("ab"),
			},
			errs: nil,
		},
		// Lte
		{
			name: "Lte int ok0",
			schema: v.Schema{
				v.F("value", 1): v.Lte(2),
			},
			errs: nil,
		},
		{
			name: "Lte int ok1",
			schema: v.Schema{
				v.F("value", 1): v.Lte(1),
			},
			errs: nil,
		},
		{
			name: "Lte int err",
			schema: v.Schema{
				v.F("value", 1): v.Lte(0),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is greater than the given value"),
		},
		{
			name: "Lte float ok0",
			schema: v.Schema{
				v.F("value", 1.0): v.Lte(2.0),
			},
			errs: nil,
		},
		{
			name: "Lte float ok1",
			schema: v.Schema{
				v.F("value", 1.0): v.Lte(1.0),
			},
		},
		{
			name: "Lte float err",
			schema: v.Schema{
				v.F("value", 1.0): v.Lte(0.0),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is greater than the given value"),
		},
		{
			name: "Lte string ok0",
			schema: v.Schema{
				v.F("value", "a"): v.Lte("ab"),
			},
			errs: nil,
		},
		{
			name: "Lte string ok1",
			schema: v.Schema{
				v.F("value", "a"): v.Lte("a"),
			},
			errs: nil,
		},
		{
			name: "Lte string err",
			schema: v.Schema{
				v.F("value", "a"): v.Lte(""),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is greater than the given value"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := v.Validate(c.schema)
			if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
				t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
			}
		})
	}
}

func TestIn_Nin(t *testing.T) {
	cases := []struct {
		name   string
		schema v.Schema
		errs   v.Errors
	}{
		// In
		{
			name: "In int err",
			schema: v.Schema{
				v.F("value", 0): v.In(1, 2),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of the given values"),
		},
		{
			name: "In int ok",
			schema: v.Schema{
				v.F("value", 1): v.In(1, 2),
			},
			errs: nil,
		},
		{
			name: "In string err",
			schema: v.Schema{
				v.F("value", ""): v.In("a", "b"),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is not one of the given values"),
		},
		{
			name: "In string ok",
			schema: v.Schema{
				v.F("value", "a"): v.In("a", "b"),
			},
			errs: nil,
		},
		// Nin
		{
			name: "Nin int ok",
			schema: v.Schema{
				v.F("value", 0): v.Nin(1, 2),
			},
			errs: nil,
		},
		{
			name: "Nin int err",
			schema: v.Schema{
				v.F("value", 1): v.Nin(1, 2),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is one of the given values"),
		},
		{
			name: "Nin string ok",
			schema: v.Schema{
				v.F("value", ""): v.Nin("a", "b"),
			},
			errs: nil,
		},
		{
			name: "Nin string err",
			schema: v.Schema{
				v.F("value", "a"): v.Nin("a", "b"),
			},
			errs: v.NewErrors("value", v.ErrInvalid, "is one of the given values"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := v.Validate(c.schema)
			if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
				t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	cases := []struct {
		name      string
		value     interface{}
		validator v.Validator
		errs      v.Errors
	}{
		{
			value:     0,
			validator: v.Match(regexp.MustCompile(``)),
			errs:      v.NewErrors("value", v.ErrUnsupported, "cannot use validator `Match` on type int"),
		},
		{
			value:     "x13012345678",
			validator: v.Match(regexp.MustCompile(`^(86)?1\d{10}$`)), // cellphone
			errs:      v.NewErrors("value", v.ErrInvalid, "does not match the given regular expression"),
		},
		{
			value:     "x13012345678",
			validator: v.Match(regexp.MustCompile(`^(86)?1\d{10}$`)).Msg("invalid cellphone"), // cellphone
			errs:      v.NewErrors("value", v.ErrInvalid, "invalid cellphone"),
		},
		{
			value:     "13012345678",
			validator: v.Match(regexp.MustCompile(`^(86)?1\d{10}$`)), // cellphone
			errs:      nil,
		},
		{
			value:     []byte("13012345678"),
			validator: v.Match(regexp.MustCompile(`^(86)?1\d{10}$`)), // cellphone
			errs:      nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := v.Validate(v.Schema{
				v.F("value", c.value): c.validator,
			})
			if !reflect.DeepEqual(makeErrsMap(errs), makeErrsMap(c.errs)) {
				t.Errorf("Got (%+v) != Want (%+v)", errs, c.errs)
			}
		})
	}
}
