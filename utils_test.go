package toy_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/infastin/toy"

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
	case []toy.Object:
		expectEqualObjects(t, expected, actual.([]toy.Object), msg...)
	case *toy.Array:
		expectEqualArray(t, expected, actual.(*toy.Array), msg...)
	case *toy.Map:
		expectEqualMap(t, expected, actual.(*toy.Map), msg...)
	case toy.Tuple:
		expectEqualTuple(t, expected, actual.(toy.Tuple), msg...)
	case *toy.Error:
		expectEqualError(t, expected, actual.(*toy.Error), msg...)
	case *toy.CompiledFunction:
		expectEqualCompiledFunction(t, expected, actual.(*toy.CompiledFunction), msg...)
	case *toy.Bytecode:
		expectEqualBytecode(t, expected, actual.(*toy.Bytecode), msg...)
	default:
		require.Equal(t, expected, actual, msg...)
	}
}

func expectEqualObjects(t testing.TB, expected, actual []toy.Object, msg ...any) {
	expectEqual(t, len(expected), len(actual), msg...)
	for i := 0; i < len(expected); i++ {
		expectEqual(t, expected[i], actual[i], msg...)
	}
}

func expectEqualArray(t testing.TB, expected, actual *toy.Array, msg ...any) {
	expectEqual(t, expected.Immutable(), actual.Immutable(), msg...)
	expectEqual(t, expected.Len(), actual.Len(), msg...)
	for i := range expected.Len() {
		expectEqual(t, expected.At(i), actual.At(i), msg...)
	}
}

func expectEqualMap(t testing.TB, expected, actual *toy.Map, msg ...any) {
	expectEqual(t, expected.Immutable(), actual.Immutable(), msg...)
	expectEqual(t, expected.Len(), actual.Len(), msg...)

	expectedItems := expected.Items()
	actualItems := actual.Items()

	for i := range expectedItems {
		expectEqual(t, expectedItems[i], actualItems[i], msg...)
	}
}

func expectEqualTuple(t testing.TB, expected, actual toy.Tuple, msg ...any) {
	expectEqual(t, expected.Len(), actual.Len(), msg...)
	for i := range expected {
		expectEqual(t, expected[i], actual[i], msg...)
	}
}

func expectEqualError(t testing.TB, expected, actual *toy.Error, msg ...any) {
	expectEqual(t, expected.Message(), actual.Message(), msg...)
	if expected.Cause() == nil {
		expectNil(t, actual.Cause(), "expected nil, but got not nil")
		return
	}
	expectEqualError(t, expected.Cause(), actual.Cause(), msg...)
}

func expectEqualCompiledFunction(t testing.TB, expected, actual *toy.CompiledFunction, msg ...any) {
	expectEqual(t, expected.Instructions(), actual.Instructions(), msg...)
	expectEqual(t, expected.NumLocals(), actual.NumLocals(), msg...)
	expectEqual(t, expected.NumParameters(), actual.NumParameters(), msg...)
	expectEqual(t, expected.NumOptionals(), actual.NumOptionals(), msg...)
	expectEqual(t, expected.VarArgs(), actual.VarArgs(), msg...)
}

func expectEqualBytecode(t testing.TB, expected, actual *toy.Bytecode, msg ...any) {
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

type (
	ARR  []any
	IARR []any
	MAP  map[string]any
	IMAP map[string]any
)

func makeMap(pairs ...any) *toy.Map {
	out := new(toy.Map)
	for i := 0; i < len(pairs); i += 2 {
		key := pairs[i].(string)
		out.IndexSet(toy.String(key), toObject(pairs[i+1]))
	}
	return out
}

func makeImmutableMap(pars ...any) *toy.Map {
	return makeMap(pars...).Freeze().(*toy.Map)
}

func makeArray(args ...any) *toy.Array {
	elems := make([]toy.Object, 0, len(args))
	for _, arg := range args {
		elems = append(elems, toObject(arg))
	}
	return toy.NewArray(elems)
}

func makeImmutableArray(args ...any) *toy.Array {
	return makeArray(args...).Freeze().(*toy.Array)
}

func makeTuple(args ...any) toy.Tuple {
	tup := make(toy.Tuple, 0, len(args))
	for _, arg := range args {
		tup = append(tup, toObject(arg))
	}
	return tup
}

func toObject(v any) toy.Object {
	switch v := v.(type) {
	case toy.Object:
		return v
	case string:
		return toy.String(v)
	case int64:
		return toy.Int(v)
	case int:
		return toy.Int(v)
	case bool:
		return toy.Bool(v)
	case rune:
		return toy.Char(v)
	case byte: // for convenience
		return toy.Char(v)
	case float64:
		return toy.Float(v)
	case []byte:
		return toy.Bytes(v)
	case ARR:
		return makeArray(v...)
	case IARR:
		return makeImmutableArray(v...)
	case MAP:
		m := new(toy.Map)
		for k, v := range v {
			m.IndexSet(toy.String(k), toObject(v))
		}
		return m
	case IMAP:
		m := new(toy.Map)
		for k, v := range v {
			m.IndexSet(toy.String(k), toObject(v))
		}
		return m.Freeze()
	case nil:
		return toy.Nil
	}
	panic(fmt.Errorf("unknown type: %T", v))
}

func objectZeroCopy(o toy.Object) toy.Object {
	switch o.(type) {
	case toy.Int:
		return toy.Int(0)
	case toy.Float:
		return toy.Float(0)
	case toy.Bool:
		return toy.Bool(false)
	case toy.Char:
		return toy.Char(0)
	case toy.String:
		return toy.String("")
	case *toy.Array:
		return new(toy.Array)
	case *toy.Map:
		return new(toy.Map)
	case toy.NilValue:
		return toy.Nil
	case *toy.Error:
		return &toy.Error{}
	case toy.Bytes:
		return toy.Bytes{}
	case toy.Tuple:
		return toy.Tuple{}
	default:
		panic(fmt.Errorf("unknown object type: %s", toy.TypeName(o)))
	}
}
