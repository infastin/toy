package toy

import (
	"slices"
	"testing"

	"github.com/infastin/toy/token"
)

func TestObject_TypeName(t *testing.T) {
	var o Object = Int(0)
	expectEqual(t, "int", o.TypeName())
	o = Float(0)
	expectEqual(t, "float", o.TypeName())
	o = Char(0)
	expectEqual(t, "char", o.TypeName())
	o = String("")
	expectEqual(t, "string", o.TypeName())
	o = Bool(false)
	expectEqual(t, "bool", o.TypeName())
	o = &Array{}
	expectEqual(t, "array", o.TypeName())
	o = &Map{}
	expectEqual(t, "map", o.TypeName())
	o = &BuiltinFunction{Name: "fn"}
	expectEqual(t, "builtin-function:fn", o.TypeName())
	o = &CompiledFunction{}
	expectEqual(t, "compiled-function", o.TypeName())
	o = UndefinedType(0)
	expectEqual(t, "undefined", o.TypeName())
	o = &Error{}
	expectEqual(t, "error", o.TypeName())
	o = Bytes{}
	expectEqual(t, "bytes", o.TypeName())
	o = Tuple{}
	expectEqual(t, "tuple", o.TypeName())
}

func TestObject_IsFalsy(t *testing.T) {
	var o Object = Int(0)
	expectTrue(t, o.IsFalsy())
	o = Int(1)
	expectFalse(t, o.IsFalsy())
	o = Float(0)
	expectFalse(t, o.IsFalsy())
	o = Float(1)
	expectFalse(t, o.IsFalsy())
	o = Char(' ')
	expectFalse(t, o.IsFalsy())
	o = Char('T')
	expectFalse(t, o.IsFalsy())
	o = String("")
	expectTrue(t, o.IsFalsy())
	o = String(" ")
	expectFalse(t, o.IsFalsy())
	o = &Array{}
	expectTrue(t, o.IsFalsy())
	o = makeArray(Undefined)
	expectFalse(t, o.IsFalsy())
	o = &Map{}
	expectTrue(t, o.IsFalsy())
	o = makeMap("a", Undefined)
	expectFalse(t, o.IsFalsy())
	o = &BuiltinFunction{}
	expectFalse(t, o.IsFalsy())
	o = &CompiledFunction{}
	expectFalse(t, o.IsFalsy())
	o = UndefinedType(0)
	expectTrue(t, o.IsFalsy())
	o = &Error{}
	expectTrue(t, o.IsFalsy())
	o = Bytes{}
	expectTrue(t, o.IsFalsy())
	o = Bytes{1, 2}
	expectFalse(t, o.IsFalsy())
	o = Tuple{}
	expectTrue(t, o.IsFalsy())
	o = Tuple{Undefined}
	expectFalse(t, o.IsFalsy())
}

func TestObject_String(t *testing.T) {
	var o Object = Int(0)
	expectEqual(t, "0", o.String())
	o = Int(1)
	expectEqual(t, "1", o.String())
	o = Float(0)
	expectEqual(t, "0", o.String())
	o = Float(1)
	expectEqual(t, "1", o.String())
	o = Char(' ')
	expectEqual(t, "' '", o.String())
	o = Char('T')
	expectEqual(t, "'T'", o.String())
	o = String("")
	expectEqual(t, `""`, o.String())
	o = String(" ")
	expectEqual(t, `" "`, o.String())
	o = &Array{}
	expectEqual(t, "[]", o.String())
	o = &Map{}
	expectEqual(t, "{}", o.String())
	o = &Error{}
	expectEqual(t, "", o.String())
	o = &Error{message: "error 1"}
	expectEqual(t, "error 1", o.String())
	o = UndefinedType(0)
	expectEqual(t, "<undefined>", o.String())
	o = Bytes{}
	expectEqual(t, "[]", o.String())
	o = Bytes{'f', 'o', 'o'}
	expectEqual(t, "[102, 111, 111]", o.String())
	o = Tuple{}
	expectEqual(t, "", o.String())
}

func TestObject_BinaryOp(t *testing.T) {
	var o Object = Char(0)
	_, err := BinaryOp(token.Add, o, Undefined)
	expectError(t, err)
	o = Bool(false)
	_, err = BinaryOp(token.Add, o, Undefined)
	expectError(t, err)
	o = &Map{}
	_, err = BinaryOp(token.Add, o, Undefined)
	expectError(t, err)
	o = &BuiltinFunction{}
	_, err = BinaryOp(token.Add, o, Undefined)
	expectError(t, err)
	o = &CompiledFunction{}
	_, err = BinaryOp(token.Add, o, Undefined)
	expectError(t, err)
	o = UndefinedType(0)
	_, err = BinaryOp(token.Add, o, Undefined)
	expectError(t, err)
	o = &Error{}
	_, err = BinaryOp(token.Add, o, Undefined)
	expectError(t, err)
	o = Tuple{}
	_, err = BinaryOp(token.Add, o, Undefined)
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
	err1 := &Error{message: "some error"}
	err2 := err1

	testCompare(t, err1, token.Equal, err2, true)
	testCompare(t, err2, token.Equal, err1, true)

	err2 = &Error{message: "some error"}

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
				testBinaryOp(t, Float(l), op.tok, Float(r), Float(op.fn(l, r)))
			}
		}
	}

	// float [+,-,*,/] int
	for _, op := range ops {
		for l := float64(-2); l <= 2.1; l += 0.4 {
			for r := int64(-2); r <= 2; r++ {
				if op.tok != token.Quo || r != 0 {
					testBinaryOp(t, Float(l), op.tok, Int(r), Float(op.fn(l, float64(r))))
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
				testCompare(t, Float(l), op.tok, Float(r), op.fn(l, r))
			}
		}
	}

	// float [<,>,<=,>=,==,!=] int
	for _, op := range ops {
		for l := float64(-2); l <= 2.1; l += 0.4 {
			for r := int64(-2); r <= 2; r++ {
				testCompare(t, Float(l), op.tok, Int(r), op.fn(l, float64(r)))
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
				testBinaryOp(t, Int(l), op.tok, Int(r), Int(op.fn(l, r)))
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
			testBinaryOp(t, Int(lhs), op.tok, Int(rhs), Int(op.fn(lhs, rhs)))
		}
	}

	shTests := []int64{0, 1, 2, -1, 2, 0xffffffff}

	// int << int
	for _, lhs := range shTests {
		for s := int64(0); s < 64; s++ {
			testBinaryOp(t, Int(lhs), token.Shl, Int(s), Int(lhs<<s))
		}
	}

	// int >> int
	for _, lhs := range shTests {
		for s := int64(0); s < 64; s++ {
			testBinaryOp(t, Int(lhs), token.Shr, Int(s), Int(lhs<<s))
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
					testBinaryOp(t, Int(l), op.tok, Float(r), Float(op.fn(float64(l), r)))
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
				testCompare(t, Int(l), op.tok, Int(r), op.fn(l, r))
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
				testCompare(t, Int(l), op.tok, Float(r), op.fn(float64(l), r))
			}
		}
	}
}

func TestMap_Index(t *testing.T) {
	m := new(Map)
	err := m.IndexSet(Int(1), String("abcdef"))
	expectNoError(t, err)
	res, err := m.IndexGet(Int(1))
	expectNoError(t, err)
	expectEqual(t, String("abcdef"), res)
}

func TestString_BinaryOp(t *testing.T) {
	lstr := "abcde"
	rstr := "01234"
	for l := 0; l < len(lstr); l++ {
		for r := 0; r < len(rstr); r++ {
			ls := lstr[l:]
			rs := rstr[r:]
			testBinaryOp(t, String(ls), token.Add, String(rs), String(ls+rs))
			rc := []rune(rstr)[r]
			testBinaryOp(t, String(ls), token.Add, Char(rc), String(ls+string(rc)))
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
			testBinaryOp(t, Bytes(lb), token.Add, Bytes(rb), Bytes(slices.Concat(lb, rb)))
		}
	}
}

func testBinaryOp(t *testing.T, lhs Object, op token.Token, rhs Object, expected Object) {
	t.Helper()
	actual, err := BinaryOp(op, lhs, rhs)
	expectNoError(t, err)
	expectEqual(t, expected, actual)
}

func testCompare(t *testing.T, lhs Object, op token.Token, rhs Object, expected bool) {
	t.Helper()
	actual, err := Compare(op, lhs, rhs)
	expectNoError(t, err)
	expectEqual(t, expected, actual)
}
