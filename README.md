# validating

A Go library for validating struct and field.


## Features

1. Simple

    Simple and stupid, no magic involved.

2. Type-safe

    Schema is defined in Go, which is type-safer (and more powerful) than traditional struct tags.

3. Flexible

    - Validators are composite.
    - Nested stuct validation is well supported.
    - Schema can be defined inside or outside struct.
    - Validator customizations are made easy.

4. No reflection


## Examples

- [Single field](examples/single-field/main.go)
- [Flag field](examples/flag-field/main.go)
- [Simple struct](examples/simple-struct/main.go)
- [Nested struct](examples/nested-struct/main.go)
- [Nested struct (schema inside)](examples/nested-struct-schema-inside/main.go)
- [Nested struct pointer](examples/nested-struct-pointer/main.go)
- [Nested struct slice](examples/nested-struct-slice/main.go)
- [Nested struct map](examples/nested-struct-map/main.go)


## Thanks

This library borrows some ideas from the following libraries:

- [mholt/binding][1]

    Prefer no reflection.

- [alecthomas/voluptuous][2]

    Support composite-style validators `All`/`And`, `Any`/`Or`.

- [go-validator/validator][3]

    Use the term `nonzero` instead of `required`/`optional`.


## License

[MIT][4]


[1]: https://github.com/mholt/binding
[2]: https://github.com/alecthomas/voluptuous
[3]: https://github.com/go-validator/validator
[4]: http://opensource.org/licenses/MIT
