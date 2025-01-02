package tengo

import "fmt"

var builtinFuncs = []*BuiltinFunction{
	{Name: "copy", Func: builtinCopy},
	{Name: "copy", Func: builtinCopy},
	{Name: "string", Func: builtinConvert[String]},
	{Name: "int", Func: builtinConvert[Int]},
	{Name: "bool", Func: builtinConvert[Bool]},
	{Name: "float", Func: builtinConvert[Float]},
	{Name: "char", Func: builtinConvert[Char]},
	{Name: "bytes", Func: builtinConvert[Bytes]},
	{Name: "is_int", Func: builtinIs[Int]},
	{Name: "is_float", Func: builtinIs[Float]},
	{Name: "is_string", Func: builtinIs[String]},
	{Name: "is_bool", Func: builtinIs[Bool]},
	{Name: "is_char", Func: builtinIs[Char]},
	{Name: "is_bytes", Func: builtinIs[Bytes]},
	{Name: "is_array", Func: builtinIs[*Array]},
	{Name: "is_immutable_array", Func: builtinIsImmutableArray},
	{Name: "is_map", Func: builtinIs[*Map]},
	{Name: "is_immutable_map", Func: builtinIsImmutableMap},
	{Name: "is_iterable", Func: builtinIs[Iterable]},
	{Name: "is_error", Func: builtinIs[*Error]},
	{Name: "is_undefined", Func: builtinIs[UndefinedType]},
	{Name: "is_function", Func: builtinIs[*CompiledFunction]},
	{Name: "is_callable", Func: builtinIs[Callable]},
	{Name: "type_name", Func: builtinTypeName},
	{Name: "format", Func: builtinFormat},
	{Name: "range", Func: builtinRange},
	{Name: "error", Func: builtinError},
}

// GetAllBuiltinFunctions returns all builtin function objects.
func GetAllBuiltinFunctions() []*BuiltinFunction {
	return append([]*BuiltinFunction{}, builtinFuncs...)
}

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
	return &Error{message: s}, nil
}
