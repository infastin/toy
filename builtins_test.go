package toy

import (
	"errors"
	"testing"
)

func Test_builtinLen(t *testing.T) {
	tests := []struct {
		name    string
		args    []Object
		want    Object
		wantErr error
	}{
		{
			name: "zero-array",
			args: []Object{makeArray()},
			want: Int(0),
		},
		{
			name: "some-array",
			args: []Object{makeArray(1, 2, 3)},
			want: Int(3),
		},
		{
			name: "zero-map",
			args: []Object{makeMap()},
			want: Int(0),
		},
		{
			name: "some-map",
			args: []Object{makeMap("a", 1, "b", 2)},
			want: Int(2),
		},
		{
			name: "zero-string",
			args: []Object{String("")},
			want: Int(0),
		},
		{
			name: "unicode-string",
			args: []Object{String("👾")},
			want: Int(1),
		},
		{
			name: "zero-tuple",
			args: []Object{Tuple{}},
			want: Int(0),
		},
		{
			name: "some-tuple",
			args: []Object{makeTuple(1, "2")},
			want: Int(2),
		},
		{
			name: "zero-bytes",
			args: []Object{Bytes{}},
			want: Int(0),
		},
		{
			name: "some-bytes",
			args: []Object{Bytes("hello")},
			want: Int(5),
		},
		{
			name:    "no-args",
			wantErr: &MissingArgumentError{Name: "value"},
		},
		{
			name: "not-sized",
			args: []Object{Int(0)},
			wantErr: &InvalidArgumentTypeError{
				Name: "value",
				Want: "sized",
				Got:  "int",
			},
		},
		{
			name: "too-many-args",
			args: []Object{Tuple{}, Int(1), Int(2)},
			wantErr: &WrongNumArgumentsError{
				WantMin: 1,
				WantMax: 1,
				Got:     3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builtinLen(tt.args...)
			if tt.wantErr != nil || err != nil {
				expectNotNil(t, err, "builtinLen: expected an error")
				expectNotNil(t, tt.wantErr, "builtinLen: encountered unexpected error")
				expectEqual(t, tt.wantErr.Error(), err.Error(), "builtinLen: wrong error message")
			}
			if tt.want != nil || got != nil {
				expectNotNil(t, got, "builtinLen: expected a result")
				expectNotNil(t, tt.want, "builtinLen: got unexpected result")
				expectEqual(t, tt.want, got, "builtinLen: wrong result")
			}
		})
	}
}

func Test_builtinAppend(t *testing.T) {
	tests := []struct {
		name    string
		args    []Object
		want    Object
		wantErr error
		target  Object
	}{
		{
			name:   "multiple",
			args:   []Object{makeArray(), Int(1), Int(2)},
			want:   makeArray(1, 2),
			target: makeArray(),
		},
		{
			name:   "zero-empty-array",
			args:   []Object{makeArray()},
			want:   makeArray(),
			target: makeArray(),
		},
		{
			name:   "zero-non-empty-array",
			args:   []Object{makeArray("not empty")},
			want:   makeArray("not empty"),
			target: makeArray("not empty"),
		},
		{
			name: "not array",
			args: []Object{Int(1)},
			wantErr: &InvalidArgumentTypeError{
				Name: "arr",
				Want: "array",
				Got:  "int",
			},
		},
		{
			name:    "no args",
			wantErr: &MissingArgumentError{Name: "arr"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builtinAppend(tt.args...)
			if tt.wantErr != nil || err != nil {
				expectNotNil(t, err, "builtinAppend: expected an error")
				expectNotNil(t, tt.wantErr, "builtinAppend: encountered unexpected error")
				expectEqual(t, tt.wantErr.Error(), err.Error(), "builtinAppend: wrong error message")
			}
			if tt.want != nil || got != nil {
				expectNotNil(t, got, "builtinAppend: expected a result")
				expectNotNil(t, tt.want, "builtinAppend: got unexpected result")
				expectEqual(t, tt.want, got, "builtinAppend: wrong result")
			}
			if tt.target != nil {
				expectEqual(t, tt.target, tt.args[0], "builtinAppend: incorrect target value")
			}
		})
	}
}

func Test_builtinDelete(t *testing.T) {
	tests := []struct {
		name    string
		args    []Object
		want    Object
		wantErr error
		target  Object
	}{
		{
			name: "invalid-arg",
			args: []Object{String(""), String("")},
			wantErr: &InvalidArgumentTypeError{
				Name: "collection",
				Want: "array or map",
				Got:  "string",
			},
		},
		{
			name: "no-args",
			wantErr: &WrongNumArgumentsError{
				WantMin: 2,
				Got:     0,
			},
		},
		{
			name: "empty-args",
			wantErr: &WrongNumArgumentsError{
				WantMin: 2,
				Got:     0,
			},
		},
		{
			name: "map-3-args",
			args: []Object{(*Map)(nil), String(""), String("")},
			wantErr: &WrongNumArgumentsError{
				WantMin: 2,
				WantMax: 2,
				Got:     3,
			},
		},
		{
			name: "nil-map-empty-key",
			args: []Object{makeMap(), String("")},
			want: Undefined,
		},
		{
			name: "nil-map-no-key",
			args: []Object{makeMap()},
			wantErr: &WrongNumArgumentsError{
				WantMin: 2,
				Got:     1,
			},
		},
		{
			name:   "map-missing-key",
			args:   []Object{makeMap("key", "value"), String("key1")},
			want:   Undefined,
			target: makeMap("key", "value"),
		},
		{
			name:   "map-emptied",
			args:   []Object{makeMap("key", "value"), String("key")},
			want:   String("value"),
			target: makeMap(),
		},
		{
			name:   "map-multi-keys",
			args:   []Object{makeMap("key1", 9, "key2", 10), String("key1")},
			want:   Int(9),
			target: makeMap("key2", 10),
		},
		{
			name: "array-4-args",
			args: []Object{(*Array)(nil), String(""), String(""), String("")},
			wantErr: &WrongNumArgumentsError{
				WantMin: 2,
				WantMax: 3,
				Got:     4,
			},
		},
		{
			name:   "array-2-args",
			args:   []Object{makeArray(1, 2), Int(0)},
			want:   makeArray(1),
			target: makeArray(2),
		},
		{
			name:   "array-3-args",
			args:   []Object{makeArray(1, 2, 3, 4), Int(1), Int(3)},
			want:   makeArray(2, 3),
			target: makeArray(1, 4),
		},
		{
			name: "array-2-invalid-args",
			args: []Object{makeArray(1, 2), String("")},
			wantErr: &InvalidArgumentTypeError{
				Name: "start",
				Want: "int",
				Got:  "string",
			},
		},
		{
			name: "array-3-invalid-args",
			args: []Object{makeArray(1, 2), Int(1), String("")},
			wantErr: &InvalidArgumentTypeError{
				Name: "stop",
				Want: "int",
				Got:  "string",
			},
		},
		{
			name:    "array-invalid-indices",
			args:    []Object{makeArray(1, 2), Int(2), Int(0)},
			wantErr: errors.New("invalid delete indices: 2 > 0"),
		},
		{
			name:    "array-out-of-bounds",
			args:    []Object{makeArray(1, 2), Int(3)},
			wantErr: errors.New("delete bounds out of range [3:2]"),
		},
		{
			name:    "array-out-of-bounds-with-len",
			args:    []Object{makeArray(1, 2), Int(1), Int(5)},
			wantErr: errors.New("delete bounds out of range [1:5] with len 2"),
		},
		{
			name:    "immutable-array",
			args:    []Object{makeImmutableArray(1, 2, 3), Int(1)},
			wantErr: errors.New("cannot delete from immutable array"),
		},
		{
			name:    "immutable-map",
			args:    []Object{makeImmutableMap("1", 1, "2", 2), String("1")},
			wantErr: errors.New("failed to delete 'string' from map: cannot delete from immutable hash table"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builtinDelete(tt.args...)
			if tt.wantErr != nil || err != nil {
				expectNotNil(t, err, "builtinDelete: expected an error")
				expectNotNil(t, tt.wantErr, "builtinDelete: encountered unexpected error")
				expectEqual(t, tt.wantErr.Error(), err.Error(), "builtinDelete: wrong error message")
			}
			if tt.want != nil || got != nil {
				expectNotNil(t, got, "builtinDelete: expected a result")
				expectNotNil(t, tt.want, "builtinDelete: got unexpected result")
				expectEqual(t, tt.want, got, "builtinDelete: wrong result")
			}
			if tt.target != nil {
				expectEqual(t, tt.target, tt.args[0], "builtinDelete: incorrect target value")
			}
		})
	}
}

func Test_builtinSplice(t *testing.T) {
	tests := []struct {
		name    string
		args    []Object
		want    Object
		wantErr error
		target  Object
	}{
		{
			name:    "no-args",
			wantErr: errors.New("missing argument for 'arr'"),
		},
		{
			name: "invalid-args",
			args: []Object{new(Map)},
			wantErr: &InvalidArgumentTypeError{
				Name: "arr",
				Want: "array",
				Got:  "map",
			},
		},
		{
			name: "invalid-args",
			args: []Object{new(Array), String("")},
			wantErr: &InvalidArgumentTypeError{
				Name: "start",
				Want: "int",
				Got:  "string",
			},
		},
		{
			name:    "negative-index",
			args:    []Object{new(Array), Int(-1)},
			wantErr: errors.New("splice bounds out of range [-1:0]"),
		},
		{
			name: "non-int-stop",
			args: []Object{new(Array), Int(0), String("")},
			wantErr: &InvalidArgumentTypeError{
				Name: "stop",
				Want: "int",
				Got:  "string",
			},
		},
		{
			name:    "negative-count",
			args:    []Object{makeArray(0, 1, 2), Int(0), Int(-1)},
			wantErr: errors.New("invalid splice indices: 0 > -1"),
		},
		{
			name:   "append",
			args:   []Object{makeArray(0, 1, 2), Int(3), Int(3), String("b")},
			want:   makeArray(),
			target: makeArray(0, 1, 2, "b"),
		},
		{
			name:   "prepend",
			args:   []Object{makeArray(0, 1, 2), Int(0), Int(0), String("b")},
			want:   makeArray(),
			target: makeArray("b", 0, 1, 2),
		},
		{
			name:   "insert-at-one",
			args:   []Object{makeArray(0, 1, 2), Int(1), Int(1), String("c"), String("d")},
			want:   makeArray(),
			target: makeArray(0, "c", "d", 1, 2),
		},
		{
			name:   "insert-with-delete",
			args:   []Object{makeArray(0, 1, 2), Int(1), Int(2), String("c"), String("d")},
			want:   makeArray(1),
			target: makeArray(0, "c", "d", 2),
		},
		{
			name:   "insert-with-delete-multiple",
			args:   []Object{makeArray(0, 1, 2), Int(1), Int(3), String("c"), String("d")},
			want:   makeArray(1, 2),
			target: makeArray(0, "c", "d"),
		},
		{
			name:   "delete-all",
			args:   []Object{makeArray(0, 1, 2), Int(0), Int(3)},
			want:   makeArray(0, 1, 2),
			target: makeArray(),
		},
		{
			name:    "delete-out-of-bounds",
			args:    []Object{makeArray(0, 1, 2), Int(0), Int(4)},
			wantErr: errors.New("splice bounds out of range [0:4] with len 3"),
		},
		{
			name:   "pop",
			args:   []Object{makeArray(0, 1, 2), Int(2)},
			want:   makeArray(2),
			target: makeArray(0, 1),
		},
		{
			name:    "immutable",
			args:    []Object{makeImmutableArray(1, 2, 3), Int(1)},
			wantErr: errors.New("cannot splice immutable array"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builtinSplice(tt.args...)
			if tt.wantErr != nil || err != nil {
				expectNotNil(t, err, "builtinSplice: expected an error")
				expectNotNil(t, tt.wantErr, "builtinSplice: encountered unexpected error")
				expectEqual(t, tt.wantErr.Error(), err.Error(), "builtinSplice: wrong error message")
			}
			if tt.want != nil || got != nil {
				expectNotNil(t, got, "builtinSplice: expected a result")
				expectNotNil(t, tt.want, "builtinSplice: got unexpected result")
				expectEqual(t, tt.want, got, "builtinSplice: wrong result")
			}
			if tt.target != nil {
				expectEqual(t, tt.target, tt.args[0], "builtinSplice: incorrect target value")
			}
		})
	}
}

func Test_builtinInsert(t *testing.T) {
	tests := []struct {
		name    string
		args    []Object
		want    Object
		wantErr error
		target  Object
	}{
		{
			name: "no-args",
			wantErr: &WrongNumArgumentsError{
				WantMin: 2,
				Got:     0,
			},
		},
		{
			name: "one-arg",
			args: []Object{Int(0)},
			wantErr: &WrongNumArgumentsError{
				WantMin: 2,
				Got:     1,
			},
		},
		{
			name: "invalid-argument",
			args: []Object{Int(0), Int(0)},
			wantErr: &InvalidArgumentTypeError{
				Name: "collection",
				Want: "array or map",
				Got:  "int",
			},
		},
		{
			name:   "array-prepend",
			args:   []Object{makeArray(0, 1, 2), Int(0), String("b")},
			want:   Undefined,
			target: makeArray("b", 0, 1, 2),
		},
		{
			name:   "array-insert-at-one",
			args:   []Object{makeArray(0, 1, 2), Int(1), String("c"), String("d")},
			want:   Undefined,
			target: makeArray(0, "c", "d", 1, 2),
		},
		{
			name:   "array-append",
			args:   []Object{makeArray(0, 1, 2), Int(3), String("c"), String("d")},
			want:   Undefined,
			target: makeArray(0, 1, 2, "c", "d"),
		},
		{
			name:   "map-insert",
			args:   []Object{makeMap(), String("1"), Int(1)},
			want:   Undefined,
			target: makeMap("1", 1),
		},
		{
			name:    "map-insert-not-hashable",
			args:    []Object{makeMap(), Tuple{}, Int(1)},
			wantErr: errors.New("failed to insert 'tuple' into map: not hashable"),
		},
		{
			name:    "immutable-array",
			args:    []Object{makeImmutableArray(1, 2, 3), Int(1), String("c")},
			wantErr: errors.New("cannot insert into immutable array"),
		},
		{
			name:    "immutable-map",
			args:    []Object{makeImmutableMap("1", 1, "2", 2), String("3"), Int(3)},
			wantErr: errors.New("failed to insert 'string' into map: cannot insert into immutable hash table"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builtinInsert(tt.args...)
			if tt.wantErr != nil || err != nil {
				expectNotNil(t, err, "builtinInsert: expected an error")
				expectNotNil(t, tt.wantErr, "builtinInsert: encountered unexpected error")
				expectEqual(t, tt.wantErr.Error(), err.Error(), "builtinInsert: wrong error message")
			}
			if tt.want != nil || got != nil {
				expectNotNil(t, got, "builtinInsert: expected a result")
				expectNotNil(t, tt.want, "builtinInsert: got unexpected result")
				expectEqual(t, tt.want, got, "builtinInsert: wrong result")
			}
			if tt.target != nil {
				expectEqual(t, tt.target, tt.args[0], "builtinInsert: incorrect target value")
			}
		})
	}
}

func Test_builtinClear(t *testing.T) {
	tests := []struct {
		name    string
		args    []Object
		want    Object
		wantErr error
		target  Object
	}{
		{
			name: "no-args",
			wantErr: &WrongNumArgumentsError{
				WantMin: 1,
				WantMax: 1,
				Got:     0,
			},
		},
		{
			name: "too-many-args",
			args: []Object{makeArray(1), Int(1), Int(2)},
			wantErr: &WrongNumArgumentsError{
				WantMin: 1,
				WantMax: 1,
				Got:     3,
			},
		},
		{
			name: "invalid-argument",
			args: []Object{Int(0)},
			wantErr: &InvalidArgumentTypeError{
				Name: "collection",
				Want: "array or map",
				Got:  "int",
			},
		},
		{
			name:   "map",
			args:   []Object{makeMap("1", 1, "2", 2)},
			want:   Undefined,
			target: makeMap(),
		},
		{
			name:   "map",
			args:   []Object{makeArray(1, 2, 3)},
			want:   Undefined,
			target: makeArray(),
		},
		{
			name:    "immutable-array",
			args:    []Object{makeImmutableArray(1, 2, 3)},
			wantErr: errors.New("cannot clear immutable array"),
		},
		{
			name:    "immutable-map",
			args:    []Object{makeImmutableMap("1", 1, "2", 2)},
			wantErr: errors.New("cannot clear immutable hash table"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builtinClear(tt.args...)
			if tt.wantErr != nil || err != nil {
				expectNotNil(t, err, "builtinClear: expected an error")
				expectNotNil(t, tt.wantErr, "builtinClear: encountered unexpected error")
				expectEqual(t, tt.wantErr.Error(), err.Error(), "builtinClear: wrong error message")
			}
			if tt.want != nil || got != nil {
				expectNotNil(t, got, "builtinClear: expected a result")
				expectNotNil(t, tt.want, "builtinClear: got unexpected result")
				expectEqual(t, tt.want, got, "builtinClear: wrong result")
			}
			if tt.target != nil {
				expectEqual(t, tt.target, tt.args[0], "builtinClear: incorrect target value")
			}
		})
	}
}

func Test_builtinRange(t *testing.T) {
	tests := []struct {
		name    string
		args    []Object
		want    Object
		wantErr error
	}{
		{
			name:    "no-args",
			wantErr: &MissingArgumentError{Name: "start"},
		},
		{
			name: "4-args",
			args: []Object{Int(0), Int(0), Int(0), Int(0)},
			wantErr: &WrongNumArgumentsError{
				WantMin: 1,
				WantMax: 3,
				Got:     4,
			},
		},
		{
			name: "invalid-start",
			args: []Object{String("")},
			wantErr: &InvalidArgumentTypeError{
				Name: "start",
				Want: "int",
				Got:  "string",
			},
		},
		{
			name: "invalid-stop",
			args: []Object{Int(0), String("")},
			wantErr: &InvalidArgumentTypeError{
				Name: "stop",
				Want: "int",
				Got:  "string",
			},
		},
		{
			name: "invalid-stop",
			args: []Object{Int(0), Int(0), String("")},
			wantErr: &InvalidArgumentTypeError{
				Name: "step",
				Want: "int",
				Got:  "string",
			},
		},
		{
			name:    "zero-step",
			args:    []Object{Int(0), Int(0), Int(0)},
			wantErr: errors.New("invalid range step: must be > 0, got 0"),
		},
		{
			name:    "negative step",
			args:    []Object{Int(0), Int(0), Int(-2)},
			wantErr: errors.New("invalid range step: must be > 0, got -2"),
		},
		{
			name: "same-bound",
			args: []Object{Int(0), Int(0)},
			want: makeArray(),
		},
		{
			name: "positive-range",
			args: []Object{Int(0), Int(5)},
			want: makeArray(0, 1, 2, 3, 4),
		},
		{
			name: "negative-range",
			args: []Object{Int(0), Int(-5)},
			want: makeArray(0, -1, -2, -3, -4),
		},
		{
			name: "positive-with-step",
			args: []Object{Int(0), Int(5), Int(2)},
			want: makeArray(0, 2, 4),
		},
		{
			name: "positive-with-step",
			args: []Object{Int(0), Int(-10), Int(2)},
			want: makeArray(0, -2, -4, -6, -8),
		},
		{
			name: "large-range",
			args: []Object{Int(-10), Int(10), Int(3)},
			want: makeArray(-10, -7, -4, -1, 2, 5, 8),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builtinRange(tt.args...)
			if tt.wantErr != nil || err != nil {
				expectNotNil(t, err, "builtinRange: expected an error")
				expectNotNil(t, tt.wantErr, "builtinRange: encountered unexpected error")
				expectEqual(t, tt.wantErr.Error(), err.Error(), "builtinRange: wrong error message")
			}
			if tt.want != nil || got != nil {
				expectNotNil(t, got, "builtinRange: expected a result")
				expectNotNil(t, tt.want, "builtinRange: got unexpected result")
				expectEqual(t, tt.want, got, "builtinRange: wrong result")
			}
		})
	}
}

func Test_builtinMin(t *testing.T) {
	tests := []struct {
		name    string
		args    []Object
		want    Object
		wantErr error
	}{
		{
			name:    "no-args",
			wantErr: &WrongNumArgumentsError{WantMin: 1, Got: 0},
		},
		{
			name: "ints",
			args: []Object{Int(0), Int(1), Int(2), Int(3)},
			want: Int(0),
		},
		{
			name: "floats",
			args: []Object{Float(-1.23), Float(0.1), Float(1.3), Float(2.5)},
			want: Float(-1.23),
		},
		{
			name: "ints-floats",
			args: []Object{Int(0), Float(-1.2), Int(2), Float(-3.33), Int(-5)},
			want: Int(-5),
		},
		{
			name:    "not-comparable",
			args:    []Object{String("hello"), Int(0)},
			wantErr: errors.New(`operation 'int < string' has failed: invalid operator`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builtinMin(tt.args...)
			if tt.wantErr != nil || err != nil {
				expectNotNil(t, err, "builtinMin: expected an error")
				expectNotNil(t, tt.wantErr, "builtinMin: encountered unexpected error")
				expectEqual(t, tt.wantErr.Error(), err.Error(), "builtinMin: wrong error message")
			}
			if tt.want != nil || got != nil {
				expectNotNil(t, got, "builtinMin: expected a result")
				expectNotNil(t, tt.want, "builtinMin: got unexpected result")
				expectEqual(t, tt.want, got, "builtinMin: wrong result")
			}
		})
	}
}

func Test_builtinMax(t *testing.T) {
	tests := []struct {
		name    string
		args    []Object
		want    Object
		wantErr error
	}{
		{
			name:    "no-args",
			wantErr: &WrongNumArgumentsError{WantMin: 1, Got: 0},
		},
		{
			name: "ints",
			args: []Object{Int(0), Int(1), Int(2), Int(3)},
			want: Int(3),
		},
		{
			name: "floats",
			args: []Object{Float(-1.23), Float(0.1), Float(1.3), Float(2.5)},
			want: Float(2.5),
		},
		{
			name: "ints-floats",
			args: []Object{Int(0), Float(-1.2), Int(2), Float(-3.33), Int(-5)},
			want: Int(2),
		},
		{
			name:    "not-comparable",
			args:    []Object{String("hello"), Int(0)},
			wantErr: errors.New(`operation 'int > string' has failed: invalid operator`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builtinMax(tt.args...)
			if tt.wantErr != nil || err != nil {
				expectNotNil(t, err, "builtinMax: expected an error")
				expectNotNil(t, tt.wantErr, "builtinMax: encountered unexpected error")
				expectEqual(t, tt.wantErr.Error(), err.Error(), "builtinMax: wrong error message")
			}
			if tt.want != nil || got != nil {
				expectNotNil(t, got, "builtinMax: expected a result")
				expectNotNil(t, tt.want, "builtinMax: got unexpected result")
				expectEqual(t, tt.want, got, "builtinMax: wrong result")
			}
		})
	}
}
