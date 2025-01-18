package ast

import (
	"strings"

	"github.com/infastin/toy/token"
)

// File represents a file unit.
type File struct {
	InputFile *token.File
	Stmts     []Stmt
}

// Pos returns the position of first character belonging to the node.
func (n *File) Pos() token.Pos {
	return token.Pos(n.InputFile.Base)
}

// End returns the position of first character immediately after the node.
func (n *File) End() token.Pos {
	return token.Pos(n.InputFile.Base + n.InputFile.Size)
}

func (n *File) String() string {
	var stmts []string
	for _, e := range n.Stmts {
		stmts = append(stmts, e.String())
	}
	return strings.Join(stmts, "; ")
}
