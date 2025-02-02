package uuid

import (
	"fmt"
	"iter"

	"github.com/infastin/toy"
	"github.com/infastin/toy/hash"
	"github.com/infastin/toy/internal/xiter"
	"github.com/infastin/toy/token"

	"github.com/google/uuid"
)

var Module = &toy.BuiltinModule{
	Name: "uuid",
	Members: map[string]toy.Value{
		"UUID":  UUIDType,
		"uuid4": toy.NewBuiltinFunction("uuid.uuid4", v4Fn),
		"uuid7": toy.NewBuiltinFunction("uuid.uuid7", v7Fn),
	},
}

type UUID uuid.UUID

var UUIDType = toy.NewType[UUID]("uuid.UUID", func(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
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
		for i, v := range xiter.Enum(x.Elements()) {
			b, ok := v.(toy.Int)
			if !ok {
				return nil, &toy.InvalidArgumentTypeError{
					Name: "value",
					Sel:  fmt.Sprintf("[%d]", i),
					Want: "int",
					Got:  toy.TypeName(v),
				}
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

func (u UUID) Type() toy.ValueType { return UUIDType }
func (u UUID) String() string      { return fmt.Sprintf("uuid.UUID(%q)", uuid.UUID(u).String()) }
func (u UUID) IsFalsy() bool       { return u == UUID(uuid.Nil) }
func (u UUID) Clone() toy.Value    { return u }
func (u UUID) Hash() uint64        { return hash.Bytes(u[:]) }

func (u UUID) Compare(op token.Token, rhs toy.Value) (bool, error) {
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
		elems := make([]toy.Value, len(u))
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

func (u UUID) Len() int           { return len(u) }
func (u UUID) At(i int) toy.Value { return toy.Int(u[i]) }

func (u UUID) Items() []toy.Value {
	elems := make([]toy.Value, len(u))
	for i, b := range u {
		elems[i] = toy.Int(b)
	}
	return elems
}

func (u UUID) Elements() iter.Seq[toy.Value] {
	return func(yield func(toy.Value) bool) {
		for _, b := range u {
			if !yield(toy.Int(b)) {
				break
			}
		}
	}
}

func v4Fn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	u, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	return UUID(u), nil
}

func v7Fn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	u, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return UUID(u), nil
}
