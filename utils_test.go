package toy

import (
	"reflect"
	"testing"

	"github.com/infastin/toy/parser"

	"github.com/stretchr/testify/require"
)

func expectNoError(t *testing.T, err error, msg ...any) {
	require.NoError(t, err, msg...)
}

func expectError(t *testing.T, err error, msg ...any) {
	require.Error(t, err, msg...)
}

func expectNil(t *testing.T, v any, msg ...any) {
	require.Nil(t, v, msg...)
}

func expectNotNil(t *testing.T, v any, msg ...any) {
	require.NotNil(t, v, msg...)
}

func expectTrue(t *testing.T, v bool, msg ...any) {
	require.True(t, v, msg...)
}

func expectFalse(t *testing.T, v bool, msg ...any) {
	require.False(t, v, msg...)
}

func expectSameType(t *testing.T, expected, actual any, msg ...any) {
	require.IsType(t, expected, actual, msg...)
}

func expectContains(t *testing.T, s, sub any, msg ...any) {
	require.Contains(t, s, sub, msg...)
}

func expectEqualBytecode(t *testing.T, expected, actual *Bytecode, msg ...any) {
	expectEqualCompiledFunction(t, expected.MainFunction, actual.MainFunction)
	expectEqualObjects(t, expected.Constants, actual.Constants, msg...)
}

func expectEqual(t *testing.T, expected, actual any, msg ...any) {
	if isNil(expected) {
		expectNil(t, actual, "expected nil, but got not nil")
		return
	}
	expectNotNil(t, actual, "expected not nil, but got nil")

	expectTrue(t, (expected != nil) == (actual != nil), msg...)
	if expected == nil && actual == nil {
		return
	}
	expectSameType(t, expected, actual, msg...)

	switch expected := expected.(type) {
	case []Object:
		expectEqualObjects(t, expected, actual.([]Object), msg...)
	case *Array:
		expectEqualArray(t, expected, actual.(*Array), msg...)
	case *Map:
		expectEqualMap(t, expected, actual.(*Map), msg...)
	case Tuple:
		expectEqualTuple(t, expected, actual.(Tuple), msg...)
	case *Error:
		expectEqualError(t, expected, actual.(*Error), msg...)
	case *CompiledFunction:
		expectEqualCompiledFunction(t, expected, actual.(*CompiledFunction))
	default:
		require.Equal(t, expected, actual, msg...)
	}
}

func expectEqualObjects(t *testing.T, expected, actual []Object, msg ...any) {
	expectEqual(t, len(expected), len(actual), msg...)
	for i := 0; i < len(expected); i++ {
		expectEqual(t, expected[i], actual[i], msg...)
	}
}

func expectEqualArray(t *testing.T, expected, actual *Array, msg ...any) {
	expectEqual(t, expected.immutable, actual.immutable, msg...)
	expectEqual(t, expected.Len(), actual.Len(), msg...)
	for i := range expected.elems {
		expectEqual(t, expected.elems[i], actual.elems[i], msg...)
	}
}

func expectEqualMap(t *testing.T, expected, actual *Map, msg ...any) {
	expectEqual(t, expected.ht.immutable, actual.ht.immutable, msg...)
	expectEqual(t, expected.Len(), actual.Len(), msg...)

	expectedItems := expected.Items()
	actualItems := actual.Items()

	for i := range expectedItems {
		expectEqual(t, expectedItems[i], actualItems[i], msg...)
	}
}

func expectEqualTuple(t *testing.T, expected, actual Tuple, msg ...any) {
	expectEqual(t, expected.Len(), actual.Len(), msg...)
	for i := range expected {
		expectEqual(t, expected[i], actual[i], msg...)
	}
}

func expectEqualError(t *testing.T, expected, actual *Error, msg ...any) {
	expectEqual(t, expected.message, actual.message, msg...)
	if expected.cause == nil {
		expectNil(t, actual.cause, "expected nil, but got not nil")
		return
	}
	expectEqualError(t, expected.cause, actual.cause, msg...)
}

func expectEqualCompiledFunction(t *testing.T, expected, actual *CompiledFunction, msg ...any) {
	expectEqual(t, expected.instructions, actual.instructions, msg...)
	expectEqual(t, expected.numParameters, actual.numParameters, msg...)
	expectEqual(t, expected.varArgs, actual.varArgs, msg...)
	expectEqual(t, expected.numLocals, actual.numLocals, msg...)
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	value := reflect.ValueOf(v)
	kind := value.Kind()
	return kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil()
}

func objectsArray(o ...Object) []Object {
	return o
}

func compiledFunction(numLocals, numParams int, insts ...[]byte) *CompiledFunction {
	return &CompiledFunction{
		instructions:  concatInsts(insts...),
		numLocals:     numLocals,
		numParameters: numParams,
	}
}

func concatInsts(instructions ...[]byte) []byte {
	var concat []byte
	for _, i := range instructions {
		concat = append(concat, i...)
	}
	return concat
}

func bytecode(
	instructions []byte,
	constants []Object,
) *Bytecode {
	return &Bytecode{
		FileSet:      parser.NewFileSet(),
		MainFunction: &CompiledFunction{instructions: instructions},
		Constants:    constants,
	}
}

func makeMap(pairs ...any) *Map {
	out := new(Map)
	for i := 0; i < len(pairs); i += 2 {
		key := pairs[i].(string)
		switch value := pairs[i+1].(type) {
		case string:
			out.IndexSet(String(key), String(value))
		case int:
			out.IndexSet(String(key), Int(value))
		case Object:
			out.IndexSet(String(key), value)
		}
	}
	return out
}

func makeImmutableMap(pars ...any) *Map {
	return makeMap(pars...).AsImmutable().(*Map)
}

func makeArray(args ...any) *Array {
	var elems []Object
	for _, arg := range args {
		switch a := arg.(type) {
		case string:
			elems = append(elems, String(a))
		case int:
			elems = append(elems, Int(a))
		case Object:
			elems = append(elems, a)
		}
	}
	return NewArray(elems)
}

func makeImmutableArray(args ...any) *Array {
	return makeArray(args...).AsImmutable().(*Array)
}

func makeTuple(args ...any) Tuple {
	var tup Tuple
	for _, arg := range args {
		switch a := arg.(type) {
		case string:
			tup = append(tup, String(a))
		case int:
			tup = append(tup, Int(a))
		case Object:
			tup = append(tup, a)
		}
	}
	return tup
}
