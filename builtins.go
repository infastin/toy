package toy

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/infastin/toy/token"
)

var Universe = []*Variable{
	NewVariable("typename", NewBuiltinFunction("typename", builtinTypeName)),
	NewVariable("clone", NewBuiltinFunction("clone", builtinClone)),
	NewVariable("freeze", NewBuiltinFunction("freeze", builtinFreeze)),
	NewVariable("satisfies", NewBuiltinFunction("satisfies", builtinSatisfies)),
	NewVariable("immutable", NewBuiltinFunction("immutable", builtinImmutable)),

	NewVariable("len", NewBuiltinFunction("len", builtinLen)),
	NewVariable("append", NewBuiltinFunction("append", builtinAppend)),
	NewVariable("copy", NewBuiltinFunction("copy", builtinCopy)),
	NewVariable("delete", NewBuiltinFunction("delete", builtinDelete)),
	NewVariable("splice", NewBuiltinFunction("splice", builtinSplice)),
	NewVariable("insert", NewBuiltinFunction("insert", builtinInsert)),
	NewVariable("clear", NewBuiltinFunction("clear", builtinClear)),
	NewVariable("contains", NewBuiltinFunction("contains", builtinContains)),

	NewVariable("format", NewBuiltinFunction("format", builtinFormat)),
	NewVariable("min", NewBuiltinFunction("min", builtinMin)),
	NewVariable("max", NewBuiltinFunction("max", builtinMax)),

	NewVariable("type", rootTypeImpl{}),
	NewVariable("bool", BoolType),
	NewVariable("float", FloatType),
	NewVariable("int", IntType),
	NewVariable("string", StringType),
	NewVariable("bytes", BytesType),
	NewVariable("char", CharType),
	NewVariable("array", ArrayType),
	NewVariable("table", TableType),
	NewVariable("tuple", TupleType),
	NewVariable("range", RangeType),
	NewVariable("function", FunctionType),
}

func builtinTypeName(_ *Runtime, args ...Value) (Value, error) {
	if len(args) != 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	return String(TypeName(args[0])), nil
}

func builtinClone(_ *Runtime, args ...Value) (Value, error) {
	if len(args) != 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	return args[0].Clone(), nil
}

func builtinFreeze(_ *Runtime, args ...Value) (Value, error) {
	if len(args) != 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	return Freeze(args[0]), nil
}

func builtinLen(_ *Runtime, args ...Value) (Value, error) {
	var value Sized
	if err := UnpackArgs(args, "value", &value); err != nil {
		return nil, err
	}
	return Int(value.Len()), nil
}

func builtinAppend(_ *Runtime, args ...Value) (Value, error) {
	var (
		arr  *Array
		rest []Value
	)
	if err := UnpackArgs(args, "arr", &arr, "...", &rest); err != nil {
		return nil, err
	}
	return &Array{
		elems:     append(arr.elems, rest...),
		immutable: arr.immutable,
		itercount: 0,
	}, nil
}

func builtinCopy(_ *Runtime, args ...Value) (Value, error) {
	var dst, src *Array
	if err := UnpackArgs(args, "dst", &dst, "src", &src); err != nil {
		return nil, err
	}
	if err := dst.checkMutable("copy to"); err != nil {
		return nil, err
	}
	n := copy(dst.elems, src.elems)
	return Int(n), nil
}

func builtinDelete(_ *Runtime, args ...Value) (Value, error) {
	if len(args) < 2 {
		return nil, &WrongNumArgumentsError{
			WantMin: 2,
			Got:     len(args),
		}
	}
	switch x := args[0].(type) {
	case *Array:
		if len(args) > 3 {
			return nil, &WrongNumArgumentsError{
				WantMin: 2,
				WantMax: 3,
				Got:     len(args),
			}
		}
		var (
			start, stop int
			stopPtr     *int
		)
		if err := UnpackArgs(args[1:], "start", &start, "stop?", &stopPtr); err != nil {
			return nil, err
		}
		if stopPtr != nil {
			stop = *stopPtr
		} else {
			stop = start + 1
		}
		if err := x.checkMutable("delete from"); err != nil {
			return nil, err
		}
		n := len(x.elems)
		if start < 0 || stop < 0 {
			neg := stop
			if start < 0 {
				neg = start
			}
			return nil, fmt.Errorf("negative delete index: %d", neg)
		}
		if start > n {
			return nil, fmt.Errorf("delete bounds out of range [%d:%d]", start, n)
		}
		if stop > n {
			return nil, fmt.Errorf("delete bounds out of range [%d:%d] with len %d", start, stop, n)
		}
		if start > stop {
			return nil, fmt.Errorf("invalid delete indices: %d > %d", start, stop)
		}
		if start == stop {
			return NewArray(nil), nil
		}
		deleted := slices.Clone(x.elems[start:stop])
		x.elems = slices.Delete(x.elems, start, stop)
		return NewArray(deleted), nil
	case *Table:
		if len(args) > 2 {
			return nil, &WrongNumArgumentsError{
				WantMin: 2,
				WantMax: 2,
				Got:     len(args),
			}
		}
		value, err := x.Delete(args[1])
		if err != nil {
			return nil, fmt.Errorf("failed to delete '%s' from table: %w", TypeName(args[1]), err)
		}
		return value, nil
	default:
		return nil, &InvalidArgumentTypeError{
			Name: "collection",
			Want: "array or table",
			Got:  TypeName(x),
		}
	}
}

func builtinSplice(_ *Runtime, args ...Value) (Value, error) {
	var (
		arr         *Array
		start, stop int
		stopPtr     *int
		rest        []Value
	)
	if err := UnpackArgs(args,
		"arr", &arr,
		"start?", &start,
		"stop?", &stopPtr,
		"...", &rest,
	); err != nil {
		return nil, err
	}
	n := len(arr.elems)
	if stopPtr != nil {
		stop = *stopPtr
	} else {
		stop = n
	}
	if err := arr.checkMutable("splice"); err != nil {
		return nil, err
	}
	if start < 0 || stop < 0 {
		neg := stop
		if start < 0 {
			neg = start
		}
		return nil, fmt.Errorf("negative splice index: %d", neg)
	}
	if start > n {
		return nil, fmt.Errorf("splice bounds out of range [%d:%d]", start, n)
	}
	if stop > n {
		return nil, fmt.Errorf("splice bounds out of range [%d:%d] with len %d", start, stop, n)
	}
	if start > stop {
		return nil, fmt.Errorf("invalid splice indices: %d > %d", start, stop)
	}
	if start == stop {
		arr.elems = slices.Insert(arr.elems, start, rest...)
		return NewArray(nil), nil
	}
	deleted := slices.Clone(arr.elems[start:stop])
	if len(rest) != 0 {
		arr.elems = slices.Concat(arr.elems[:start], rest, arr.elems[stop:])
	} else {
		arr.elems = slices.Delete(arr.elems, start, stop)
	}
	return NewArray(deleted), nil
}

func builtinInsert(_ *Runtime, args ...Value) (Value, error) {
	if len(args) < 2 {
		return nil, &WrongNumArgumentsError{
			WantMin: 2,
			Got:     len(args),
		}
	}
	switch x := args[0].(type) {
	case *Array:
		var (
			index int
			rest  []Value
		)
		if err := UnpackArgs(args[1:], "index", &index, "...", &rest); err != nil {
			return nil, err
		}
		if err := x.checkMutable("insert into"); err != nil {
			return nil, err
		}
		n := len(x.elems)
		if index < 0 {
			return nil, fmt.Errorf("negative insert index: %d", index)
		}
		if index > n {
			return nil, fmt.Errorf("insert index %d out of range [:%d]", index, n)
		}
		x.elems = slices.Insert(x.elems, index, rest...)
		return Nil, nil
	case *Table:
		if len(args) != 3 {
			return nil, &WrongNumArgumentsError{
				WantMin: 2,
				WantMax: 3,
				Got:     len(args),
			}
		}
		var index, value Value
		if err := UnpackArgs(args[1:], "index", &index, "value", &value); err != nil {
			return nil, err
		}
		if err := x.SetProperty(index, value); err != nil {
			return nil, fmt.Errorf("failed to insert '%s' into table: %w", TypeName(index), err)
		}
		return Nil, nil
	default:
		return nil, &InvalidArgumentTypeError{
			Name: "collection",
			Want: "array or table",
			Got:  TypeName(x),
		}
	}
}

func builtinClear(_ *Runtime, args ...Value) (Value, error) {
	if len(args) != 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	switch x := args[0].(type) {
	case *Array:
		if err := x.Clear(); err != nil {
			return nil, err
		}
	case *Table:
		if err := x.Clear(); err != nil {
			return nil, err
		}
	default:
		return nil, &InvalidArgumentTypeError{
			Name: "collection",
			Want: "array or table",
			Got:  TypeName(x),
		}
	}
	return Nil, nil
}

func builtinContains(_ *Runtime, args ...Value) (Value, error) {
	var (
		container Container
		value     Value
	)
	if err := UnpackArgs(args, "container", &container, "value", &value); err != nil {
		return nil, err
	}
	contains, err := container.Contains(value)
	if err != nil {
		if e := (*InvalidValueTypeError)(nil); errors.As(err, &e) {
			err = &InvalidArgumentTypeError{
				Name: "value",
				Want: e.Want,
				Got:  e.Got,
			}
		}
		return nil, err
	}
	return Bool(contains), nil
}

func builtinFormat(_ *Runtime, args ...Value) (Value, error) {
	var (
		format string
		rest   []Value
	)
	if err := UnpackArgs(args, "format", &format, "...", &rest); err != nil {
		return nil, err
	}
	if len(rest) == 0 {
		return String(format), nil
	}
	s, err := Format(string(format), rest...)
	if err != nil {
		return nil, err
	}
	return String(s), nil
}

func builtinMin(_ *Runtime, args ...Value) (Value, error) {
	if len(args) < 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			Got:     len(args),
		}
	}
	min := args[0]
	for _, arg := range args[1:] {
		less, err := Compare(token.Less, arg, min)
		if err != nil {
			return nil, err
		}
		if less {
			min = arg
		}
	}
	return min, nil
}

func builtinMax(_ *Runtime, args ...Value) (Value, error) {
	if len(args) < 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			Got:     len(args),
		}
	}
	max := args[0]
	for _, arg := range args[1:] {
		greater, err := Compare(token.Greater, arg, max)
		if err != nil {
			return nil, err
		}
		if greater {
			max = arg
		}
	}
	return max, nil
}

func builtinSatisfies(_ *Runtime, args ...Value) (Value, error) {
	if len(args) < 2 {
		return nil, &WrongNumArgumentsError{
			WantMin: 2,
			Got:     len(args),
		}
	}
	x := args[0]
	ifaces := args[1:]
	for i, iface := range ifaces {
		name, ok := iface.(String)
		if !ok {
			return nil, &InvalidArgumentTypeError{
				Name: fmt.Sprintf("ifaces[%d]", i),
				Want: "string",
				Got:  TypeName(iface),
			}
		}

		trait := strings.TrimSpace(string(name))
		trait = strings.Map(func(r rune) rune {
			if r == '-' {
				return ' '
			}
			return unicode.ToLower(r)
		}, trait)

		switch trait {
		case "hashable":
			_, ok = x.(Hashable)
		case "freezable":
			_, ok = x.(Freezable)
		case "comparable":
			_, ok = x.(Comparable)
		case "binary-op":
			_, ok = x.(HasBinaryOp)
		case "unary-op":
			_, ok = x.(HasUnaryOp)
		case "property accessible":
			_, ok = x.(PropertyAccessible)
		case "property assignable":
			_, ok = x.(PropertyAssignable)
		case "sized":
			_, ok = x.(Sized)
		case "index accessible":
			_, ok = x.(IndexAccessible)
		case "index assignable":
			_, ok = x.(IndexAssignable)
		case "sliceable":
			_, ok = x.(Sliceable)
		case "convertible":
			_, ok = x.(Convertible)
		case "callable":
			_, ok = x.(Callable)
		case "container":
			_, ok = x.(Container)
		case "iterable":
			_, ok = x.(Iterable)
		case "sequence":
			_, ok = x.(Sequence)
		case "kv iterable":
			_, ok = x.(KVIterable)
		case "mapping":
			_, ok = x.(Mapping)
		default:
			return nil, fmt.Errorf("ifaces[%d]: invalid interface '%s'", i, string(name))
		}
		if !ok {
			return False, nil
		}
	}
	return True, nil
}

func builtinImmutable(_ *Runtime, args ...Value) (Value, error) {
	if len(args) != 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	return Bool(Immutable(args[0])), nil
}
