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
			wantErr: errors.New("missing argument for 'value'"),
		},
		{
			name: "not-sized",
			wantErr: &ErrInvalidArgumentType{
				Name:     "value",
				Expected: "sized",
				Found:    "int",
			},
		},
		{
			name:    "too-many-args",
			args:    []Object{Tuple{}, Int(1), Int(2)},
			wantErr: errors.New("want at most 1 argument(s), got %d"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builtinLen(tt.args...)
			expectEqual(t, tt.wantErr, err, "builtinLen() error")
			expectEqual(t, tt.want, got, "builtinLen() result")
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
			want:   makeArray(),
			target: makeArray(),
		},
		{
			name: "not array",
			args: []Object{Int(1)},
			wantErr: &ErrInvalidArgumentType{
				Name:     "arr",
				Expected: "array",
				Found:    "int",
			},
		},
		{
			name:    "no args",
			wantErr: errors.New("missing argument for 'arr'"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builtinAppend(tt.args...)
			expectEqual(t, tt.wantErr, err, "builtinAppend() error")
			expectEqual(t, tt.want, got, "builtinAppend() result")
			if tt.target != nil {
				expectEqual(t, tt.target, tt.args[0], "builtinAppend() target")
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
			wantErr: &ErrInvalidArgumentType{
				Name:     "collection",
				Expected: "array or map",
				Found:    "string",
			},
		},
		{
			name:    "no-args",
			wantErr: errors.New("want at least 2 arguments, got 0"),
		},
		{
			name:    "empty-args",
			wantErr: errors.New("want at least 2 arguments, got 0"),
		},
		{
			name:    "3-args",
			args:    []Object{(*Map)(nil), String(""), String("")},
			wantErr: errors.New("want at most 2 argument(s), got 3"),
		},
		{
			name: "nil-map-empty-key",
			args: []Object{makeMap(), String("")},
			want: Undefined,
		},
		{
			name:    "nil-map-no-key",
			args:    []Object{makeMap()},
			wantErr: errors.New("want at least 2 arguments, got 1"),
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
			want:   Undefined,
			target: makeMap(),
		},
		{
			name:   "map-multi-keys",
			args:   []Object{makeMap("key1", 9, "key2", 10), String("key1")},
			want:   Undefined,
			target: makeMap("key2", 10),
		},
		{
			name:    "array-4-args",
			args:    []Object{(*Array)(nil), String(""), String(""), String("")},
			wantErr: errors.New("want at most 3 argument(s), got 4"),
		},
		{
			name:   "array-2-args",
			args:   []Object{makeArray(1, 2), Int(0)},
			want:   makeArray(1),
			target: makeArray(2),
		},
		{
			name:   "array-3-args",
			args:   []Object{makeArray(1, 2, 3, 4), Int(1), Int(2)},
			want:   makeArray(2, 3),
			target: makeArray(1, 4),
		},
		{
			name: "array-2-invalid-args",
			args: []Object{makeArray(1, 2), String("")},
			wantErr: &ErrInvalidArgumentType{
				Name:     "start",
				Expected: "int",
				Found:    "string",
			},
		},
		{
			name: "array-3-invalid-args",
			args: []Object{makeArray(1, 2), Int(1), String("")},
			wantErr: &ErrInvalidArgumentType{
				Name:     "stop",
				Expected: "int",
				Found:    "string",
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
			expectEqual(t, tt.wantErr, err, "builtinDelete() error")
			expectEqual(t, tt.want, got, "builtinDelete() result")
			if tt.target != nil {
				expectEqual(t, tt.target, tt.args[0], "builtinDelete() target")
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
			args: []Object{(*Map)(nil)},
			wantErr: &ErrInvalidArgumentType{
				Name:     "arr",
				Expected: "array",
				Found:    "map",
			},
		},
		{
			name: "invalid-args",
			args: []Object{(*Array)(nil), String("")},
			wantErr: &ErrInvalidArgumentType{
				Name:     "start",
				Expected: "int",
				Found:    "string",
			},
		},
		{
			name:    "negative-index",
			args:    []Object{(*Array)(nil), Int(-1)},
			wantErr: errors.New("splice bounds out of range [-1:0]"),
		},
		{
			name: "non-int-stop",
			args: []Object{(*Array)(nil), Int(0), String("")},
			wantErr: &ErrInvalidArgumentType{
				Name:     "stop",
				Expected: "int",
				Found:    "string",
			},
		},
		{
			name:    "negative-count",
			args:    []Object{makeArray(0, 1, 2), Int(0), Int(-1)},
			wantErr: errors.New("splice bounds out of range [0:-1] with len 3"),
		},
		{
			name:   "append",
			args:   []Object{makeArray(0, 1, 2), Int(3), Int(3), String("b")},
			want:   makeArray(),
			target: makeArray(makeArray(0, 1, 2, "b")),
		},
		{
			name:   "prepend",
			args:   []Object{makeArray(0, 1, 2), Int(0), Int(0), String("b")},
			want:   makeArray(),
			target: makeArray(makeArray("b", 0, 1, 2)),
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
			expectEqual(t, tt.wantErr, err, "builtinSplice() error")
			expectEqual(t, tt.want, got, "builtinSplice() result")
			if tt.target != nil {
				expectEqual(t, tt.target, tt.args[0], "builtinSplice() target")
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
			name: "invalid-argument",
			args: []Object{Int(0)},
			wantErr: &ErrInvalidArgumentType{
				Name:     "collection",
				Expected: "array or map",
				Found:    "int",
			},
		},
		{
			name:   "array-prepend",
			args:   []Object{makeArray(0, 1, 2), Int(0), String("b")},
			want:   makeArray(),
			target: makeArray(makeArray("b", 0, 1, 2)),
		},
		{
			name:   "array-insert-at-one",
			args:   []Object{makeArray(0, 1, 2), Int(1), String("c"), String("d")},
			want:   makeArray(),
			target: makeArray(0, "c", "d", 1, 2),
		},
		{
			name:   "array-append",
			args:   []Object{makeArray(0, 1, 2), Int(3), String("c"), String("d")},
			want:   makeArray(),
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
			expectEqual(t, tt.wantErr, err, "builtinInsert() error")
			expectEqual(t, tt.want, got, "builtinInsert() result")
			if tt.target != nil {
				expectEqual(t, tt.target, tt.args[0], "builtinInsert() target")
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
			name:    "no-args",
			wantErr: errors.New("want 1 arguments, got 0"),
		},
		{
			name:    "too-many-args",
			args:    []Object{makeArray(1), Int(1), Int(2)},
			wantErr: errors.New("want 1 argument, got 3"),
		},
		{
			name: "invalid-argument",
			args: []Object{Int(0)},
			wantErr: &ErrInvalidArgumentType{
				Name:     "collection",
				Expected: "array or map",
				Found:    "int",
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
			expectEqual(t, tt.wantErr, err, "builtinClear() error")
			expectEqual(t, tt.want, got, "builtinClear() result")
			if tt.target != nil {
				expectEqual(t, tt.target, tt.args[0], "builtinClear() target")
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
			name:    "no args",
			wantErr: errors.New("missing argument for 'start'"),
		},
		{
			name:    "4 args",
			args:    []Object{Int(0), Int(0), Int(0), Int(0)},
			wantErr: errors.New("nant at most 3 argument(s), got 4"),
		},
		{
			name: "invalid-start",
			args: []Object{String("")},
			wantErr: &ErrInvalidArgumentType{
				Name:     "start",
				Expected: "int",
				Found:    "string",
			},
		},
		{
			name: "invalid-stop",
			args: []Object{Int(0), String("")},
			wantErr: &ErrInvalidArgumentType{
				Name:     "stop",
				Expected: "int",
				Found:    "string",
			},
		},
		{
			name: "invalid-stop",
			args: []Object{Int(0), Int(0), String("")},
			wantErr: &ErrInvalidArgumentType{
				Name:     "step",
				Expected: "int",
				Found:    "string",
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
			expectEqual(t, tt.wantErr, err, "builtinRange() error")
			expectEqual(t, tt.want, got, "builtinRange() result")
		})
	}
}
