package ast

import (
	"strings"

	"github.com/infastin/toy/token"
)

const (
	nullRep = "<null>"
)

// Node represents a node in the AST.
type Node interface {
	// Pos returns the position of first character belonging to the node.
	Pos() token.Pos
	// End returns the position of first character immediately after the node.
	End() token.Pos
	// String returns a string representation of the node.
	String() string
}

// IdentList represents a list of identifiers.
type IdentList struct {
	LParen       token.Pos
	List         []*Ident
	NumOptionals int
	VarArgs      bool
	RParen       token.Pos
}

// Pos returns the position of first character belonging to the node.
func (n *IdentList) Pos() token.Pos {
	if n.LParen.IsValid() {
		return n.LParen
	}
	if len(n.List) > 0 {
		return n.List[0].Pos()
	}
	return token.NoPos
}

// End returns the position of first character immediately after the node.
func (n *IdentList) End() token.Pos {
	if n.RParen.IsValid() {
		return n.RParen + 1
	}
	if l := len(n.List); l > 0 {
		return n.List[l-1].End()
	}
	return token.NoPos
}

func (n *IdentList) String() string {
	numParams := len(n.List)
	numRequired := numParams - n.NumOptionals
	if n.VarArgs {
		numRequired--
	}
	var b strings.Builder
	b.WriteByte('(')
	i := 0
	for ; i < numRequired; i++ {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(n.List[i].String())
	}
	for ; i < numParams; i++ {
		if i != 0 {
			b.WriteString(", ")
		}
		if i == numParams-1 {
			b.WriteString("...")
			b.WriteString(n.List[i].String())
		} else {
			b.WriteString(n.List[i].String())
			b.WriteByte('?')
		}
	}
	b.WriteByte(')')
	return b.String()
}
