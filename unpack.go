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
	// Unpack unpacks the given object into Unpacker.
	// If the given Object can't be unpacked into using Unpacker,
	// it is recommended to return InvalidValueTypeError.
	Unpack(o Object) error
}

// UnpackArgs unpacks function arguments into the supplied parameter variables.
// pairs is an alternating list of names and pointers to variables.
//
// The variable must either be one of Go's primitive types (int*, float*, bool and string),
// implement Unpacker or Object, be one of Toy's interfaces (Object, Iterable, etc.),
// or be a pointer to any of these.
//
// If the parameter name ends with "?", it is optional.
// If a parameter is marked optional, then all following parameters are
// implicitly optional whether or not they are marked.
//
// If the parameter name is "...", all remaining arguments
// are unpacked into the supplied pointer to []Object.
//
// If the variable implements Unpacker, its Unpack argument is called with the argument value,
// allowing an application to define its own argument validation and conversion.
//
// If the variable implements Value, UnpackArgs may call its Type()
// method while constructing the error message.
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
	if len(args) > nparams && !slices.Contains(pairs, "...") {
		i := 0
		for ; i < nparams; i++ {
			name := pairs[2*i].(string)
			if strings.HasSuffix(name, "?") {
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
			panic(fmt.Sprintf("expected *[]Object type for remaining arguments, got %T", pairs[2*i+1]))
		}
		if err := unpackArg(pairs[2*i+1], arg); err != nil {
			if e := (*InvalidValueTypeError)(nil); errors.As(err, &e) {
				err = &InvalidArgumentTypeError{
					Name: name,
					Want: e.Want,
					Got:  e.Got,
				}
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

// UnpackMapping unpacks Mapping entries into the supplied parameter variables
// with the names corresponding to the keys of the entries.
//
// If the parameter name ends with "?", it is optional.
// If a parameter is marked optional, then all following parameters are
// implicitly optional whether or not they are marked.
//
// If the parameter name is "...", all remaining entries
// are unpacked into the supplied pointer to []Tuple.
func UnpackMapping(m Mapping, pairs ...any) error {
	var defined big.Int
	nparams := len(pairs) / 2
	paramName := func(x any) string {
		name := x.(string)
		if name[len(name)-1] == '?' {
			name = name[:len(name)-1]
		}
		return name
	}
	var rest *[]Tuple
loop:
	for key, value := range Entries(m) {
		name, ok := key.(String)
		if !ok {
			return &InvalidKeyTypeError{
				Want: "string",
				Got:  TypeName(key),
			}
		}
		for i := 0; i < nparams; i++ {
			pName := paramName(pairs[2*i])
			if pName == "..." {
				if rest != nil {
					break
				}
				if p, ok := pairs[2*i+1].(*[]Tuple); ok {
					rest = p
					*rest = (*rest)[:0]
				} else {
					panic(fmt.Sprintf("expected *[]Tuple type for remaining arguments, got %T", pairs[2*i+1]))
				}
			} else if pName == string(name) {
				// found it
				defined.SetBit(&defined, i, 1)
				if err := unpackArg(pairs[2*i+1], value); err != nil {
					if e := (*InvalidValueTypeError)(nil); errors.As(err, &e) {
						err = &InvalidEntryValueTypeError{
							Name: pName,
							Want: e.Want,
							Got:  e.Got,
						}
					}
					return err
				}
				continue loop
			}
		}
		if rest == nil {
			return &UnexpectedEntryError{Name: string(name)}
		}
		*rest = append(*rest, Tuple{key, value})
	}
	for i := 0; i < nparams; i++ {
		name := pairs[2*i].(string)
		if name == "..." || strings.HasSuffix(name, "?") {
			break // optional
		}
		if defined.Bit(i) == 0 {
			return &MissingEntryError{Name: name}
		}
	}
	return nil
}

// UnpackMapArgs expects first function argument to be Map and then
// unpacks its entries into the supplied parameter variables
// with the names corresponding to the keys of the entries.
//
// If the parameter name ends with "?", it is optional.
// If a parameter is marked optional, then all following parameters are
// implicitly optional whether or not they are marked.
//
// If the parameter name is "...", all remaining entries
// are unpacked into the supplied pointer to []Tuple.
func UnpackMapArgs(args []Object, pairs ...any) error {
	if len(args) != 1 {
		return &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	m, ok := args[0].(*Map)
	if !ok {
		return &InvalidArgumentTypeError{
			Name: "args",
			Want: "map",
			Got:  TypeName(args[0]),
		}
	}
	var defined big.Int
	nparams := len(pairs) / 2
	paramName := func(x any) string {
		name := x.(string)
		if name[len(name)-1] == '?' {
			name = name[:len(name)-1]
		}
		return name
	}
	if m.Len() > nparams && !slices.Contains(pairs, "...") {
		i := 0
		for ; i < nparams; i++ {
			name := pairs[2*i].(string)
			if strings.HasSuffix(name, "?") {
				break // optional
			}
		}
		return &WrongNumArgumentsError{
			WantMin: i,
			WantMax: nparams,
			Got:     m.Len(),
		}
	}
	var rest *[]Tuple
loop:
	for key, value := range Entries(m) {
		name, ok := key.(String)
		if !ok {
			return &InvalidKeyTypeError{
				Want: "string",
				Got:  TypeName(key),
			}
		}
		for i := 0; i < nparams; i++ {
			pName := paramName(pairs[2*i])
			if pName == "..." {
				if rest != nil {
					break
				}
				if p, ok := pairs[2*i+1].(*[]Tuple); ok {
					rest = p
					*rest = (*rest)[:0]
				} else {
					panic(fmt.Sprintf("expected *[]Tuple type for remaining arguments, got %T", pairs[2*i+1]))
				}
			} else if pName == string(name) {
				// found it
				defined.SetBit(&defined, i, 1)
				if err := unpackArg(pairs[2*i+1], value); err != nil {
					if e := (*InvalidValueTypeError)(nil); errors.As(err, &e) {
						err = &InvalidArgumentTypeError{
							Name: pName,
							Want: e.Want,
							Got:  e.Got,
						}
					}
					return err
				}
				continue loop
			}
		}
		if rest == nil {
			return &UnexpectedArgumentError{Name: string(name)}
		}
		*rest = append(*rest, Tuple{key, value})
	}
	for i := 0; i < nparams; i++ {
		name := pairs[2*i].(string)
		if name == "..." || strings.HasSuffix(name, "?") {
			break // optional
		}
		// We needn't check the first m.Len().
		if i < m.Len() {
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
			return &InvalidValueTypeError{
				Want: "string",
				Got:  TypeName(o),
			}
		}
		*ptr = string(s)
	case *bool:
		b, ok := o.(Bool)
		if !ok {
			return &InvalidValueTypeError{
				Want: "bool",
				Got:  TypeName(o),
			}
		}
		*ptr = bool(b)
	case *int, *int8, *int16, *int32, *int64,
		*uint, *uint8, *uint16, *uint32, *uint64, *uintptr:
		i, ok := o.(Int)
		if !ok {
			return &InvalidValueTypeError{
				Want: "int",
				Got:  TypeName(o),
			}
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
			return &InvalidValueTypeError{
				Want: "float",
				Got:  TypeName(o),
			}
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
			return &InvalidValueTypeError{
				Want: "hashable",
				Got:  TypeName(o),
			}
		}
		*ptr = h
	case *Freezable:
		f, ok := o.(Freezable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "freezable",
				Got:  TypeName(o),
			}
		}
		*ptr = f
	case *Comparable:
		c, ok := o.(Comparable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "comparable",
				Got:  TypeName(o),
			}
		}
		*ptr = c
	case *HasBinaryOp:
		b, ok := o.(HasBinaryOp)
		if !ok {
			return &InvalidValueTypeError{
				Want: "object supporting binary operations",
				Got:  TypeName(o),
			}
		}
		*ptr = b
	case *HasUnaryOp:
		u, ok := o.(HasUnaryOp)
		if !ok {
			return &InvalidValueTypeError{
				Want: "object supporting unary operations",
				Got:  TypeName(o),
			}
		}
		*ptr = u
	case *IndexAccessible:
		i, ok := o.(IndexAccessible)
		if !ok {
			return &InvalidValueTypeError{
				Want: "index accesible",
				Got:  TypeName(o),
			}
		}
		*ptr = i
	case *IndexAssignable:
		i, ok := o.(IndexAssignable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "index assignable",
				Got:  TypeName(o),
			}
		}
		*ptr = i
	case *FieldAccessible:
		f, ok := o.(FieldAccessible)
		if !ok {
			return &InvalidValueTypeError{
				Want: "field accesible",
				Got:  TypeName(o),
			}
		}
		*ptr = f
	case *FieldAssignable:
		f, ok := o.(FieldAssignable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "field assignable",
				Got:  TypeName(o),
			}
		}
		*ptr = f
	case *Sized:
		s, ok := o.(Sized)
		if !ok {
			return &InvalidValueTypeError{
				Want: "sized",
				Got:  TypeName(o),
			}
		}
		*ptr = s
	case *Indexable:
		i, ok := o.(Indexable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "indexable",
				Got:  TypeName(o),
			}
		}
		*ptr = i
	case *Sliceable:
		s, ok := o.(Sliceable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "sliceable",
				Got:  TypeName(o),
			}
		}
		*ptr = s
	case *Convertible:
		c, ok := o.(Convertible)
		if !ok {
			return &InvalidValueTypeError{
				Want: "convertible",
				Got:  TypeName(o),
			}
		}
		*ptr = c
	case *Callable:
		f, ok := o.(Callable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "callable",
				Got:  TypeName(o),
			}
		}
		*ptr = f
	case *Container:
		c, ok := o.(Container)
		if !ok {
			return &InvalidValueTypeError{
				Want: "container",
				Got:  TypeName(o),
			}
		}
		*ptr = c
	case *Iterable:
		it, ok := o.(Iterable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "iterable",
				Got:  TypeName(o),
			}
		}
		*ptr = it
	case *Sequence:
		seq, ok := o.(Sequence)
		if !ok {
			return &InvalidValueTypeError{
				Want: "sequence",
				Got:  TypeName(o),
			}
		}
		*ptr = seq
	case *IndexableSequence:
		iseq, ok := o.(IndexableSequence)
		if !ok {
			return &InvalidValueTypeError{
				Want: "indexable sequence",
				Got:  TypeName(o),
			}
		}
		*ptr = iseq
	case *Mapping:
		m, ok := o.(Mapping)
		if !ok {
			return &InvalidValueTypeError{
				Want: "mapping",
				Got:  TypeName(o),
			}
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
			return &InvalidValueTypeError{
				Want: TypeName(paramVar.Interface().(Object)),
				Got:  TypeName(o),
			}
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
