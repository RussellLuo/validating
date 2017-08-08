# validating

A Go library for validating structs and fields.


## Features

1. Simple

    Simple and stupid, no magic involved.

2. Type-safe

    Schema is defined in Go, which is type-safer (and more powerful) than traditional struct tags.

3. Flexible

    - Validators are composite.
    - Nested struct validation is well supported.
    - Schema can be defined inside or outside struct.
    - Validator customizations are made easy.

4. No reflection


## Validator factories and validators

To be strict, this library has a conceptual distinction between `validator factory` and `validator`.

A validator factory is a function to create a validator, which will do the actual validation.

### Built-in validator factories

Below are the built-in validator factories:

- [FromFunc](https://godoc.org/github.com/RussellLuo/validating#FromFunc)
- [All/And](https://godoc.org/github.com/RussellLuo/validating#All)
- [Any/Or](https://godoc.org/github.com/RussellLuo/validating#All)
- [Not](https://godoc.org/github.com/RussellLuo/validating#Not)
- [Nested](https://godoc.org/github.com/RussellLuo/validating#Nested)
- [NestedMulti](https://godoc.org/github.com/RussellLuo/validating#NestedMulti)
- [Lazy](https://godoc.org/github.com/RussellLuo/validating#Lazy)
- [Assert](https://godoc.org/github.com/RussellLuo/validating#Assert)
- [Nonzero](https://godoc.org/github.com/RussellLuo/validating#Nonzero)
- [Len](https://godoc.org/github.com/RussellLuo/validating#Len)
- [Gt](https://godoc.org/github.com/RussellLuo/validating#Gt)
- [Gte](https://godoc.org/github.com/RussellLuo/validating#Gte)
- [Lt](https://godoc.org/github.com/RussellLuo/validating#Lt)
- [Lte](https://godoc.org/github.com/RussellLuo/validating#Lte)
- [Range](https://godoc.org/github.com/RussellLuo/validating#Range)
- [In](https://godoc.org/github.com/RussellLuo/validating#In)

### Validator customizations

1. From a boolean expression

    ```go
    validator := v.Assert(value != nil, "is empty")

    // do validation
    var value map[string]time.Time
    v.Validate(v.Schema{
        v.F("value", &value): validator,
    })
    ```

2. From a function

    ```go
	validator := v.FromFunc(func(field Field) Errors {
		switch t := field.ValuePtr.(type) {
		case *map[string]time.Time:
		    if *t == nil {
		        return v.NewErrors(field.Name, v.ErrInvalid, "is empty")
		    }
		    return nil
		default:
		    return v.NewErrors(field.Name, v.ErrUnsupported, "is unsupported")
		}
	})

    // do validation
    var value map[string]time.Time
    v.Validate(v.Schema{
        v.F("value", &value): validator,
    })
    ```

3. From a struct

    ```go
    type MyValidator struct{}

    func (mv *MyValidator) Validate(field Field) Errors {
		switch t := field.ValuePtr.(type) {
		case *map[string]time.Time:
		    if *t == nil {
		        return v.NewErrors(field.Name, v.ErrInvalid, "is empty")
		    }
		    return nil
		default:
		    return v.NewErrors(field.Name, v.ErrUnsupported, "is unsupported")
		}
    }

    validator := &MyValidator{}

    // do validation
    var value map[string]time.Time
    v.Validate(v.Schema{
        v.F("value", &value): validator,
    })
    ```


## Examples

- [Single field](examples/single-field/main.go)
- [Flag field](examples/flag-field/main.go)
- [Simple struct](examples/simple-struct/main.go)
- [Nested struct](examples/nested-struct/main.go)
- [Nested struct (schema inside)](examples/nested-struct-schema-inside/main.go)
- [Nested struct pointer](examples/nested-struct-pointer/main.go)
- [Nested struct slice](examples/nested-struct-slice/main.go)
- [Nested struct map](examples/nested-struct-map/main.go)


## Documentation

Check out the [Godoc][1].


## Thanks

This library borrows some ideas from the following libraries:

- [mholt/binding][2]

    Prefer no reflection.

- [alecthomas/voluptuous][3]

    Support composite-style validator factories `All`/`And`, `Any`/`Or`.

- [go-validator/validator][4]

    Use the term `nonzero` instead of `required`/`optional`.


## License

[MIT][5]


[1]: https://godoc.org/github.com/RussellLuo/validating
[2]: https://github.com/mholt/binding
[3]: https://github.com/alecthomas/voluptuous
[4]: https://github.com/go-validator/validator
[5]: http://opensource.org/licenses/MIT
