package tengo

import (
	"fmt"
	"slices"
)

var (
	BuiltinFuncs = []*BuiltinFunction{
		{Name: "copy", Func: builtinCopy},
		{Name: "len", Func: builtinLen},
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
	}

	ArrayMethods = map[string]*BuiltinFunction{
		"append":   {Name: "append", Func: builtinArrayAppend},
		"clear":    {Name: "clear", Func: builtinArrayClear},
		"index":    {Name: "index", Func: builtinArrayIndex},
		"contains": {},
		"insert":   {Name: "insert", Func: builtinArrayInsert},
		"pop":      {Name: "pop", Func: builtinArrayPop},
		"remove":   {Name: "remove", Func: builtinArrayRemove},
	}

	MapMethods = map[string]*BuiltinFunction{
		"clear":  {},
		"pop":    {},
		"keys":   {},
		"values": {},
		"items":  {},
	}

	StringMethods = map[string]*BuiltinFunction{
		"count":        {},
		"contains":     {},
		"containsAny":  {},
		"index":        {},
		"indexAny":     {},
		"lastIndex":    {},
		"lastIndexAny": {},
		"split":        {},
		"splitAfter":   {},
		"splitN":       {},
		"splitAfterN":  {},
		"fields":       {},
		"join":         {},
		"hasPrefix":    {},
		"hasSuffix":    {},
		"repeat":       {},
		"upper":        {},
		"lower":        {},
		"title":        {},
		"trim":         {},
		"trimLeft":     {},
		"trimRight":    {},
		"trimSpace":    {},
		"trimPrefix":   {},
		"trimSuffix":   {},
		"replace":      {},
		"replaceAll":   {},
		"quote":        {},
		"unquote":      {},
	}

	ErrorMethods = map[string]*BuiltinFunction{
		"wrap": {Name: "wrap", Func: builtinErrorWrap},
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

func builtinArrayAppend(args ...Object) (ret Object, err error) {
	recv := args[0].(*Array)
	args = args[1:]
	if err := recv.Append(args...); err != nil {
		return nil, err
	}
	return recv, nil
}

func builtinArrayClear(args ...Object) (ret Object, err error) {
	recv := args[0].(*Array)
	args = args[1:]
	if len(args) != 0 {
		return nil, fmt.Errorf("got %d arguments, want none", len(args))
	}
	if err := recv.Clear(); err != nil {
		return nil, err
	}
	return Undefined, nil
}

func builtinArrayIndex(args ...Object) (ret Object, err error) {
	var (
		recv = args[0].(*Array)
		x    Object
	)
	if err := UnpackArgs(args[1:], "x", &x); err != nil {
		return nil, err
	}
	for i, elem := range recv.Entries() {
		if eq, err := Equals(elem, x); err != nil {
			return nil, err
		} else if eq {
			return i, nil
		}
	}
	return Undefined, nil
}

func builtinArrayInsert(args ...Object) (ret Object, err error) {
	var (
		recv = args[0].(*Array)
		i    int
		rest []Object
	)
	if err := UnpackArgs(args[1:], "i", &i, "...", rest); err != nil {
		return nil, err
	}
	if i < 0 || i > len(recv.elems) {
		return nil, ErrIndexOutOfBounds
	}
	if err := recv.checkMutable("insert into"); err != nil {
		return nil, err
	}
	recv.elems = slices.Insert(recv.elems, i, rest...)
	return recv, nil
}

func builtinArrayPop(args ...Object) (ret Object, err error) {
	var (
		recv = args[0].(*Array)
		i    = len(recv.elems) - 1
	)
	if err := UnpackArgs(args[1:], "i?", &i); err != nil {
		return nil, err
	}
	if i < 0 || i >= len(recv.elems) {
		return nil, ErrIndexOutOfBounds
	}
	if err := recv.checkMutable("pop from"); err != nil {
		return nil, err
	}
	elem := recv.elems[i]
	recv.elems = slices.Delete(recv.elems, i, i+1)
	return elem, nil
}

func builtinArrayRemove(args ...Object) (ret Object, err error) {
	var (
		recv = args[0].(*Array)
		x    Object
	)
	if err := UnpackArgs(args[1:], "x", &x); err != nil {
		return nil, err
	}
	if err := recv.checkMutable("remove from"); err != nil {
		return nil, err
	}
	for i, elem := range enumerate(recv.Elements()) {
		if eq, err := Equals(elem, x); err != nil {
			return nil, err
		} else if eq {
			recv.elems = slices.Delete(recv.elems, i, i+1)
			return elem, nil
		}
	}
	return Undefined, nil
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
