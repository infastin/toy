package toy_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/infastin/toy"
	"github.com/infastin/toy/stdlib"
	"github.com/infastin/toy/token"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TestScript_Add(t *testing.T) {
	s := toy.NewScript([]byte(`a := b; c := test(b); d := test(5)`))
	s.Add("b", toy.Int(5))
	s.Add("b", toy.String("foo"))
	s.Add("test", toy.NewBuiltinFunction("test",
		func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
			if len(args) > 0 {
				switch arg := args[0].(type) {
				case toy.Int:
					return arg + 1, nil
				}
			}
			return toy.Int(0), nil
		}))
	c, err := s.Compile()
	expectNoError(t, err)
	expectNoError(t, c.Run())
	expectEqual(t, toy.String("foo"), c.Get("a").Value())
	expectEqual(t, toy.String("foo"), c.Get("b").Value())
	expectEqual(t, toy.Int(0), c.Get("c").Value())
	expectEqual(t, toy.Int(6), c.Get("d").Value())
}

func TestScript_Remove(t *testing.T) {
	s := toy.NewScript([]byte(`a := b`))
	s.Add("b", toy.Int(5))
	expectTrue(t, s.Remove("b")) // b is removed
	_, err := s.Compile()        // should not compile because b is undefined
	expectError(t, err)
}

func TestScript_Run(t *testing.T) {
	s := toy.NewScript([]byte(`a := b`))
	s.Add("b", toy.Int(5))
	c, err := s.Run()
	expectNoError(t, err)
	expectNotNil(t, c)
	compiledGet(t, c, "a", int64(5))
}

func TestScript_BuiltinModules(t *testing.T) {
	s := toy.NewScript([]byte(`math := import("math"); a := math.abs(-19.84)`))
	s.SetImports(toy.ModuleMap{"math": stdlib.MathModule})
	c, err := s.Run()
	expectNoError(t, err)
	expectNotNil(t, c)
	compiledGet(t, c, "a", 19.84)

	c, err = s.Run()
	expectNoError(t, err)
	expectNotNil(t, c)
	compiledGet(t, c, "a", 19.84)

	s.SetImports(nil)
	_, err = s.Run()
	expectError(t, err)

	s.SetImports(nil)
	_, err = s.Run()
	expectError(t, err)
}

func TestScriptConcurrency(t *testing.T) {
	solve := func(a, b, c int) (d, e int) {
		a += 2
		b += c
		a += b * 2
		d = a + b + c
		e = 0
		for i := 1; i <= d; i++ {
			e += i
		}
		e *= 2
		return
	}

	code := []byte(`
mod1 := import("mod1")

a += 2
b += c
a += 2*b

arr := [a, b, c]
arrStr := string(arr)
m := {a: a, b: b, c: c}

d := a+b+c
s := 0

for i := 1; i <= d; i++ {
	s += i
}

e := mod1.double(s)
`)

	scr := toy.NewScript(code)

	scr.Add("a", toy.Int(0))
	scr.Add("b", toy.Int(0))
	scr.Add("c", toy.Int(0))

	mods := make(toy.ModuleMap)
	mods.AddBuiltinModule("mod1", map[string]toy.Object{
		"double": toy.NewBuiltinFunction("double",
			func(_ *toy.VM, args ...toy.Object) (ret toy.Object, err error) {
				var arg int
				if err := toy.UnpackArgs(args, "arg", &arg); err != nil {
					return nil, err
				}
				return 2 * toy.Int(arg), nil
			}),
	})
	scr.SetImports(mods)

	compiled, err := scr.Compile()
	expectNoError(t, err)

	executeFn := func(compiled *toy.Compiled, a, b, c int) (d, e int) {
		err = compiled.Set("a", toy.Int(a))
		expectNoError(t, err)
		err = compiled.Set("b", toy.Int(b))
		expectNoError(t, err)
		err = compiled.Set("c", toy.Int(c))
		expectNoError(t, err)
		err := compiled.Run()
		expectNoError(t, err)
		di := compiled.Get("d").Value().(toy.Int)
		ei := compiled.Get("e").Value().(toy.Int)
		return int(di), int(ei)
	}

	const concurrency = 500
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(compiled *toy.Compiled) {
			time.Sleep(time.Duration(rand.Int63n(50)) * time.Millisecond)
			defer wg.Done()

			a := rand.Intn(10)
			b := rand.Intn(10)
			c := rand.Intn(10)

			d, e := executeFn(compiled, a, b, c)
			expectedD, expectedE := solve(a, b, c)

			expectEqual(t, expectedD, d, "input: %d, %d, %d", a, b, c)
			expectEqual(t, expectedE, e, "input: %d, %d, %d", a, b, c)
		}(compiled.Clone())
	}
	wg.Wait()
}

type Counter int64

var CounterType = toy.NewType[Counter]("counter", nil)

func (o Counter) Type() toy.ObjectType { return CounterType }
func (o Counter) String() string       { return fmt.Sprintf("Counter(%d)", int64(o)) }
func (o Counter) IsFalsy() bool        { return o == 0 }
func (o Counter) Clone() toy.Object    { return o }

func (o Counter) Compare(op token.Token, rhs toy.Object) (bool, error) {
	switch y := rhs.(type) {
	case Counter:
		return o == y, nil
	}
	return false, toy.ErrInvalidOperator
}

func (o Counter) BinaryOp(op token.Token, other toy.Object, right bool) (toy.Object, error) {
	switch rhs := other.(type) {
	case Counter:
		switch op {
		case token.Add:
			return o + rhs, nil
		case token.Sub:
			return o - rhs, nil
		}
	case toy.Int:
		switch op {
		case token.Add:
			return o + Counter(rhs), nil
		case token.Sub:
			if right {
				return Counter(rhs) - o, nil
			}
			return o - Counter(rhs), nil
		}
	}
	return nil, toy.ErrInvalidOperator
}

func (o Counter) Call(*toy.VM, ...toy.Object) (toy.Object, error) { return toy.Int(o), nil }

func TestScript_CustomObjects(t *testing.T) {
	c := compile(t, `a := c1(); s := string(c1); c2 := c1; c2++`, MAP{"c1": Counter(5)})
	compiledRun(t, c)
	compiledGet(t, c, "a", int64(5))
	compiledGet(t, c, "s", "Counter(5)")
	compiledGetCounter(t, c, "c2", Counter(6))

	c = compile(t, `
arr := [1, 2, 3, 4]
for x in arr {
	c1 += x
}
out := c1()
`, MAP{"c1": Counter(5)})
	compiledRun(t, c)
	compiledGet(t, c, "out", int64(15))

	c = compile(t, `
arr := [1, 2, 3, 4]
for x in arr {
	c1 = x - c1
}
out := c1()
`, MAP{"c1": Counter(42)})
	compiledRun(t, c)
	compiledGet(t, c, "out", int64(44))
}

func compiledGetCounter(t *testing.T, c *toy.Compiled, name string, expected Counter) {
	v := c.Get(name)
	expectNotNil(t, v)

	actual := v.Value().(Counter)
	expectNotNil(t, actual)
	expectEqual(t, expected, actual)
}

func TestScriptSourceModule(t *testing.T) {
	// script1 imports "mod1"
	scr := toy.NewScript([]byte(`out := import("mod")`))

	mods := make(toy.ModuleMap)
	mods.AddSourceModule("mod", []byte(`export 5`))
	scr.SetImports(mods)

	c, err := scr.Run()
	expectNoError(t, err)
	expectEqual(t, toy.Int(5), c.Get("out").Value())

	// executing module function
	scr = toy.NewScript([]byte(`mod := import("mod"); out := mod()`))

	mods = make(toy.ModuleMap)
	mods.AddSourceModule("mod", []byte(`a := 3; export fn() => a + 5`))
	scr.SetImports(mods)

	c, err = scr.Run()
	expectNoError(t, err)
	expectEqual(t, toy.Int(8), c.Get("out").Value())

	// source module imports builtin module
	scr = toy.NewScript([]byte(`out := import("mod")`))

	mods = make(toy.ModuleMap)
	mods.AddSourceModule("mod", []byte(`text := import("text"); export text.title("foo")`))
	mods.AddBuiltinModule("text", map[string]toy.Object{
		"title": toy.NewBuiltinFunction("title",
			func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
				var s string
				if err := toy.UnpackArgs(args, "s", &s); err != nil {
					return nil, err
				}
				caser := cases.Title(language.Und, cases.NoLower)
				return toy.String(caser.String(s)), nil
			}),
	})
	scr.SetImports(mods)

	c, err = scr.Run()
	expectNoError(t, err)
	expectEqual(t, toy.String("Foo"), c.Get("out").Value())

	scr.SetImports(nil)
	_, err = scr.Run()
	expectError(t, err)
}

func BenchmarkArrayIndex(b *testing.B) {
	bench(b, `
		a := [1, 2, 3, 4, 5, 6, 7, 8, 9];
    for i := 0; i < 1000; i++ {
      a[0]; a[1]; a[2]; a[3]; a[4]; a[5]; a[6]; a[7]; a[7];
    }
	`)
}

func BenchmarkArrayIndexCompare(b *testing.B) {
	bench(b, `
		a := [1, 2, 3, 4, 5, 6, 7, 8, 9];
    for i := 0; i < 1000; i++ {
			1; 2; 3; 4; 5; 6; 7; 8; 9;
    }
  `)
}

func bench(b *testing.B, input string) {
	b.Helper()

	s := toy.NewScript([]byte(input))
	c, err := s.Compile()
	expectNoError(b, err)

	for i := 0; i < b.N; i++ {
		err := c.Run()
		expectNoError(b, err)
	}
}

func TestCompiled_Get(t *testing.T) {
	// simple script
	c := compile(t, `a := 5`, nil)
	compiledRun(t, c)
	compiledGet(t, c, "a", int64(5))

	// user-defined variables
	compileError(t, `a := b`, nil)            // compile error because "b" is not defined
	c = compile(t, `a := b`, MAP{"b": "foo"}) // now compile with b = "foo" defined
	compiledGet(t, c, "a", nil)               // a = nil; because it's before Compiled.Run()
	compiledRun(t, c)                         // Compiled.Run()
	compiledGet(t, c, "a", "foo")             // a = "foo"
}

func TestCompiled_GetAll(t *testing.T) {
	c := compile(t, `a := 5`, nil)
	compiledRun(t, c)
	compiledGetAll(t, c, MAP{"a": int64(5)})

	c = compile(t, `a := b`, MAP{"b": "foo"})
	compiledRun(t, c)
	compiledGetAll(t, c, MAP{"a": "foo", "b": "foo"})

	c = compile(t, `a := b; b = 5`, MAP{"b": "foo"})
	compiledRun(t, c)
	compiledGetAll(t, c, MAP{"a": "foo", "b": int64(5)})
}

func TestCompiled_IsDefined(t *testing.T) {
	c := compile(t, `a := 5`, nil)
	compiledIsDefined(t, c, "a", false) // a is not defined before Run()
	compiledRun(t, c)
	compiledIsDefined(t, c, "a", true)
	compiledIsDefined(t, c, "b", false)
}

func TestCompiled_Set(t *testing.T) {
	c := compile(t, `a := b`, MAP{"b": "foo"})
	compiledRun(t, c)
	compiledGet(t, c, "a", "foo")

	// replace value of 'b'
	err := c.Set("b", toy.String("bar"))
	expectNoError(t, err)
	compiledRun(t, c)
	compiledGet(t, c, "a", "bar")

	// try to replace undefined variable
	err = c.Set("c", toy.Int(1984))
	expectError(t, err) // 'c' is not defined

	// case #2
	c = compile(t, `
a := fn() {
	return fn() {
		return b + 5
	}()
}()`,
		MAP{"b": 5},
	)
	compiledRun(t, c)
	compiledGet(t, c, "a", int64(10))

	err = c.Set("b", toy.Int(10))
	expectNoError(t, err)
	compiledRun(t, c)
	compiledGet(t, c, "a", int64(15))
}

func TestCompiled_RunContext(t *testing.T) {
	// machine completes normally
	c := compile(t, `a := 5`, nil)
	err := c.RunContext(context.Background())
	expectNoError(t, err)
	compiledGet(t, c, "a", int64(5))

	// timeout
	c = compile(t, `for true {}`, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	err = c.RunContext(ctx)
	expectEqual(t, context.DeadlineExceeded, err)
}

// CustomNumber is a user defined object that can compare to toy.Int.
type CustomNumber float64

var CustomNumberType = toy.NewType[CustomNumber]("Number", nil)

func (n CustomNumber) Type() toy.ObjectType { return CustomNumberType }
func (n CustomNumber) String() string       { return strconv.FormatFloat(float64(n), 'g', -1, 64) }
func (n CustomNumber) IsFalsy() bool        { return n == 0 }
func (n CustomNumber) Clone() toy.Object    { return n }

func (n CustomNumber) Compare(op token.Token, rhs toy.Object) (bool, error) {
	switch y := rhs.(type) {
	case toy.Int:
		switch op {
		case token.Equal:
			return n == CustomNumber(y), nil
		case token.NotEqual:
			return n != CustomNumber(y), nil
		case token.Less:
			return n < CustomNumber(y), nil
		case token.Greater:
			return n > CustomNumber(y), nil
		case token.LessEq:
			return n <= CustomNumber(y), nil
		case token.GreaterEq:
			return n >= CustomNumber(y), nil
		}
	}
	return false, toy.ErrInvalidOperator
}

func TestCompiled_CustomObject(t *testing.T) {
	c := compile(t, `r := t < 130`, MAP{"t": CustomNumber(123)})
	compiledRun(t, c)
	compiledGet(t, c, "r", true)

	c = compile(t, `r := t > 13`, MAP{"t": CustomNumber(123)})
	compiledRun(t, c)
	compiledGet(t, c, "r", true)
}

func TestScript_ImportError(t *testing.T) {
	m := `
	exp := import("expression")
	r := exp(ctx)
`

	src := `
export fn(ctx) {
	closure := fn() {
		return ctx.actiontimes < 0 // an error is thrown here because actiontimes is undefined
	}
	return closure()
}`

	s := toy.NewScript([]byte(m))

	mods := make(toy.ModuleMap)
	mods.AddSourceModule("expression", []byte(src))
	s.SetImports(mods)

	s.Add("ctx", makeMap("ctx", 12))
	_, err := s.Run()
	expectContains(t, err.Error(), "expression:4:14")
}

func compile(t *testing.T, input string, vars MAP) *toy.Compiled {
	s := toy.NewScript([]byte(input))
	for vn, vv := range vars {
		switch vv := vv.(type) {
		case string:
			s.Add(vn, toy.String(vv))
		case int:
			s.Add(vn, toy.Int(vv))
		case toy.Object:
			s.Add(vn, vv)
		}
	}

	c, err := s.Compile()
	expectNoError(t, err)
	expectNotNil(t, c)

	return c
}

func compileError(t *testing.T, input string, vars MAP) {
	s := toy.NewScript([]byte(input))
	for vn, vv := range vars {
		switch vv := vv.(type) {
		case string:
			s.Add(vn, toy.String(vv))
		case int:
			s.Add(vn, toy.Int(vv))
		case toy.Object:
			s.Add(vn, vv)
		}
	}
	_, err := s.Compile()
	expectError(t, err)
}

func compiledRun(t *testing.T, c *toy.Compiled) {
	err := c.Run()
	expectNoError(t, err)
}

func compiledGet(t *testing.T, c *toy.Compiled, name string, expected any) {
	v := c.Get(name)
	expectNotNil(t, v)
	expectEqual(t, toObject(expected), v.Value())
}

func compiledGetAll(t *testing.T, c *toy.Compiled, expected MAP) {
	vars := c.GetAll()
	expectEqual(t, len(expected), len(vars))

	for k, v := range expected {
		var found bool
		for _, e := range vars {
			if e.Name() == k {
				expectEqual(t, toObject(v), e.Value())
				found = true
			}
		}
		expectTrue(t, found, "variable '%s' not found", k)
	}
}

func compiledIsDefined(t *testing.T, c *toy.Compiled, name string, expected bool) {
	expectEqual(t, expected, c.IsDefined(name))
}

func TestCompiled_Clone(t *testing.T) {
	script := toy.NewScript([]byte(`
count += 1
data["b"] = 2
`))

	script.Add("data", makeMap("a", 1))
	script.Add("count", toy.Int(1000))

	compiled, err := script.Compile()
	expectNoError(t, err)

	clone := compiled.Clone()
	err = clone.RunContext(context.Background())
	expectNoError(t, err)

	expectEqual(t, toy.Int(1000), compiled.Get("count").Value())
	expectEqual(t, 1, compiled.Get("data").Value().(*toy.Map).Len())

	expectEqual(t, toy.Int(1001), clone.Get("count").Value())
	expectEqual(t, 2, clone.Get("data").Value().(*toy.Map).Len())
}
