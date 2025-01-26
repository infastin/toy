package uuid

import (
	"fmt"

	"github.com/infastin/toy"
	"github.com/infastin/toy/hash"
	"github.com/infastin/toy/token"

	"github.com/google/uuid"
)

var Module = &toy.BuiltinModule{
	Name: "uuid",
	Members: map[string]toy.Object{
		"UUID":      UUIDType,
		"uuid4":     toy.NewBuiltinFunction("uuid.uuid4", v4Fn),
		"uuid7":     toy.NewBuiltinFunction("uuid.uuid7", v7Fn),
		"parse":     toy.NewBuiltinFunction("uuid.parse", parseFn),
		"fromBytes": toy.NewBuiltinFunction("uuid.fromBytes", fromBytesFn),
	},
}

type UUID uuid.UUID

var UUIDType = toy.NewType[UUID]("uuid.UUID", func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	switch x := args[0].(type) {
	case toy.String:
		u, err := uuid.Parse(string(x))
		if err != nil {
			return nil, err
		}
		return UUID(u), nil
	case toy.Bytes:
		if len(x) != 16 {
			return nil, &toy.InvalidArgumentTypeError{
				Name: "value",
				Want: "bytes[16]",
				Got:  fmt.Sprintf("bytes[%d]", len(x)),
			}
		}
		return UUID(x), nil
	case toy.Sequence:
		if x.Len() != 16 {
			return nil, &toy.InvalidArgumentTypeError{
				Name: "value",
				Want: "sequence[int, 16]",
				Got:  fmt.Sprintf("sequence[object, %d]", x.Len()),
			}
		}
		var u UUID
		for i, v := range toy.Enumerate(x) {
			b, ok := v.(toy.Int)
			if !ok {
				return nil, fmt.Errorf("value[%d]: want 'int', got '%s'", i, toy.TypeName(v))
			}
			u[i] = byte(b)
		}
		return u, nil
	default:
		var u UUID
		if err := toy.Convert(&u, x); err != nil {
			return nil, err
		}
		return u, nil
	}
})

func (u UUID) Type() toy.ObjectType { return UUIDType }
func (u UUID) String() string       { return fmt.Sprintf("uuid.UUID(%q)", uuid.UUID(u).String()) }
func (u UUID) IsFalsy() bool        { return u == UUID(uuid.Nil) }
func (u UUID) Clone() toy.Object    { return u }
func (u UUID) Hash() uint64         { return hash.Bytes(u[:]) }

func (u UUID) Compare(op token.Token, rhs toy.Object) (bool, error) {
	y, ok := rhs.(UUID)
	if !ok {
		return false, toy.ErrInvalidOperation
	}
	switch op {
	case token.Equal:
		return u == y, nil
	case token.NotEqual:
		return u == y, nil
	}
	return false, toy.ErrInvalidOperation
}

func (u UUID) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String(uuid.UUID(u).String())
	case *toy.Bytes:
		*p = u[:]
	case **toy.Array:
		elems := make([]toy.Object, len(u))
		for i, b := range u {
			elems[i] = toy.Int(b)
		}
		*p = toy.NewArray(elems)
	case *toy.Tuple:
		tup := make(toy.Tuple, len(u))
		for i, b := range u {
			tup[i] = toy.Int(b)
		}
		*p = tup
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func (u UUID) Len() int            { return len(u) }
func (u UUID) At(i int) toy.Object { return toy.Int(u[i]) }

func (u UUID) Items() []toy.Object {
	elems := make([]toy.Object, len(u))
	for i, b := range u {
		elems[i] = toy.Int(b)
	}
	return elems
}

func (u UUID) Iterate() toy.Iterator { return &uuidIterator{u: u, i: 0} }

type uuidIterator struct {
	u UUID
	i int
}

var uuidIteratorType = toy.NewType[*uuidIterator]("uuid.UUID-iterator", nil)

func (it *uuidIterator) Type() toy.ObjectType { return uuidIteratorType }
func (it *uuidIterator) String() string       { return "<uuid.UUID-iterator>" }
func (it *uuidIterator) IsFalsy() bool        { return true }
func (it *uuidIterator) Clone() toy.Object    { return &uuidIterator{u: it.u, i: it.i} }

func (it *uuidIterator) Next(key, value *toy.Object) bool {
	if it.i < len(it.u) {
		if key != nil {
			*key = toy.Int(it.i)
		}
		if value != nil {
			*value = toy.Int(it.u[it.i])
		}
		it.i++
		return true
	}
	return false
}

func v4Fn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	u, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	return UUID(u), nil
}

func v7Fn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	u, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return UUID(u), nil
}

func parseFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s string
	if err := toy.UnpackArgs(args, "s", &s); err != nil {
		return nil, err
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{UUID(u), toy.Nil}, nil
}

func fromBytesFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var data toy.Bytes
	if err := toy.UnpackArgs(args, "data", &data); err != nil {
		return nil, err
	}
	u, err := uuid.FromBytes(data)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{UUID(u), toy.Nil}, nil
}
