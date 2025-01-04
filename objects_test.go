package toy_test

import (
	"testing"

	"github.com/infastin/toy"
	"github.com/infastin/toy/token"
)

func TestObject_TypeName(t *testing.T) {
	var o toy.Object = toy.Int(0)
	require.Equal(t, "int", o.TypeName())
	o = toy.Float(0)
	require.Equal(t, "float", o.TypeName())
	o = toy.Char(0)
	require.Equal(t, "char", o.TypeName())
	o = toy.String("")
	require.Equal(t, "string", o.TypeName())
	o = toy.Bool(false)
	require.Equal(t, "bool", o.TypeName())
	o = &toy.Array{}
	require.Equal(t, "array", o.TypeName())
	o = &toy.Map{}
	require.Equal(t, "map", o.TypeName())
	o = &toy.BuiltinFunction{Name: "fn"}
	require.Equal(t, "builtin-function:fn", o.TypeName())
	o = &toy.CompiledFunction{}
	require.Equal(t, "compiled-function", o.TypeName())
	o = toy.UndefinedType(0)
	require.Equal(t, "undefined", o.TypeName())
	o = &toy.Error{}
	require.Equal(t, "error", o.TypeName())
	o = &toy.Bytes{}
	require.Equal(t, "bytes", o.TypeName())
}

func TestObject_IsFalsy(t *testing.T) {
	var o toy.Object = &toy.Int{Value: 0}
	require.True(t, o.IsFalsy())
	o = &toy.Int{Value: 1}
	require.False(t, o.IsFalsy())
	o = &toy.Float{Value: 0}
	require.False(t, o.IsFalsy())
	o = &toy.Float{Value: 1}
	require.False(t, o.IsFalsy())
	o = &toy.Char{Value: ' '}
	require.False(t, o.IsFalsy())
	o = &toy.Char{Value: 'T'}
	require.False(t, o.IsFalsy())
	o = &toy.String{Value: ""}
	require.True(t, o.IsFalsy())
	o = &toy.String{Value: " "}
	require.False(t, o.IsFalsy())
	o = &toy.Array{Value: nil}
	require.True(t, o.IsFalsy())
	o = &toy.Array{Value: []toy.Object{nil}} // nil is not valid but still count as 1 element
	require.False(t, o.IsFalsy())
	o = &toy.Map{Value: nil}
	require.True(t, o.IsFalsy())
	o = &toy.Map{Value: map[string]toy.Object{"a": nil}} // nil is not valid but still count as 1 element
	require.False(t, o.IsFalsy())
	o = &toy.StringIterator{}
	require.True(t, o.IsFalsy())
	o = &toy.ArrayIterator{}
	require.True(t, o.IsFalsy())
	o = &toy.MapIterator{}
	require.True(t, o.IsFalsy())
	o = &toy.BuiltinFunction{}
	require.False(t, o.IsFalsy())
	o = &toy.CompiledFunction{}
	require.False(t, o.IsFalsy())
	o = &toy.UndefinedType{}
	require.True(t, o.IsFalsy())
	o = &toy.Error{}
	require.True(t, o.IsFalsy())
	o = &toy.Bytes{}
	require.True(t, o.IsFalsy())
	o = &toy.Bytes{Value: []byte{1, 2}}
	require.False(t, o.IsFalsy())
}

func TestObject_String(t *testing.T) {
	var o toy.Object = &toy.Int{Value: 0}
	require.Equal(t, "0", o.String())
	o = &toy.Int{Value: 1}
	require.Equal(t, "1", o.String())
	o = &toy.Float{Value: 0}
	require.Equal(t, "0", o.String())
	o = &toy.Float{Value: 1}
	require.Equal(t, "1", o.String())
	o = &toy.Char{Value: ' '}
	require.Equal(t, " ", o.String())
	o = &toy.Char{Value: 'T'}
	require.Equal(t, "T", o.String())
	o = &toy.String{Value: ""}
	require.Equal(t, `""`, o.String())
	o = &toy.String{Value: " "}
	require.Equal(t, `" "`, o.String())
	o = &toy.Array{Value: nil}
	require.Equal(t, "[]", o.String())
	o = &toy.Map{Value: nil}
	require.Equal(t, "{}", o.String())
	o = &toy.Error{Value: nil}
	require.Equal(t, "error", o.String())
	o = &toy.Error{Value: &toy.String{Value: "error 1"}}
	require.Equal(t, `error: "error 1"`, o.String())
	o = &toy.StringIterator{}
	require.Equal(t, "<string-iterator>", o.String())
	o = &toy.ArrayIterator{}
	require.Equal(t, "<array-iterator>", o.String())
	o = &toy.MapIterator{}
	require.Equal(t, "<map-iterator>", o.String())
	o = &toy.UndefinedType{}
	require.Equal(t, "<undefined>", o.String())
	o = &toy.Bytes{}
	require.Equal(t, "", o.String())
	o = &toy.Bytes{Value: []byte("foo")}
	require.Equal(t, "foo", o.String())
}

func TestObject_BinaryOp(t *testing.T) {
	var o toy.Object = &toy.Char{}
	_, err := o.BinaryOp(token.Add, toy.Undefined)
	require.Error(t, err)
	o = &toy.Bool{}
	_, err = o.BinaryOp(token.Add, toy.Undefined)
	require.Error(t, err)
	o = &toy.Map{}
	_, err = o.BinaryOp(token.Add, toy.Undefined)
	require.Error(t, err)
	o = &toy.ArrayIterator{}
	_, err = o.BinaryOp(token.Add, toy.Undefined)
	require.Error(t, err)
	o = &toy.StringIterator{}
	_, err = o.BinaryOp(token.Add, toy.Undefined)
	require.Error(t, err)
	o = &toy.MapIterator{}
	_, err = o.BinaryOp(token.Add, toy.Undefined)
	require.Error(t, err)
	o = &toy.BuiltinFunction{}
	_, err = o.BinaryOp(token.Add, toy.Undefined)
	require.Error(t, err)
	o = &toy.CompiledFunction{}
	_, err = o.BinaryOp(token.Add, toy.Undefined)
	require.Error(t, err)
	o = &toy.UndefinedType{}
	_, err = o.BinaryOp(token.Add, toy.Undefined)
	require.Error(t, err)
	o = &toy.Error{}
	_, err = o.BinaryOp(token.Add, toy.Undefined)
	require.Error(t, err)
}

func TestArray_BinaryOp(t *testing.T) {
	testBinaryOp(t, &toy.Array{Value: nil}, token.Add,
		&toy.Array{Value: nil}, &toy.Array{Value: nil})
	testBinaryOp(t, &toy.Array{Value: nil}, token.Add,
		&toy.Array{Value: []toy.Object{}}, &toy.Array{Value: nil})
	testBinaryOp(t, &toy.Array{Value: []toy.Object{}}, token.Add,
		&toy.Array{Value: nil}, &toy.Array{Value: []toy.Object{}})
	testBinaryOp(t, &toy.Array{Value: []toy.Object{}}, token.Add,
		&toy.Array{Value: []toy.Object{}},
		&toy.Array{Value: []toy.Object{}})
	testBinaryOp(t, &toy.Array{Value: nil}, token.Add,
		&toy.Array{Value: []toy.Object{
			&toy.Int{Value: 1},
		}}, &toy.Array{Value: []toy.Object{
			&toy.Int{Value: 1},
		}})
	testBinaryOp(t, &toy.Array{Value: nil}, token.Add,
		&toy.Array{Value: []toy.Object{
			&toy.Int{Value: 1},
			&toy.Int{Value: 2},
			&toy.Int{Value: 3},
		}}, &toy.Array{Value: []toy.Object{
			&toy.Int{Value: 1},
			&toy.Int{Value: 2},
			&toy.Int{Value: 3},
		}})
	testBinaryOp(t, &toy.Array{Value: []toy.Object{
		&toy.Int{Value: 1},
		&toy.Int{Value: 2},
		&toy.Int{Value: 3},
	}}, token.Add, &toy.Array{Value: nil},
		&toy.Array{Value: []toy.Object{
			&toy.Int{Value: 1},
			&toy.Int{Value: 2},
			&toy.Int{Value: 3},
		}})
	testBinaryOp(t, &toy.Array{Value: []toy.Object{
		&toy.Int{Value: 1},
		&toy.Int{Value: 2},
		&toy.Int{Value: 3},
	}}, token.Add, &toy.Array{Value: []toy.Object{
		&toy.Int{Value: 4},
		&toy.Int{Value: 5},
		&toy.Int{Value: 6},
	}}, &toy.Array{Value: []toy.Object{
		&toy.Int{Value: 1},
		&toy.Int{Value: 2},
		&toy.Int{Value: 3},
		&toy.Int{Value: 4},
		&toy.Int{Value: 5},
		&toy.Int{Value: 6},
	}})
}

func TestError_Equals(t *testing.T) {
	err1 := &toy.Error{Value: &toy.String{Value: "some error"}}
	err2 := err1
	require.True(t, err1.Equals(err2))
	require.True(t, err2.Equals(err1))

	err2 = &toy.Error{Value: &toy.String{Value: "some error"}}
	require.False(t, err1.Equals(err2))
	require.False(t, err2.Equals(err1))
}

func TestFloat_BinaryOp(t *testing.T) {
	// float + float
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := float64(-2); r <= 2.1; r += 0.4 {
			testBinaryOp(t, &toy.Float{Value: l}, token.Add,
				&toy.Float{Value: r}, &toy.Float{Value: l + r})
		}
	}

	// float - float
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := float64(-2); r <= 2.1; r += 0.4 {
			testBinaryOp(t, &toy.Float{Value: l}, token.Sub,
				&toy.Float{Value: r}, &toy.Float{Value: l - r})
		}
	}

	// float * float
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := float64(-2); r <= 2.1; r += 0.4 {
			testBinaryOp(t, &toy.Float{Value: l}, token.Mul,
				&toy.Float{Value: r}, &toy.Float{Value: l * r})
		}
	}

	// float / float
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := float64(-2); r <= 2.1; r += 0.4 {
			if r != 0 {
				testBinaryOp(t, &toy.Float{Value: l}, token.Quo,
					&toy.Float{Value: r}, &toy.Float{Value: l / r})
			}
		}
	}

	// float < float
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := float64(-2); r <= 2.1; r += 0.4 {
			testBinaryOp(t, &toy.Float{Value: l}, token.Less,
				&toy.Float{Value: r}, boolValue(l < r))
		}
	}

	// float > float
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := float64(-2); r <= 2.1; r += 0.4 {
			testBinaryOp(t, &toy.Float{Value: l}, token.Greater,
				&toy.Float{Value: r}, boolValue(l > r))
		}
	}

	// float <= float
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := float64(-2); r <= 2.1; r += 0.4 {
			testBinaryOp(t, &toy.Float{Value: l}, token.LessEq,
				&toy.Float{Value: r}, boolValue(l <= r))
		}
	}

	// float >= float
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := float64(-2); r <= 2.1; r += 0.4 {
			testBinaryOp(t, &toy.Float{Value: l}, token.GreaterEq,
				&toy.Float{Value: r}, boolValue(l >= r))
		}
	}

	// float + int
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Float{Value: l}, token.Add,
				&toy.Int{Value: r}, &toy.Float{Value: l + float64(r)})
		}
	}

	// float - int
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Float{Value: l}, token.Sub,
				&toy.Int{Value: r}, &toy.Float{Value: l - float64(r)})
		}
	}

	// float * int
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Float{Value: l}, token.Mul,
				&toy.Int{Value: r}, &toy.Float{Value: l * float64(r)})
		}
	}

	// float / int
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := int64(-2); r <= 2; r++ {
			if r != 0 {
				testBinaryOp(t, &toy.Float{Value: l}, token.Quo,
					&toy.Int{Value: r},
					&toy.Float{Value: l / float64(r)})
			}
		}
	}

	// float < int
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Float{Value: l}, token.Less,
				&toy.Int{Value: r}, boolValue(l < float64(r)))
		}
	}

	// float > int
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Float{Value: l}, token.Greater,
				&toy.Int{Value: r}, boolValue(l > float64(r)))
		}
	}

	// float <= int
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Float{Value: l}, token.LessEq,
				&toy.Int{Value: r}, boolValue(l <= float64(r)))
		}
	}

	// float >= int
	for l := float64(-2); l <= 2.1; l += 0.4 {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Float{Value: l}, token.GreaterEq,
				&toy.Int{Value: r}, boolValue(l >= float64(r)))
		}
	}
}

func TestInt_BinaryOp(t *testing.T) {
	// int + int
	for l := int64(-2); l <= 2; l++ {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Int{Value: l}, token.Add,
				&toy.Int{Value: r}, &toy.Int{Value: l + r})
		}
	}

	// int - int
	for l := int64(-2); l <= 2; l++ {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Int{Value: l}, token.Sub,
				&toy.Int{Value: r}, &toy.Int{Value: l - r})
		}
	}

	// int * int
	for l := int64(-2); l <= 2; l++ {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Int{Value: l}, token.Mul,
				&toy.Int{Value: r}, &toy.Int{Value: l * r})
		}
	}

	// int / int
	for l := int64(-2); l <= 2; l++ {
		for r := int64(-2); r <= 2; r++ {
			if r != 0 {
				testBinaryOp(t, &toy.Int{Value: l}, token.Quo,
					&toy.Int{Value: r}, &toy.Int{Value: l / r})
			}
		}
	}

	// int % int
	for l := int64(-4); l <= 4; l++ {
		for r := -int64(-4); r <= 4; r++ {
			if r == 0 {
				testBinaryOp(t, &toy.Int{Value: l}, token.Rem,
					&toy.Int{Value: r}, &toy.Int{Value: l % r})
			}
		}
	}

	// int & int
	testBinaryOp(t,
		&toy.Int{Value: 0}, token.And, &toy.Int{Value: 0},
		&toy.Int{Value: int64(0)})
	testBinaryOp(t,
		&toy.Int{Value: 1}, token.And, &toy.Int{Value: 0},
		&toy.Int{Value: int64(1) & int64(0)})
	testBinaryOp(t,
		&toy.Int{Value: 0}, token.And, &toy.Int{Value: 1},
		&toy.Int{Value: int64(0) & int64(1)})
	testBinaryOp(t,
		&toy.Int{Value: 1}, token.And, &toy.Int{Value: 1},
		&toy.Int{Value: int64(1)})
	testBinaryOp(t,
		&toy.Int{Value: 0}, token.And, &toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(0) & int64(0xffffffff)})
	testBinaryOp(t,
		&toy.Int{Value: 1}, token.And, &toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(1) & int64(0xffffffff)})
	testBinaryOp(t,
		&toy.Int{Value: int64(0xffffffff)}, token.And,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(0xffffffff)})
	testBinaryOp(t,
		&toy.Int{Value: 1984}, token.And,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(1984) & int64(0xffffffff)})
	testBinaryOp(t, &toy.Int{Value: -1984}, token.And,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(-1984) & int64(0xffffffff)})

	// int | int
	testBinaryOp(t,
		&toy.Int{Value: 0}, token.Or, &toy.Int{Value: 0},
		&toy.Int{Value: int64(0)})
	testBinaryOp(t,
		&toy.Int{Value: 1}, token.Or, &toy.Int{Value: 0},
		&toy.Int{Value: int64(1) | int64(0)})
	testBinaryOp(t,
		&toy.Int{Value: 0}, token.Or, &toy.Int{Value: 1},
		&toy.Int{Value: int64(0) | int64(1)})
	testBinaryOp(t,
		&toy.Int{Value: 1}, token.Or, &toy.Int{Value: 1},
		&toy.Int{Value: int64(1)})
	testBinaryOp(t,
		&toy.Int{Value: 0}, token.Or, &toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(0) | int64(0xffffffff)})
	testBinaryOp(t,
		&toy.Int{Value: 1}, token.Or, &toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(1) | int64(0xffffffff)})
	testBinaryOp(t,
		&toy.Int{Value: int64(0xffffffff)}, token.Or,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(0xffffffff)})
	testBinaryOp(t,
		&toy.Int{Value: 1984}, token.Or,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(1984) | int64(0xffffffff)})
	testBinaryOp(t,
		&toy.Int{Value: -1984}, token.Or,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(-1984) | int64(0xffffffff)})

	// int ^ int
	testBinaryOp(t,
		&toy.Int{Value: 0}, token.Xor, &toy.Int{Value: 0},
		&toy.Int{Value: int64(0)})
	testBinaryOp(t,
		&toy.Int{Value: 1}, token.Xor, &toy.Int{Value: 0},
		&toy.Int{Value: int64(1) ^ int64(0)})
	testBinaryOp(t,
		&toy.Int{Value: 0}, token.Xor, &toy.Int{Value: 1},
		&toy.Int{Value: int64(0) ^ int64(1)})
	testBinaryOp(t,
		&toy.Int{Value: 1}, token.Xor, &toy.Int{Value: 1},
		&toy.Int{Value: int64(0)})
	testBinaryOp(t,
		&toy.Int{Value: 0}, token.Xor, &toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(0) ^ int64(0xffffffff)})
	testBinaryOp(t,
		&toy.Int{Value: 1}, token.Xor, &toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(1) ^ int64(0xffffffff)})
	testBinaryOp(t,
		&toy.Int{Value: int64(0xffffffff)}, token.Xor,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(0)})
	testBinaryOp(t,
		&toy.Int{Value: 1984}, token.Xor,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(1984) ^ int64(0xffffffff)})
	testBinaryOp(t,
		&toy.Int{Value: -1984}, token.Xor,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(-1984) ^ int64(0xffffffff)})

	// int &^ int
	testBinaryOp(t,
		&toy.Int{Value: 0}, token.AndNot, &toy.Int{Value: 0},
		&toy.Int{Value: int64(0)})
	testBinaryOp(t,
		&toy.Int{Value: 1}, token.AndNot, &toy.Int{Value: 0},
		&toy.Int{Value: int64(1) &^ int64(0)})
	testBinaryOp(t,
		&toy.Int{Value: 0}, token.AndNot,
		&toy.Int{Value: 1}, &toy.Int{Value: int64(0) &^ int64(1)})
	testBinaryOp(t,
		&toy.Int{Value: 1}, token.AndNot, &toy.Int{Value: 1},
		&toy.Int{Value: int64(0)})
	testBinaryOp(t,
		&toy.Int{Value: 0}, token.AndNot,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(0) &^ int64(0xffffffff)})
	testBinaryOp(t,
		&toy.Int{Value: 1}, token.AndNot,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(1) &^ int64(0xffffffff)})
	testBinaryOp(t,
		&toy.Int{Value: int64(0xffffffff)}, token.AndNot,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(0)})
	testBinaryOp(t,
		&toy.Int{Value: 1984}, token.AndNot,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(1984) &^ int64(0xffffffff)})
	testBinaryOp(t,
		&toy.Int{Value: -1984}, token.AndNot,
		&toy.Int{Value: int64(0xffffffff)},
		&toy.Int{Value: int64(-1984) &^ int64(0xffffffff)})

	// int << int
	for s := int64(0); s < 64; s++ {
		testBinaryOp(t,
			&toy.Int{Value: 0}, token.Shl, &toy.Int{Value: s},
			&toy.Int{Value: int64(0) << uint(s)})
		testBinaryOp(t,
			&toy.Int{Value: 1}, token.Shl, &toy.Int{Value: s},
			&toy.Int{Value: int64(1) << uint(s)})
		testBinaryOp(t,
			&toy.Int{Value: 2}, token.Shl, &toy.Int{Value: s},
			&toy.Int{Value: int64(2) << uint(s)})
		testBinaryOp(t,
			&toy.Int{Value: -1}, token.Shl, &toy.Int{Value: s},
			&toy.Int{Value: int64(-1) << uint(s)})
		testBinaryOp(t,
			&toy.Int{Value: -2}, token.Shl, &toy.Int{Value: s},
			&toy.Int{Value: int64(-2) << uint(s)})
		testBinaryOp(t,
			&toy.Int{Value: int64(0xffffffff)}, token.Shl,
			&toy.Int{Value: s},
			&toy.Int{Value: int64(0xffffffff) << uint(s)})
	}

	// int >> int
	for s := int64(0); s < 64; s++ {
		testBinaryOp(t,
			&toy.Int{Value: 0}, token.Shr, &toy.Int{Value: s},
			&toy.Int{Value: int64(0) >> uint(s)})
		testBinaryOp(t,
			&toy.Int{Value: 1}, token.Shr, &toy.Int{Value: s},
			&toy.Int{Value: int64(1) >> uint(s)})
		testBinaryOp(t,
			&toy.Int{Value: 2}, token.Shr, &toy.Int{Value: s},
			&toy.Int{Value: int64(2) >> uint(s)})
		testBinaryOp(t,
			&toy.Int{Value: -1}, token.Shr, &toy.Int{Value: s},
			&toy.Int{Value: int64(-1) >> uint(s)})
		testBinaryOp(t,
			&toy.Int{Value: -2}, token.Shr, &toy.Int{Value: s},
			&toy.Int{Value: int64(-2) >> uint(s)})
		testBinaryOp(t,
			&toy.Int{Value: int64(0xffffffff)}, token.Shr,
			&toy.Int{Value: s},
			&toy.Int{Value: int64(0xffffffff) >> uint(s)})
	}

	// int < int
	for l := int64(-2); l <= 2; l++ {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Int{Value: l}, token.Less,
				&toy.Int{Value: r}, boolValue(l < r))
		}
	}

	// int > int
	for l := int64(-2); l <= 2; l++ {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Int{Value: l}, token.Greater,
				&toy.Int{Value: r}, boolValue(l > r))
		}
	}

	// int <= int
	for l := int64(-2); l <= 2; l++ {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Int{Value: l}, token.LessEq,
				&toy.Int{Value: r}, boolValue(l <= r))
		}
	}

	// int >= int
	for l := int64(-2); l <= 2; l++ {
		for r := int64(-2); r <= 2; r++ {
			testBinaryOp(t, &toy.Int{Value: l}, token.GreaterEq,
				&toy.Int{Value: r}, boolValue(l >= r))
		}
	}

	// int + float
	for l := int64(-2); l <= 2; l++ {
		for r := float64(-2); r <= 2.1; r += 0.5 {
			testBinaryOp(t, &toy.Int{Value: l}, token.Add,
				&toy.Float{Value: r},
				&toy.Float{Value: float64(l) + r})
		}
	}

	// int - float
	for l := int64(-2); l <= 2; l++ {
		for r := float64(-2); r <= 2.1; r += 0.5 {
			testBinaryOp(t, &toy.Int{Value: l}, token.Sub,
				&toy.Float{Value: r},
				&toy.Float{Value: float64(l) - r})
		}
	}

	// int * float
	for l := int64(-2); l <= 2; l++ {
		for r := float64(-2); r <= 2.1; r += 0.5 {
			testBinaryOp(t, &toy.Int{Value: l}, token.Mul,
				&toy.Float{Value: r},
				&toy.Float{Value: float64(l) * r})
		}
	}

	// int / float
	for l := int64(-2); l <= 2; l++ {
		for r := float64(-2); r <= 2.1; r += 0.5 {
			if r != 0 {
				testBinaryOp(t, &toy.Int{Value: l}, token.Quo,
					&toy.Float{Value: r},
					&toy.Float{Value: float64(l) / r})
			}
		}
	}

	// int < float
	for l := int64(-2); l <= 2; l++ {
		for r := float64(-2); r <= 2.1; r += 0.5 {
			testBinaryOp(t, &toy.Int{Value: l}, token.Less,
				&toy.Float{Value: r}, boolValue(float64(l) < r))
		}
	}

	// int > float
	for l := int64(-2); l <= 2; l++ {
		for r := float64(-2); r <= 2.1; r += 0.5 {
			testBinaryOp(t, &toy.Int{Value: l}, token.Greater,
				&toy.Float{Value: r}, boolValue(float64(l) > r))
		}
	}

	// int <= float
	for l := int64(-2); l <= 2; l++ {
		for r := float64(-2); r <= 2.1; r += 0.5 {
			testBinaryOp(t, &toy.Int{Value: l}, token.LessEq,
				&toy.Float{Value: r}, boolValue(float64(l) <= r))
		}
	}

	// int >= float
	for l := int64(-2); l <= 2; l++ {
		for r := float64(-2); r <= 2.1; r += 0.5 {
			testBinaryOp(t, &toy.Int{Value: l}, token.GreaterEq,
				&toy.Float{Value: r}, boolValue(float64(l) >= r))
		}
	}
}

func TestMap_Index(t *testing.T) {
	m := &toy.Map{Value: make(map[string]toy.Object)}
	k := &toy.Int{Value: 1}
	v := &toy.String{Value: "abcdef"}
	err := m.IndexSet(k, v)

	require.NoError(t, err)

	res, err := m.IndexGet(k)
	require.NoError(t, err)
	require.Equal(t, v, res)
}

func TestString_BinaryOp(t *testing.T) {
	lstr := "abcde"
	rstr := "01234"
	for l := 0; l < len(lstr); l++ {
		for r := 0; r < len(rstr); r++ {
			ls := lstr[l:]
			rs := rstr[r:]
			testBinaryOp(t, &toy.String{Value: ls}, token.Add,
				&toy.String{Value: rs},
				&toy.String{Value: ls + rs})

			rc := []rune(rstr)[r]
			testBinaryOp(t, &toy.String{Value: ls}, token.Add,
				&toy.Char{Value: rc},
				&toy.String{Value: ls + string(rc)})
		}
	}
}

func testBinaryOp(
	t *testing.T,
	lhs toy.Object,
	op token.Token,
	rhs toy.Object,
	expected toy.Object,
) {
	t.Helper()
	actual, err := lhs.BinaryOp(op, rhs)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func boolValue(b bool) toy.Object {
	if b {
		return toy.True
	}
	return toy.False
}
