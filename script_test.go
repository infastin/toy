package toy_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/infastin/toy"
	"github.com/infastin/toy/token"

	"github.com/stretchr/testify/require"
)

func TestScript_Add(t *testing.T) {
	s := toy.NewScript([]byte(`a := b; c := test(b); d := test(5)`))
	s.Add("b", toy.Int(5))
	s.Add("b", toy.String("foo"))
	s.Add("test", &toy.BuiltinFunction{
		Name: "test",
		Func: func(args ...toy.Object) (ret toy.Object, err error) {
			if len(args) > 0 {
				switch arg := args[0].(type) {
				case toy.Int:
					return arg + 1, nil
				}
			}
			return toy.Int(0), nil
		},
	})
	c, err := s.Compile()
	require.NoError(t, err)
	require.NoError(t, c.Run())
	require.Equal(t, "foo", c.Get("a").Value())
	require.Equal(t, "foo", c.Get("b").Value())
	require.Equal(t, int64(0), c.Get("c").Value())
	require.Equal(t, int64(6), c.Get("d").Value())
}

func TestScript_Remove(t *testing.T) {
	s := toy.NewScript([]byte(`a := b`))
	s.Add("b", toy.Int(5))
	require.True(t, s.Remove("b")) // b is removed
	_, err := s.Compile()          // should not compile because b is undefined
	require.Error(t, err)
}

func TestScript_Run(t *testing.T) {
	s := toy.NewScript([]byte(`a := b`))
	s.Add("b", toy.Int(5))
	c, err := s.Run()
	require.NoError(t, err)
	require.NotNil(t, c)
	compiledGet(t, c, "a", int64(5))
}

// func TestScript_BuiltinModules(t *testing.T) {
// 	s := toy.NewScript([]byte(`math := import("math"); a := math.abs(-19.84)`))
// 	s.SetImports(stdlib.GetModuleMap("math"))
// 	c, err := s.Run()
// 	require.NoError(t, err)
// 	require.NotNil(t, c)
// 	compiledGet(t, c, "a", 19.84)
//
// 	c, err = s.Run()
// 	require.NoError(t, err)
// 	require.NotNil(t, c)
// 	compiledGet(t, c, "a", 19.84)
//
// 	s.SetImports(stdlib.GetModuleMap("os"))
// 	_, err = s.Run()
// 	require.Error(t, err)
//
// 	s.SetImports(nil)
// 	_, err = s.Run()
// 	require.Error(t, err)
// }
//
// func TestScript_SourceModules(t *testing.T) {
// 	s := toy.NewScript([]byte(`
// enum := import("enum")
// a := enum.all([1,2,3], func(_, v) {
// 	return v > 0
// })
// `))
// 	s.SetImports(stdlib.GetModuleMap("enum"))
// 	c, err := s.Run()
// 	require.NoError(t, err)
// 	require.NotNil(t, c)
// 	compiledGet(t, c, "a", true)
//
// 	s.SetImports(nil)
// 	_, err = s.Run()
// 	require.Error(t, err)
// }

func TestScript_SetMaxConstObjects(t *testing.T) {
	// one constant '5'
	s := toy.NewScript([]byte(`a := 5`))
	s.SetMaxConstObjects(1) // limit = 1
	_, err := s.Compile()
	require.NoError(t, err)
	s.SetMaxConstObjects(0) // limit = 0
	_, err = s.Compile()
	require.Error(t, err)
	require.Equal(t, "exceeding constant objects limit: 1", err.Error())

	// two constants '5' and '1'
	s = toy.NewScript([]byte(`a := 5 + 1`))
	s.SetMaxConstObjects(2) // limit = 2
	_, err = s.Compile()
	require.NoError(t, err)
	s.SetMaxConstObjects(1) // limit = 1
	_, err = s.Compile()
	require.Error(t, err)
	require.Equal(t, "exceeding constant objects limit: 2", err.Error())

	// duplicates will be removed
	s = toy.NewScript([]byte(`a := 5 + 5`))
	s.SetMaxConstObjects(1) // limit = 1
	_, err = s.Compile()
	require.NoError(t, err)
	s.SetMaxConstObjects(0) // limit = 0
	_, err = s.Compile()
	require.Error(t, err)
	require.Equal(t, "exceeding constant objects limit: 1", err.Error())

	// no limit set
	s = toy.NewScript([]byte(`a := 1 + 2 + 3 + 4 + 5`))
	_, err = s.Compile()
	require.NoError(t, err)
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
a += b * 2

arr := [a, b, c]
arrstr := string(arr)
map := {a: a, b: b, c: c}

d := a + b + c
s := 0

for i:=1; i<=d; i++ {
	s += i
}

e := mod1.double(s)
`)
	mod1 := map[string]toy.Object{
		"double": &toy.UserFunction{
			Func: func(args ...toy.Object) (
				ret toy.Object,
				err error,
			) {
				arg0, _ := toy.ToInt64(args[0])
				ret = &toy.Int{Value: arg0 * 2}
				return
			},
		},
	}

	scr := toy.NewScript(code)
	_ = scr.Add("a", 0)
	_ = scr.Add("b", 0)
	_ = scr.Add("c", 0)
	mods := toy.NewModuleMap()
	mods.AddBuiltinModule("mod1", mod1)
	scr.SetImports(mods)
	compiled, err := scr.Compile()
	require.NoError(t, err)

	executeFn := func(compiled *toy.Compiled, a, b, c int) (d, e int) {
		_ = compiled.Set("a", a)
		_ = compiled.Set("b", b)
		_ = compiled.Set("c", c)
		err := compiled.Run()
		require.NoError(t, err)
		d = compiled.Get("d").Int()
		e = compiled.Get("e").Int()
		return
	}

	concurrency := 500
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

			require.Equal(t, expectedD, d, "input: %d, %d, %d", a, b, c)
			require.Equal(t, expectedE, e, "input: %d, %d, %d", a, b, c)
		}(compiled.Clone())
	}
	wg.Wait()
}

type Counter int64

func (o Counter) TypeName() string { return "counter" }
func (o Counter) String() string   { return fmt.Sprintf("Counter(%d)", int64(o)) }
func (o Counter) IsFalsy() bool    { return o == 0 }
func (o Counter) Copy() toy.Object { return o }

func (o Counter) Compare(op token.Token, rhs toy.Object) (bool, error) {
	switch y := rhs.(type) {
	case Counter:
		return o == y, nil
	}
	return false, toy.ErrInvalidOperator
}

func (o Counter) BinaryOp(op token.Token, rhs toy.Object) (toy.Object, error) {
	switch rhs := rhs.(type) {
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
			return o - Counter(rhs), nil
		}
	}
	return nil, toy.ErrInvalidOperator
}

func (o Counter) Call(_ ...toy.Object) (toy.Object, error) { return toy.Int(o), nil }

func TestScript_CustomObjects(t *testing.T) {
	c := compile(t, `a := c1(); s := string(c1); c2 := c1; c2++`, M{
		"c1": &Counter{value: 5},
	})
	compiledRun(t, c)
	compiledGet(t, c, "a", int64(5))
	compiledGet(t, c, "s", "Counter(5)")
	compiledGetCounter(t, c, "c2", &Counter{value: 6})

	c = compile(t, `
arr := [1, 2, 3, 4]
for x in arr {
	c1 += x
}
out := c1()
`, M{
		"c1": &Counter{value: 5},
	})
	compiledRun(t, c)
	compiledGet(t, c, "out", int64(15))
}

func compiledGetCounter(
	t *testing.T,
	c *toy.Compiled,
	name string,
	expected *Counter,
) {
	v := c.Get(name)
	require.NotNil(t, v)

	actual := v.Value().(*Counter)
	require.NotNil(t, actual)
	require.Equal(t, expected.value, actual.value)
}

func TestScriptSourceModule(t *testing.T) {
	// script1 imports "mod1"
	scr := toy.NewScript([]byte(`out := import("mod")`))
	mods := toy.NewModuleMap()
	mods.AddSourceModule("mod", []byte(`export 5`))
	scr.SetImports(mods)
	c, err := scr.Run()
	require.NoError(t, err)
	require.Equal(t, int64(5), c.Get("out").Value())

	// executing module function
	scr = toy.NewScript([]byte(`fn := import("mod"); out := fn()`))
	mods = toy.NewModuleMap()
	mods.AddSourceModule("mod",
		[]byte(`a := 3; export func() { return a + 5 }`))
	scr.SetImports(mods)
	c, err = scr.Run()
	require.NoError(t, err)
	require.Equal(t, int64(8), c.Get("out").Value())

	scr = toy.NewScript([]byte(`out := import("mod")`))
	mods = toy.NewModuleMap()
	mods.AddSourceModule("mod",
		[]byte(`text := import("text"); export text.title("foo")`))
	mods.AddBuiltinModule("text",
		map[string]toy.Object{
			"title": &toy.UserFunction{
				Name: "title",
				Func: func(args ...toy.Object) (toy.Object, error) {
					s, _ := toy.ToString(args[0])
					return &toy.String{Value: strings.Title(s)}, nil
				}},
		})
	scr.SetImports(mods)
	c, err = scr.Run()
	require.NoError(t, err)
	require.Equal(t, "Foo", c.Get("out").Value())
	scr.SetImports(nil)
	_, err = scr.Run()
	require.Error(t, err)
}

func BenchmarkArrayIndex(b *testing.B) {
	bench(b.N, `a := [1, 2, 3, 4, 5, 6, 7, 8, 9];
        for i := 0; i < 1000; i++ {
            a[0]; a[1]; a[2]; a[3]; a[4]; a[5]; a[6]; a[7]; a[7];
        }
    `)
}

func BenchmarkArrayIndexCompare(b *testing.B) {
	bench(b.N, `a := [1, 2, 3, 4, 5, 6, 7, 8, 9];
        for i := 0; i < 1000; i++ {
            1; 2; 3; 4; 5; 6; 7; 8; 9;
        }
    `)
}

func bench(n int, input string) {
	s := toy.NewScript([]byte(input))
	c, err := s.Compile()
	if err != nil {
		panic(err)
	}

	for i := 0; i < n; i++ {
		if err := c.Run(); err != nil {
			panic(err)
		}
	}
}

type M map[string]any

func TestCompiled_Get(t *testing.T) {
	// simple script
	c := compile(t, `a := 5`, nil)
	compiledRun(t, c)
	compiledGet(t, c, "a", int64(5))

	// user-defined variables
	compileError(t, `a := b`, nil)          // compile error because "b" is not defined
	c = compile(t, `a := b`, M{"b": "foo"}) // now compile with b = "foo" defined
	compiledGet(t, c, "a", nil)             // a = undefined; because it's before Compiled.Run()
	compiledRun(t, c)                       // Compiled.Run()
	compiledGet(t, c, "a", "foo")           // a = "foo"
}

func TestCompiled_GetAll(t *testing.T) {
	c := compile(t, `a := 5`, nil)
	compiledRun(t, c)
	compiledGetAll(t, c, M{"a": int64(5)})

	c = compile(t, `a := b`, M{"b": "foo"})
	compiledRun(t, c)
	compiledGetAll(t, c, M{"a": "foo", "b": "foo"})

	c = compile(t, `a := b; b = 5`, M{"b": "foo"})
	compiledRun(t, c)
	compiledGetAll(t, c, M{"a": "foo", "b": int64(5)})
}

func TestCompiled_IsDefined(t *testing.T) {
	c := compile(t, `a := 5`, nil)
	compiledIsDefined(t, c, "a", false) // a is not defined before Run()
	compiledRun(t, c)
	compiledIsDefined(t, c, "a", true)
	compiledIsDefined(t, c, "b", false)
}

func TestCompiled_Set(t *testing.T) {
	c := compile(t, `a := b`, M{"b": "foo"})
	compiledRun(t, c)
	compiledGet(t, c, "a", "foo")

	// replace value of 'b'
	err := c.Set("b", "bar")
	require.NoError(t, err)
	compiledRun(t, c)
	compiledGet(t, c, "a", "bar")

	// try to replace undefined variable
	err = c.Set("c", 1984)
	require.Error(t, err) // 'c' is not defined

	// case #2
	c = compile(t, `
a := func() {
	return func() {
		return b + 5
	}()
}()`, M{"b": 5})
	compiledRun(t, c)
	compiledGet(t, c, "a", int64(10))
	err = c.Set("b", 10)
	require.NoError(t, err)
	compiledRun(t, c)
	compiledGet(t, c, "a", int64(15))
}

func TestCompiled_RunContext(t *testing.T) {
	// machine completes normally
	c := compile(t, `a := 5`, nil)
	err := c.RunContext(context.Background())
	require.NoError(t, err)
	compiledGet(t, c, "a", int64(5))

	// timeout
	c = compile(t, `for true {}`, nil)
	ctx, cancel := context.WithTimeout(context.Background(),
		1*time.Millisecond)
	defer cancel()
	err = c.RunContext(ctx)
	require.Equal(t, context.DeadlineExceeded, err)
}

func TestCompiled_CustomObject(t *testing.T) {
	c := compile(t, `r := (t<130)`, M{"t": &customNumber{value: 123}})
	compiledRun(t, c)
	compiledGet(t, c, "r", true)

	c = compile(t, `r := (t>13)`, M{"t": &customNumber{value: 123}})
	compiledRun(t, c)
	compiledGet(t, c, "r", true)
}

// customNumber is a user defined object that can compare to toy.Int
// very shitty implementation, just to test that token.Less and token.Greater in BinaryOp works
type customNumber struct {
	toy.ObjectImpl
	value int64
}

func (n *customNumber) TypeName() string {
	return "Number"
}

func (n *customNumber) String() string {
	return strconv.FormatInt(n.value, 10)
}

func (n *customNumber) BinaryOp(op token.Token, rhs toy.Object) (toy.Object, error) {
	toy.nt, ok := rhs.(*toy.Int)
	if !ok {
		return nil, toy.ErrInvalidOperator
	}
	return n.binaryOpInt(op, toy.nt)
}

func (n *customNumber) binaryOpInt(op token.Token, rhs *toy.Int) (toy.Object, error) {
	i := n.value

	switch op {
	case token.Less:
		if i < rhs.Value {
			return toy.True, nil
		}
		return toy.False, nil
	case token.Greater:
		if i > rhs.Value {
			return toy.True, nil
		}
		return toy.False, nil
	case token.LessEq:
		if i <= rhs.Value {
			return toy.True, nil
		}
		return toy.False, nil
	case token.GreaterEq:
		if i >= rhs.Value {
			return toy.True, nil
		}
		return toy.False, nil
	}
	return nil, toy.ErrInvalidOperator
}

func TestScript_ImportError(t *testing.T) {
	m := `
	exp := import("expression")
	r := exp(ctx)
`

	src := `
export func(ctx) {
	closure := func() {
		if ctx.actiontimes < 0 { // an error is thrown here because actiontimes is undefined
			return true
		}
		return false
	}

	return closure()
}`

	s := toy.NewScript([]byte(m))
	mods := toy.NewModuleMap()
	mods.AddSourceModule("expression", []byte(src))
	s.SetImports(mods)

	err := s.Add("ctx", map[string]any{
		"ctx": 12,
	})
	require.NoError(t, err)

	_, err = s.Run()
	require.True(t, strings.Contains(err.Error(), "expression:4:6"))
}

func compile(t *testing.T, input string, vars M) *toy.Compiled {
	s := toy.NewScript([]byte(input))
	for vn, vv := range vars {
		err := s.Add(vn, vv)
		require.NoError(t, err)
	}

	c, err := s.Compile()
	require.NoError(t, err)
	require.NotNil(t, c)
	return c
}

func compileError(t *testing.T, input string, vars M) {
	s := toy.NewScript([]byte(input))
	for vn, vv := range vars {
		err := s.Add(vn, vv)
		require.NoError(t, err)
	}
	_, err := s.Compile()
	require.Error(t, err)
}

func compiledRun(t *testing.T, c *toy.Compiled) {
	err := c.Run()
	require.NoError(t, err)
}

func compiledGet(t *testing.T, c *toy.Compiled, name string, expected any) {
	v := c.Get(name)
	require.NotNil(t, v)
	require.Equal(t, expected, v.Value())
}

func compiledGetAll(t *testing.T, c *toy.Compiled, expected M) {
	vars := c.GetAll()
	require.Equal(t, len(expected), len(vars))

	for k, v := range expected {
		var found bool
		for _, e := range vars {
			if e.Name() == k {
				require.Equal(t, v, e.Value())
				found = true
			}
		}
		require.True(t, found, "variable '%s' not found", k)
	}
}

func compiledIsDefined(
	t *testing.T,
	c *toy.Compiled,
	name string,
	expected bool,
) {
	require.Equal(t, expected, c.IsDefined(name))
}
func TestCompiled_Clone(t *testing.T) {
	script := toy.NewScript([]byte(`
count += 1
data["b"] = 2
`))

	err := script.Add("data", map[string]any{"a": 1})
	require.NoError(t, err)

	err = script.Add("count", 1000)
	require.NoError(t, err)

	compiled, err := script.Compile()
	require.NoError(t, err)

	clone := compiled.Clone()
	err = clone.RunContext(context.Background())
	require.NoError(t, err)

	require.Equal(t, 1000, compiled.Get("count").Int())
	require.Equal(t, 1, len(compiled.Get("data").Map()))

	require.Equal(t, 1001, clone.Get("count").Int())
	require.Equal(t, 2, len(clone.Get("data").Map()))
}
