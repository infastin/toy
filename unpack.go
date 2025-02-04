package toy

import (
	"fmt"
	"math/big"
	"reflect"
	"slices"
	"strings"

	"github.com/infastin/toy/internal/xiter"
)

var UnpackStructTag = "toy"

// An Unpacker defines custom argument unpacking behavior.
// Is is generally not recommended to define custom unpacking behavior on values,
// since it results in implicit behavior.
type Unpacker interface {
	// Unpack unpacks the given value into Unpacker.
	// If the given Value can't be unpacked into using Unpacker,
	// it is recommended to return InvalidValueTypeError.
	//
	// Client code should not call this method.
	// Instead, use the standalone Unpack function.
	Unpack(v Value) error
}

// UnpackArgs unpacks function arguments into the supplied parameter variables.
// pairs is an alternating list of names and pointers to variables.
//
// The variable must either be one of Go's primitive types (int*, float*, bool and string),
// implement Unpacker or Value, be one of Toy's interfaces (Value, Iterable, etc.),
// or be a pointer to any of these.
//
// If the parameter name ends with "?", it is optional.
// If a parameter is marked optional, then all following parameters are
// implicitly optional whether or not they are marked.
//
// If the parameter name is "...", all remaining arguments
// are unpacked into the supplied pointer to []Value.
//
// If the variable implements Unpacker, its Unpack argument is called with the argument value,
// allowing an application to define its own argument validation and conversion.
//
// If the variable implements Value, UnpackArgs may call its Type()
// method while constructing the error message.
func UnpackArgs(args []Value, pairs ...any) error {
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
			if p, ok := pairs[2*i+1].(*[]Value); ok {
				*p = args[i:]
				break
			}
			panic(fmt.Sprintf("expected *[]Value type for remaining arguments, got %T", pairs[2*i+1]))
		}
		if err := Unpack(pairs[2*i+1], arg); err != nil {
			if e, ok := err.(*InvalidValueTypeError); ok {
				err = &InvalidArgumentTypeError{
					Name: name,
					Sel:  e.Sel,
					Want: e.Want,
					Got:  e.Got,
				}
			} else {
				err = fmt.Errorf("invalid value for argument '%s': %w", name, err)
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

// Unpack tries to unpack the given Value into a pointer value.
func Unpack(ptr any, v Value) error {
	switch ptr := ptr.(type) {
	case Unpacker:
		return ptr.Unpack(v)
	case *Value:
		*ptr = v
	case *string:
		s, ok := v.(String)
		if !ok {
			return &InvalidValueTypeError{
				Want: "string",
				Got:  TypeName(v),
			}
		}
		*ptr = string(s)
	case *bool:
		b, ok := v.(Bool)
		if !ok {
			return &InvalidValueTypeError{
				Want: "bool",
				Got:  TypeName(v),
			}
		}
		*ptr = bool(b)
	case *int, *int8, *int16, *int32, *int64,
		*uint, *uint8, *uint16, *uint32, *uint64, *uintptr:
		i, ok := v.(Int)
		if !ok {
			return &InvalidValueTypeError{
				Want: "int",
				Got:  TypeName(v),
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
		f, ok := v.(Float)
		if !ok {
			return &InvalidValueTypeError{
				Want: "float",
				Got:  TypeName(v),
			}
		}
		switch p := ptr.(type) {
		case *float32:
			*p = float32(f)
		case *float64:
			*p = float64(f)
		}
	case *Hashable:
		h, ok := v.(Hashable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "hashable",
				Got:  TypeName(v),
			}
		}
		*ptr = h
	case *Freezable:
		f, ok := v.(Freezable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "freezable",
				Got:  TypeName(v),
			}
		}
		*ptr = f
	case *Comparable:
		c, ok := v.(Comparable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "comparable",
				Got:  TypeName(v),
			}
		}
		*ptr = c
	case *HasBinaryOp:
		b, ok := v.(HasBinaryOp)
		if !ok {
			return &InvalidValueTypeError{
				Want: "binary-op",
				Got:  TypeName(v),
			}
		}
		*ptr = b
	case *HasUnaryOp:
		u, ok := v.(HasUnaryOp)
		if !ok {
			return &InvalidValueTypeError{
				Want: "unary-op",
				Got:  TypeName(v),
			}
		}
		*ptr = u
	case *PropertyAccessible:
		i, ok := v.(PropertyAccessible)
		if !ok {
			return &InvalidValueTypeError{
				Want: "property-accessible",
				Got:  TypeName(v),
			}
		}
		*ptr = i
	case *PropertyAssignable:
		i, ok := v.(PropertyAssignable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "property-assignable",
				Got:  TypeName(v),
			}
		}
		*ptr = i
	case *Sized:
		s, ok := v.(Sized)
		if !ok {
			return &InvalidValueTypeError{
				Want: "sized",
				Got:  TypeName(v),
			}
		}
		*ptr = s
	case *IndexAccessible:
		i, ok := v.(IndexAccessible)
		if !ok {
			return &InvalidValueTypeError{
				Want: "index-accessible",
				Got:  TypeName(v),
			}
		}
		*ptr = i
	case *IndexAssignable:
		i, ok := v.(IndexAssignable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "index-assignable",
				Got:  TypeName(v),
			}
		}
		*ptr = i
	case *Sliceable:
		s, ok := v.(Sliceable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "sliceable",
				Got:  TypeName(v),
			}
		}
		*ptr = s
	case *Convertible:
		c, ok := v.(Convertible)
		if !ok {
			return &InvalidValueTypeError{
				Want: "convertible",
				Got:  TypeName(v),
			}
		}
		*ptr = c
	case *Callable:
		f, ok := v.(Callable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "callable",
				Got:  TypeName(v),
			}
		}
		*ptr = f
	case *Container:
		c, ok := v.(Container)
		if !ok {
			return &InvalidValueTypeError{
				Want: "container",
				Got:  TypeName(v),
			}
		}
		*ptr = c
	case *Iterable:
		it, ok := v.(Iterable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "iterable",
				Got:  TypeName(v),
			}
		}
		*ptr = it
	case *Sequence:
		seq, ok := v.(Sequence)
		if !ok {
			return &InvalidValueTypeError{
				Want: "sequence",
				Got:  TypeName(v),
			}
		}
		*ptr = seq
	case *KVIterable:
		it, ok := v.(KVIterable)
		if !ok {
			return &InvalidValueTypeError{
				Want: "kv-iterable",
				Got:  TypeName(v),
			}
		}
		*ptr = it
	case *Mapping:
		m, ok := v.(Mapping)
		if !ok {
			return &InvalidValueTypeError{
				Want: "mapping",
				Got:  TypeName(v),
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
		if reflect.TypeOf(v).AssignableTo(paramVar.Type()) {
			// *ptr = o is valid here.
			paramVar.Set(reflect.ValueOf(v))
			break
		}
		// If *ptr implements Value, return an error.
		if paramVar.Type().Implements(reflect.TypeFor[Value]()) {
			// It should be safe to call TypeName on potentially nil value.
			return &InvalidValueTypeError{
				Want: TypeName(paramVar.Interface().(Value)),
				Got:  TypeName(v),
			}
		}
		switch paramVar.Kind() {
		case reflect.Slice:
			return unpackToSlice(paramVar, v)
		case reflect.Array:
			return unpackToArray(paramVar, v)
		case reflect.Map:
			return unpackToMap(paramVar, v)
		case reflect.Struct:
			return unpackToStruct(paramVar, v)
		case reflect.Pointer:
			// Maybe ptr is a pointer to a pointer that implements Value.
			if paramVar.IsNil() {
				elem := reflect.New(paramVar.Type().Elem())
				paramVar.Set(elem)
			}
			// Unwrap ptr and call Unpack recursively.
			return Unpack(paramVar.Interface(), v)
		}
		// Nothing worked, panic.
		panic(fmt.Sprintf("pointer element type does not implement Value: %T", ptr))
	}
	return nil
}

// unpacks v into rv (go slice).
func unpackToSlice(rv reflect.Value, v Value) error {
	seq, ok := v.(Sequence)
	if !ok {
		return &InvalidValueTypeError{
			Want: "sequence",
			Got:  TypeName(v),
		}
	}
	rt := rv.Type()
	rte := rt.Elem()
	for i, elem := range xiter.Enum(seq.Elements()) {
		sv := reflect.New(rte)
		if err := Unpack(sv.Interface(), elem); err != nil {
			if e, ok := err.(*InvalidValueTypeError); ok {
				err = &InvalidValueTypeError{
					Sel:  fmt.Sprintf("[%d]%s", i, e.Sel),
					Want: e.Want,
					Got:  e.Got,
				}
			} else {
				err = fmt.Errorf("invalid value for '[%d]': %w", i, err)
			}
			return err
		}
		rv.Set(reflect.Append(rv, sv.Elem()))
	}
	return nil
}

// unpacks v into rv (go array).
func unpackToArray(rv reflect.Value, v Value) error {
	seq, ok := v.(Sequence)
	if !ok {
		return &InvalidValueTypeError{
			Want: "sequence",
			Got:  TypeName(v),
		}
	}
	rvLen := rv.Len()
	i := 0
	for elem := range seq.Elements() {
		if i == rvLen {
			i++
			break // too many elements
		}
		if err := Unpack(rv.Index(i).Addr().Interface(), elem); err != nil {
			if e, ok := err.(*InvalidValueTypeError); ok {
				err = &InvalidValueTypeError{
					Sel:  fmt.Sprintf("[%d]%s", i, e.Sel),
					Want: e.Want,
					Got:  e.Got,
				}
			} else {
				err = fmt.Errorf("invalid value for '[%d]': %w", i, err)
			}
			return err
		}
		i++
	}
	if i != rvLen { // wrong number of elements
		return &InvalidValueTypeError{
			Want: fmt.Sprintf("sequence[%d]", rvLen),
			Got:  fmt.Sprintf("%s[%d]", TypeName(v), seq.Len()),
		}
	}
	return nil
}

// unpacks v into rv (go map).
func unpackToMap(rv reflect.Value, v Value) error {
	mapping, ok := v.(Mapping)
	if !ok {
		return &InvalidValueTypeError{
			Want: "mapping",
			Got:  TypeName(v),
		}
	}
	rt := rv.Type()
	if rv.IsNil() {
		rv.Set(reflect.MakeMap(rt))
	}
	rtk := rt.Key()
	rte := rt.Elem()
	for key, value := range mapping.Entries() {
		kv := reflect.New(rtk)
		if err := Unpack(kv.Interface(), key); err != nil {
			if e, ok := err.(*InvalidValueTypeError); ok {
				err = &InvalidValueTypeError{
					Want: fmt.Sprintf("mapping[%s, ...]", e.Want),
					Got:  fmt.Sprintf("%s[%s, ...]", TypeName(v), e.Got),
				}
			} else {
				err = fmt.Errorf("invalid key '[%s]': %w", key.String(), err)
			}
			return err
		}
		ev := reflect.New(rte)
		if err := Unpack(ev.Interface(), value); err != nil {
			if e, ok := err.(*InvalidValueTypeError); ok {
				err = &InvalidValueTypeError{
					Sel:  fmt.Sprintf("[%s]%s", key.String(), e.Sel),
					Want: e.Want,
					Got:  e.Got,
				}
			} else {
				err = fmt.Errorf("invalid value for '[%s]': %w", key.String(), err)
			}
			return err
		}
		rv.SetMapIndex(kv.Elem(), ev.Elem())
	}
	return nil
}

// unpacks v into rv (struct).
func unpackToStruct(rv reflect.Value, v Value) error {
	mapping, ok := v.(Mapping)
	if !ok {
		return &InvalidValueTypeError{
			Want: "mapping",
			Got:  TypeName(v),
		}
	}

	fields := make(map[string]reflect.Value)
	required := make(map[string]struct{})
	var rest *[]Tuple

	var extract func(reflect.Value)
	extract = func(rv reflect.Value) {
		rt := rv.Type()
		for i := range rt.NumField() {
			field := rt.Field(i)
			name := field.Tag.Get(UnpackStructTag)
			if name == "-" {
				continue
			}
			if field.Anonymous && name == "" {
				fv := rv.Field(i)
				// check that field is a struct first
				ft := fv.Type()
				for ft.Kind() == reflect.Pointer {
					ft = ft.Elem()
				}
				if ft.Kind() != reflect.Struct {
					return
				}
				// dive deep and allocate space for data if necessary
				for fv.Kind() == reflect.Pointer {
					if fv.IsNil() {
						elem := reflect.New(fv.Type().Elem())
						fv.Set(elem)
					}
					fv = fv.Elem()
				}
				extract(fv)
			} else if name != "" && field.IsExported() {
				fv := rv.Field(i)
				if name == "..." {
					fvp := fv.Addr()
					p, ok := fvp.Interface().(*[]Tuple)
					if !ok {
						panic(fmt.Sprintf("expected *[]Tuple type for remaining key-value pairs, got %s", fvp.Type().String()))
					}
					rest = p
				} else {
					if name[len(name)-1] != '?' {
						required[name] = struct{}{}
					} else {
						name = name[:len(name)-1]
					}
					fields[name] = fv
				}
			}
		}
	}
	extract(rv)

	for key, value := range mapping.Entries() {
		keyStr, ok := key.(String)
		if !ok {
			if rest != nil {
				*rest = append(*rest, Tuple{key, value})
			}
			continue
		}
		fv, ok := fields[string(keyStr)]
		if !ok {
			if rest != nil {
				*rest = append(*rest, Tuple{key, value})
			}
			continue
		}
		delete(required, string(keyStr))
		if err := Unpack(fv.Addr().Interface(), value); err != nil {
			if e, ok := err.(*InvalidValueTypeError); ok {
				err = &InvalidValueTypeError{
					Sel:  fmt.Sprintf(".%s%s", string(keyStr), e.Sel),
					Want: e.Want,
					Got:  e.Got,
				}
			} else {
				err = fmt.Errorf("invalid value for '%s': %w", string(keyStr), err)
			}
			return err
		}
	}

	if len(required) != 0 {
		for key := range required {
			return fmt.Errorf("missing key '%s'", key)
		}
	}

	return nil
}
