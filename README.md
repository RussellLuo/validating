# validating

A Go library for validating structs, maps and slices.


## Features

1. Simple

    Simple and stupid, no magic involved.

2. Type-safe

    Schema is defined in Go, which is type-safer (and more powerful) than traditional struct tags.

3. Flexible

    - Validators are composable.
    - Nested struct validation is well supported.
    - Schema can be defined inside or outside struct.
    - Validator customizations are made easy.

4. No reflection


## Installation


```bash
$ go get github.com/RussellLuo/validating/v3@latest
```


## Quick Start

```go
package main

import (
	"fmt"

	v "github.com/RussellLuo/validating/v3"
)

type Address struct {
	Country string
	City    string
}

func (a Address) Schema() v.Schema {
	return v.Schema{
		v.F("country", a.Country): v.Nonzero[string]().Msg("empty country"),
		v.F("city", a.City):       v.In("A", "B", "C").Msg("must be A or B or C"),
	}
}

type Person struct {
	Name    string
	Age     int
	Hobbies []string
	Address Address
}

func (p Person) Schema() v.Schema {
	return v.Schema{
		v.F("name", p.Name): v.All(
			v.LenString(5, 10).Msg("bad name length"),
			v.Match(`\w+`).Msg("bad name pattern"),
		),
		v.F("age", p.Age):         v.Gte(10).Msg("must be at least 10 years old"),
		v.F("hobbies", p.Hobbies): v.EachSlice[[]string](v.In("Music", "Sports").Msg("unknown hobby")),
		v.F("address", p.Address): p.Address.Schema(),
	}
}

func main() {
	p := Person{
		Name:    "Foo",
		Age:     5,
		Hobbies: []string{"Nothing"},
		Address: Address{City: "D"},
	}
	errs := v.Validate(p.Schema())
	for _, err := range errs {
		fmt.Println(err)
	}
}
```

```bash
$ go run main.go
name: INVALID(bad name length)
age: INVALID(must be at least 10 years old)
hobbies[0]: INVALID(unknown hobby)
address.country: INVALID(empty country)
address.city: INVALID(must be A or B or C)
```


## Validator factories and validators

To be strict, this library has a conceptual distinction between `validator factory` and `validator`.

A validator factory is a function used to create a validator, which will do the actual validation.

### Built-in validator factories

- [Func](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Func)
- [Schema](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Schema)
- [Value](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Value)
- [Nested](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Nested)
- [EachMap](https://pkg.go.dev/github.com/RussellLuo/validating/v3#EachMap)
- [EachSlice](https://pkg.go.dev/github.com/RussellLuo/validating/v3#EachSlice)
- [Map](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Map)
- [Slice/Array](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Slice)
- [All/And](https://pkg.go.dev/github.com/RussellLuo/validating/v3#All)
- [Any/Or](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Any)
- [Not](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Not)
- [Is](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Is)
- [Nonzero](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Nonzero)
- [Zero](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Zero)
- [ZeroOr](https://pkg.go.dev/github.com/RussellLuo/validating/v3#ZeroOr)
- [LenString](https://pkg.go.dev/github.com/RussellLuo/validating/v3#LenString)
- [LenSlice](https://pkg.go.dev/github.com/RussellLuo/validating/v3#LenSlice)
- [RuneCount](https://pkg.go.dev/github.com/RussellLuo/validating/v3#RuneCount)
- [Eq](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Eq)
- [Ne](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Ne)
- [Gt](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Gt)
- [Gte](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Gte)
- [Lt](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Lt)
- [Lte](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Lte)
- [Range](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Range)
- [In](https://pkg.go.dev/github.com/RussellLuo/validating/v3#In)
- [Nin](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Nin)
- [Match](https://pkg.go.dev/github.com/RussellLuo/validating/v3#Match)

### Extension validator factories

- [vext](https://github.com/RussellLuo/vext)

### Validator customizations

- [From a boolean function](example_simple_string_isip_test.go#L16)
- [From a normal function](example_customizations_test.go#L30-L32)
- [From a struct](example_customizations_test.go#L35-L37)


## Examples

- [Simple value](example_simple_value_test.go)
- [Simple pointer](example_simple_pointer_test.go)
- [Simple string (is IP?)](example_simple_string_isip_test.go)
- [Simple struct](example_simple_struct_test.go)
- [Simple slice](example_simple_slice_test.go)
- [Simple map](example_simple_map_test.go)
- [Nested struct](example_nested_struct_test.go)
- [Nested struct (schema inside)](example_nested_struct_schema_inside_test.go)
- [Nested struct pointer](example_nested_struct_pointer_test.go)
- [Nested struct slice](example_nested_struct_slice_test.go)
- [Nested struct map](example_nested_struct_map_test.go)


## Documentation

Check out the [Godoc][1].


## Thanks

This library borrows some ideas from the following libraries:

- [mholt/binding][2]

    Prefer no reflection.

- [alecthomas/voluptuous][3]

    Support composite validator factories `All`/`And`, `Any`/`Or`.

- [go-validator/validator][4]

    Use the term `nonzero` instead of `required`/`optional`.


## License

[MIT][5]


[1]: https://pkg.go.dev/github.com/RussellLuo/validating/v3
[2]: https://github.com/mholt/binding
[3]: https://github.com/alecthomas/voluptuous
[4]: https://github.com/go-validator/validator
[5]: http://opensource.org/licenses/MIT
