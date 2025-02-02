package toy

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"

	"github.com/infastin/toy/ast"
	"github.com/infastin/toy/bytecode"
	"github.com/infastin/toy/parser"
	"github.com/infastin/toy/token"
)

// compilationScope represents a compiled instructions
// and the last two instructions that were emitted.
type compilationScope struct {
	instructions []byte
	sourceMap    map[int]token.Pos
	deferMap     []token.Pos
	labels       map[string]int
}

// loop represents a loop construct that
// the compiler uses to track the current loop.
type loop struct {
	label     string
	continues []int
	breaks    []int
}

// CompilerError represents a compiler error.
type CompilerError struct {
	FileSet *token.FileSet
	Node    ast.Node
	Err     error
}

func (e *CompilerError) Error() string {
	filePos := e.FileSet.Position(e.Node.Pos())
	return fmt.Sprintf("compile error: %s\n└─ at %s", e.Err.Error(), filePos)
}

// Compiler compiles the AST into a bytecode.
type Compiler struct {
	file            *token.File
	parent          *Compiler
	modulePath      string
	importDir       string
	importFileExt   []string
	constants       []Value
	symbolTable     *SymbolTable
	scopes          []compilationScope
	scopeIndex      int
	modules         ModuleGetter
	compiledModules map[string]*CompiledFunction
	allowFileImport bool
	loops           []*loop
	loopIndex       int
	trace           io.Writer
	indent          int
}

// NewCompiler creates a Compiler.
func NewCompiler(
	file *token.File,
	symbolTable *SymbolTable,
	constants []Value,
	modules ModuleGetter,
	trace io.Writer,
) *Compiler {
	mainScope := compilationScope{
		sourceMap: make(map[int]token.Pos),
		labels:    make(map[string]int),
	}

	// symbol table
	if symbolTable == nil {
		symbolTable = NewSymbolTable()
	}

	// add builtin values to the symbol table
	for i, v := range Universe {
		symbolTable.DefineBuiltin(i, v.name)
	}

	// builtin modules
	if modules == nil {
		modules = make(ModuleMap)
	}

	return &Compiler{
		file:            file,
		symbolTable:     symbolTable,
		constants:       constants,
		scopes:          []compilationScope{mainScope},
		scopeIndex:      0,
		loopIndex:       -1,
		trace:           trace,
		modules:         modules,
		compiledModules: make(map[string]*CompiledFunction),
		importFileExt:   []string{SourceFileExtDefault},
	}
}

// Compile compiles the AST node.
func (c *Compiler) Compile(node ast.Node) error {
	if c.trace != nil {
		if node != nil {
			defer untracec(tracec(c, fmt.Sprintf("%s (%s)",
				node.String(), reflect.TypeOf(node).Elem().Name())))
		} else {
			defer untracec(tracec(c, "<nil>"))
		}
	}

	switch node := node.(type) {
	case *ast.File:
		for _, stmt := range node.Stmts {
			if err := c.Compile(stmt); err != nil {
				return err
			}
		}
		// code optimization
		c.optimizeFunc(node)
	case *ast.ExprStmt:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		c.emit(node, bytecode.OpPop)
	case *ast.IncDecStmt:
		op := token.AddAssign
		if node.Token == token.Dec {
			op = token.SubAssign
		}
		return c.compileAssign(node, []ast.Expr{node.Expr},
			[]ast.Expr{&ast.IntLit{Value: 1}}, op)
	case *ast.ParenExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
	case *ast.BinaryExpr:
		if node.Token == token.LAnd || node.Token == token.LOr {
			return c.compileLogical(node)
		}

		if err := c.Compile(node.LHS); err != nil {
			return err
		}
		if err := c.Compile(node.RHS); err != nil {
			return err
		}

		switch node.Token {
		case token.Equal, token.NotEqual,
			token.Greater, token.GreaterEq,
			token.Less, token.LessEq:
			c.emit(node, bytecode.OpCompare, int(node.Token))
		case token.Add, token.Sub, token.Mul, token.Quo, token.Rem,
			token.And, token.Or, token.Xor, token.AndNot,
			token.Shl, token.Shr, token.Nullish:
			c.emit(node, bytecode.OpBinaryOp, int(node.Token))
		default:
			return c.errorf(node, "invalid binary operator: %s",
				node.Token.String())
		}
	case *ast.IntLit:
		c.emit(node, bytecode.OpConstant, c.addConstant(Int(node.Value)))
	case *ast.FloatLit:
		c.emit(node, bytecode.OpConstant, c.addConstant(Float(node.Value)))
	case *ast.BoolLit:
		if node.Value {
			c.emit(node, bytecode.OpTrue)
		} else {
			c.emit(node, bytecode.OpFalse)
		}
	case *ast.StringLit:
		if len(node.Exprs) == 0 {
			// empty string
			c.emit(node, bytecode.OpConstant, c.addConstant(String("")))
			return nil
		}
		if len(node.Exprs) == 1 {
			// maybe we don't need to build a string
			switch expr := node.Exprs[0].(type) {
			case *ast.PlainText:
				// we don't need to build a string here,
				// since it's already a string
				str := expr.Value
				if node.Kind == token.DoubleSingleQuote {
					// unindent indented string
					str = unindentString(str)
				}
				c.emit(expr, bytecode.OpConstant, c.addConstant(String(str)))
				return nil
			case *ast.StringInterpolationExpr:
				if str, ok := expr.Expr.(*ast.StringLit); ok {
					// we don't need to build a string here,
					// since it will be built by the nested string literal
					if err := c.Compile(str); err != nil {
						return err
					}
					return nil
				}
			}
		}
		// build a string here
		for _, expr := range node.Exprs {
			switch expr := expr.(type) {
			case *ast.PlainText:
				c.emit(expr, bytecode.OpConstant, c.addConstant(String(expr.Value)))
			case *ast.StringInterpolationExpr:
				if err := c.Compile(expr.Expr); err != nil {
					return err
				}
			}
		}
		var unindent int
		if node.Kind == token.DoubleSingleQuote {
			unindent = 1
		}
		c.emit(node, bytecode.OpString, len(node.Exprs), unindent)
	case *ast.CharLit:
		c.emit(node, bytecode.OpConstant, c.addConstant(Char(node.Value)))
	case *ast.NilLit:
		c.emit(node, bytecode.OpNull)
	case *ast.UnaryExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		switch node.Token {
		case token.Add, token.Sub, token.Not, token.Xor:
			c.emit(node, bytecode.OpUnaryOp, int(node.Token))
		default:
			return c.errorf(node, "invalid unary operator: %s", node.Token.String())
		}
	case *ast.IfStmt:
		// open new symbol table for the statement
		c.symbolTable = c.symbolTable.Fork(true)
		defer func() {
			c.symbolTable = c.symbolTable.Parent(false)
		}()

		if node.Init != nil {
			if err := c.Compile(node.Init); err != nil {
				return err
			}
		}
		if err := c.Compile(node.Cond); err != nil {
			return err
		}

		// first jump placeholder
		jumpPos1 := c.emit(node, bytecode.OpJumpFalsy, 0)
		if err := c.Compile(node.Body); err != nil {
			return err
		}
		if node.Else != nil {
			// second jump placeholder
			jumpPos2 := c.emit(node, bytecode.OpJump, 0)

			// update first jump offset
			curPos := len(c.currentInstructions())
			c.changeOperand(jumpPos1, curPos)
			if err := c.Compile(node.Else); err != nil {
				return err
			}

			// update second jump offset
			curPos = len(c.currentInstructions())
			c.changeOperand(jumpPos2, curPos)
		} else {
			// update first jump offset
			curPos := len(c.currentInstructions())
			c.changeOperand(jumpPos1, curPos)
		}
	case *ast.ForStmt:
		return c.compileForStmt(node, "")
	case *ast.ForInStmt:
		return c.compileForInStmt(node, "")
	case *ast.BranchStmt:
		if c.loopIndex < 0 {
			return c.errorf(node, "%s not allowed outside loop", node.Token.String())
		}

		var curLoop *loop
		if node.Label != nil {
			for i := c.loopIndex; i >= 0; i-- {
				curLoop = c.loops[i]
				if curLoop.label == node.Label.Name {
					break
				}
			}
			if curLoop.label != node.Label.Name {
				return c.errorf(node, "invalid %s label: %s", node.Token.String(), node.Label.Name)
			}
		} else {
			curLoop = c.loops[c.loopIndex]
		}

		switch node.Token {
		case token.Break:
			pos := c.emit(node, bytecode.OpJump, 0)
			curLoop.breaks = append(curLoop.breaks, pos)
		case token.Continue:
			pos := c.emit(node, bytecode.OpJump, 0)
			curLoop.continues = append(curLoop.continues, pos)
		default:
			panic(fmt.Errorf("invalid branch statement: %s", node.Token.String()))
		}
	case *ast.LabeledStmt:
		if _, ok := c.scopes[c.scopeIndex].labels[node.Label.Name]; ok {
			return c.errorf(node, "label '%s' already declared", node.Label.Name)
		} else {
			c.scopes[c.scopeIndex].labels[node.Label.Name] = len(c.currentInstructions())
		}
		switch stmt := node.Stmt.(type) {
		case *ast.ForStmt:
			if err := c.compileForStmt(stmt, node.Label.Name); err != nil {
				return err
			}
		case *ast.ForInStmt:
			if err := c.compileForInStmt(stmt, node.Label.Name); err != nil {
				return err
			}
		default:
			if err := c.Compile(node.Stmt); err != nil {
				return err
			}
		}
	case *ast.BlockStmt:
		if len(node.Stmts) == 0 {
			return nil
		}

		c.symbolTable = c.symbolTable.Fork(true)
		defer func() {
			c.symbolTable = c.symbolTable.Parent(false)
		}()

		for _, stmt := range node.Stmts {
			if err := c.Compile(stmt); err != nil {
				return err
			}
		}
	case *ast.ShortFuncBodyStmt:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		c.emit(node, bytecode.OpReturn, 1)
	case *ast.AssignStmt:
		err := c.compileAssign(node, node.LHS, node.RHS, node.Token)
		if err != nil {
			return err
		}
	case *ast.Ident:
		symbol, _, ok := c.symbolTable.Resolve(node.Name, false)
		if !ok {
			return c.errorf(node, "unresolved reference '%s'", node.Name)
		}
		switch symbol.Scope {
		case ScopeGlobal:
			c.emit(node, bytecode.OpGetGlobal, symbol.Index)
		case ScopeLocal:
			c.emit(node, bytecode.OpGetLocal, symbol.Index)
		case ScopeBuiltin:
			c.emit(node, bytecode.OpGetBuiltin, symbol.Index)
		case ScopeFree:
			c.emit(node, bytecode.OpGetFree, symbol.Index)
		}
	case *ast.ArrayLit:
		splat, err := c.compileListElements(node.Elements)
		if err != nil {
			return err
		}
		c.emit(node, bytecode.OpArray, len(node.Elements), splat)
	case *ast.TableLit:
		for _, elt := range node.Elements {
			switch key := elt.Key.(type) {
			case *ast.Ident:
				c.emit(key, bytecode.OpConstant, c.addConstant(String(key.Name)))
			case *ast.TableKeyExpr:
				if err := c.Compile(key.Expr); err != nil {
					return err
				}
			}
			if err := c.Compile(elt.Value); err != nil {
				return err
			}
		}
		c.emit(node, bytecode.OpTable, len(node.Elements))
	case *ast.SelectorExpr: // selector on RHS side
		if err := c.compileSelectorExpr(node, false); err != nil {
			return err
		}
	case *ast.IndexExpr:
		if err := c.compileIndexExpr(node, false); err != nil {
			return err
		}
	case *ast.SliceExpr:
		var indices int
		if node.Low != nil {
			if err := c.Compile(node.Low); err != nil {
				return err
			}
			indices |= 0x1
		}
		if node.High != nil {
			if err := c.Compile(node.High); err != nil {
				return err
			}
			indices |= 0x2
		}
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		c.emit(node, bytecode.OpSliceIndex, indices)
	case *ast.FuncLit:
		c.enterScope()

		for _, p := range node.Type.Params.List {
			// maybe such parameter has been already defined
			_, depth, exists := c.symbolTable.Resolve(p.Name, false)
			if depth == 0 && exists {
				return c.errorf(p, "'%s' redeclared in this block", p.Name)
			}
			s := c.symbolTable.Define(p.Name)
			// function arguments is not assigned directly.
			s.LocalAssigned = true
		}

		if err := c.Compile(node.Body); err != nil {
			return err
		}

		// code optimization
		numRemovedLocals := c.optimizeFunc(node)

		freeSymbols := c.symbolTable.FreeSymbols()
		numLocals := c.symbolTable.MaxSymbols() - numRemovedLocals
		scope := c.leaveScope()

		for _, s := range freeSymbols {
			switch s.Scope {
			case ScopeLocal:
				if !s.LocalAssigned {
					// Here, the closure is capturing a local variable that's
					// not yet assigned its value. One example is a local
					// recursive function:
					//
					//   fn() {
					//     foo := fn(x) {
					//       // ..
					//       return foo(x-1)
					//     }
					//   }
					//
					// which translate into
					//
					//   0000 GETL    0
					//   0002 CLOSURE ?     1
					//   0006 DEFL    0
					//
					// . So the local variable (0) is being captured before
					// it's assigned the value.
					//
					// Solution is to transform the code into something like
					// this:
					//
					//   fn() {
					//     foo := nil
					//     foo = fn(x) {
					//       // ..
					//       return foo(x-1)
					//     }
					//   }
					//
					// that is equivalent to
					//
					//   0000 NULL
					//   0001 DEFL    0
					//   0003 GETL    0
					//   0005 CLOSURE ?     1
					//   0009 SETL    0
					//
					c.emit(node, bytecode.OpNull)
					c.emit(node, bytecode.OpDefineLocal, s.Index)
					s.LocalAssigned = true
				}
				c.emit(node, bytecode.OpGetLocalPtr, s.Index)
			case ScopeFree:
				c.emit(node, bytecode.OpGetFreePtr, s.Index)
			}
		}

		compiledFunction := &CompiledFunction{
			instructions:  scope.instructions,
			numLocals:     numLocals,
			numParameters: len(node.Type.Params.List),
			numOptionals:  node.Type.Params.NumOptionals,
			varArgs:       node.Type.Params.VarArgs,
			sourceMap:     scope.sourceMap,
			deferMap:      scope.deferMap,
		}

		if len(freeSymbols) > 0 {
			c.emit(node, bytecode.OpClosure,
				c.addConstant(compiledFunction), len(freeSymbols))
		} else {
			c.emit(node, bytecode.OpConstant, c.addConstant(compiledFunction))
		}
	case *ast.ReturnStmt:
		for _, result := range node.Results {
			if err := c.Compile(result); err != nil {
				return err
			}
		}
		if len(node.Results) > 1 {
			c.emit(node, bytecode.OpTuple, len(node.Results), 0)
		}
		var hasResults int
		if len(node.Results) != 0 {
			hasResults = 1
		}
		c.emit(node, bytecode.OpReturn, hasResults)
	case *ast.DeferStmt:
		if err := c.Compile(node.CallExpr.Func); err != nil {
			return err
		}
		splat, err := c.compileListElements(node.CallExpr.Args)
		if err != nil {
			return err
		}
		deferIdx := c.addDeferPos(node.CallExpr.Pos())
		c.emit(node, bytecode.OpDefer, len(node.CallExpr.Args), splat, deferIdx)
	case *ast.SplatExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		c.emit(node, bytecode.OpSplat)
	case *ast.CallExpr:
		if err := c.Compile(node.Func); err != nil {
			return err
		}
		splat, err := c.compileListElements(node.Args)
		if err != nil {
			return err
		}
		c.emit(node, bytecode.OpCall, len(node.Args), splat)
	case *ast.ImportExpr:
		if node.ModuleName == "" {
			return c.errorf(node, "empty module name")
		}
		if mod := c.modules.Get(node.ModuleName); mod != nil {
			switch mod := mod.(type) {
			case SourceModule:
				compiled, err := c.compileModule(node, node.ModuleName, mod, false)
				if err != nil {
					return err
				}
				c.emit(node, bytecode.OpConstant, c.addConstant(compiled))
				c.emit(node, bytecode.OpCall, 0, 0)
			case *BuiltinModule:
				c.emit(node, bytecode.OpConstant, c.addConstant(mod))
			default:
				panic(fmt.Errorf("invalid import value type: %T", mod))
			}
		} else if c.allowFileImport {
			moduleName := node.ModuleName

			modulePath, err := c.getPathModule(moduleName)
			if err != nil {
				return c.errorf(node, "module file path error: %s",
					err.Error())
			}

			moduleSrc, err := os.ReadFile(modulePath)
			if err != nil {
				return c.errorf(node, "module file read error: %s",
					err.Error())
			}

			compiled, err := c.compileModule(node, modulePath, moduleSrc, true)
			if err != nil {
				return err
			}

			c.emit(node, bytecode.OpConstant, c.addConstant(compiled))
			c.emit(node, bytecode.OpCall, 0, 0)
		} else {
			return c.errorf(node, "module '%s' not found", node.ModuleName)
		}
	case *ast.TryExpr:
		if err := c.Compile(node.CallExpr.Func); err != nil {
			return err
		}
		splat, err := c.compileListElements(node.CallExpr.Args)
		if err != nil {
			return err
		}
		c.emit(node, bytecode.OpTry, len(node.CallExpr.Args), splat)
	case *ast.ThrowStmt:
		for _, e := range node.Errors {
			if err := c.Compile(e); err != nil {
				return err
			}
		}
		if len(node.Errors) > 1 {
			c.emit(node, bytecode.OpTuple, len(node.Errors), 0)
		}
		var hasErrors int
		if len(node.Errors) != 0 {
			hasErrors = 1
		}
		c.emit(node, bytecode.OpThrow, hasErrors)
	case *ast.CondExpr:
		if err := c.Compile(node.Cond); err != nil {
			return err
		}

		// first jump placeholder
		jumpPos1 := c.emit(node, bytecode.OpJumpFalsy, 0)
		if err := c.Compile(node.True); err != nil {
			return err
		}

		// second jump placeholder
		jumpPos2 := c.emit(node, bytecode.OpJump, 0)

		// update first jump offset
		curPos := len(c.currentInstructions())
		c.changeOperand(jumpPos1, curPos)
		if err := c.Compile(node.False); err != nil {
			return err
		}

		// update second jump offset
		curPos = len(c.currentInstructions())
		c.changeOperand(jumpPos2, curPos)
	}
	return nil
}

// Bytecode returns a compiled bytecode.
func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		FileSet: c.file.Set(),
		MainFunction: &CompiledFunction{
			instructions: c.currentInstructions(),
			sourceMap:    c.currentSourceMap(),
			deferMap:     c.currentDeferMap(),
		},
		Constants: c.constants,
	}
}

// EnableFileImport enables or disables module loading from local files.
// Local file modules are disabled by default.
func (c *Compiler) EnableFileImport(enable bool) {
	c.allowFileImport = enable
}

// SetImportDir sets the initial import directory path for file imports.
func (c *Compiler) SetImportDir(dir string) {
	c.importDir = dir
}

// SetImportFileExt sets the extension name of the source file for loading
// local module files.
//
// Use this method if you want other source file extension than ".toy".
//
//	// this will search for *.toy. *.foo, *.bar
//	err := c.SetImportFileExt(".toy", ".foo", ".bar")
//
// This function requires at least one argument, since it will replace the
// current list of extension name.
func (c *Compiler) SetImportFileExt(exts ...string) error {
	if len(exts) == 0 {
		return fmt.Errorf("missing arg: at least one argument is required")
	}

	for _, ext := range exts {
		if ext != filepath.Ext(ext) || ext == "" {
			return fmt.Errorf("invalid file extension: %s", ext)
		}
	}

	c.importFileExt = exts // Replace the hole current extension list

	return nil
}

// GetImportFileExt returns the current list of extension name.
// Thease are the complementary suffix of the source file to search and load
// local module files.
func (c *Compiler) GetImportFileExt() []string {
	return c.importFileExt
}

func (c *Compiler) compileListElements(elems []ast.Expr) (splat int, err error) {
	splat = 0
	for _, elem := range elems {
		if _, ok := elem.(*ast.SplatExpr); ok {
			splat = 1
		}
		if err := c.Compile(elem); err != nil {
			return 0, err
		}
	}
	return splat, nil
}

func (c *Compiler) compileSelectorExpr(node *ast.SelectorExpr, withOk bool) error {
	if err := c.Compile(node.Expr); err != nil {
		return err
	}
	c.emit(node.Sel, bytecode.OpConstant, c.addConstant(String(node.Sel.Name)))
	returnBool := 0
	if withOk {
		returnBool = 1
	}
	c.emit(node, bytecode.OpIndex, returnBool)
	return nil
}

func (c *Compiler) compileIndexExpr(node *ast.IndexExpr, withOk bool) error {
	if err := c.Compile(node.Expr); err != nil {
		return err
	}
	if err := c.Compile(node.Index); err != nil {
		return err
	}
	returnBool := 0
	if withOk {
		returnBool = 1
	}
	c.emit(node, bytecode.OpIndex, returnBool)
	return nil
}

func (c *Compiler) compileAssign(
	node ast.Node,
	lhs, rhs []ast.Expr,
	op token.Token,
) error {
	if op == token.Assign || op == token.Define {
		return c.compileAssignDefine(node, lhs, rhs, op)
	}

	ident, expr, sel := resolveAssignLHS(lhs[0])
	// we disallow statements like 1 += 1
	// and allow statements like [1, 2, 3][0] += 1
	if ident == "" && sel == nil {
		return c.errorf(node, "cannot assign to '%s': not a selector expression", expr.String())
	}

	// it's fine for symbol to be nil,
	// since it won't be used in selector expressions
	var symbol *Symbol
	if ident != "" {
		var exists bool
		symbol, _, exists = c.symbolTable.Resolve(ident, false)
		if !exists {
			return c.errorf(node, "unresolved reference '%s'", ident)
		}
	}

	// +=, -=, *=, /=
	if err := c.Compile(lhs[0]); err != nil {
		return err
	}

	// compile RHSs
	if err := c.Compile(rhs[0]); err != nil {
		return err
	}

	switch op {
	case token.AddAssign:
		c.emit(node, bytecode.OpBinaryOp, int(token.Add))
	case token.SubAssign:
		c.emit(node, bytecode.OpBinaryOp, int(token.Sub))
	case token.MulAssign:
		c.emit(node, bytecode.OpBinaryOp, int(token.Mul))
	case token.QuoAssign:
		c.emit(node, bytecode.OpBinaryOp, int(token.Quo))
	case token.RemAssign:
		c.emit(node, bytecode.OpBinaryOp, int(token.Rem))
	case token.AndAssign:
		c.emit(node, bytecode.OpBinaryOp, int(token.And))
	case token.OrAssign:
		c.emit(node, bytecode.OpBinaryOp, int(token.Or))
	case token.XorAssign:
		c.emit(node, bytecode.OpBinaryOp, int(token.Xor))
	case token.AndNotAssign:
		c.emit(node, bytecode.OpBinaryOp, int(token.AndNot))
	case token.ShlAssign:
		c.emit(node, bytecode.OpBinaryOp, int(token.Shl))
	case token.ShrAssign:
		c.emit(node, bytecode.OpBinaryOp, int(token.Shr))
	case token.NullishAssign:
		c.emit(node, bytecode.OpBinaryOp, int(token.Nullish))
	}

	if sel != nil {
		// compile left side of the selector expression
		if err := c.Compile(expr); err != nil {
			return err
		}
		// compile selector
		if ident, ok := sel.(*ast.Ident); ok {
			c.emit(sel, bytecode.OpConstant, c.addConstant(String(ident.Name)))
		} else if err := c.Compile(sel); err != nil {
			return err
		}
	}

	if sel != nil {
		c.emit(node, bytecode.OpSetIndex)
	} else {
		switch symbol.Scope {
		case ScopeGlobal:
			c.emit(node, bytecode.OpSetGlobal, symbol.Index)
		case ScopeLocal:
			if op == token.Define && !symbol.LocalAssigned {
				c.emit(node, bytecode.OpDefineLocal, symbol.Index)
			} else {
				c.emit(node, bytecode.OpSetLocal, symbol.Index)
			}
			// mark the symbol as local-assigned
			symbol.LocalAssigned = true
		case ScopeFree:
			c.emit(node, bytecode.OpSetFree, symbol.Index)
		default:
			panic(fmt.Errorf("invalid assignment variable scope: %s", symbol.Scope))
		}
	}

	return nil
}

// compileAssignDefine only handles = and := operations.
func (c *Compiler) compileAssignDefine(
	node ast.Node,
	lhs, rhs []ast.Expr,
	op token.Token,
) error {
	var unpacking bool
	if len(lhs) != len(rhs) {
		unpacking = true
		if len(rhs) != 1 {
			return c.errorf(node, "trying to assign %d value(s) to %d variable(s)", len(rhs), len(lhs))
		}
	}

	type lhsResolved struct {
		ident  string
		expr   ast.Expr
		sel    ast.Expr
		isFunc bool
		symbol *Symbol
		exists bool
		depth  int
	}

	var redecl int
	resolved := make([]*lhsResolved, 0, len(lhs))

	// we have to resolve everything first
	for j := range lhs {
		ident, expr, sel := resolveAssignLHS(lhs[j])
		// we disallow statements like 1 := 1 or 1 = 1
		// and allow statements like [1, 2, 3][0] = 1
		if ident == "" && sel == nil {
			if op == token.Define {
				return c.errorf(node, "non-name '%s' on left side of :=", expr.String())
			}
			return c.errorf(node, "cannot assign to '%s': not a selector expression", expr.String())
		}
		// we also disallow using selectors with define operator
		if op == token.Define && sel != nil {
			return c.errorf(node, "operator := not allowed with selector")
		}

		var isFunc bool
		if !unpacking {
			_, isFunc = rhs[j].(*ast.FuncLit)
		}

		var (
			// it's fine for symbol to be nil,
			// since it won't be used with selector expressions
			symbol *Symbol
			exists bool
			depth  int
		)

		if ident != "" {
			symbol, depth, exists = c.symbolTable.Resolve(ident, false)
			if op == token.Define {
				if depth == 0 && exists {
					redecl++ // increment the number of variable redeclarations
				}
				if isFunc {
					symbol = c.symbolTable.Define(ident)
				}
			} else if !exists {
				return c.errorf(node, "unresolved reference '%s'", ident)
			}
			if redecl == len(lhs) {
				// if all variables have been redeclared, return an error
				if redecl == 1 {
					return c.errorf(node, "'%s' redeclared in this block", ident)
				}
				return c.errorf(node, "no new variables on the left side of :=")
			}
		}

		resolved = append(resolved, &lhsResolved{
			ident:  ident,
			expr:   expr,
			sel:    sel,
			isFunc: isFunc,
			symbol: symbol,
			exists: exists,
			depth:  depth,
		})
	}

	if !unpacking {
		for j := len(rhs) - 1; j >= 0; j-- {
			// compile RHSs
			if err := c.Compile(rhs[j]); err != nil {
				return err
			}
		}
	} else {
		// rhs should be an indexable
		switch rhs0 := rhs[0].(type) {
		case *ast.SelectorExpr:
			// x, ok := xs[0]
			if err := c.compileSelectorExpr(rhs0, true); err != nil {
				return err
			}
		case *ast.IndexExpr:
			// x, ok := xs[0]
			if err := c.compileIndexExpr(rhs0, true); err != nil {
				return err
			}
		default:
			if err := c.Compile(rhs0); err != nil {
				return err
			}
		}
		// since we can't check how many values an indexable contains at compile time
		// we have to check it at the runtime
		c.emit(node, bytecode.OpIdxAssignAssert, len(lhs))
	}

	for j, lr := range resolved {
		if op == token.Define && (!lr.exists || lr.depth > 0) && !lr.isFunc {
			lr.symbol = c.symbolTable.Define(lr.ident)
		}

		if lr.sel != nil {
			// compile left side of the selector expression
			if err := c.Compile(lr.expr); err != nil {
				return err
			}
			// compile selector
			if ident, ok := lr.sel.(*ast.Ident); ok {
				c.emit(lr.sel, bytecode.OpConstant, c.addConstant(String(ident.Name)))
			} else if err := c.Compile(lr.sel); err != nil {
				return err
			}
		}

		if unpacking {
			// push value from the indexable to the top of the stack
			c.emit(node, bytecode.OpIdxElem, j)
		}

		if lr.sel != nil {
			c.emit(node, bytecode.OpSetIndex)
		} else {
			// symbol can't be nil here,
			// since it can only be nil in selector expressions
			switch lr.symbol.Scope {
			case ScopeGlobal:
				c.emit(node, bytecode.OpSetGlobal, lr.symbol.Index)
			case ScopeLocal:
				if op == token.Define && !lr.symbol.LocalAssigned {
					c.emit(node, bytecode.OpDefineLocal, lr.symbol.Index)
				} else {
					c.emit(node, bytecode.OpSetLocal, lr.symbol.Index)
				}
				// mark the symbol as local-assigned
				lr.symbol.LocalAssigned = true
			case ScopeFree:
				c.emit(node, bytecode.OpSetFree, lr.symbol.Index)
			default:
				panic(fmt.Errorf("invalid assignment variable scope: %s", lr.symbol.Scope))
			}
		}
	}

	if unpacking {
		// there will be indexable left, so we pop it
		c.emit(node, bytecode.OpPop)
	}

	return nil
}

func (c *Compiler) compileLogical(node *ast.BinaryExpr) error {
	// left side term
	if err := c.Compile(node.LHS); err != nil {
		return err
	}

	// jump position
	var jumpPos int
	if node.Token == token.LAnd {
		jumpPos = c.emit(node, bytecode.OpAndJump, 0)
	} else {
		jumpPos = c.emit(node, bytecode.OpOrJump, 0)
	}

	// right side term
	if err := c.Compile(node.RHS); err != nil {
		return err
	}

	c.changeOperand(jumpPos, len(c.currentInstructions()))
	return nil
}

func (c *Compiler) compileForStmt(stmt *ast.ForStmt, label string) error {
	c.symbolTable = c.symbolTable.Fork(true)
	defer func() {
		c.symbolTable = c.symbolTable.Parent(false)
	}()

	// init statement
	if stmt.Init != nil {
		if err := c.Compile(stmt.Init); err != nil {
			return err
		}
	}

	// pre-condition position
	preCondPos := len(c.currentInstructions())

	// condition expression
	postCondPos := -1
	if stmt.Cond != nil {
		if err := c.Compile(stmt.Cond); err != nil {
			return err
		}
		// condition jump position
		postCondPos = c.emit(stmt, bytecode.OpJumpFalsy, 0)
	}

	// enter loop
	loop := c.enterLoop(label)

	// body statement
	if err := c.Compile(stmt.Body); err != nil {
		c.leaveLoop()
		return err
	}

	c.leaveLoop()

	// post-body position
	postBodyPos := len(c.currentInstructions())

	// post statement
	if stmt.Post != nil {
		if err := c.Compile(stmt.Post); err != nil {
			return err
		}
	}

	// back to condition
	c.emit(stmt, bytecode.OpJump, preCondPos)

	// post-statement position
	postStmtPos := len(c.currentInstructions())
	if postCondPos >= 0 {
		c.changeOperand(postCondPos, postStmtPos)
	}

	// update all break/continue jump positions
	for _, pos := range loop.breaks {
		c.changeOperand(pos, postStmtPos)
	}
	for _, pos := range loop.continues {
		c.changeOperand(pos, postBodyPos)
	}

	return nil
}

func (c *Compiler) compileForInStmt(stmt *ast.ForInStmt, label string) error {
	c.symbolTable = c.symbolTable.Fork(true)
	defer func() {
		c.symbolTable = c.symbolTable.Parent(false)
	}()

	// for-in statement is compiled like following:
	//
	//   :it := iterator(iterable)
	//   for {
	//     k, v, ok := :it.next()
	//     if !ok {
	//       break
	//     }
	//     ... body ...
	//   }
	//   :it.close() // some iterators might implement CloseableIterator
	//
	// ":it" is a local variable but it will not conflict with other user variables
	// because character ":" is not allowed in the variable names.

	// init
	//   :it = iterator(iterable)
	itSymbol := c.symbolTable.Define(":it")
	if err := c.Compile(stmt.Iterable); err != nil {
		return err
	}
	c.emit(stmt, bytecode.OpIteratorInit)
	if itSymbol.Scope == ScopeGlobal {
		c.emit(stmt, bytecode.OpSetGlobal, itSymbol.Index)
	} else {
		c.emit(stmt, bytecode.OpDefineLocal, itSymbol.Index)
	}

	var setKeyVal int

	// define key variable
	var keySymbol *Symbol
	if stmt.Key.Name != "_" {
		keySymbol = c.symbolTable.Define(stmt.Key.Name)
		setKeyVal |= 0x1
	}

	// define value variable
	var valueSymbol *Symbol
	if stmt.Value.Name != "_" {
		valueSymbol = c.symbolTable.Define(stmt.Value.Name)
		setKeyVal |= 0x2
	}

	// pre-condition position
	preCondPos := len(c.currentInstructions())

	// enter loop
	loop := c.enterLoop(label)

	// get next entry
	// k, v, ok := :it.next()
	if itSymbol.Scope == ScopeGlobal {
		c.emit(stmt, bytecode.OpGetGlobal, itSymbol.Index)
	} else {
		c.emit(stmt, bytecode.OpGetLocal, itSymbol.Index)
	}
	c.emit(stmt, bytecode.OpIteratorNext, setKeyVal)

	// condition jump position
	postCondPos := c.emit(stmt, bytecode.OpJumpFalsy, 0)

	// assign value variable
	if valueSymbol != nil {
		if valueSymbol.Scope == ScopeGlobal {
			c.emit(stmt, bytecode.OpSetGlobal, valueSymbol.Index)
		} else {
			valueSymbol.LocalAssigned = true
			c.emit(stmt, bytecode.OpDefineLocal, valueSymbol.Index)
		}
	}

	// assign key variable
	if keySymbol != nil {
		if keySymbol.Scope == ScopeGlobal {
			c.emit(stmt, bytecode.OpSetGlobal, keySymbol.Index)
		} else {
			keySymbol.LocalAssigned = true
			c.emit(stmt, bytecode.OpDefineLocal, keySymbol.Index)
		}
	}

	// body statement
	if err := c.Compile(stmt.Body); err != nil {
		c.leaveLoop()
		return err
	}

	c.leaveLoop()

	// post-body position
	postBodyPos := len(c.currentInstructions())

	// back to condition
	c.emit(stmt, bytecode.OpJump, preCondPos)

	// post-statement position
	postStmtPos := len(c.currentInstructions())
	c.changeOperand(postCondPos, postStmtPos)

	// update all break/continue jump positions
	for _, pos := range loop.breaks {
		c.changeOperand(pos, postStmtPos)
	}
	for _, pos := range loop.continues {
		c.changeOperand(pos, postBodyPos)
	}

	// deinit
	// :it.close()
	if itSymbol.Scope == ScopeGlobal {
		c.emit(stmt, bytecode.OpGetGlobal, itSymbol.Index)
	} else {
		c.emit(stmt, bytecode.OpGetLocal, itSymbol.Index)
	}
	c.emit(stmt, bytecode.OpIteratorClose)

	return nil
}

func (c *Compiler) checkCyclicImports(
	node ast.Node,
	modulePath string,
) error {
	if c.modulePath == modulePath {
		return c.errorf(node, "cyclic module import: %s", modulePath)
	} else if c.parent != nil {
		return c.parent.checkCyclicImports(node, modulePath)
	}
	return nil
}

func (c *Compiler) compileModule(
	node ast.Node,
	modulePath string,
	src []byte,
	isFile bool,
) (*CompiledFunction, error) {
	if err := c.checkCyclicImports(node, modulePath); err != nil {
		return nil, err
	}

	compiledModule, exists := c.loadCompiledModule(modulePath)
	if exists {
		return compiledModule, nil
	}

	modFile := c.file.Set().AddFile(modulePath, -1, len(src))
	p := parser.NewParser(modFile, src, nil)
	file, err := p.ParseFile()
	if err != nil {
		return nil, err
	}

	// inherit builtin functions
	symbolTable := NewSymbolTable()
	for _, sym := range c.symbolTable.BuiltinSymbols() {
		symbolTable.DefineBuiltin(sym.Index, sym.Name)
	}

	// no global scope for the module
	symbolTable = symbolTable.Fork(false)

	// compile module
	moduleCompiler := c.fork(modFile, modulePath, symbolTable, isFile)
	if err := moduleCompiler.Compile(file); err != nil {
		return nil, err
	}

	// code optimization
	numRemovedLocals := moduleCompiler.optimizeFunc(node)

	compiledFunc := moduleCompiler.Bytecode().MainFunction
	compiledFunc.numLocals = symbolTable.MaxSymbols() - numRemovedLocals
	c.storeCompiledModule(modulePath, compiledFunc)

	return compiledFunc, nil
}

func (c *Compiler) loadCompiledModule(modulePath string) (mod *CompiledFunction, ok bool) {
	if c.parent != nil {
		return c.parent.loadCompiledModule(modulePath)
	}
	mod, ok = c.compiledModules[modulePath]
	return mod, ok
}

func (c *Compiler) storeCompiledModule(
	modulePath string,
	module *CompiledFunction,
) {
	if c.parent != nil {
		c.parent.storeCompiledModule(modulePath, module)
	}
	c.compiledModules[modulePath] = module
}

func (c *Compiler) enterLoop(label string) *loop {
	loop := &loop{label: label}
	c.loops = append(c.loops, loop)
	c.loopIndex++
	if c.trace != nil {
		c.printTrace("LOOPE", c.loopIndex)
	}
	return loop
}

func (c *Compiler) leaveLoop() {
	if c.trace != nil {
		c.printTrace("LOOPL", c.loopIndex)
	}
	c.loops = c.loops[:len(c.loops)-1]
	c.loopIndex--
}

func (c *Compiler) currentInstructions() []byte {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) currentSourceMap() map[int]token.Pos {
	return c.scopes[c.scopeIndex].sourceMap
}

func (c *Compiler) currentDeferMap() []token.Pos {
	return c.scopes[c.scopeIndex].deferMap
}

func (c *Compiler) addDeferPos(pos token.Pos) int {
	idx := len(c.scopes[c.scopeIndex].deferMap)
	c.scopes[c.scopeIndex].deferMap = append(c.scopes[c.scopeIndex].deferMap, pos)
	return idx
}

func (c *Compiler) enterScope() {
	scope := compilationScope{
		sourceMap: make(map[int]token.Pos),
		labels:    make(map[string]int),
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.symbolTable = c.symbolTable.Fork(false)
	if c.trace != nil {
		c.printTrace("SCOPE", c.scopeIndex)
	}
}

func (c *Compiler) leaveScope() compilationScope {
	scope := c.scopes[c.scopeIndex]
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Parent(true)
	if c.trace != nil {
		c.printTrace("SCOPL", c.scopeIndex)
	}
	return scope
}

func (c *Compiler) fork(
	file *token.File,
	modulePath string,
	symbolTable *SymbolTable,
	isFile bool,
) *Compiler {
	child := NewCompiler(file, symbolTable, nil, c.modules, c.trace)
	child.modulePath = modulePath // module file path
	child.parent = c              // parent to set to current compiler
	child.allowFileImport = c.allowFileImport
	child.importDir = c.importDir
	child.importFileExt = c.importFileExt
	if isFile && c.importDir != "" {
		child.importDir = filepath.Dir(modulePath)
	}
	return child
}

func (c *Compiler) error(node ast.Node, err error) error {
	return &CompilerError{
		FileSet: c.file.Set(),
		Node:    node,
		Err:     err,
	}
}

func (c *Compiler) errorf(node ast.Node, format string, args ...any) error {
	return &CompilerError{
		FileSet: c.file.Set(),
		Node:    node,
		Err:     fmt.Errorf(format, args...),
	}
}

func (c *Compiler) addConstant(v Value) int {
	if c.parent != nil {
		// module compilers will use their parent's constants array
		return c.parent.addConstant(v)
	}
	c.constants = append(c.constants, v)
	if c.trace != nil {
		c.printTrace(fmt.Sprintf("CONST %04d %s", len(c.constants)-1, v))
	}
	return len(c.constants) - 1
}

func (c *Compiler) addInstruction(b []byte) int {
	posNewIns := len(c.currentInstructions())
	c.scopes[c.scopeIndex].instructions = append(c.currentInstructions(), b...)
	return posNewIns
}

func (c *Compiler) replaceInstruction(pos int, inst []byte) {
	copy(c.currentInstructions()[pos:], inst)
	if c.trace != nil {
		c.printTrace(fmt.Sprintf("REPLACE: %s",
			bytecode.FormatInstructions(c.scopes[c.scopeIndex].instructions[pos:], pos)[0]))
	}
}

func (c *Compiler) changeOperand(opPos int, operand ...int) {
	op := c.currentInstructions()[opPos]
	inst := bytecode.MakeInstruction(op, operand...)
	c.replaceInstruction(opPos, inst)
}

// optimizeFunc performs some code-level optimization for the current function
// instructions. It also removes unreachable (dead code) instructions and adds
// "return" instruction if needed.
// Returns the number of removed local variables.
func (c *Compiler) optimizeFunc(node ast.Node) int {
	// any instructions between RETURN and the function end
	// or instructions between RETURN and jump target position
	// are considered as unreachable.

	// pass 1. identify all jump destinations
	dsts := make(map[int]bool)
	iterateInstructions(c.scopes[c.scopeIndex].instructions,
		func(pos int, opcode bytecode.Opcode, operands []int) bool {
			switch opcode {
			case bytecode.OpJump, bytecode.OpJumpFalsy,
				bytecode.OpAndJump, bytecode.OpOrJump:
				dsts[operands[0]] = true
			}
			return true
		})

	// pass 2. eliminate dead code
	var newInsts []byte
	posMap := make(map[int]int) // old position to new position
	var dstIdx int
	var deadCode bool
	deadLocals := make(map[int]struct{}) // set of indices of dead local variables
	iterateInstructions(c.scopes[c.scopeIndex].instructions,
		func(pos int, opcode bytecode.Opcode, operands []int) bool {
			if dsts[pos] {
				dstIdx++
				deadCode = false
			}
			switch {
			case opcode == bytecode.OpDefineLocal:
				if deadCode {
					deadLocals[operands[0]] = struct{}{}
					return true
				}
				delete(deadLocals, operands[0])
			case deadCode:
				return true
			case opcode == bytecode.OpReturn:
				deadCode = true
			}
			posMap[pos] = len(newInsts)
			newInsts = append(newInsts, bytecode.MakeInstruction(opcode, operands...)...)
			return true
		})

	// pass 3. update jump positions
	var lastOp bytecode.Opcode
	var appendReturn bool
	endPos := len(c.scopes[c.scopeIndex].instructions)
	newEndPost := len(newInsts)
	iterateInstructions(newInsts,
		func(pos int, opcode bytecode.Opcode, operands []int) bool {
			switch opcode {
			case bytecode.OpJump, bytecode.OpJumpFalsy, bytecode.OpAndJump,
				bytecode.OpOrJump:
				newDst, ok := posMap[operands[0]]
				if ok {
					copy(newInsts[pos:], bytecode.MakeInstruction(opcode, newDst))
				} else if endPos == operands[0] {
					// there's a jump instruction that jumps to the end of
					// function compiler should append "return".
					copy(newInsts[pos:], bytecode.MakeInstruction(opcode, newEndPost))
					appendReturn = true
				} else {
					panic(fmt.Errorf("invalid jump position: %d", newDst))
				}
			}
			lastOp = opcode
			return true
		})
	if lastOp != bytecode.OpReturn {
		appendReturn = true
	}

	// pass 4. update source map
	newSourceMap := make(map[int]token.Pos)
	for pos, srcPos := range c.scopes[c.scopeIndex].sourceMap {
		newPos, ok := posMap[pos]
		if ok {
			newSourceMap[newPos] = srcPos
		}
	}
	c.scopes[c.scopeIndex].instructions = newInsts
	c.scopes[c.scopeIndex].sourceMap = newSourceMap

	// append "return"
	if appendReturn {
		c.emit(node, bytecode.OpReturn, 0)
	}

	return len(deadLocals)
}

func (c *Compiler) emit(node ast.Node, opcode bytecode.Opcode, operands ...int) int {
	inst := bytecode.MakeInstruction(opcode, operands...)
	pos := c.addInstruction(inst)
	if node != nil {
		c.scopes[c.scopeIndex].sourceMap[pos] = node.Pos()
	}
	if c.trace != nil {
		c.printTrace(fmt.Sprintf("EMIT: %s",
			bytecode.FormatInstructions(c.scopes[c.scopeIndex].instructions[pos:], pos)[0]))
	}
	return pos
}

func (c *Compiler) printTrace(a ...any) {
	const (
		dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
		n    = len(dots)
	)
	i := 2 * c.indent
	for i > n {
		fmt.Fprint(c.trace, dots)
		i -= n
	}
	fmt.Fprint(c.trace, dots[0:i])
	fmt.Fprintln(c.trace, a...)
}

func (c *Compiler) getPathModule(moduleName string) (pathFile string, err error) {
	for _, ext := range c.importFileExt {
		nameFile := moduleName
		if !strings.HasSuffix(nameFile, ext) {
			nameFile += ext
		}
		pathFile, err = filepath.Abs(filepath.Join(c.importDir, nameFile))
		if err != nil {
			continue
		}
		// Check if file exists
		if _, err := os.Stat(pathFile); !errors.Is(err, os.ErrNotExist) {
			return pathFile, nil
		}
	}
	return "", fmt.Errorf("module '%s' not found at: %s", moduleName, pathFile)
}

// For the given expression, returns the name of the leftmost identifier,
// an expression containing all but the last selector,
// and the last selector for the given expression.
// If the given expression is just an identifier, returns the name
// of the identifier, and the other results are set to nil.
// If the leftmost expression is not an identifier, then name is set to "".
func resolveAssignLHS(expr ast.Expr) (name string, x, sel ast.Expr) {
	switch term := expr.(type) {
	case *ast.SelectorExpr:
		x, sel = term.Expr, term.Sel
	case *ast.IndexExpr:
		x, sel = term.Expr, term.Index
	case *ast.Ident:
		return term.Name, nil, nil
	default:
		return "", expr, nil
	}
	left := x
	for {
		switch term := left.(type) {
		case *ast.SelectorExpr:
			left = term.Expr
		case *ast.IndexExpr:
			left = term.Expr
		case *ast.Ident:
			return term.Name, x, sel
		default:
			return "", x, sel
		}
	}
}

// unindentString strips indentation from the start of each line.
// Number of spaces to be stripped is equal to the minimal
// indentation of the string as a whole (ignoring lines with no non-space text).
// Completely strips the first line if there is no non-space text.
func unindentString(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) > 0 {
		trim := strings.TrimSpace(lines[0])
		if len(trim) == 0 {
			lines = lines[1:]
		}
	}
	if len(lines) == 0 {
		return ""
	}

	indent := -1
	for _, line := range lines {
		j := strings.IndexFunc(line, func(r rune) bool {
			return !unicode.IsSpace(r)
		})
		if j == -1 {
			continue
		}
		if indent == -1 || j < indent {
			indent = j
		}
	}

	var b strings.Builder
	for i, line := range lines {
		if i != 0 {
			b.WriteByte('\n')
		}
		if len(line) > indent {
			b.WriteString(line[indent:])
		}
	}

	return b.String()
}

func iterateInstructions(b []byte, fn func(pos int, opcode bytecode.Opcode, operands []int) bool) {
	for i := 0; i < len(b); i++ {
		numOperands := bytecode.OpcodeOperands[b[i]]
		operands, read := bytecode.ReadOperands(numOperands, b[i+1:])
		if !fn(i, b[i], operands) {
			break
		}
		i += read
	}
}

func tracec(c *Compiler, msg string) *Compiler {
	c.printTrace(msg, "{")
	c.indent++
	return c
}

func untracec(c *Compiler) {
	c.indent--
	c.printTrace("}")
}
