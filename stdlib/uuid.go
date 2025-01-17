package stdlib

import (
	"fmt"

	"github.com/infastin/toy"
	"github.com/infastin/toy/hash"
	"github.com/infastin/toy/token"

	"github.com/google/uuid"
)

var UUIDModule = &toy.BuiltinModule{
	Name: "uuid",
	Members: map[string]toy.Object{
		"UUID":      UUIDType,
		"uuid4":     toy.NewBuiltinFunction("uuid.uuid4", uuidV4),
		"uuid7":     toy.NewBuiltinFunction("uuid.uuid7", uuidV7),
		"parse":     toy.NewBuiltinFunction("uuid.parse", uuidParse),
		"fromBytes": toy.NewBuiltinFunction("uuid.fromBytes", uuidFromBytes),
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
		u, err := uuid.ParseBytes(x)
		if err != nil {
			return nil, err
		}
		return UUID(u), nil
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
		return false, toy.ErrInvalidOperator
	}
	switch op {
	case token.Equal:
		return u == y, nil
	case token.NotEqual:
		return u == y, nil
	}
	return false, toy.ErrInvalidOperator
}

func (u UUID) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String(uuid.UUID(u).String())
	case *toy.Bytes:
		*p = u[:]
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func uuidV4(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	u, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	return UUID(u), nil
}

func uuidV7(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	u, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return UUID(u), nil
}

func uuidParse(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func uuidFromBytes(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
