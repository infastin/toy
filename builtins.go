package toy

import (
	"errors"
	"fmt"
	"slices"

	"github.com/infastin/toy/token"
)

var (
	BuiltinFuncs = []*BuiltinFunction{
		{Name: "typename", Func: builtinTypeName},
		{Name: "clone", Func: builtinClone},

		{Name: "len", Func: builtinLen},
		{Name: "append", Func: builtinAppend},
		{Name: "copy", Func: builtinCopy},
		{Name: "delete", Func: builtinDelete},
		{Name: "splice", Func: builtinSplice},
		{Name: "insert", Func: builtinInsert},
		{Name: "clear", Func: builtinClear},
		{Name: "has", Func: builtinHas},

		{Name: "format", Func: builtinFormat},
		{Name: "fail", Func: builtinFail},
		{Name: "min", Func: builtinMin},
		{Name: "max", Func: builtinMax},

		{Name: "error", Func: builtinError},
		{Name: "range", Func: builtinRange},

		{Name: "string", Func: builtinConvert[String]},
		{Name: "int", Func: builtinConvert[Int]},
		{Name: "bool", Func: builtinConvert[Bool]},
		{Name: "float", Func: builtinConvert[Float]},
		{Name: "char", Func: builtinConvert[Char]},
		{Name: "bytes", Func: builtinBytes},

		{Name: "isBool", Func: builtinIs[Bool]},
		{Name: "isFloat", Func: builtinIs[Float]},
		{Name: "isInt", Func: builtinIs[Int]},
		{Name: "isString", Func: builtinIs[String]},
		{Name: "isBytes", Func: builtinIs[Bytes]},
		{Name: "isChar", Func: builtinIs[Char]},
		{Name: "isArray", Func: builtinIs[*Array]},
		{Name: "isMap", Func: builtinIs[*Map]},
		{Name: "isTuple", Func: builtinIs[Tuple]},
		{Name: "isError", Func: builtinIs[*Error]},
		{Name: "isImmutable", Func: builtinIsImmutable},
		{Name: "isBuiltinFunction", Func: builtinIs[*BuiltinFunction]},
		{Name: "isCompiledFunction", Func: builtinIs[*CompiledFunction]},
		{Name: "isFunction", Func: builtinIsFunction},
		{Name: "isBuiltinModule", Func: builtinIs[*BuiltinModule]},

		{Name: "isHashable", Func: builtinIs[Hashable]},
		{Name: "isFreezable", Func: builtinIs[Freezable]},
		{Name: "isComparable", Func: builtinIs[Comparable]},
		{Name: "hasBinaryOp", Func: builtinIs[HasBinaryOp]},
		{Name: "hasUnaryOp", Func: builtinIs[HasUnaryOp]},
		{Name: "isIndexAccessible", Func: builtinIs[IndexAccessible]},
		{Name: "isIndexAssignable", Func: builtinIs[IndexAssignable]},
		{Name: "isFieldAccessible", Func: builtinIs[FieldAccessible]},
		{Name: "isFieldAssignable", Func: builtinIs[FieldAssignable]},
		{Name: "isSized", Func: builtinIs[Sized]},
		{Name: "isIndexable", Func: builtinIs[Indexable]},
		{Name: "isSliceable", Func: builtinIs[Sliceable]},
		{Name: "isConvertible", Func: builtinIs[Convertible]},
		{Name: "isCallable", Func: builtinIs[Callable]},
		{Name: "isIterable", Func: builtinIs[Iterable]},
		{Name: "isSequence", Func: builtinIs[Sequence]},
		{Name: "isIndexableSequence", Func: builtinIs[IndexableSequence]},
		{Name: "isMapping", Func: builtinIs[Mapping]},
	}
)

func builtinTypeName(_ *VM, args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	return String(args[0].TypeName()), nil
}

func builtinClone(_ *VM, args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	return args[0].Clone(), nil
}

func builtinLen(_ *VM, args ...Object) (Object, error) {
	var value Sized
	if err := UnpackArgs(args, "value", &value); err != nil {
		return nil, err
	}
	return Int(value.Len()), nil
}

func builtinAppend(_ *VM, args ...Object) (Object, error) {
	var (
		arr  *Array
		rest []Object
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

func builtinCopy(_ *VM, args ...Object) (Object, error) {
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

func builtinDelete(_ *VM, args ...Object) (Object, error) {
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
		if start > stop {
			return nil, fmt.Errorf("invalid delete indices: %d > %d", start, stop)
		}
		if start < 0 || start > n {
			return nil, fmt.Errorf("delete bounds out of range [%d:%d]", start, n)
		}
		if stop < 0 || stop > n {
			return nil, fmt.Errorf("delete bounds out of range [%d:%d] with len %d", start, stop, n)
		}
		if start == stop {
			return NewArray(nil), nil
		}
		deleted := slices.Clone(x.elems[start:stop])
		x.elems = slices.Delete(x.elems, start, stop)
		return NewArray(deleted), nil
	case *Map:
		if len(args) > 2 {
			return nil, &WrongNumArgumentsError{
				WantMin: 2,
				WantMax: 2,
				Got:     len(args),
			}
		}
		value, err := x.Delete(args[1])
		if err != nil {
			return nil, fmt.Errorf("failed to delete '%s' from map: %w", args[1].TypeName(), err)
		}
		return value, nil
	default:
		return nil, &InvalidArgumentTypeError{
			Name: "collection",
			Want: "array or map",
			Got:  x.TypeName(),
		}
	}
}

func builtinSplice(_ *VM, args ...Object) (Object, error) {
	var (
		arr         *Array
		start, stop int
		stopPtr     *int
		rest        []Object
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
	if start > stop {
		return nil, fmt.Errorf("invalid splice indices: %d > %d", start, stop)
	}
	if start < 0 || start > n {
		return nil, fmt.Errorf("splice bounds out of range [%d:%d]", start, n)
	}
	if stop < 0 || stop > n {
		return nil, fmt.Errorf("splice bounds out of range [%d:%d] with len %d", start, stop, n)
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

func builtinInsert(_ *VM, args ...Object) (Object, error) {
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
			rest  []Object
		)
		if err := UnpackArgs(args[1:], "index", &index, "...", &rest); err != nil {
			return nil, err
		}
		if err := x.checkMutable("insert into"); err != nil {
			return nil, err
		}
		n := len(x.elems)
		if index < 0 || index > n {
			return nil, fmt.Errorf("insert index %d out of range [:%d]", index, n)
		}
		x.elems = slices.Insert(x.elems, index, rest...)
		return Nil, nil
	case *Map:
		if len(args) != 3 {
			return nil, &WrongNumArgumentsError{
				WantMin: 2,
				WantMax: 3,
				Got:     len(args),
			}
		}
		var index, value Object
		if err := UnpackArgs(args[1:], "index", &index, "value", &value); err != nil {
			return nil, err
		}
		if err := x.IndexSet(index, value); err != nil {
			return nil, fmt.Errorf("failed to insert '%s' into map: %w", index.TypeName(), err)
		}
		return Nil, nil
	default:
		return nil, &InvalidArgumentTypeError{
			Name: "collection",
			Want: "array or map",
			Got:  x.TypeName(),
		}
	}
}

func builtinClear(_ *VM, args ...Object) (Object, error) {
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
	case *Map:
		if err := x.Clear(); err != nil {
			return nil, err
		}
	default:
		return nil, &InvalidArgumentTypeError{
			Name: "collection",
			Want: "array or map",
			Got:  x.TypeName(),
		}
	}
	return Nil, nil
}

func builtinHas(vm *VM, args ...Object) (Object, error) {
	var (
		container Container
		value     Object
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

func builtinFormat(_ *VM, args ...Object) (Object, error) {
	var (
		format string
		rest   []Object
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

func builtinRange(_ *VM, args ...Object) (Object, error) {
	var (
		start, stop int
		step        = 1
	)
	if err := UnpackArgs(args,
		"start", &start,
		"stop", &stop,
		"step?", &step,
	); err != nil {
		return nil, err
	}
	if step <= 0 {
		return nil, fmt.Errorf("invalid range step: must be > 0, got %d", step)
	}
	return &rangeType{
		start: start,
		stop:  stop,
		step:  step,
	}, nil
}

func builtinError(_ *VM, args ...Object) (_ Object, err error) {
	if len(args) < 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			Got:     len(args),
		}
	}
	var cause *Error
	if e, ok := args[0].(*Error); ok {
		if len(args) == 1 {
			return e, nil
		}
		args = args[1:]
		cause = e
	}
	var (
		format string
		rest   []Object
	)
	if err := UnpackArgs(args, "format", &format, "...", &rest); err != nil {
		return nil, err
	}
	var s string
	if len(rest) != 0 {
		s, err = Format(format, rest...)
		if err != nil {
			return nil, err
		}
	} else {
		s = format
	}
	return &Error{message: s, cause: cause}, nil
}

func builtinFail(v *VM, args ...Object) (Object, error) {
	var (
		format string
		rest   []Object
	)
	if err := UnpackArgs(args, "format", &format, "...", &rest); err != nil {
		return nil, err
	}
	if len(rest) == 0 {
		v.err = errors.New(format)
		return nil, nil
	}
	s, err := Format(string(format), rest...)
	if err != nil {
		return nil, err
	}
	v.err = errors.New(s)
	return nil, nil
}

func builtinMin(_ *VM, args ...Object) (Object, error) {
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

func builtinMax(_ *VM, args ...Object) (Object, error) {
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

func builtinConvert[T Object](_ *VM, args ...Object) (Object, error) {
	argsLen := len(args)
	if argsLen == 0 || argsLen > 2 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 2,
			Got:     len(args),
		}
	}
	var o T
	if err := Convert(&o, args[0]); err == nil {
		return o, nil
	}
	if argsLen == 2 {
		return args[1], nil
	}
	return Nil, nil
}

func builtinBytes(_ *VM, args ...Object) (Object, error) {
	argsLen := len(args)
	if argsLen == 0 || argsLen > 2 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 2,
			Got:     len(args),
		}
	}
	if i, ok := args[0].(Int); ok {
		return make(Bytes, i), nil
	}
	var b Bytes
	if err := Convert(&b, args[0]); err == nil {
		return b, nil
	}
	if argsLen == 2 {
		return args[1], nil
	}
	return Nil, nil
}

func builtinIs[T Object](_ *VM, args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	_, ok := args[0].(T)
	return Bool(ok), nil
}

func builtinIsFunction(_ *VM, args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	switch args[0].(type) {
	case *BuiltinFunction, *CompiledFunction:
		return Bool(true), nil
	}
	return Bool(false), nil
}

func builtinIsImmutable(_ *VM, args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	return Bool(!Mutable(args[0])), nil
}
