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
	paramName := func(x any) string { // (no free variables)
		name := x.(string)
		if name[len(name)-1] == '?' {
			name = name[:len(name)-1]
		}
		return name
	}
	if !slices.Contains(pairs, "...") && len(args) > nparams {
		return fmt.Errorf("want at most %d argument(s), got %d", nparams, len(args))
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
			var e *ErrInvalidArgumentType
			if errors.As(err, &e) {
				e.Name = name
				e.Found = arg.TypeName()
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
			return fmt.Errorf("missing argument for '%s'", name)
		}
	}
	return nil
}

func unpackArg(ptr any, o Object) error {
	switch p := ptr.(type) {
	case Unpacker:
		return p.Unpack(o)
	case *Object:
		*p = o
	case *string:
		s, ok := o.(String)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "string"}
		}
		*p = string(s)
	case *bool:
		b, ok := o.(Bool)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "bool"}
		}
		*p = bool(b)
	case *int, *int8, *int16, *int32, *int64,
		*uint, *uint8, *uint16, *uint32, *uint64, *uintptr:
		i, ok := o.(Int)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "int"}
		}
		switch p := p.(type) {
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
			return &ErrInvalidArgumentType{Expected: "float"}
		}
		switch p := p.(type) {
		case *float32:
			*p = float32(f)
		case *float64:
			*p = float64(f)
		}
	case *Hashable:
		h, ok := o.(Hashable)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "hashable"}
		}
		*p = h
	case *Freezable:
		f, ok := o.(Freezable)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "freezable"}
		}
		*p = f
	case *Comparable:
		c, ok := o.(Comparable)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "comparable"}
		}
		*p = c
	case *HasBinaryOp:
		b, ok := o.(HasBinaryOp)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "object supporting binary operations"}
		}
		*p = b
	case *HasUnaryOp:
		u, ok := o.(HasUnaryOp)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "object supporting unary operations"}
		}
		*p = u
	case *IndexAccessible:
		i, ok := o.(IndexAccessible)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "index accesible"}
		}
		*p = i
	case *IndexAssignable:
		i, ok := o.(IndexAssignable)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "index assignable"}
		}
		*p = i
	case *FieldAccessible:
		f, ok := o.(FieldAccessible)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "field accesible"}
		}
		*p = f
	case *FieldAssignable:
		f, ok := o.(FieldAssignable)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "field assignable"}
		}
		*p = f
	case *Sized:
		s, ok := o.(Sized)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "sized"}
		}
		*p = s
	case *Indexable:
		i, ok := o.(Indexable)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "indexable"}
		}
		*p = i
	case *Sliceable:
		s, ok := o.(Sliceable)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "sliceable"}
		}
		*p = s
	case *Convertible:
		c, ok := o.(Convertible)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "convertible"}
		}
		*p = c
	case *Callable:
		f, ok := o.(Callable)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "callable"}
		}
		*p = f
	case *Iterable:
		it, ok := o.(Iterable)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "iterable"}
		}
		*p = it
	case *Sequence:
		seq, ok := o.(Sequence)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "sequence"}
		}
		*p = seq
	case *Mapping:
		m, ok := o.(Mapping)
		if !ok {
			return &ErrInvalidArgumentType{Expected: "mapping"}
		}
		*p = m
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
		paramVarType := paramVar.Type()
		// Maybe ptr is a pointer to a pointer.
		if paramVarType.Kind() == reflect.Pointer {
			// Unwrap ptr and call unpackArg recursively.
			if paramVar.IsNil() {
				elem := reflect.New(paramVarType.Elem())
				paramVar.Set(elem)
			}
			return unpackArg(paramVar.Interface(), o)
		}
		// Nothing worked, so we need to return an error.
		// For than we need to make sure that ptr points
		// to a value of a type that implements Object.
		if paramVarType.Implements(reflect.TypeFor[Object]()) {
			// It should be safe to call TypeName on potentially nil object.
			return &ErrInvalidArgumentType{Expected: paramVar.Interface().(Object).TypeName()}
		}
		// If *ptr doesn't implement Object then panic.
		panic(fmt.Sprintf("pointer element type does not implement Object: %T", ptr))
	}
	return nil
}
