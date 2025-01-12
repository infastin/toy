package toy_test

import (
	"slices"
	"testing"

	"github.com/infastin/toy"
	"github.com/infastin/toy/token"
)

func TestObject_TypeName(t *testing.T) {
	var o toy.Object = toy.Int(0)
	expectEqual(t, "int", o.TypeName())
	o = toy.Float(0)
	expectEqual(t, "float", o.TypeName())
	o = toy.Char(0)
	expectEqual(t, "char", o.TypeName())
	o = toy.String("")
	expectEqual(t, "string", o.TypeName())
	o = toy.Bool(false)
	expectEqual(t, "bool", o.TypeName())
	o = &toy.Array{}
	expectEqual(t, "array", o.TypeName())
	o = &toy.Map{}
	expectEqual(t, "map", o.TypeName())
	o = &toy.BuiltinFunction{Name: "fn"}
	expectEqual(t, "builtin-function:fn", o.TypeName())
	o = &toy.CompiledFunction{}
	expectEqual(t, "compiled-function", o.TypeName())
	o = toy.NilType(0)
	expectEqual(t, "nil", o.TypeName())
	o = &toy.Error{}
	expectEqual(t, "error", o.TypeName())
	o = toy.Bytes{}
	expectEqual(t, "bytes", o.TypeName())
	o = toy.Tuple{}
	expectEqual(t, "tuple", o.TypeName())
}

func TestObject_IsFalsy(t *testing.T) {
	var o toy.Object = toy.Int(0)
	expectTrue(t, o.IsFalsy())
	o = toy.Int(1)
	expectFalse(t, o.IsFalsy())
	o = toy.Float(0)
	expectFalse(t, o.IsFalsy())
	o = toy.Float(1)
	expectFalse(t, o.IsFalsy())
	o = toy.Char(' ')
	expectFalse(t, o.IsFalsy())
	o = toy.Char('T')
	expectFalse(t, o.IsFalsy())
	o = toy.String("")
	expectTrue(t, o.IsFalsy())
	o = toy.String(" ")
	expectFalse(t, o.IsFalsy())
	o = &toy.Array{}
	expectTrue(t, o.IsFalsy())
	o = makeArray(toy.Nil)
	expectFalse(t, o.IsFalsy())
	o = &toy.Map{}
	expectTrue(t, o.IsFalsy())
	o = makeMap("a", toy.Nil)
	expectFalse(t, o.IsFalsy())
	o = &toy.BuiltinFunction{}
	expectFalse(t, o.IsFalsy())
	o = &toy.CompiledFunction{}
	expectFalse(t, o.IsFalsy())
	o = toy.NilType(0)
	expectTrue(t, o.IsFalsy())
	o = &toy.Error{}
	expectTrue(t, o.IsFalsy())
	o = toy.Bytes{}
	expectTrue(t, o.IsFalsy())
	o = toy.Bytes{1, 2}
	expectFalse(t, o.IsFalsy())
	o = toy.Tuple{}
	expectTrue(t, o.IsFalsy())
	o = toy.Tuple{toy.Nil}
	expectFalse(t, o.IsFalsy())
}

func TestObject_String(t *testing.T) {
	var o toy.Object = toy.Int(0)
	expectEqual(t, "0", o.String())
	o = toy.Int(1)
	expectEqual(t, "1", o.String())
	o = toy.Float(0)
	expectEqual(t, "0", o.String())
	o = toy.Float(1)
	expectEqual(t, "1", o.String())
	o = toy.Char(' ')
	expectEqual(t, "' '", o.String())
	o = toy.Char('T')
	expectEqual(t, "'T'", o.String())
	o = toy.String("")
	expectEqual(t, `""`, o.String())
	o = toy.String(" ")
	expectEqual(t, `" "`, o.String())
	o = &toy.Array{}
	expectEqual(t, "[]", o.String())
	o = &toy.Map{}
	expectEqual(t, "{}", o.String())
	o = &toy.Error{}
	expectEqual(t, `error("")`, o.String())
	o = toy.NewError("error 1")
	expectEqual(t, `error("error 1")`, o.String())
	o = toy.NilType(0)
	expectEqual(t, "<nil>", o.String())
	o = toy.Bytes{}
	expectEqual(t, `Bytes("")`, o.String())
	o = toy.Bytes{'f', 'o', 'o'}
	expectEqual(t, `Bytes("foo")`, o.String())
	o = toy.Tuple{}
	expectEqual(t, "tuple()", o.String())
}

func TestObject_BinaryOp(t *testing.T) {
	var o toy.Object = toy.Char(0)
	_, err := toy.BinaryOp(token.Add, o, toy.Nil)
	expectError(t, err)
	o = toy.Bool(false)
	_, err = toy.BinaryOp(token.Add, o, toy.Nil)
	expectError(t, err)
	o = &toy.Map{}
	_, err = toy.BinaryOp(token.Add, o, toy.Nil)
	expectError(t, err)
	o = &toy.BuiltinFunction{}
	_, err = toy.BinaryOp(token.Add, o, toy.Nil)
	expectError(t, err)
	o = &toy.CompiledFunction{}
	_, err = toy.BinaryOp(token.Add, o, toy.Nil)
	expectError(t, err)
	o = toy.NilType(0)
	_, err = toy.BinaryOp(token.Add, o, toy.Nil)
	expectError(t, err)
	o = &toy.Error{}
	_, err = toy.BinaryOp(token.Add, o, toy.Nil)
	expectError(t, err)
	o = toy.Tuple{}
	_, err = toy.BinaryOp(token.Add, o, toy.Nil)
	expectError(t, err)
}

func TestArray_BinaryOp(t *testing.T) {
	testBinaryOp(t, makeArray(), token.Add, makeArray(), makeArray())
	testBinaryOp(t, makeArray(), token.Add, makeArray(1), makeArray(1))
	testBinaryOp(t, makeArray(), token.Add, makeArray(1, 2, 3), makeArray(1, 2, 3))
	testBinaryOp(t, makeArray(1, 2, 3), token.Add, makeArray(), makeArray(1, 2, 3))
	testBinaryOp(t, makeArray(1, 2, 3), token.Add, makeArray(4, 5, 6), makeArray(1, 2, 3, 4, 5, 6))
}

func TestError_Compare(t *testing.T) {
	err1 := toy.NewError("some error")
	err2 := err1

	testCompare(t, err1, token.Equal, err2, true)
	testCompare(t, err2, token.Equal, err1, true)

	err2 = toy.NewError("some error")

	testCompare(t, err1, token.Equal, err2, false)
	testCompare(t, err2, token.Equal, err1, false)
}

func TestFloat_BinaryOp(t *testing.T) {
	ops := []struct {
		tok token.Token
		fn  func(lhs, rhs float64) float64
	}{
		{tok: token.Add, fn: func(lhs, rhs float64) float64 { return lhs + rhs }},
		{tok: token.Sub, fn: func(lhs, rhs float64) float64 { return lhs - rhs }},
		{tok: token.Mul, fn: func(lhs, rhs float64) float64 { return lhs * rhs }},
		{tok: token.Quo, fn: func(lhs, rhs float64) float64 { return lhs / rhs }},
	}

	// float [+,-,*,/] float
	for _, op := range ops {
		for l := float64(-2); l <= 2.1; l += 0.4 {
			for r := float64(-2); r <= 2.1; r += 0.4 {
				testBinaryOp(t, toy.Float(l), op.tok, toy.Float(r), toy.Float(op.fn(l, r)))
			}
		}
	}

	// float [+,-,*,/] int
	for _, op := range ops {
		for l := float64(-2); l <= 2.1; l += 0.4 {
			for r := int64(-2); r <= 2; r++ {
				if op.tok != token.Quo || r != 0 {
					testBinaryOp(t, toy.Float(l), op.tok, toy.Int(r), toy.Float(op.fn(l, float64(r))))
				}
			}
		}
	}
}

func TestFloat_Compare(t *testing.T) {
	ops := []struct {
		tok token.Token
		fn  func(lhs, rhs float64) bool
	}{
		{tok: token.Less, fn: func(lhs, rhs float64) bool { return lhs < rhs }},
		{tok: token.Greater, fn: func(lhs, rhs float64) bool { return lhs > rhs }},
		{tok: token.LessEq, fn: func(lhs, rhs float64) bool { return lhs <= rhs }},
		{tok: token.GreaterEq, fn: func(lhs, rhs float64) bool { return lhs >= rhs }},
		{tok: token.Equal, fn: func(lhs, rhs float64) bool { return lhs == rhs }},
		{tok: token.NotEqual, fn: func(lhs, rhs float64) bool { return lhs != rhs }},
	}

	// float [<,>,<=,>=,==,!=] float
	for _, op := range ops {
		for l := float64(-2); l <= 2.1; l += 0.4 {
			for r := float64(-2); r <= 2.1; r += 0.4 {
				testCompare(t, toy.Float(l), op.tok, toy.Float(r), op.fn(l, r))
			}
		}
	}

	// float [<,>,<=,>=,==,!=] int
	for _, op := range ops {
		for l := float64(-2); l <= 2.1; l += 0.4 {
			for r := int64(-2); r <= 2; r++ {
				testCompare(t, toy.Float(l), op.tok, toy.Int(r), op.fn(l, float64(r)))
			}
		}
	}
}

func TestInt_BinaryOp(t *testing.T) {
	iiOps := []struct {
		tok token.Token
		fn  func(lhs, rhs int64) int64
	}{
		{tok: token.Add, fn: func(lhs, rhs int64) int64 { return lhs + rhs }},
		{tok: token.Sub, fn: func(lhs, rhs int64) int64 { return lhs - rhs }},
		{tok: token.Mul, fn: func(lhs, rhs int64) int64 { return lhs * rhs }},
		{tok: token.Quo, fn: func(lhs, rhs int64) int64 { return lhs / rhs }},
	}

	// int [+,-,*,/] int
	for _, op := range iiOps {
		for l := int64(-2); l <= 2; l++ {
			for r := int64(-2); r <= 2; r++ {
				if op.tok != token.Quo || r != 0 {
					testBinaryOp(t, toy.Int(l), op.tok, toy.Int(r), toy.Int(op.fn(l, r)))
				}
			}
		}
	}

	bTests := []struct {
		lhs int64
		rhs int64
	}{
		{lhs: 0, rhs: 0},
		{lhs: 1, rhs: 0},
		{lhs: 0, rhs: 1},
		{lhs: 1, rhs: 1},
		{lhs: 0, rhs: 0xffffffff},
		{lhs: 1, rhs: 0xffffffff},
		{lhs: 0xffffffff, rhs: 0xffffffff},
		{lhs: 1984, rhs: 0xffffffff},
		{lhs: -1984, rhs: 0xffffffff},
	}

	bOps := []struct {
		tok token.Token
		fn  func(lhs, rhs int64) int64
	}{
		{tok: token.And, fn: func(lhs, rhs int64) int64 { return lhs & rhs }},
		{tok: token.Or, fn: func(lhs, rhs int64) int64 { return lhs | rhs }},
		{tok: token.Xor, fn: func(lhs, rhs int64) int64 { return lhs ^ rhs }},
		{tok: token.AndNot, fn: func(lhs, rhs int64) int64 { return lhs &^ rhs }},
	}

	// int [&,|,^,&^] int
	for _, op := range bOps {
		for _, tt := range bTests {
			lhs, rhs := tt.lhs, tt.rhs
			testBinaryOp(t, toy.Int(lhs), op.tok, toy.Int(rhs), toy.Int(op.fn(lhs, rhs)))
		}
	}

	shTests := []int64{0, 1, 2, -1, 2, 0xffffffff}

	// int << int
	for _, lhs := range shTests {
		for s := int64(0); s < 64; s++ {
			testBinaryOp(t, toy.Int(lhs), token.Shl, toy.Int(s), toy.Int(lhs<<s))
		}
	}

	// int >> int
	for _, lhs := range shTests {
		for s := int64(0); s < 64; s++ {
			testBinaryOp(t, toy.Int(lhs), token.Shr, toy.Int(s), toy.Int(lhs>>s))
		}
	}

	ifOps := []struct {
		tok token.Token
		fn  func(lhs, rhs float64) float64
	}{
		{tok: token.Add, fn: func(lhs, rhs float64) float64 { return lhs + rhs }},
		{tok: token.Sub, fn: func(lhs, rhs float64) float64 { return lhs - rhs }},
		{tok: token.Mul, fn: func(lhs, rhs float64) float64 { return lhs * rhs }},
		{tok: token.Quo, fn: func(lhs, rhs float64) float64 { return lhs / rhs }},
	}

	// int [+,-,*,/] float
	for _, op := range ifOps {
		for l := int64(-2); l <= 2; l++ {
			for r := float64(-2); r <= 2.1; r += 0.5 {
				if op.tok != token.Quo || l != 0 {
					testBinaryOp(t, toy.Int(l), op.tok, toy.Float(r), toy.Float(op.fn(float64(l), r)))
				}
			}
		}
	}
}

func TestInt_Compare(t *testing.T) {
	iiOps := []struct {
		tok token.Token
		fn  func(lhs, rhs int64) bool
	}{
		{tok: token.Less, fn: func(lhs, rhs int64) bool { return lhs < rhs }},
		{tok: token.Greater, fn: func(lhs, rhs int64) bool { return lhs > rhs }},
		{tok: token.LessEq, fn: func(lhs, rhs int64) bool { return lhs <= rhs }},
		{tok: token.GreaterEq, fn: func(lhs, rhs int64) bool { return lhs >= rhs }},
		{tok: token.Equal, fn: func(lhs, rhs int64) bool { return lhs == rhs }},
		{tok: token.NotEqual, fn: func(lhs, rhs int64) bool { return lhs != rhs }},
	}

	// int [<,>,<=,>=,==,!=] int
	for _, op := range iiOps {
		for l := int64(-2); l <= 2; l++ {
			for r := int64(-2); r <= 2; r++ {
				testCompare(t, toy.Int(l), op.tok, toy.Int(r), op.fn(l, r))
			}
		}
	}

	ifOps := []struct {
		tok token.Token
		fn  func(lhs, rhs float64) bool
	}{
		{tok: token.Less, fn: func(lhs, rhs float64) bool { return lhs < rhs }},
		{tok: token.Greater, fn: func(lhs, rhs float64) bool { return lhs > rhs }},
		{tok: token.LessEq, fn: func(lhs, rhs float64) bool { return lhs <= rhs }},
		{tok: token.GreaterEq, fn: func(lhs, rhs float64) bool { return lhs >= rhs }},
		{tok: token.Equal, fn: func(lhs, rhs float64) bool { return lhs == rhs }},
		{tok: token.NotEqual, fn: func(lhs, rhs float64) bool { return lhs != rhs }},
	}

	// int [<,>,<=,>=,==,!=] float
	for _, op := range ifOps {
		for l := int64(-2); l <= 2; l++ {
			for r := float64(-2); r <= 2.1; r += 0.5 {
				testCompare(t, toy.Int(l), op.tok, toy.Float(r), op.fn(float64(l), r))
			}
		}
	}
}

func TestMap_Index(t *testing.T) {
	m := new(toy.Map)
	err := m.IndexSet(toy.Int(1), toy.String("abcdef"))
	expectNoError(t, err)
	res, _, err := m.IndexGet(toy.Int(1))
	expectNoError(t, err)
	expectEqual(t, toy.String("abcdef"), res)
}

func TestString_BinaryOp(t *testing.T) {
	lstr := "abcde"
	rstr := "01234"
	for l := 0; l < len(lstr); l++ {
		for r := 0; r < len(rstr); r++ {
			ls := lstr[l:]
			rs := rstr[r:]
			testBinaryOp(t, toy.String(ls), token.Add, toy.String(rs), toy.String(ls+rs))
			rc := []rune(rstr)[r]
			testBinaryOp(t, toy.String(ls), token.Add, toy.Char(rc), toy.String(ls+string(rc)))
		}
	}
}

func TestBytes_BinaryOp(t *testing.T) {
	lbytes := []byte("abcde")
	rbytes := []byte("01234")
	for l := 0; l < len(lbytes); l++ {
		for r := 0; r < len(rbytes); r++ {
			lb := lbytes[l:]
			rb := rbytes[r:]
			testBinaryOp(t, toy.Bytes(lb), token.Add, toy.Bytes(rb), toy.Bytes(slices.Concat(lb, rb)))
		}
	}
}

func testBinaryOp(t *testing.T, lhs toy.Object, op token.Token, rhs toy.Object, expected toy.Object) {
	t.Helper()
	actual, err := toy.BinaryOp(op, lhs, rhs)
	expectNoError(t, err)
	expectEqual(t, expected, actual)
}

func testCompare(t *testing.T, lhs toy.Object, op token.Token, rhs toy.Object, expected bool) {
	t.Helper()
	actual, err := toy.Compare(op, lhs, rhs)
	expectNoError(t, err)
	expectEqual(t, expected, actual)
}
