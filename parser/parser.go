package parser

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/infastin/toy/ast"
	"github.com/infastin/toy/token"
)

type bailout struct{}

var stmtStart = map[token.Token]bool{
	token.Break:    true,
	token.Continue: true,
	token.For:      true,
	token.If:       true,
	token.Return:   true,
	token.Defer:    true,
	token.Export:   true,
}

// Error represents a parser error.
type Error struct {
	Pos token.FilePos
	Msg string
}

func (e Error) Error() string {
	if e.Pos.Filename != "" || e.Pos.IsValid() {
		return fmt.Sprintf("Parse Error: %s\n\tat %s", e.Msg, e.Pos)
	}
	return fmt.Sprintf("Parse Error: %s", e.Msg)
}

// ErrorList is a collection of parser errors.
type ErrorList []*Error

// Add adds a new parser error to the collection.
func (p *ErrorList) Add(pos token.FilePos, msg string) {
	*p = append(*p, &Error{pos, msg})
}

// Len returns the number of elements in the collection.
func (p ErrorList) Len() int {
	return len(p)
}

func (p ErrorList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p ErrorList) Less(i, j int) bool {
	e := &p[i].Pos
	f := &p[j].Pos
	if e.Filename != f.Filename {
		return e.Filename < f.Filename
	}
	if e.Line != f.Line {
		return e.Line < f.Line
	}
	if e.Column != f.Column {
		return e.Column < f.Column
	}
	return p[i].Msg < p[j].Msg
}

// Sort sorts the collection.
func (p ErrorList) Sort() {
	sort.Sort(p)
}

func (p ErrorList) Error() string {
	switch len(p) {
	case 0:
		return "no errors"
	case 1:
		return p[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", p[0], len(p)-1)
}

// Err returns an error.
func (p ErrorList) Err() error {
	if len(p) == 0 {
		return nil
	}
	return p
}

// Parser parses the Toy source files.
// It's based on Go's parser implementation.
type Parser struct {
	file      *token.File
	errors    ErrorList
	scanner   *Scanner
	pos       token.Pos
	token     token.Token
	tokenLit  string
	exprLevel int       // < 0: in control clause, >= 0: in expression
	syncPos   token.Pos // last sync position
	syncCount int       // number of advance calls without progress
	trace     bool
	indent    int
	traceOut  io.Writer
}

// NewParser creates a Parser.
func NewParser(file *token.File, src []byte, trace io.Writer) *Parser {
	p := &Parser{
		file:     file,
		trace:    trace != nil,
		traceOut: trace,
	}
	p.scanner = NewScanner(p.file, src,
		func(pos token.FilePos, msg string) {
			p.errors.Add(pos, msg)
		}, 0)
	p.next()
	return p
}

// ParseFile parses the source and returns an AST file unit.
func (p *Parser) ParseFile() (file *ast.File, err error) {
	defer func() {
		if e := recover(); e != nil {
			if _, ok := e.(bailout); !ok {
				panic(e)
			}
		}
		p.errors.Sort()
		err = p.errors.Err()
	}()

	if p.trace {
		defer untracep(tracep(p, "File"))
	}

	if p.errors.Len() > 0 {
		return nil, p.errors.Err()
	}

	stmts := p.parseStmtList()
	p.expect(token.EOF)
	if p.errors.Len() > 0 {
		return nil, p.errors.Err()
	}

	return &ast.File{
		InputFile: p.file,
		Stmts:     stmts,
	}, nil
}

func (p *Parser) parseExpr() ast.Expr {
	if p.trace {
		defer untracep(tracep(p, "Expression"))
	}
	expr := p.parseBinaryExpr(token.LowestPrec + 1)
	// ternary conditional expression
	if p.token == token.Question {
		return p.parseCondExpr(expr)
	}
	return expr
}

func (p *Parser) parseBinaryExpr(prec1 int) ast.Expr {
	if p.trace {
		defer untracep(tracep(p, "BinaryExpression"))
	}
	x := p.parseUnaryExpr()
	for {
		op, prec := p.token, p.token.Precedence()
		if prec < prec1 {
			return x
		}
		pos := p.expect(op)
		y := p.parseBinaryExpr(prec + 1)
		x = &ast.BinaryExpr{
			LHS:      x,
			RHS:      y,
			Token:    op,
			TokenPos: pos,
		}
	}
}

func (p *Parser) parseCondExpr(cond ast.Expr) ast.Expr {
	questionPos := p.expect(token.Question)
	trueExpr := p.parseExpr()
	colonPos := p.expect(token.Colon)
	falseExpr := p.parseExpr()
	return &ast.CondExpr{
		Cond:        cond,
		True:        trueExpr,
		False:       falseExpr,
		QuestionPos: questionPos,
		ColonPos:    colonPos,
	}
}

func (p *Parser) parseUnaryExpr() ast.Expr {
	if p.trace {
		defer untracep(tracep(p, "UnaryExpression"))
	}
	switch p.token {
	case token.Add, token.Sub, token.Not, token.Xor:
		pos, op := p.pos, p.token
		p.next()
		x := p.parseUnaryExpr()
		return &ast.UnaryExpr{
			Token:    op,
			TokenPos: pos,
			Expr:     x,
		}
	}
	return p.parsePrimaryExpr()
}

func (p *Parser) parsePrimaryExpr() ast.Expr {
	if p.trace {
		defer untracep(tracep(p, "PrimaryExpression"))
	}
	x := p.parseOperand()
loop:
	for {
		switch p.token {
		case token.Period:
			p.next()
			switch p.token {
			case token.Ident:
				x = p.parseSelector(x)
			default:
				pos := p.pos
				p.errorExpected(pos, "selector")
				p.advance(stmtStart)
				return &ast.BadExpr{From: pos, To: p.pos}
			}
		case token.LBrack:
			x = p.parseIndexOrSlice(x)
		case token.LParen:
			x = p.parseCall(x)
		default:
			break loop
		}
	}
	return x
}

// parseListElement parses an element of a call argument list
// or constructor argument list.
func (p *Parser) parseListElement() ast.Expr {
	if p.trace {
		defer untracep(tracep(p, "ListElement"))
	}
	if p.token != token.Ellipsis {
		return p.parseExpr()
	}
	ellipsis := p.pos
	p.next()
	x := p.parseExpr()
	return &ast.SplatExpr{
		Ellipsis: ellipsis,
		Expr:     x,
	}
}

func (p *Parser) parseCall(x ast.Expr) *ast.CallExpr {
	if p.trace {
		defer untracep(tracep(p, "Call"))
	}

	lparen := p.expect(token.LParen)
	p.exprLevel++

	var list []ast.Expr
	for p.token != token.RParen && p.token != token.EOF {
		list = append(list, p.parseListElement())
		if !p.expectComma("call argument") {
			break
		}
	}

	p.exprLevel--
	rparen := p.expect(token.RParen)

	return &ast.CallExpr{
		Func:   x,
		LParen: lparen,
		RParen: rparen,
		Args:   list,
	}
}

func (p *Parser) parseCallExpr(callType string) *ast.CallExpr {
	x := p.parsePrimaryExpr()
	if call, isCall := x.(*ast.CallExpr); isCall {
		return call
	}
	if _, isBad := x.(*ast.BadExpr); !isBad {
		// only report error if it's a new one
		p.error(p.safePos(x.End()), fmt.Sprintf("expression in %s must be function call", callType))
	}
	return nil
}

func (p *Parser) expectComma(want string) bool {
	if p.token == token.Comma {
		p.next()
		if p.token == token.Comma {
			p.errorExpected(p.pos, want)
			return false
		}
		return true
	}
	if p.token == token.Semicolon && p.tokenLit == "\n" {
		p.next()
	}
	return false
}

func (p *Parser) parseIndexOrSlice(x ast.Expr) ast.Expr {
	if p.trace {
		defer untracep(tracep(p, "IndexOrSlice"))
	}

	lbrack := p.expect(token.LBrack)
	p.exprLevel++

	var index [2]ast.Expr
	if p.token != token.Colon {
		index[0] = p.parseExpr()
	}
	numColons := 0
	if p.token == token.Colon {
		numColons++
		p.next()

		if p.token != token.RBrack && p.token != token.EOF {
			index[1] = p.parseExpr()
		}
	}

	p.exprLevel--
	rbrack := p.expect(token.RBrack)

	if numColons > 0 {
		// slice expression
		return &ast.SliceExpr{
			Expr:   x,
			LBrack: lbrack,
			RBrack: rbrack,
			Low:    index[0],
			High:   index[1],
		}
	}

	return &ast.IndexExpr{
		Expr:   x,
		LBrack: lbrack,
		RBrack: rbrack,
		Index:  index[0],
	}
}

func (p *Parser) parseSelector(x ast.Expr) ast.Expr {
	if p.trace {
		defer untracep(tracep(p, "Selector"))
	}
	return &ast.SelectorExpr{Expr: x, Sel: p.parseIdent()}
}

func (p *Parser) parseOperand() ast.Expr {
	if p.trace {
		defer untracep(tracep(p, "Operand"))
	}
	switch p.token {
	case token.Ident:
		return p.parseIdent()
	case token.Int:
		v, err := strconv.ParseInt(p.tokenLit, 0, 64)
		if err == strconv.ErrRange {
			p.error(p.pos, "number out of range")
		} else if err != nil {
			p.error(p.pos, "invalid integer")
		}
		x := &ast.IntLit{
			Value:    v,
			ValuePos: p.pos,
			Literal:  p.tokenLit,
		}
		p.next()
		return x
	case token.Float:
		v, err := strconv.ParseFloat(p.tokenLit, 64)
		if err == strconv.ErrRange {
			p.error(p.pos, "number out of range")
		} else if err != nil {
			p.error(p.pos, "invalid float")
		}
		x := &ast.FloatLit{
			Value:    v,
			ValuePos: p.pos,
			Literal:  p.tokenLit,
		}
		p.next()
		return x
	case token.Char:
		return p.parseCharLit()
	case token.DoubleQuote: // string literal
		return p.parseStringLit(token.DoubleQuote)
	case token.Backtick: // raw string literal
		return p.parseStringLit(token.Backtick)
	case token.DoubleSingleQuote: // indented string literal
		return p.parseStringLit(token.DoubleSingleQuote)
	case token.True:
		x := &ast.BoolLit{
			Value:    true,
			ValuePos: p.pos,
			Literal:  p.tokenLit,
		}
		p.next()
		return x
	case token.False:
		x := &ast.BoolLit{
			Value:    false,
			ValuePos: p.pos,
			Literal:  p.tokenLit,
		}
		p.next()
		return x
	case token.Nil:
		x := &ast.NilLit{TokenPos: p.pos}
		p.next()
		return x
	case token.Import:
		return p.parseImportExpr()
	case token.LParen:
		lparen := p.pos
		p.next()
		p.exprLevel++
		x := p.parseExpr()
		p.exprLevel--
		rparen := p.expect(token.RParen)
		return &ast.ParenExpr{
			LParen: lparen,
			Expr:   x,
			RParen: rparen,
		}
	case token.LBrack: // array literal
		return p.parseArrayLit()
	case token.LBrace: // map literal
		return p.parseMapLit()
	case token.Func: // function literal
		return p.parseFuncLit()
	default:
		p.errorExpected(p.pos, "operand")
	}
	pos := p.pos
	p.advance(stmtStart)
	return &ast.BadExpr{From: pos, To: p.pos}
}

func (p *Parser) parseImportExpr() ast.Expr {
	pos := p.expect(token.Import)
	lparen := p.expect(token.LParen)
	if p.token != token.DoubleQuote {
		p.errorExpected(p.pos, "module name")
		p.advance(stmtStart)
		return &ast.BadExpr{From: pos, To: p.pos}
	}
	moduleName := p.parseSimpleString(token.DoubleQuote)
	rparen := p.expect(token.RParen)
	return &ast.ImportExpr{
		ModuleName: moduleName,
		ImportPos:  pos,
		LParen:     lparen,
		RParen:     rparen,
	}
}

func (p *Parser) parseCharLit() ast.Expr {
	if n := len(p.tokenLit); n >= 3 {
		code, _, _, err := strconv.UnquoteChar(p.tokenLit[1:n-1], '\'')
		if err == nil {
			x := &ast.CharLit{
				Value:    code,
				ValuePos: p.pos,
				Literal:  p.tokenLit,
			}
			p.next()
			return x
		}
	}
	pos := p.pos
	p.error(pos, "illegal char literal")
	p.next()
	return &ast.BadExpr{
		From: pos,
		To:   p.pos,
	}
}

func (p *Parser) parseStringLit(kind token.Token) ast.Expr {
	if p.trace {
		defer untracep(tracep(p, "StringLit"))
	}

	var unescape func(string) string
	switch kind {
	case token.DoubleQuote:
		unescape = unescapeString
	case token.Backtick:
		unescape = unescapeRawString
	case token.DoubleSingleQuote:
		unescape = unescapeIndentedString
	}

	lquote := p.expect(kind)
	var exprs []ast.Expr
loop:
	for p.token != kind && p.token != token.EOF && p.token != token.Semicolon {
		switch p.token {
		case token.PlainText:
			exprs = append(exprs, &ast.PlainText{
				Value:    unescape(p.tokenLit),
				ValuePos: p.pos,
			})
			p.next()
		case token.LBrace:
			lbrace := p.pos
			p.next()
			x := p.parseExpr()
			if _, isBad := x.(*ast.BadExpr); isBad {
				break loop
			}
			rbrace := p.expect(token.RBrace)
			exprs = append(exprs, &ast.StringInterpolationExpr{
				LBrace: lbrace,
				Expr:   x,
				RBrace: rbrace,
			})
		}
	}
	rquote := p.expect(kind)

	return &ast.StringLit{
		Kind:   kind,
		LQuote: lquote,
		Exprs:  exprs,
		RQuote: rquote,
	}
}

func (p *Parser) parseSimpleString(kind token.Token) string {
	str := p.parseStringLit(kind).(*ast.StringLit)
	if len(str.Exprs) == 0 {
		return ""
	}
	if len(str.Exprs) == 1 {
		if plain, ok := str.Exprs[0].(*ast.PlainText); ok {
			return plain.Value
		}
	}
	p.error(str.Pos(), "cannot use string interpolation")
	return ""
}

func (p *Parser) parseFuncLit() ast.Expr {
	if p.trace {
		defer untracep(tracep(p, "FuncLit"))
	}
	typ := p.parseFuncType()
	p.exprLevel++
	body := p.parseBody()
	p.exprLevel--
	return &ast.FuncLit{
		Type: typ,
		Body: body,
	}
}

func (p *Parser) parseArrayLit() ast.Expr {
	if p.trace {
		defer untracep(tracep(p, "ArrayLit"))
	}

	lbrack := p.expect(token.LBrack)
	p.exprLevel++

	var elements []ast.Expr
	for p.token != token.RBrack && p.token != token.EOF {
		elements = append(elements, p.parseListElement())
		if !p.expectComma("array element") {
			break
		}
	}

	p.exprLevel--
	rbrack := p.expect(token.RBrack)

	return &ast.ArrayLit{
		Elements: elements,
		LBrack:   lbrack,
		RBrack:   rbrack,
	}
}

func (p *Parser) parseFuncType() *ast.FuncType {
	if p.trace {
		defer untracep(tracep(p, "FuncType"))
	}
	pos := p.expect(token.Func)
	params := p.parseIdentList()
	return &ast.FuncType{
		FuncPos: pos,
		Params:  params,
	}
}

func (p *Parser) parseBody() ast.FuncBodyStmt {
	if p.trace {
		defer untracep(tracep(p, "Body"))
	}
	if p.token == token.Arrow {
		p.next()
		return &ast.ShortFuncBodyStmt{Expr: p.parseExpr()}
	}
	lbrace := p.expect(token.LBrace)
	list := p.parseStmtList()
	rbrace := p.expect(token.RBrace)
	return &ast.BlockStmt{
		LBrace: lbrace,
		RBrace: rbrace,
		Stmts:  list,
	}
}

func (p *Parser) parseStmtList() (list []ast.Stmt) {
	if p.trace {
		defer untracep(tracep(p, "StatementList"))
	}
	for p.token != token.RBrace && p.token != token.EOF {
		list = append(list, p.parseStmt())
	}
	return list
}

func (p *Parser) parseIdent() *ast.Ident {
	pos := p.pos
	name := "_"
	if p.token == token.Ident {
		name = p.tokenLit
		p.next()
	} else {
		p.expect(token.Ident)
	}
	return &ast.Ident{
		NamePos: pos,
		Name:    name,
	}
}

func (p *Parser) parseIdentList() *ast.IdentList {
	if p.trace {
		defer untracep(tracep(p, "IdentList"))
	}

	var params []*ast.Ident
	lparen := p.expect(token.LParen)
	isVarArgs := false
	if p.token != token.RParen {
		if p.token == token.Ellipsis {
			isVarArgs = true
			p.next()
		}

		params = append(params, p.parseIdent())
		for !isVarArgs && p.token == token.Comma {
			p.next()
			if p.token == token.Ellipsis {
				isVarArgs = true
				p.next()
			}
			params = append(params, p.parseIdent())
		}
	}

	rparen := p.expect(token.RParen)
	return &ast.IdentList{
		LParen:  lparen,
		RParen:  rparen,
		VarArgs: isVarArgs,
		List:    params,
	}
}

func (p *Parser) parseStmt() (stmt ast.Stmt) {
	if p.trace {
		defer untracep(tracep(p, "Statement"))
	}
	switch p.token {
	case // simple statements
		token.Func, token.Ident, token.Int, token.Float, token.Char,
		token.DoubleQuote, token.Backtick, token.DoubleSingleQuote,
		token.True, token.False, token.Nil,
		token.Import, token.LParen, token.LBrace,
		token.LBrack, token.Add, token.Sub, token.Mul, token.And, token.Xor,
		token.Not:
		s := p.parseSimpleStmt(labelOk)
		// because of the required look-ahead, labeled statements are
		// parsed by parseSimpleStmt - don't expect a semicolon after
		// them
		if _, isLabeledStmt := s.(*ast.LabeledStmt); !isLabeledStmt {
			p.expectSemi()
		}
		return s
	case token.Return:
		return p.parseReturnStmt()
	case token.Defer:
		return p.parseDeferStmt()
	case token.Export:
		return p.parseExportStmt()
	case token.If:
		return p.parseIfStmt()
	case token.For:
		return p.parseForStmt()
	case token.Break, token.Continue:
		return p.parseBranchStmt(p.token)
	case token.Semicolon:
		s := &ast.EmptyStmt{Semicolon: p.pos, Implicit: p.tokenLit == "\n"}
		p.next()
		return s
	case token.RBrace:
		// semicolon may be omitted before a closing "}"
		return &ast.EmptyStmt{Semicolon: p.pos, Implicit: true}
	default:
		pos := p.pos
		p.errorExpected(pos, "statement")
		p.advance(stmtStart)
		return &ast.BadStmt{From: pos, To: p.pos}
	}
}

func (p *Parser) parseForStmt() ast.Stmt {
	if p.trace {
		defer untracep(tracep(p, "ForStmt"))
	}

	pos := p.expect(token.For)

	// for {}
	if p.token == token.LBrace {
		body := p.parseBlockStmt()
		p.expectSemi()

		return &ast.ForStmt{
			ForPos: pos,
			Body:   body,
		}
	}

	prevLevel := p.exprLevel
	p.exprLevel = -1

	var s1 ast.Stmt
	if p.token != token.Semicolon { // skipping init
		s1 = p.parseSimpleStmt(forInOk)
	}

	// for _ in seq {}            or
	// for value in seq {}        or
	// for key, value in seq {}
	if forInStmt, isForIn := s1.(*ast.ForInStmt); isForIn {
		forInStmt.ForPos = pos
		p.exprLevel = prevLevel
		forInStmt.Body = p.parseBlockStmt()
		p.expectSemi()
		return forInStmt
	}

	// for init; cond; post {}
	var s2, s3 ast.Stmt
	if p.token == token.Semicolon {
		p.next()
		if p.token != token.Semicolon {
			s2 = p.parseSimpleStmt(0) // cond
		}
		p.expect(token.Semicolon)
		if p.token != token.LBrace {
			s3 = p.parseSimpleStmt(0) // post
		}
	} else {
		// for cond {}
		s2 = s1
		s1 = nil
	}

	// body
	p.exprLevel = prevLevel
	body := p.parseBlockStmt()
	p.expectSemi()
	cond := p.makeExpr(s2, "condition expression")

	return &ast.ForStmt{
		ForPos: pos,
		Init:   s1,
		Cond:   cond,
		Post:   s3,
		Body:   body,
	}
}

func (p *Parser) parseBranchStmt(tok token.Token) ast.Stmt {
	if p.trace {
		defer untracep(tracep(p, "BranchStmt"))
	}

	pos := p.expect(tok)

	var label *ast.Ident
	if p.token == token.Ident {
		label = p.parseIdent()
	}
	p.expectSemi()
	return &ast.BranchStmt{
		Token:    tok,
		TokenPos: pos,
		Label:    label,
	}
}

func (p *Parser) parseIfStmt() ast.Stmt {
	if p.trace {
		defer untracep(tracep(p, "IfStmt"))
	}

	pos := p.expect(token.If)
	init, cond := p.parseIfHeader()
	body := p.parseBlockStmt()

	var elseStmt ast.Stmt
	if p.token == token.Else {
		p.next()
		switch p.token {
		case token.If:
			elseStmt = p.parseIfStmt()
		case token.LBrace:
			elseStmt = p.parseBlockStmt()
			p.expectSemi()
		default:
			p.errorExpected(p.pos, "if or {")
			elseStmt = &ast.BadStmt{From: p.pos, To: p.pos}
		}
	} else {
		p.expectSemi()
	}

	return &ast.IfStmt{
		IfPos: pos,
		Init:  init,
		Cond:  cond,
		Body:  body,
		Else:  elseStmt,
	}
}

func (p *Parser) parseBlockStmt() *ast.BlockStmt {
	if p.trace {
		defer untracep(tracep(p, "BlockStmt"))
	}
	lbrace := p.expect(token.LBrace)
	list := p.parseStmtList()
	rbrace := p.expect(token.RBrace)
	return &ast.BlockStmt{
		LBrace: lbrace,
		RBrace: rbrace,
		Stmts:  list,
	}
}

func (p *Parser) parseIfHeader() (init ast.Stmt, cond ast.Expr) {
	if p.token == token.LBrace {
		p.error(p.pos, "missing condition in if statement")
		cond = &ast.BadExpr{From: p.pos, To: p.pos}
		return
	}

	outer := p.exprLevel
	p.exprLevel = -1
	if p.token == token.Semicolon {
		p.error(p.pos, "missing init in if statement")
		return
	}
	init = p.parseSimpleStmt(0)

	var condStmt ast.Stmt
	switch p.token {
	case token.LBrace:
		condStmt = init
		init = nil
	case token.Semicolon:
		p.next()
		condStmt = p.parseSimpleStmt(0)
	default:
		p.error(p.pos, "missing condition in if statement")
	}

	if condStmt != nil {
		cond = p.makeExpr(condStmt, "boolean expression")
	}
	if cond == nil {
		cond = &ast.BadExpr{From: p.pos, To: p.pos}
	}
	p.exprLevel = outer

	return init, cond
}

func (p *Parser) makeExpr(s ast.Stmt, want string) ast.Expr {
	if s == nil {
		return nil
	}
	if es, isExpr := s.(*ast.ExprStmt); isExpr {
		return es.Expr
	}
	found := "simple statement"
	if _, isAss := s.(*ast.AssignStmt); isAss {
		found = "assignment"
	}
	p.error(s.Pos(), fmt.Sprintf("expected %s, found %s", want, found))
	return &ast.BadExpr{From: s.Pos(), To: p.safePos(s.End())}
}

func (p *Parser) parseReturnStmt() ast.Stmt {
	if p.trace {
		defer untracep(tracep(p, "ReturnStmt"))
	}

	pos := p.expect(token.Return)

	var results []ast.Expr
	if p.token != token.Semicolon && p.token != token.RBrace {
		results = p.parseExprList()
	}

	p.expectSemi()

	return &ast.ReturnStmt{
		ReturnPos: pos,
		Results:   results,
	}
}

func (p *Parser) parseDeferStmt() ast.Stmt {
	if p.trace {
		defer untracep(tracep(p, "DeferStmt"))
	}
	pos := p.expect(token.Defer)
	call := p.parseCallExpr("defer")
	p.expectSemi()
	if call == nil {
		return &ast.BadStmt{From: pos, To: pos + 5} // len("defer")
	}
	return &ast.DeferStmt{DeferPos: pos, CallExpr: call}
}

func (p *Parser) parseExportStmt() ast.Stmt {
	if p.trace {
		defer untracep(tracep(p, "ExportStmt"))
	}
	pos := p.expect(token.Export)
	x := p.parseExpr()
	p.expectSemi()
	return &ast.ExportStmt{
		ExportPos: pos,
		Result:    x,
	}
}

const (
	basic = iota
	labelOk
	forInOk
)

func (p *Parser) parseSimpleStmt(mode int) ast.Stmt {
	if p.trace {
		defer untracep(tracep(p, "SimpleStmt"))
	}

	x := p.parseExprList()

	switch p.token {
	case token.Assign, token.Define: // assignment statement
		pos, tok := p.pos, p.token
		p.next()
		return &ast.AssignStmt{
			LHS:      x,
			RHS:      p.parseExprList(),
			Token:    tok,
			TokenPos: pos,
		}
	case token.In:
		if mode == forInOk {
			p.next()
			y := p.parseExpr()
			var (
				key, value *ast.Ident
				ok         bool
			)
			switch len(x) {
			case 1:
				key = &ast.Ident{Name: "_", NamePos: x[0].Pos()}
				value, ok = x[0].(*ast.Ident)
				if !ok {
					p.errorExpected(x[0].Pos(), "identifier")
					value = &ast.Ident{Name: "_", NamePos: x[0].Pos()}
				}
			case 2:
				key, ok = x[0].(*ast.Ident)
				if !ok {
					p.errorExpected(x[0].Pos(), "identifier")
					key = &ast.Ident{Name: "_", NamePos: x[0].Pos()}
				}
				value, ok = x[1].(*ast.Ident)
				if !ok {
					p.errorExpected(x[1].Pos(), "identifier")
					value = &ast.Ident{Name: "_", NamePos: x[1].Pos()}
				}
			}
			return &ast.ForInStmt{
				Key:      key,
				Value:    value,
				Iterable: y,
			}
		}
	}

	if len(x) > 1 {
		p.errorExpected(x[0].Pos(), "1 expression")
		// continue with first expression
	}

	switch p.token {
	case token.Colon:
		// labeled statement
		colon := p.pos
		p.next()
		if label, isIdent := x[0].(*ast.Ident); mode == labelOk && isIdent {
			return &ast.LabeledStmt{Label: label, Colon: colon, Stmt: p.parseStmt()}
		}
		p.error(colon, "illegal label declaration")
		return &ast.BadStmt{From: x[0].Pos(), To: colon + 1}
	case token.Define,
		token.AddAssign, token.SubAssign, token.MulAssign, token.QuoAssign,
		token.RemAssign, token.AndAssign, token.OrAssign, token.XorAssign, token.AndNotAssign,
		token.ShlAssign, token.ShrAssign, token.NullishAssign:
		// define or assign statement
		pos, tok := p.pos, p.token
		p.next()
		y := p.parseExpr()
		return &ast.AssignStmt{
			LHS:      []ast.Expr{x[0]},
			RHS:      []ast.Expr{y},
			Token:    tok,
			TokenPos: pos,
		}
	case token.Inc, token.Dec:
		// increment or decrement statement
		s := &ast.IncDecStmt{Expr: x[0], Token: p.token, TokenPos: p.pos}
		p.next()
		return s
	}
	return &ast.ExprStmt{Expr: x[0]}
}

func (p *Parser) parseExprList() (list []ast.Expr) {
	if p.trace {
		defer untracep(tracep(p, "ExpressionList"))
	}
	list = append(list, p.parseExpr())
	for p.token == token.Comma {
		p.next()
		list = append(list, p.parseExpr())
	}
	return list
}

func (p *Parser) parseMapElementLit() *ast.MapElementLit {
	if p.trace {
		defer untracep(tracep(p, "MapElementLit"))
	}
	var key ast.Expr
	switch p.token {
	case token.Ident:
		key = &ast.Ident{
			Name:    p.tokenLit,
			NamePos: p.pos,
		}
		p.next()
	case token.LBrack:
		lbrack := p.pos
		p.next()
		expr := p.parseExpr()
		rbrack := p.expect(token.RBrack)
		key = &ast.MapKeyExpr{
			LBrack: lbrack,
			Expr:   expr,
			RBrack: rbrack,
		}
	default:
		p.errorExpected(p.pos, "map key")
	}
	colonPos := p.expect(token.Colon)
	valueExpr := p.parseExpr()
	return &ast.MapElementLit{
		Key:      key,
		ColonPos: colonPos,
		Value:    valueExpr,
	}
}

func (p *Parser) parseMapLit() *ast.MapLit {
	if p.trace {
		defer untracep(tracep(p, "MapLit"))
	}

	lbrace := p.expect(token.LBrace)
	p.exprLevel++

	var elements []*ast.MapElementLit
	for p.token != token.RBrace && p.token != token.EOF {
		elements = append(elements, p.parseMapElementLit())
		if !p.expectComma("map element") {
			break
		}
	}

	p.exprLevel--
	rbrace := p.expect(token.RBrace)
	return &ast.MapLit{
		LBrace:   lbrace,
		RBrace:   rbrace,
		Elements: elements,
	}
}

func (p *Parser) expect(token token.Token) token.Pos {
	pos := p.pos
	if p.token != token {
		p.errorExpected(pos, "'"+token.String()+"'")
	}
	p.next()
	return pos
}

func (p *Parser) expectSemi() {
	switch p.token {
	case token.RParen, token.RBrace:
		// semicolon is optional before a closing ')' or '}'
	case token.Comma:
		// permit a ',' instead of a ';' but complain
		p.errorExpected(p.pos, "';'")
		fallthrough
	case token.Semicolon:
		p.next()
	default:
		p.errorExpected(p.pos, "';'")
		p.advance(stmtStart)
	}
}

func (p *Parser) advance(to map[token.Token]bool) {
	for ; p.token != token.EOF; p.next() {
		if to[p.token] {
			if p.pos == p.syncPos && p.syncCount < 10 {
				p.syncCount++
				return
			}
			if p.pos > p.syncPos {
				p.syncPos = p.pos
				p.syncCount = 0
				return
			}
		}
	}
}

func (p *Parser) error(pos token.Pos, msg string) {
	filePos := p.file.Position(pos)
	n := len(p.errors)
	if n > 0 && p.errors[n-1].Pos.Line == filePos.Line {
		// discard errors reported on the same line
		return
	}
	if n > 10 {
		// too many errors; terminate early
		panic(bailout{})
	}
	p.errors.Add(filePos, msg)
}

func (p *Parser) errorExpected(pos token.Pos, msg string) {
	msg = "expected " + msg
	if pos == p.pos {
		// error happened at the current position;
		// make the error message more specific
		switch {
		case p.token == token.Semicolon && p.tokenLit == "\n":
			msg += ", found newline"
		case p.token.IsLiteral():
			msg += ", found " + p.tokenLit
		default:
			msg += ", found '" + p.token.String() + "'"
		}
	}
	p.error(pos, msg)
}

func (p *Parser) next() {
	if p.trace && p.pos.IsValid() {
		s := p.token.String()
		switch {
		case p.token.IsLiteral():
			p.printTrace(s, p.tokenLit)
		case p.token.IsOperator(), p.token.IsKeyword():
			p.printTrace(strconv.Quote(s))
		default:
			p.printTrace(s)
		}
	}
	p.token, p.tokenLit, p.pos = p.scanner.Scan()
}

func (p *Parser) printTrace(a ...interface{}) {
	const (
		dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
		n    = len(dots)
	)
	filePos := p.file.Position(p.pos)
	fmt.Fprintf(p.traceOut, "%5d: %5d:%3d: ", p.pos, filePos.Line, filePos.Column)
	i := 2 * p.indent
	for i > n {
		fmt.Fprint(p.traceOut, dots)
		i -= n
	}
	fmt.Fprint(p.traceOut, dots[0:i])
	fmt.Fprintln(p.traceOut, a...)
}

func (p *Parser) safePos(pos token.Pos) token.Pos {
	fileBase := p.file.Base
	fileSize := p.file.Size
	if int(pos) < fileBase || int(pos) > fileBase+fileSize {
		return token.Pos(fileBase + fileSize)
	}
	return pos
}

func unescapeString(s string) string {
	if strings.IndexByte(s, '\\') == -1 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	in := s
	for len(in) > 0 {
		if in[0] != '\\' {
			_, sz := utf8.DecodeRuneInString(in)
			b.WriteString(in[:sz])
			in = in[sz:]
			continue
		}
		switch ch := in[1]; ch {
		case '"':
			b.WriteByte('"')
			in = in[2:]
		case '{':
			b.WriteByte('{')
			in = in[2:]
		default:
			r, multibyte, rem, _ := strconv.UnquoteChar(in, '"')
			if r < utf8.RuneSelf || !multibyte {
				b.WriteByte(byte(r))
			} else {
				b.WriteRune(r)
			}
			in = rem
		}
	}
	return b.String()
}

func unescapeRawString(s string) string {
	if strings.IndexByte(s, '\\') == -1 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	in := s
	for len(in) > 0 {
		if in[0] != '\\' {
			_, sz := utf8.DecodeRuneInString(in)
			b.WriteString(in[:sz])
			in = in[sz:]
			continue
		}
		switch ch := in[1]; ch {
		case '`', '{', '\\':
			b.WriteByte(ch)
			in = in[2:]
		default:
			// skip
			in = in[2:]
		}
	}
	return b.String()
}

func unescapeIndentedString(s string) string {
	if strings.IndexByte(s, '\\') == -1 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	in := s
	for len(in) > 0 {
		if in[0] != '\\' {
			_, sz := utf8.DecodeRuneInString(in)
			b.WriteString(in[:sz])
			in = in[sz:]
			continue
		}
		switch ch := in[1]; ch {
		case '\'', '{', '\\':
			b.WriteByte(ch)
			in = in[2:]
		default:
			// skip
			in = in[2:]
		}
	}
	return b.String()
}

func tracep(p *Parser, msg string) *Parser {
	p.printTrace(msg, "(")
	p.indent++
	return p
}

func untracep(p *Parser) {
	p.indent--
	p.printTrace(")")
}
