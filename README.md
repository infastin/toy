# The Toy Language

[![Go Reference](https://pkg.go.dev/badge/github.com/infastin/toy.svg)](https://pkg.go.dev/github.com/infastin/toy)

**The Toy Language is a small, dynamic, toy scripting language written in Go.**

Toy is somewhat similar to Go, but was created as a configuration language
to ditch another YAML DSL. Its codebase is based on [The Tengo Language](https://github.com/d5/tengo),
but it has been modified to make the language more comfortable to work with and add missing features and functionality.
Do not expect this language to be _blazingly fast_, it relies heavily on interfaces and reflection,
and it is even slower than Tengo.

```
fmt := import("fmt")

fib := fn(n) {
  if n == 0 {
    return 0
  } else if n == 1 {
    return 1
  }
  return fib(n-1) + fib(n-2)
}

fmt.println(fib(35)) // 9227465
```

## References

- [Language Overview](./docs/overview.md)
- [Standard Library](./docs/stdlib.md)
- [Interoperability](./docs/interoperability.md)
