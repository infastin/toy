package ast

import (
	"strings"

	"github.com/infastin/toy/token"
)

// Stmt represents a statement in the AST.
type Stmt interface {
	Node
	stmtNode()
}

// FuncBodyStmt represents a function body in the AST.
type FuncBodyStmt interface {
	Stmt
	funcBodyStmtNode()
}

// AssignStmt represents an assignment statement.
type AssignStmt struct {
	LHS      []Expr
	RHS      []Expr
	Token    token.Token
	TokenPos token.Pos
}

func (s *AssignStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *AssignStmt) Pos() token.Pos {
	return s.LHS[0].Pos()
}

// End returns the position of first character immediately after the node.
func (s *AssignStmt) End() token.Pos {
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
	From token.Pos
	To   token.Pos
}

func (s *BadStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *BadStmt) Pos() token.Pos {
	return s.From
}

// End returns the position of first character immediately after the node.
func (s *BadStmt) End() token.Pos {
	return s.To
}

func (s *BadStmt) String() string {
	return "<bad statement>"
}

// BlockStmt represents a block statement.
type BlockStmt struct {
	Stmts  []Stmt
	LBrace token.Pos
	RBrace token.Pos
}

func (s *BlockStmt) stmtNode()         {}
func (s *BlockStmt) funcBodyStmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *BlockStmt) Pos() token.Pos {
	return s.LBrace
}

// End returns the position of first character immediately after the node.
func (s *BlockStmt) End() token.Pos {
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

// ShortFuncBodyStmt represents a body of a shorthand function.
type ShortFuncBodyStmt struct {
	Expr Expr
}

func (s *ShortFuncBodyStmt) stmtNode()         {}
func (s *ShortFuncBodyStmt) funcBodyStmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ShortFuncBodyStmt) Pos() token.Pos {
	return s.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (s *ShortFuncBodyStmt) End() token.Pos {
	return s.Expr.End()
}

func (s *ShortFuncBodyStmt) String() string {
	return s.Expr.String()
}

// BranchStmt represents a branch statement.
type BranchStmt struct {
	Token    token.Token
	TokenPos token.Pos
	Label    *Ident
}

func (s *BranchStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *BranchStmt) Pos() token.Pos {
	return s.TokenPos
}

// End returns the position of first character immediately after the node.
func (s *BranchStmt) End() token.Pos {
	if s.Label != nil {
		return s.Label.End()
	}
	return token.Pos(int(s.TokenPos) + len(s.Token.String()))
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
	Semicolon token.Pos
	Implicit  bool
}

func (s *EmptyStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *EmptyStmt) Pos() token.Pos {
	return s.Semicolon
}

// End returns the position of first character immediately after the node.
func (s *EmptyStmt) End() token.Pos {
	if s.Implicit {
		return s.Semicolon
	}
	return s.Semicolon + 1
}

func (s *EmptyStmt) String() string {
	return ";"
}

// LabeledStmt represents a labeled statement.
type LabeledStmt struct {
	Label *Ident
	Colon token.Pos
	Stmt  Stmt
}

func (s *LabeledStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *LabeledStmt) Pos() token.Pos {
	return s.Label.Pos()
}

// End returns the position of first character immediately after the node.
func (s *LabeledStmt) End() token.Pos {
	return s.Stmt.End()
}

func (s *LabeledStmt) String() string {
	return s.Label.String() + ": " + s.Stmt.String()
}

// ExportStmt represents an export statement.
type ExportStmt struct {
	ExportPos token.Pos
	Result    Expr
}

func (s *ExportStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ExportStmt) Pos() token.Pos {
	return s.ExportPos
}

// End returns the position of first character immediately after the node.
func (s *ExportStmt) End() token.Pos {
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
func (s *ExprStmt) Pos() token.Pos {
	return s.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (s *ExprStmt) End() token.Pos {
	return s.Expr.End()
}

func (s *ExprStmt) String() string {
	return s.Expr.String()
}

// ForInStmt represents a for-in statement.
type ForInStmt struct {
	ForPos   token.Pos
	Key      *Ident
	Value    *Ident
	Iterable Expr
	Body     *BlockStmt
}

func (s *ForInStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ForInStmt) Pos() token.Pos {
	return s.ForPos
}

// End returns the position of first character immediately after the node.
func (s *ForInStmt) End() token.Pos {
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
	ForPos token.Pos
	Init   Stmt
	Cond   Expr
	Post   Stmt
	Body   *BlockStmt
}

func (s *ForStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ForStmt) Pos() token.Pos {
	return s.ForPos
}

// End returns the position of first character immediately after the node.
func (s *ForStmt) End() token.Pos {
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
	IfPos token.Pos
	Init  Stmt
	Cond  Expr
	Body  *BlockStmt
	Else  Stmt // else branch; or nil
}

func (s *IfStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *IfStmt) Pos() token.Pos {
	return s.IfPos
}

// End returns the position of first character immediately after the node.
func (s *IfStmt) End() token.Pos {
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
	TokenPos token.Pos
}

func (s *IncDecStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *IncDecStmt) Pos() token.Pos {
	return s.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (s *IncDecStmt) End() token.Pos {
	return token.Pos(int(s.TokenPos) + 2)
}

func (s *IncDecStmt) String() string {
	return s.Expr.String() + s.Token.String()
}

// ReturnStmt represents a return statement.
type ReturnStmt struct {
	ReturnPos token.Pos
	Results   []Expr
}

func (s *ReturnStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *ReturnStmt) Pos() token.Pos {
	return s.ReturnPos
}

// End returns the position of first character immediately after the node.
func (s *ReturnStmt) End() token.Pos {
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

// DeferStmt represents a defer statement.
type DeferStmt struct {
	DeferPos token.Pos
	CallExpr *CallExpr
}

func (s *DeferStmt) stmtNode() {}

// Pos returns the position of first character belonging to the node.
func (s *DeferStmt) Pos() token.Pos {
	return s.DeferPos
}

// End returns the position of first character immediately after the node.
func (s *DeferStmt) End() token.Pos {
	return s.CallExpr.End()
}

func (s *DeferStmt) String() string {
	return "defer " + s.CallExpr.String()
}
