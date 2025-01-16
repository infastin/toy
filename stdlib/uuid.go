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
		"uuid4":     &toy.BuiltinFunction{Name: "uuid.uuid4", Func: uuidV4},
		"uuid7":     &toy.BuiltinFunction{Name: "uuid.uuid7", Func: uuidV7},
		"parse":     &toy.BuiltinFunction{Name: "uuid.parse", Func: uuidParse},
		"fromBytes": &toy.BuiltinFunction{Name: "uuid.fromBytes", Func: uuidFromBytes},
	},
}

type UUID uuid.UUID

func (u UUID) TypeName() string { return "uuid.UUID" }
func (u UUID) String() string   { return fmt.Sprintf("uuid.UUID(%q)", uuid.UUID(u).String()) }
func (u UUID) IsFalsy() bool    { return u == UUID(uuid.Nil) }
func (u UUID) Copy() toy.Object { return u }
func (u UUID) Hash() uint64     { return hash.Bytes(u[:]) }

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
