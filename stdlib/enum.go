package stdlib

import (
	"fmt"

	"github.com/infastin/toy"
)

type Enum struct {
	Name     string
	Variants map[string]toy.Object
}

func (e *Enum) TypeName() string { return fmt.Sprintf("enum:%s", e.Name) }
func (e *Enum) String() string   { return fmt.Sprintf("<enum:%s>", e.Name) }
func (e *Enum) IsFalsy() bool    { return false }

func (e *Enum) Copy() toy.Object {
	variants := make(map[string]toy.Object)
	for name, variant := range e.Variants {
		variants[name] = variant.Copy()
	}
	return &Enum{Name: e.Name, Variants: variants}
}

func (e *Enum) FieldGet(name string) (toy.Object, error) {
	variant, ok := e.Variants[name]
	if !ok {
		return nil, toy.ErrNoSuchField
	}
	return variant, nil
}
