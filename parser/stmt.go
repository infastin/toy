package parser

import (
	"strings"

	"github.com/infastin/toy/token"
)

// Stmt represents a statement in the AST.
type Stmt interface {
	Node
	stmtNode()
}

// BodyStmt represents a function body in the AST.
type BodyStmt interface {
	Stmt
	bodyStmtNode()
}

// AssignStmt represents an assignment statement.
type AssignStmt struct {
	LHS      []Expr
	RHS      []Expr
	Token    token.Token
	TokenPos Pos
}

func (s *AssignStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *AssignStmt) Pos() Pos {
	return s.LHS[0].Pos()
}

// End returns the position of first character immediately after the node.
func (s *AssignStmt) End() Pos {
	return s.RHS[len(s.RHS)-1].End()
}

func (s *AssignStmt) String() string {
	var b strings.Builder
	for i, elem := range s.LHS {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(elem.String())
	}
	b.WriteByte(' ')
	b.WriteString(s.Token.String())
	b.WriteByte(' ')
	for i, elem := range s.RHS {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(elem.String())
	}
	return b.String()
}

// BadStmt represents a bad statement.
type BadStmt struct {
	From Pos
	To   Pos
}

func (s *BadStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *BadStmt) Pos() Pos {
	return s.From
}

// End returns the position of first character immediately after the node.
func (s *BadStmt) End() Pos {
	return s.To
}

func (s *BadStmt) String() string {
	return "<bad statement>"
}

// BlockStmt represents a block statement.
type BlockStmt struct {
	Stmts  []Stmt
	LBrace Pos
	RBrace Pos
}

func (s *BlockStmt) stmtNode()     {}
func (s *BlockStmt) bodyStmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *BlockStmt) Pos() Pos {
	return s.LBrace
}

// End returns the position of first character immediately after the node.
func (s *BlockStmt) End() Pos {
	return s.RBrace + 1
}

func (s *BlockStmt) String() string {
	var b strings.Builder
	b.WriteByte('{')
	for i, e := range s.Stmts {
		if i != 0 {
			b.WriteString("; ")
		}
		b.WriteString(e.String())
	}
	b.WriteByte('}')
	return b.String()
}

// ShortBodyStmt represents a body of a shorthand function.
type ShortBodyStmt struct {
	Exprs []Expr
}

func (s *ShortBodyStmt) stmtNode()     {}
func (s *ShortBodyStmt) bodyStmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ShortBodyStmt) Pos() Pos {
	return s.Exprs[0].Pos()
}

// End returns the position of first character immediately after the node.
func (s *ShortBodyStmt) End() Pos {
	return s.Exprs[len(s.Exprs)-1].Pos()
}

func (s *ShortBodyStmt) String() string {
	var b strings.Builder
	for i, e := range s.Exprs {
		if i != 0 {
			b.WriteString("; ")
		}
		b.WriteString(e.String())
	}
	return b.String()
}

// BranchStmt represents a branch statement.
type BranchStmt struct {
	Token    token.Token
	TokenPos Pos
	Label    *Ident
}

func (s *BranchStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *BranchStmt) Pos() Pos {
	return s.TokenPos
}

// End returns the position of first character immediately after the node.
func (s *BranchStmt) End() Pos {
	if s.Label != nil {
		return s.Label.End()
	}
	return Pos(int(s.TokenPos) + len(s.Token.String()))
}

func (s *BranchStmt) String() string {
	var label string
	if s.Label != nil {
		label = " " + s.Label.Name
	}
	return s.Token.String() + label
}

// EmptyStmt represents an empty statement.
type EmptyStmt struct {
	Semicolon Pos
	Implicit  bool
}

func (s *EmptyStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *EmptyStmt) Pos() Pos {
	return s.Semicolon
}

// End returns the position of first character immediately after the node.
func (s *EmptyStmt) End() Pos {
	if s.Implicit {
		return s.Semicolon
	}
	return s.Semicolon + 1
}

func (s *EmptyStmt) String() string {
	return ";"
}

// ExportStmt represents an export statement.
type ExportStmt struct {
	ExportPos Pos
	Result    Expr
}

func (s *ExportStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ExportStmt) Pos() Pos {
	return s.ExportPos
}

// End returns the position of first character immediately after the node.
func (s *ExportStmt) End() Pos {
	return s.Result.End()
}

func (s *ExportStmt) String() string {
	return "export " + s.Result.String()
}

// ExprStmt represents an expression statement.
type ExprStmt struct {
	Expr Expr
}

func (s *ExprStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ExprStmt) Pos() Pos {
	return s.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (s *ExprStmt) End() Pos {
	return s.Expr.End()
}

func (s *ExprStmt) String() string {
	return s.Expr.String()
}

// ForInStmt represents a for-in statement.
type ForInStmt struct {
	ForPos   Pos
	Key      *Ident
	Value    *Ident
	Iterable Expr
	Body     *BlockStmt
}

func (s *ForInStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ForInStmt) Pos() Pos {
	return s.ForPos
}

// End returns the position of first character immediately after the node.
func (s *ForInStmt) End() Pos {
	return s.Body.End()
}

func (s *ForInStmt) String() string {
	var b strings.Builder
	b.WriteString("for ")
	b.WriteString(s.Key.String())
	if s.Value != nil {
		b.WriteString(", ")
		b.WriteString(s.Value.String())
	}
	b.WriteString(" in ")
	b.WriteString(s.Iterable.String())
	b.WriteByte(' ')
	b.WriteString(s.Body.String())
	return b.String()
}

// ForStmt represents a for statement.
type ForStmt struct {
	ForPos Pos
	Init   Stmt
	Cond   Expr
	Post   Stmt
	Body   *BlockStmt
}

func (s *ForStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ForStmt) Pos() Pos {
	return s.ForPos
}

// End returns the position of first character immediately after the node.
func (s *ForStmt) End() Pos {
	return s.Body.End()
}

func (s *ForStmt) String() string {
	var b strings.Builder
	b.WriteString("for ")
	if s.Init != nil || s.Post != nil {
		if s.Init != nil {
			b.WriteString(s.Init.String())
		}
		b.WriteString(" ; ")
	}
	if s.Cond != nil {
		b.WriteString(s.Cond.String())
	}
	if s.Init != nil || s.Post != nil {
		b.WriteString(" ; ")
		if s.Post != nil {
			b.WriteString(s.Post.String())
		}
	}
	b.WriteString(s.Body.String())
	return b.String()
}

// IfStmt represents an if statement.
type IfStmt struct {
	IfPos Pos
	Init  Stmt
	Cond  Expr
	Body  *BlockStmt
	Else  Stmt // else branch; or nil
}

func (s *IfStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *IfStmt) Pos() Pos {
	return s.IfPos
}

// End returns the position of first character immediately after the node.
func (s *IfStmt) End() Pos {
	if s.Else != nil {
		return s.Else.End()
	}
	return s.Body.End()
}

func (s *IfStmt) String() string {
	var b strings.Builder
	b.WriteString("if ")
	if s.Init != nil {
		b.WriteString(s.Init.String())
		b.WriteString("; ")
	}
	b.WriteString(s.Cond.String())
	b.WriteByte(' ')
	b.WriteString(s.Body.String())
	if s.Else != nil {
		b.WriteString(" else ")
		b.WriteString(s.Else.String())
	}
	return b.String()
}

// IncDecStmt represents increment or decrement statement.
type IncDecStmt struct {
	Expr     Expr
	Token    token.Token
	TokenPos Pos
}

func (s *IncDecStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *IncDecStmt) Pos() Pos {
	return s.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (s *IncDecStmt) End() Pos {
	return Pos(int(s.TokenPos) + 2)
}

func (s *IncDecStmt) String() string {
	return s.Expr.String() + s.Token.String()
}

// ReturnStmt represents a return statement.
type ReturnStmt struct {
	ReturnPos Pos
	Results   []Expr
}

func (s *ReturnStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ReturnStmt) Pos() Pos {
	return s.ReturnPos
}

// End returns the position of first character immediately after the node.
func (s *ReturnStmt) End() Pos {
	end := s.ReturnPos + 6
	for _, r := range s.Results {
		end = r.End()
	}
	return end
}

func (s *ReturnStmt) String() string {
	var b strings.Builder
	b.WriteString("return")
	if len(s.Results) != 0 {
		b.WriteByte(' ')
		for i, r := range s.Results {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(r.String())
		}
	}
	return b.String()
}
