package tengo

import (
	"fmt"
	"slices"
)

var (
	BuiltinFuncs = []*BuiltinFunction{
		{Name: "copy", Func: builtinCopy},
		{Name: "len", Func: builtinLen},
		{Name: "append", Func: builtinAppend},
		{Name: "splice", Func: builtinSplice},
		{Name: "insert", Func: builtinInsert},
		{Name: "clear", Func: builtinClear},
		{Name: "delete", Func: builtinDelete},
		{Name: "string", Func: builtinConvert[String]},
		{Name: "int", Func: builtinConvert[Int]},
		{Name: "bool", Func: builtinConvert[Bool]},
		{Name: "float", Func: builtinConvert[Float]},
		{Name: "char", Func: builtinConvert[Char]},
		{Name: "bytes", Func: builtinConvert[Bytes]},
		{Name: "isInt", Func: builtinIs[Int]},
		{Name: "isFloat", Func: builtinIs[Float]},
		{Name: "isString", Func: builtinIs[String]},
		{Name: "isBool", Func: builtinIs[Bool]},
		{Name: "isChar", Func: builtinIs[Char]},
		{Name: "isBytes", Func: builtinIs[Bytes]},
		{Name: "isArray", Func: builtinIs[*Array]},
		{Name: "isImmutableArray", Func: builtinIsImmutableArray},
		{Name: "isMap", Func: builtinIs[*Map]},
		{Name: "isImmutableMap", Func: builtinIsImmutableMap},
		{Name: "isIterable", Func: builtinIs[Iterable]},
		{Name: "isError", Func: builtinIs[*Error]},
		{Name: "isUndefined", Func: builtinIs[UndefinedType]},
		{Name: "isFunction", Func: builtinIs[*CompiledFunction]},
		{Name: "isCallable", Func: builtinIs[Callable]},
		{Name: "typename", Func: builtinTypeName},
		{Name: "format", Func: builtinFormat},
		{Name: "range", Func: builtinRange},
		{Name: "error", Func: builtinError},
		{Name: "tuple", Func: builtinTuple},
	}
)

func builtinTypeName(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("want 1 argument, got %d", len(args))
	}
	return String(args[0].TypeName()), nil
}

func builtinIs[T Object](args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("want 1 argument, got %d", len(args))
	}
	_, ok := args[0].(T)
	return Bool(ok), nil
}

func builtinIsImmutableArray(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("want 1 argument, got %d", len(args))
	}
	a, ok := args[0].(*Array)
	if !ok {
		return False, nil
	}
	return Bool(a.immutable), nil
}

func builtinIsImmutableMap(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("want 1 argument, got %d", len(args))
	}
	m, ok := args[0].(*Map)
	if !ok {
		return False, nil
	}
	return Bool(m.ht.immutable), nil
}

func builtinLen(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("want 1 argument, got %d", len(args))
	}
	l, ok := args[0].(HasLen)
	if !ok {
		return nil, &ErrInvalidArgumentType{
			Name:     "value",
			Expected: "sized object",
			Found:    args[0].TypeName(),
		}
	}
	return Int(l.Len()), nil
}

func builtinAppend(args ...Object) (Object, error) {
	var (
		arr  *Array
		rest []Object
	)
	if err := UnpackArgs(args, "array", &arr, "...", &rest); err != nil {
		return nil, err
	}
	return &Array{
		elems:     append(arr.elems, rest...),
		immutable: arr.immutable,
		itercount: 0,
	}, nil
}

func builtinSplice(args ...Object) (Object, error) {
	var (
		arr         *Array
		start, stop int
		stopPtr     *int
		rest        []Object
	)
	if err := UnpackArgs(args,
		"array", &arr,
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

func builtinInsert(args ...Object) (Object, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("want at least 2 arguments, got %d", len(args))
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
		return Undefined, nil
	case *Map:
		if len(args) != 3 {
			return nil, fmt.Errorf("want 3 arguments, got %d", len(args))
		}
		var index, value Object
		if err := UnpackArgs(args[1:], "index", &index, "value", &value); err != nil {
			return nil, err
		}
		if err := x.IndexSet(index, value); err != nil {
			return nil, err
		}
		return Undefined, nil
	default:
		return nil, &ErrInvalidArgumentType{
			Name:     "collection",
			Expected: "array or dict",
			Found:    x.TypeName(),
		}
	}
}

func builtinErrorWrap(args ...Object) (ret Object, err error) {
	var (
		recv   = args[0].(*Error)
		format string
		rest   []Object
	)
	if err := UnpackArgs(args[1:], "error", &format, "...", &rest); err != nil {
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
	return &Error{message: s, cause: recv}, nil
}

func builtinClear(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("want 1 argument, got %d", len(args))
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
		return nil, &ErrInvalidArgumentType{
			Name:     "collection",
			Expected: "array or dict",
			Found:    x.TypeName(),
		}
	}
	return Undefined, nil
}

func builtinDelete(args ...Object) (Object, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("want at least 2 arguments, got %d", len(args))
	}
	switch x := args[0].(type) {
	case *Array:
		if len(args) > 3 {
			return nil, fmt.Errorf("want at most 3 arguments, got %d", len(args))
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
		if err := x.checkMutable("remove from"); err != nil {
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
		value, err := x.Delete(args[1])
		if err != nil {
			return nil, err
		}
		return value, nil
	default:
		return nil, &ErrInvalidArgumentType{
			Name:     "collection",
			Expected: "array or dict",
			Found:    x.TypeName(),
		}
	}
}

func builtinRange(args ...Object) (Object, error) {
	var (
		start, stop int64
		step        int64 = 1
	)
	if err := UnpackArgs(args,
		"start", &start,
		"stop", &stop,
		"step?", &step,
	); err != nil {
		return nil, err
	}
	if step < 0 {
		return nil, fmt.Errorf("invalid range step: must be > 0, got %d", step)
	}
	elems := make([]Object, 0, (stop-start)/step)
	if start <= stop {
		for i := start; i < stop; i += step {
			elems = append(elems, Int(i))
		}
	} else {
		for i := start; i > stop; i -= step {
			elems = append(elems, Int(i))
		}
	}
	return NewArray(elems), nil
}

func builtinFormat(args ...Object) (Object, error) {
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

func builtinCopy(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("want 1 argument, got %d", len(args))
	}
	return args[0].Copy(), nil
}

func builtinConvert[T Object](args ...Object) (Object, error) {
	argsLen := len(args)
	if argsLen == 0 || argsLen > 2 {
		return nil, fmt.Errorf("want 1 or 2 arguments, got %d", len(args))
	}
	var v T
	if err := Convert(&v, args[0]); err == nil {
		return v, nil
	}
	if argsLen == 2 {
		return args[1], nil
	}
	return Undefined, nil
}

func builtinError(args ...Object) (ret Object, err error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("want at least 1 argument, got %d", len(args))
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

func builtinTuple(args ...Object) (ret Object, err error) {
	return Tuple(args), nil
}
