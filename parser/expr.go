package parser

import (
	"strconv"
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
	LBrack   Pos
	RBrack   Pos
}

func (e *ArrayLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *ArrayLit) Pos() Pos {
	return e.LBrack
}

// End returns the position of first character immediately after the node.
func (e *ArrayLit) End() Pos {
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
	From Pos
	To   Pos
}

func (e *BadExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *BadExpr) Pos() Pos {
	return e.From
}

// End returns the position of first character immediately after the node.
func (e *BadExpr) End() Pos {
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
	TokenPos Pos
}

func (e *BinaryExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *BinaryExpr) Pos() Pos {
	return e.LHS.Pos()
}

// End returns the position of first character immediately after the node.
func (e *BinaryExpr) End() Pos {
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
	ValuePos Pos
	Literal  string
}

func (e *BoolLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *BoolLit) Pos() Pos {
	return e.ValuePos
}

// End returns the position of first character immediately after the node.
func (e *BoolLit) End() Pos {
	return Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *BoolLit) String() string {
	return e.Literal
}

// CallExpr represents a splat expression.
type SplatExpr struct {
	Ellipsis Pos
	Expr     Expr
}

func (e *SplatExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *SplatExpr) Pos() Pos {
	return e.Ellipsis
}

// End returns the position of first character immediately after the node.
func (e *SplatExpr) End() Pos {
	return e.Expr.End()
}

func (e *SplatExpr) String() string {
	return "..." + e.Expr.String()
}

// CallExpr represents a function call expression.
type CallExpr struct {
	Func   Expr
	LParen Pos
	Args   []Expr
	RParen Pos
}

func (e *CallExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *CallExpr) Pos() Pos {
	return e.Func.Pos()
}

// End returns the position of first character immediately after the node.
func (e *CallExpr) End() Pos {
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
	ValuePos Pos
	Literal  string
}

func (e *CharLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *CharLit) Pos() Pos {
	return e.ValuePos
}

// End returns the position of first character immediately after the node.
func (e *CharLit) End() Pos {
	return Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *CharLit) String() string {
	return e.Literal
}

// CondExpr represents a ternary conditional expression.
type CondExpr struct {
	Cond        Expr
	True        Expr
	False       Expr
	QuestionPos Pos
	ColonPos    Pos
}

func (e *CondExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *CondExpr) Pos() Pos {
	return e.Cond.Pos()
}

// End returns the position of first character immediately after the node.
func (e *CondExpr) End() Pos {
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
	ValuePos Pos
	Literal  string
}

func (e *FloatLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *FloatLit) Pos() Pos {
	return e.ValuePos
}

// End returns the position of first character immediately after the node.
func (e *FloatLit) End() Pos {
	return Pos(int(e.ValuePos) + len(e.Literal))
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
func (e *FuncLit) Pos() Pos {
	return e.Type.Pos()
}

// End returns the position of first character immediately after the node.
func (e *FuncLit) End() Pos {
	return e.Body.End()
}

func (e *FuncLit) String() string {
	return "fn" + e.Type.Params.String() + " " + e.Body.String()
}

// FuncType represents a function type definition.
type FuncType struct {
	FuncPos Pos
	Params  *IdentList
}

func (e *FuncType) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *FuncType) Pos() Pos {
	return e.FuncPos
}

// End returns the position of first character immediately after the node.
func (e *FuncType) End() Pos {
	return e.Params.End()
}

func (e *FuncType) String() string {
	return "fn" + e.Params.String()
}

// Ident represents an identifier.
type Ident struct {
	Name    string
	NamePos Pos
}

func (e *Ident) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *Ident) Pos() Pos {
	return e.NamePos
}

// End returns the position of first character immediately after the node.
func (e *Ident) End() Pos {
	return Pos(int(e.NamePos) + len(e.Name))
}

func (e *Ident) String() string {
	if e != nil {
		return e.Name
	}
	return nullRep
}

// ImmutableExpr represents an immutable expression.
type ImmutableExpr struct {
	Expr         Expr
	ImmutablePos Pos
	LParen       Pos
	RParen       Pos
}

func (e *ImmutableExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *ImmutableExpr) Pos() Pos {
	return e.ImmutablePos
}

// End returns the position of first character immediately after the node.
func (e *ImmutableExpr) End() Pos {
	return e.RParen
}

func (e *ImmutableExpr) String() string {
	return "immutable(" + e.Expr.String() + ")"
}

// TupleLit represents a tuple literal.
type TupleLit struct {
	Elements []Expr
	TuplePos Pos
	LParen   Pos
	RParen   Pos
}

func (e *TupleLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *TupleLit) Pos() Pos {
	return e.TuplePos
}

// End returns the position of first character immediately after the node.
func (e *TupleLit) End() Pos {
	return e.RParen
}

func (e *TupleLit) String() string {
	var b strings.Builder
	b.WriteString("tuple(")
	for i, elem := range e.Elements {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(elem.String())
	}
	b.WriteByte(')')
	return b.String()
}

// UnpackExpr represents an unpack expression.
type UnpackExpr struct {
	Expr      Expr
	UnpackPos Pos
	LParen    Pos
	RParen    Pos
}

func (e *UnpackExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *UnpackExpr) Pos() Pos {
	return e.UnpackPos
}

// End returns the position of first character immediately after the node.
func (e *UnpackExpr) End() Pos {
	return e.RParen
}

func (e *UnpackExpr) String() string {
	return "unpack(" + e.Expr.String() + ")"
}

// ImportExpr represents an import expression.
type ImportExpr struct {
	ModuleName string
	Token      token.Token
	TokenPos   Pos
}

func (e *ImportExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *ImportExpr) Pos() Pos {
	return e.TokenPos
}

// End returns the position of first character immediately after the node.
func (e *ImportExpr) End() Pos {
	// import("moduleName")
	return Pos(int(e.TokenPos) + 10 + len(e.ModuleName))
}

func (e *ImportExpr) String() string {
	return `import("` + e.ModuleName + `")`
}

// IndexExpr represents an index expression.
type IndexExpr struct {
	Expr   Expr
	LBrack Pos
	Index  Expr
	RBrack Pos
}

func (e *IndexExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *IndexExpr) Pos() Pos {
	return e.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (e *IndexExpr) End() Pos {
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
	ValuePos Pos
	Literal  string
}

func (e *IntLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *IntLit) Pos() Pos {
	return e.ValuePos
}

// End returns the position of first character immediately after the node.
func (e *IntLit) End() Pos {
	return Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *IntLit) String() string {
	return e.Literal
}

// MapElementLit represents a map element.
type MapElementLit struct {
	LBrack   Pos
	Key      Expr
	ColonPos Pos
	Value    Expr
}

func (e *MapElementLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *MapElementLit) Pos() Pos {
	if e.LBrack.IsValid() {
		return e.LBrack
	}
	return e.Key.Pos()
}

// End returns the position of first character immediately after the node.
func (e *MapElementLit) End() Pos {
	return e.Value.End()
}

func (e *MapElementLit) String() string {
	var b strings.Builder
	if e.LBrack.IsValid() {
		b.WriteByte('[')
		b.WriteString(e.Key.String())
		b.WriteByte(']')
	} else {
		key, _ := strconv.Unquote(e.Key.String())
		b.WriteString(key)
	}
	b.WriteString(": ")
	b.WriteString(e.Value.String())
	return b.String()
}

// MapLit represents a map literal.
type MapLit struct {
	LBrace   Pos
	Elements []*MapElementLit
	RBrace   Pos
}

func (e *MapLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *MapLit) Pos() Pos {
	return e.LBrace
}

// End returns the position of first character immediately after the node.
func (e *MapLit) End() Pos {
	return e.RBrace + 1
}

func (e *MapLit) String() string {
	var b strings.Builder
	b.WriteByte('{')
	for i, elem := range e.Elements {
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
	LParen Pos
	RParen Pos
}

func (e *ParenExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *ParenExpr) Pos() Pos {
	return e.LParen
}

// End returns the position of first character immediately after the node.
func (e *ParenExpr) End() Pos {
	return e.RParen + 1
}

func (e *ParenExpr) String() string {
	return "(" + e.Expr.String() + ")"
}

// SelectorExpr represents a selector expression.
type SelectorExpr struct {
	Expr Expr
	Sel  Expr
}

func (e *SelectorExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *SelectorExpr) Pos() Pos {
	return e.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (e *SelectorExpr) End() Pos {
	return e.Sel.End()
}

func (e *SelectorExpr) String() string {
	return e.Expr.String() + "." + e.Sel.String()
}

// SliceExpr represents a slice expression.
type SliceExpr struct {
	Expr   Expr
	LBrack Pos
	Low    Expr
	High   Expr
	RBrack Pos
}

func (e *SliceExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *SliceExpr) Pos() Pos {
	return e.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (e *SliceExpr) End() Pos {
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

// StringLit represents a string literal.
type StringLit struct {
	Value    string
	ValuePos Pos
	Literal  string
}

func (e *StringLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *StringLit) Pos() Pos {
	return e.ValuePos
}

// End returns the position of first character immediately after the node.
func (e *StringLit) End() Pos {
	return Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *StringLit) String() string {
	return e.Literal
}

// UnaryExpr represents an unary operator expression.
type UnaryExpr struct {
	Expr     Expr
	Token    token.Token
	TokenPos Pos
}

func (e *UnaryExpr) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *UnaryExpr) Pos() Pos {
	return e.Expr.Pos()
}

// End returns the position of first character immediately after the node.
func (e *UnaryExpr) End() Pos {
	return e.Expr.End()
}

func (e *UnaryExpr) String() string {
	return "(" + e.Token.String() + e.Expr.String() + ")"
}

// UndefinedLit represents an undefined literal.
type UndefinedLit struct {
	TokenPos Pos
}

func (e *UndefinedLit) exprNode() {}

// Pos returns the position of first character belonging to the node.
func (e *UndefinedLit) Pos() Pos {
	return e.TokenPos
}

// End returns the position of first character immediately after the node.
func (e *UndefinedLit) End() Pos {
	return e.TokenPos + 9 // len(undefined) == 9
}

func (e *UndefinedLit) String() string {
	return "undefined"
}
