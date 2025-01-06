package toy

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/infastin/toy/parser"

	"github.com/stretchr/testify/require"
)

func expectNoError(t testing.TB, err error, msg ...any) {
	require.NoError(t, err, msg...)
}

func expectError(t testing.TB, err error, msg ...any) {
	require.Error(t, err, msg...)
}

func expectNil(t testing.TB, v any, msg ...any) {
	require.Nil(t, v, msg...)
}

func expectNotNil(t testing.TB, v any, msg ...any) {
	require.NotNil(t, v, msg...)
}

func expectTrue(t testing.TB, v bool, msg ...any) {
	require.True(t, v, msg...)
}

func expectFalse(t testing.TB, v bool, msg ...any) {
	require.False(t, v, msg...)
}

func expectSameType(t testing.TB, expected, actual any, msg ...any) {
	require.IsType(t, expected, actual, msg...)
}

func expectContains(t testing.TB, s, sub any, msg ...any) {
	require.Contains(t, s, sub, msg...)
}

func expectEqual(t testing.TB, expected, actual any, msg ...any) {
	if isNil(expected) {
		expectNil(t, actual, "expected nil, but got not nil")
		return
	}
	expectNotNil(t, actual, "expected not nil, but got nil")

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
		expectEqualCompiledFunction(t, expected, actual.(*CompiledFunction), msg...)
	case *Bytecode:
		expectEqualBytecode(t, expected, actual.(*Bytecode), msg...)
	default:
		require.Equal(t, expected, actual, msg...)
	}
}

func expectEqualObjects(t testing.TB, expected, actual []Object, msg ...any) {
	expectEqual(t, len(expected), len(actual), msg...)
	for i := 0; i < len(expected); i++ {
		expectEqual(t, expected[i], actual[i], msg...)
	}
}

func expectEqualArray(t testing.TB, expected, actual *Array, msg ...any) {
	expectEqual(t, expected.Mutable(), actual.Mutable(), msg...)
	expectEqual(t, expected.Len(), actual.Len(), msg...)
	for i := range expected.Len() {
		expectEqual(t, expected.At(i), actual.At(i), msg...)
	}
}

func expectEqualMap(t testing.TB, expected, actual *Map, msg ...any) {
	expectEqual(t, expected.Mutable(), actual.Mutable(), msg...)
	expectEqual(t, expected.Len(), actual.Len(), msg...)

	expectedItems := expected.Items()
	actualItems := actual.Items()

	for i := range expectedItems {
		expectEqual(t, expectedItems[i], actualItems[i], msg...)
	}
}

func expectEqualTuple(t testing.TB, expected, actual Tuple, msg ...any) {
	expectEqual(t, expected.Len(), actual.Len(), msg...)
	for i := range expected {
		expectEqual(t, expected[i], actual[i], msg...)
	}
}

func expectEqualError(t testing.TB, expected, actual *Error, msg ...any) {
	expectEqual(t, expected.Message(), actual.Message(), msg...)
	if expected.Cause() == nil {
		expectNil(t, actual.Cause(), "expected nil, but got not nil")
		return
	}
	expectEqualError(t, expected.Cause(), actual.Cause(), msg...)
}

func expectEqualCompiledFunction(t testing.TB, expected, actual *CompiledFunction, msg ...any) {
	expectEqual(t, expected.Instructions(), actual.Instructions(), msg...)
	expectEqual(t, expected.NumParameters(), actual.NumParameters(), msg...)
	expectEqual(t, expected.VarArgs(), actual.VarArgs(), msg...)
	expectEqual(t, expected.NumLocals(), actual.NumLocals(), msg...)
}

func expectEqualBytecode(t testing.TB, expected, actual *Bytecode, msg ...any) {
	expectEqualCompiledFunction(t, expected.MainFunction, actual.MainFunction)
	expectEqualObjects(t, expected.Constants, actual.Constants, msg...)
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

func compiledFunction(numLocals, numParams int, varArgs bool, insts ...[]byte) *CompiledFunction {
	return &CompiledFunction{
		instructions:  concatInsts(insts...),
		numLocals:     numLocals,
		numParameters: numParams,
		varArgs:       varArgs,
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
		out.IndexSet(String(key), toObject(pairs[i+1])) //nolint:errcheck
	}
	return out
}

func makeImmutableMap(pars ...any) *Map {
	return makeMap(pars...).AsImmutable().(*Map)
}

func makeArray(args ...any) *Array {
	var elems []Object
	for _, arg := range args {
		elems = append(elems, toObject(arg))
	}
	return NewArray(elems)
}

func makeImmutableArray(args ...any) *Array {
	return makeArray(args...).AsImmutable().(*Array)
}

func makeTuple(args ...any) Tuple {
	var tup Tuple
	for _, arg := range args {
		tup = append(tup, toObject(arg))
	}
	return tup
}

type ARR []any
type IARR []any
type MAP map[string]any
type IMAP map[string]any

func toObject(v any) Object {
	switch v := v.(type) {
	case Object:
		return v
	case string:
		return String(v)
	case int64:
		return Int(v)
	case int:
		return Int(v)
	case bool:
		return Bool(v)
	case rune:
		return Char(v)
	case byte: // for convenience
		return Char(v)
	case float64:
		return Float(v)
	case []byte:
		return Bytes(v)
	case ARR:
		var elems []Object
		for _, e := range v {
			elems = append(elems, toObject(e))
		}
		return NewArray(elems)
	case IARR:
		var elems []Object
		for _, e := range v {
			elems = append(elems, toObject(e))
		}
		return NewArray(elems).AsImmutable()
	case MAP:
		m := new(Map)
		for k, v := range v {
			m.IndexSet(String(k), toObject(v)) //nolint:errcheck
		}
		return m
	case IMAP:
		m := new(Map)
		for k, v := range v {
			m.IndexSet(String(k), toObject(v)) //nolint:errcheck
		}
		return m.AsImmutable()
	}
	panic(fmt.Errorf("unknown type: %T", v))
}
