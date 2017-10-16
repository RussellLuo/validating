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


## Installation


```bash
$ go get github.com/RussellLuo/validating
```


## Validator factories and validators

To be strict, this library has a conceptual distinction between `validator factory` and `validator`.

A validator factory is a function to create a validator, which will do the actual validation.

### Built-in validator factories

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

- [From a boolean expression](example_nested_struct_pointer_test.go#L24)
- [From a function](example_customizations_test.go#L32-L34)
- [From a struct](example_customizations_test.go#L22-L26)


## Examples

- [Single field](example_single_field_test.go)
- [Flag field](example_flag_field_test.go)
- [Simple struct](example_simple_struct_test.go)
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
