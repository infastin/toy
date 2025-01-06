package toy

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"slices"
	"strings"
)

// An Unpacker defines custom argument unpacking behavior.
type Unpacker interface {
	Unpack(o Object) error
}

func UnpackArgs(args []Object, pairs ...any) error {
	var defined big.Int
	nparams := len(pairs) / 2
	paramName := func(x any) string {
		name := x.(string)
		if name[len(name)-1] == '?' {
			name = name[:len(name)-1]
		}
		return name
	}
	if !slices.Contains(pairs, "...") && len(args) > nparams {
		i := 0
		for ; i < nparams; i++ {
			name := pairs[2*i].(string)
			if name == "..." || strings.HasSuffix(name, "?") {
				break // optional
			}
		}
		return &WrongNumArgumentsError{
			WantMin: i,
			WantMax: nparams,
			Got:     len(args),
		}
	}
	for i, arg := range args {
		defined.SetBit(&defined, i, 1)
		name := paramName(pairs[2*i])
		if name == "..." {
			if p, ok := pairs[2*i+1].(*[]Object); ok {
				*p = args[i:]
				break
			}
			panic(fmt.Errorf("expected *[]Object type for variadic parameters, got %T", pairs[2*i+1]))
		}
		if err := unpackArg(pairs[2*i+1], arg); err != nil {
			var e *InvalidArgumentTypeError
			if errors.As(err, &e) {
				e.Name = name
				e.Got = arg.TypeName()
			}
			return err
		}
	}
	for i := 0; i < nparams; i++ {
		name := pairs[2*i].(string)
		if name == "..." || strings.HasSuffix(name, "?") {
			break // optional
		}
		// We needn't check the first len(args).
		if i < len(args) {
			continue
		}
		if defined.Bit(i) == 0 {
			return &MissingArgumentError{Name: name}
		}
	}
	return nil
}

func unpackArg(ptr any, o Object) error {
	switch ptr := ptr.(type) {
	case Unpacker:
		return ptr.Unpack(o)
	case *Object:
		*ptr = o
	case *string:
		s, ok := o.(String)
		if !ok {
			return &InvalidArgumentTypeError{Want: "string"}
		}
		*ptr = string(s)
	case *bool:
		b, ok := o.(Bool)
		if !ok {
			return &InvalidArgumentTypeError{Want: "bool"}
		}
		*ptr = bool(b)
	case *int, *int8, *int16, *int32, *int64,
		*uint, *uint8, *uint16, *uint32, *uint64, *uintptr:
		i, ok := o.(Int)
		if !ok {
			return &InvalidArgumentTypeError{Want: "int"}
		}
		switch p := ptr.(type) {
		case *int:
			*p = int(i)
		case *int8:
			*p = int8(i)
		case *int16:
			*p = int16(i)
		case *int32:
			*p = int32(i)
		case *int64:
			*p = int64(i)
		case *uint:
			*p = uint(i)
		case *uint8:
			*p = uint8(i)
		case *uint16:
			*p = uint16(i)
		case *uint32:
			*p = uint32(i)
		case *uint64:
			*p = uint64(i)
		case *uintptr:
			*p = uintptr(i)
		}
	case *float32, *float64:
		f, ok := o.(Float)
		if !ok {
			return &InvalidArgumentTypeError{Want: "float"}
		}
		switch p := ptr.(type) {
		case *float32:
			*p = float32(f)
		case *float64:
			*p = float64(f)
		}
	case *Hashable:
		h, ok := o.(Hashable)
		if !ok {
			return &InvalidArgumentTypeError{Want: "hashable"}
		}
		*ptr = h
	case *Freezable:
		f, ok := o.(Freezable)
		if !ok {
			return &InvalidArgumentTypeError{Want: "freezable"}
		}
		*ptr = f
	case *Comparable:
		c, ok := o.(Comparable)
		if !ok {
			return &InvalidArgumentTypeError{Want: "comparable"}
		}
		*ptr = c
	case *HasBinaryOp:
		b, ok := o.(HasBinaryOp)
		if !ok {
			return &InvalidArgumentTypeError{Want: "object supporting binary operations"}
		}
		*ptr = b
	case *HasUnaryOp:
		u, ok := o.(HasUnaryOp)
		if !ok {
			return &InvalidArgumentTypeError{Want: "object supporting unary operations"}
		}
		*ptr = u
	case *IndexAccessible:
		i, ok := o.(IndexAccessible)
		if !ok {
			return &InvalidArgumentTypeError{Want: "index accesible"}
		}
		*ptr = i
	case *IndexAssignable:
		i, ok := o.(IndexAssignable)
		if !ok {
			return &InvalidArgumentTypeError{Want: "index assignable"}
		}
		*ptr = i
	case *FieldAccessible:
		f, ok := o.(FieldAccessible)
		if !ok {
			return &InvalidArgumentTypeError{Want: "field accesible"}
		}
		*ptr = f
	case *FieldAssignable:
		f, ok := o.(FieldAssignable)
		if !ok {
			return &InvalidArgumentTypeError{Want: "field assignable"}
		}
		*ptr = f
	case *Sized:
		s, ok := o.(Sized)
		if !ok {
			return &InvalidArgumentTypeError{Want: "sized"}
		}
		*ptr = s
	case *Indexable:
		i, ok := o.(Indexable)
		if !ok {
			return &InvalidArgumentTypeError{Want: "indexable"}
		}
		*ptr = i
	case *Sliceable:
		s, ok := o.(Sliceable)
		if !ok {
			return &InvalidArgumentTypeError{Want: "sliceable"}
		}
		*ptr = s
	case *Convertible:
		c, ok := o.(Convertible)
		if !ok {
			return &InvalidArgumentTypeError{Want: "convertible"}
		}
		*ptr = c
	case *Callable:
		f, ok := o.(Callable)
		if !ok {
			return &InvalidArgumentTypeError{Want: "callable"}
		}
		*ptr = f
	case *Iterable:
		it, ok := o.(Iterable)
		if !ok {
			return &InvalidArgumentTypeError{Want: "iterable"}
		}
		*ptr = it
	case *Sequence:
		seq, ok := o.(Sequence)
		if !ok {
			return &InvalidArgumentTypeError{Want: "sequence"}
		}
		*ptr = seq
	case *Mapping:
		m, ok := o.(Mapping)
		if !ok {
			return &InvalidArgumentTypeError{Want: "mapping"}
		}
		*ptr = m
	default:
		// ptr must be a pointer.
		ptrv := reflect.ValueOf(ptr)
		if ptrv.Kind() != reflect.Ptr {
			panic(fmt.Sprintf("not a pointer: %T", ptr))
		}
		paramVar := ptrv.Elem()
		// Check if *ptr = o is valid.
		if reflect.TypeOf(o).AssignableTo(paramVar.Type()) {
			// *ptr = o is valid here.
			paramVar.Set(reflect.ValueOf(o))
			break
		}
		// If *ptr implements Object, return an error.
		if paramVar.Type().Implements(reflect.TypeFor[Object]()) {
			// It should be safe to call TypeName on potentially nil object.
			return &InvalidArgumentTypeError{Want: paramVar.Interface().(Object).TypeName()}
		}
		// Maybe ptr is a pointer to a pointer that implements Object.
		if paramVar.Type().Kind() == reflect.Pointer {
			// Unwrap ptr and call unpackArg recursively.
			if paramVar.IsNil() {
				elem := reflect.New(paramVar.Type().Elem())
				paramVar.Set(elem)
			}
			return unpackArg(paramVar.Interface(), o)
		}
		// Nothing worked, panic.
		panic(fmt.Sprintf("pointer element type does not implement Object: %T", ptr))
	}
	return nil
}
