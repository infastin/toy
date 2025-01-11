package stdlib

import (
	"fmt"

	"github.com/infastin/toy"
)

type EnumVariant struct {
	Name  string
	Value toy.Object
}

type Enum struct {
	Name     string
	Variants []EnumVariant
}

func (e *Enum) TypeName() string { return fmt.Sprintf("enum:%s", e.Name) }
func (e *Enum) String() string   { return fmt.Sprintf("<enum:%s>", e.Name) }
func (e *Enum) IsFalsy() bool    { return false }

func (e *Enum) Copy() toy.Object {
	variants := make([]EnumVariant, 0, len(e.Variants))
	for _, variant := range e.Variants {
		variants = append(variants, EnumVariant{Name: variant.Name, Value: variant.Value.Copy()})
	}
	return &Enum{Name: e.Name, Variants: variants}
}

func (e *Enum) FieldGet(name string) (toy.Object, error) {
	for _, variant := range e.Variants {
		if variant.Name == name {
			return variant.Value, nil
		}
	}
	return nil, toy.ErrNoSuchField
}
