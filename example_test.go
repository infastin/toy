package toy_test

import (
	"context"
	"fmt"

	"github.com/infastin/toy"
)

func Example() {
	// Toy script code
	src := `
each := fn(seq, f) {
	for x in seq {
		f(x)
	}
}
sum, mul := 0, 1
each([a, b, c, d], fn(x) {
	sum += x
	mul *= x
})`

	// create a new Script instance
	script := toy.NewScript([]byte(src))

	// set values
	script.Add("a", toy.Int(1))
	script.Add("b", toy.Int(9))
	script.Add("c", toy.Int(8))
	script.Add("d", toy.Int(4))

	// run the script
	compiled, err := script.RunContext(context.Background())
	if err != nil {
		panic(err)
	}

	// retrieve values
	sum := compiled.Get("sum")
	mul := compiled.Get("mul")
	fmt.Println(sum.Value(), mul.Value())

	// Output:
	// 22 288
}
