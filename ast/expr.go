package ast

import (
	"strings"

	"github.com/infastin/toy/token"
)

// Expr represents an expression node in the AST.
type Expr interface {
	Node
	exprNode()
}

// ArrayLit represents an array literal.
type ArrayLit struct {
	Elements []Expr
	LBrack   token.Pos
	RBrack   token.Pos
}

func (e *ArrayLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *ArrayLit) Pos() token.Pos {
	return e.LBrack
}

// End returns the position of first character immediately after the node.
func (e *ArrayLit) End() token.Pos {
	return e.RBrack + 1
}

func (e *ArrayLit) String() string {
	var b strings.Builder
	b.WriteByte('[')
	for i, elem := range e.Elements {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(elem.String())
	}
	b.WriteByte(']')
	return b.String()
}

// BadExpr represents a bad expression.
type BadExpr struct {
	From token.Pos
	To   token.Pos
}

func (e *BadExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *BadExpr) Pos() token.Pos {
	return e.From
}

// End returns the position of first character immediately after the node.
func (e *BadExpr) End() token.Pos {
	return e.To
}

func (e *BadExpr) String() string {
	return "<bad expression>"
}

// BinaryExpr represents a binary operator expression.
type BinaryExpr struct {
	LHS      Expr
	RHS      Expr
	Token    token.Token
	TokenPos token.Pos
}

func (e *BinaryExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *BinaryExpr) Pos() token.Pos {
	return e.LHS.Pos()
}

// End returns the position of first character immediately after the node.
func (e *BinaryExpr) End() token.Pos {
	return e.RHS.End()
}

func (e *BinaryExpr) String() string {
	var b strings.Builder
	b.WriteByte('(')
	b.WriteString(e.LHS.String())
	b.WriteByte(' ')
	b.WriteString(e.Token.String())
	b.WriteByte(' ')
	b.WriteString(e.RHS.String())
	b.WriteByte(')')
	return b.String()
}

// BoolLit represents a boolean literal.
type BoolLit struct {
	Value    bool
	ValuePos token.Pos
	Literal  string
}

func (e *BoolLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *BoolLit) Pos() token.Pos {
	return e.ValuePos
}

// End returns the position of first character immediately after the node.
func (e *BoolLit) End() token.Pos {
	return token.Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *BoolLit) String() string {
	return e.Literal
}

// CallExpr represents a splat expression.
type SplatExpr struct {
	Ellipsis token.Pos
	Expr     Expr
}

func (e *SplatExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *SplatExpr) Pos() token.Pos {
	return e.Ellipsis
}

// End returns the position of first character immediately after the node.
func (e *SplatExpr) End() token.Pos {
	return e.Expr.End()
}

func (e *SplatExpr) String() string {
	return "..." + e.Expr.String()
}

// CallExpr represents a function call expression.
type CallExpr struct {
	Func   Expr
	LParen token.Pos
	Args   []Expr
	RParen token.Pos
}

func (e *CallExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *CallExpr) Pos() token.Pos {
	return e.Func.Pos()
}

// End returns the position of first character immediately after the node.
func (e *CallExpr) End() token.Pos {
	return e.RParen + 1
}

func (e *CallExpr) String() string {
	var b strings.Builder
	b.WriteString(e.Func.String())
	b.WriteByte('(')
	for i, arg := range e.Args {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(arg.String())
	}
	b.WriteByte(')')
	return b.String()
}

// CharLit represents a character literal.
type CharLit struct {
	Value    rune
	ValuePos token.Pos
	Literal  string
}

func (e *CharLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *CharLit) Pos() token.Pos {
	return e.ValuePos
}

// End returns the position of first character immediately after the node.
func (e *CharLit) End() token.Pos {
	return token.Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *CharLit) String() string {
	return e.Literal
}

// CondExpr represents a ternary conditional expression.
type CondExpr struct {
	Cond        Expr
	True        Expr
	False       Expr
	QuestionPos token.Pos
	ColonPos    token.Pos
}

func (e *CondExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *CondExpr) Pos() token.Pos {
	return e.Cond.Pos()
}

// End returns the position of first character immediately after the node.
func (e *CondExpr) End() token.Pos {
	return e.False.End()
}

func (e *CondExpr) String() string {
	var b strings.Builder
	b.WriteByte('(')
	b.WriteString(e.Cond.String())
	b.WriteString(" ? ")
	b.WriteString(e.True.String())
	b.WriteString(" : ")
	b.WriteString(e.False.String())
	b.WriteByte(')')
	return b.String()
}

// FloatLit represents a floating point literal.
type FloatLit struct {
	Value    float64
	ValuePos token.Pos
	Literal  string
}

func (e *FloatLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *FloatLit) Pos() token.Pos {
	return e.ValuePos
}

// End returns the position of first character immediately after the node.
func (e *FloatLit) End() token.Pos {
	return token.Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *FloatLit) String() string {
	return e.Literal
}

// FuncLit represents a function literal.
type FuncLit struct {
	Type *FuncType
	Body FuncBodyStmt
}

func (e *FuncLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *FuncLit) Pos() token.Pos {
	return e.Type.Pos()
}

// End returns the position of first character immediately after the node.
func (e *FuncLit) End() token.Pos {
	return e.Body.End()
}

func (e *FuncLit) String() string {
	return "fn" + e.Type.Params.String() + " " + e.Body.String()
}

// FuncType represents a function type definition.
type FuncType struct {
	FuncPos token.Pos
	Params  *IdentList
}

func (e *FuncType) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *FuncType) Pos() token.Pos {
	return e.FuncPos
}

// End returns the position of first character immediately after the node.
func (e *FuncType) End() token.Pos {
	return e.Params.End()
}

func (e *FuncType) String() string {
	return "fn" + e.Params.String()
}

// Ident represents an identifier.
type Ident struct {
	Name    string
	NamePos token.Pos
}

func (e *Ident) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *Ident) Pos() token.Pos {
	return e.NamePos
}

// End returns the position of first character immediately after the node.
func (e *Ident) End() token.Pos {
	return token.Pos(int(e.NamePos) + len(e.Name))
}

func (e *Ident) String() string {
	if e != nil {
		return e.Name
	}
	return nullRep
}

// ImportExpr represents an import expression.
type ImportExpr struct {
	ModuleName string
	ImportPos  token.Pos
	LParen     token.Pos
	RParen     token.Pos
}

func (e *ImportExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *ImportExpr) Pos() token.Pos {
	return e.ImportPos
}

// End returns the position of first character immediately after the node.
func (e *ImportExpr) End() token.Pos {
	return e.RParen + 1
}

func (e *ImportExpr) String() string {
	return `import("` + e.ModuleName + `")`
}

// TryExpr represents a try expression.
type TryExpr struct {
	TryPos   token.Pos
	CallExpr *CallExpr
}

func (e *TryExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *TryExpr) Pos() token.Pos {
	return e.TryPos
}

// End returns the position of first character immediately after the node.
func (e *TryExpr) End() token.Pos {
	return e.CallExpr.End()
}

func (e *TryExpr) String() string {
	return "try " + e.CallExpr.String()
}

// IndexExpr represents an index expression.
type IndexExpr struct {
	Expr   Expr
	LBrack token.Pos
	Index  Expr
	RBrack token.Pos
}

func (e *IndexExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *IndexExpr) Pos() token.Pos {
	return e.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (e *IndexExpr) End() token.Pos {
	return e.RBrack + 1
}

func (e *IndexExpr) String() string {
	var index string
	if e.Index != nil {
		index = e.Index.String()
	}
	return e.Expr.String() + "[" + index + "]"
}

// IntLit represents an integer literal.
type IntLit struct {
	Value    int64
	ValuePos token.Pos
	Literal  string
}

func (e *IntLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *IntLit) Pos() token.Pos {
	return e.ValuePos
}

// End returns the position of first character immediately after the node.
func (e *IntLit) End() token.Pos {
	return token.Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *IntLit) String() string {
	return e.Literal
}

// TableKeyExpr represents a table key expression.
type TableKeyExpr struct {
	LBrack token.Pos
	Expr   Expr
	RBrack token.Pos
}

func (e *TableKeyExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *TableKeyExpr) Pos() token.Pos {
	return e.LBrack
}

// End returns the position of first character immediately after the node.
func (e *TableKeyExpr) End() token.Pos {
	return e.RBrack + 1
}

func (e *TableKeyExpr) String() string {
	return "[" + e.Expr.String() + "]"
}

// TableElement represents a table element.
type TableElement struct {
	Key      Expr
	ColonPos token.Pos
	Value    Expr
}

func (e *TableElement) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *TableElement) Pos() token.Pos {
	return e.Key.Pos()
}

// End returns the position of first character immediately after the node.
func (e *TableElement) End() token.Pos {
	return e.Value.End()
}

func (e *TableElement) String() string {
	return e.Key.String() + ": " + e.Value.String()
}

// TableLit represents a table literal.
type TableLit struct {
	LBrace token.Pos
	Exprs  []Expr
	RBrace token.Pos
}

func (e *TableLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *TableLit) Pos() token.Pos {
	return e.LBrace
}

// End returns the position of first character immediately after the node.
func (e *TableLit) End() token.Pos {
	return e.RBrace + 1
}

func (e *TableLit) String() string {
	var b strings.Builder
	b.WriteByte('{')
	for i, elem := range e.Exprs {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(elem.String())
	}
	b.WriteByte('}')
	return b.String()
}

// ParenExpr represents a parenthesis wrapped expression.
type ParenExpr struct {
	Expr   Expr
	LParen token.Pos
	RParen token.Pos
}

func (e *ParenExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *ParenExpr) Pos() token.Pos {
	return e.LParen
}

// End returns the position of first character immediately after the node.
func (e *ParenExpr) End() token.Pos {
	return e.RParen + 1
}

func (e *ParenExpr) String() string {
	return "(" + e.Expr.String() + ")"
}

// SelectorExpr represents a selector expression.
type SelectorExpr struct {
	Expr Expr
	Sel  *Ident
}

func (e *SelectorExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *SelectorExpr) Pos() token.Pos {
	return e.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (e *SelectorExpr) End() token.Pos {
	return e.Sel.End()
}

func (e *SelectorExpr) String() string {
	return e.Expr.String() + "." + e.Sel.String()
}

// SliceExpr represents a slice expression.
type SliceExpr struct {
	Expr   Expr
	LBrack token.Pos
	Low    Expr
	High   Expr
	RBrack token.Pos
}

func (e *SliceExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *SliceExpr) Pos() token.Pos {
	return e.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (e *SliceExpr) End() token.Pos {
	return e.RBrack + 1
}

func (e *SliceExpr) String() string {
	var b strings.Builder
	b.WriteString(e.Expr.String())
	b.WriteByte('[')
	if e.Low != nil {
		b.WriteString(e.Low.String())
	}
	b.WriteByte(':')
	if e.High != nil {
		b.WriteString(e.High.String())
	}
	b.WriteByte(']')
	return b.String()
}

// StringFragment represents a string fragment.
type StringFragment struct {
	Value    string
	ValuePos token.Pos
}

func (e *StringFragment) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *StringFragment) Pos() token.Pos {
	return e.ValuePos
}

// End returns the position of first character immediately after the node.
func (e *StringFragment) End() token.Pos {
	return token.Pos(int(e.ValuePos) + len(e.Value))
}

func (e *StringFragment) String() string {
	return e.Value
}

// StringInterpolationExpr represents string interpolation in string literals.
type StringInterpolationExpr struct {
	LBrace token.Pos
	Expr   Expr
	RBrace token.Pos
}

func (e *StringInterpolationExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *StringInterpolationExpr) Pos() token.Pos {
	return e.LBrace
}

// End returns the position of first character immediately after the node.
func (e *StringInterpolationExpr) End() token.Pos {
	return e.RBrace + 1
}

func (e *StringInterpolationExpr) String() string {
	return "{" + e.Expr.String() + "}"
}

// StringLit represents a string literal.
type StringLit struct {
	Kind   token.Token
	LQuote token.Pos
	Exprs  []Expr
	RQuote token.Pos
}

func (e *StringLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *StringLit) Pos() token.Pos {
	return e.LQuote
}

// End returns the position of first character immediately after the node.
func (e *StringLit) End() token.Pos {
	return token.Pos(int(e.RQuote) + len(e.Kind.String()))
}

func (e *StringLit) String() string {
	var b strings.Builder
	b.WriteString(e.Kind.String())
	for _, expr := range e.Exprs {
		b.WriteString(expr.String())
	}
	b.WriteString(e.Kind.String())
	return b.String()
}

// UnaryExpr represents an unary operator expression.
type UnaryExpr struct {
	Expr     Expr
	Token    token.Token
	TokenPos token.Pos
}

func (e *UnaryExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *UnaryExpr) Pos() token.Pos {
	return e.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (e *UnaryExpr) End() token.Pos {
	return e.Expr.End()
}

func (e *UnaryExpr) String() string {
	return "(" + e.Token.String() + e.Expr.String() + ")"
}

// NilLit represents a nil literal.
type NilLit struct {
	TokenPos token.Pos
}

func (e *NilLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *NilLit) Pos() token.Pos {
	return e.TokenPos
}

// End returns the position of first character immediately after the node.
func (e *NilLit) End() token.Pos {
	return e.TokenPos + 3 // len(nil) == 3
}

func (e *NilLit) String() string {
	return "nil"
}
