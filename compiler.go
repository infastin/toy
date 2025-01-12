package toy

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/infastin/toy/parser"
	"github.com/infastin/toy/token"
)

// compilationScope represents a compiled instructions
// and the last two instructions that were emitted.
type compilationScope struct {
	instructions []byte
	sourceMap    map[int]parser.Pos
	deferMap     []parser.Pos
	labels       map[string]parser.Pos
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
	FileSet *parser.SourceFileSet
	Node    parser.Node
	Err     error
}

func (e *CompilerError) Error() string {
	filePos := e.FileSet.Position(e.Node.Pos())
	return fmt.Sprintf("Compile Error: %s\n\tat %s", e.Err.Error(), filePos)
}

// Compiler compiles the AST into a bytecode.
type Compiler struct {
	file            *parser.SourceFile
	parent          *Compiler
	modulePath      string
	importDir       string
	importFileExt   []string
	constants       []Object
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
	file *parser.SourceFile,
	symbolTable *SymbolTable,
	constants []Object,
	modules ModuleGetter,
	trace io.Writer,
) *Compiler {
	mainScope := compilationScope{
		sourceMap: make(map[int]parser.Pos),
		labels:    make(map[string]parser.Pos),
	}

	// symbol table
	if symbolTable == nil {
		symbolTable = NewSymbolTable()
	}

	// add builtin functions to the symbol table
	for idx, fn := range BuiltinFuncs {
		symbolTable.DefineBuiltin(idx, fn.Name)
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
func (c *Compiler) Compile(node parser.Node) error {
	if c.trace != nil {
		if node != nil {
			defer untracec(tracec(c, fmt.Sprintf("%s (%s)",
				node.String(), reflect.TypeOf(node).Elem().Name())))
		} else {
			defer untracec(tracec(c, "<nil>"))
		}
	}

	switch node := node.(type) {
	case *parser.File:
		for _, stmt := range node.Stmts {
			if err := c.Compile(stmt); err != nil {
				return err
			}
		}
	case *parser.ExprStmt:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		c.emit(node, parser.OpPop)
	case *parser.IncDecStmt:
		op := token.AddAssign
		if node.Token == token.Dec {
			op = token.SubAssign
		}
		return c.compileAssign(node, []parser.Expr{node.Expr},
			[]parser.Expr{&parser.IntLit{Value: 1}}, op)
	case *parser.ParenExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
	case *parser.BinaryExpr:
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
			c.emit(node, parser.OpCompare, int(node.Token))
		case token.Add, token.Sub, token.Mul, token.Quo, token.Rem,
			token.And, token.Or, token.Xor, token.AndNot,
			token.Shl, token.Shr:
			c.emit(node, parser.OpBinaryOp, int(node.Token))
		default:
			return c.errorf(node, "invalid binary operator: %s",
				node.Token.String())
		}
	case *parser.IntLit:
		c.emit(node, parser.OpConstant,
			c.addConstant(Int(node.Value)))
	case *parser.FloatLit:
		c.emit(node, parser.OpConstant,
			c.addConstant(Float(node.Value)))
	case *parser.BoolLit:
		if node.Value {
			c.emit(node, parser.OpTrue)
		} else {
			c.emit(node, parser.OpFalse)
		}
	case *parser.StringLit:
		c.emit(node, parser.OpConstant,
			c.addConstant(String(node.Value)))
	case *parser.CharLit:
		c.emit(node, parser.OpConstant,
			c.addConstant(Char(node.Value)))
	case *parser.NilLit:
		c.emit(node, parser.OpNull)
	case *parser.UnaryExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		switch node.Token {
		case token.Add, token.Sub, token.Not, token.Xor:
			c.emit(node, parser.OpUnaryOp, int(node.Token))
		default:
			return c.errorf(node, "invalid unary operator: %s", node.Token.String())
		}
	case *parser.IfStmt:
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
		jumpPos1 := c.emit(node, parser.OpJumpFalsy, 0)
		if err := c.Compile(node.Body); err != nil {
			return err
		}
		if node.Else != nil {
			// second jump placeholder
			jumpPos2 := c.emit(node, parser.OpJump, 0)

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
	case *parser.ForStmt:
		return c.compileForStmt(node, "")
	case *parser.ForInStmt:
		return c.compileForInStmt(node, "")
	case *parser.BranchStmt:
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
			pos := c.emit(node, parser.OpJump, 0)
			curLoop.breaks = append(curLoop.breaks, pos)
		case token.Continue:
			pos := c.emit(node, parser.OpJump, 0)
			curLoop.continues = append(curLoop.continues, pos)
		default:
			panic(fmt.Errorf("invalid branch statement: %s", node.Token.String()))
		}
	case *parser.LabeledStmt:
		if _, ok := c.scopes[c.scopeIndex].labels[node.Label.Name]; ok {
			return c.errorf(node, "label '%s' already declared", node.Label.Name)
		} else {
			c.scopes[c.scopeIndex].labels[node.Label.Name] = node.Label.Pos()
		}
		switch stmt := node.Stmt.(type) {
		case *parser.ForStmt:
			if err := c.compileForStmt(stmt, node.Label.Name); err != nil {
				return err
			}
		case *parser.ForInStmt:
			if err := c.compileForInStmt(stmt, node.Label.Name); err != nil {
				return err
			}
		default:
			if err := c.Compile(node.Stmt); err != nil {
				return err
			}
		}
	case *parser.BlockStmt:
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
	case *parser.ShortFuncBodyStmt:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		c.emit(node, parser.OpReturn, 1)
	case *parser.AssignStmt:
		err := c.compileAssign(node, node.LHS, node.RHS, node.Token)
		if err != nil {
			return err
		}
	case *parser.Ident:
		symbol, _, ok := c.symbolTable.Resolve(node.Name, false)
		if !ok {
			return c.errorf(node, "unresolved reference '%s'", node.Name)
		}
		switch symbol.Scope {
		case ScopeGlobal:
			c.emit(node, parser.OpGetGlobal, symbol.Index)
		case ScopeLocal:
			c.emit(node, parser.OpGetLocal, symbol.Index)
		case ScopeBuiltin:
			c.emit(node, parser.OpGetBuiltin, symbol.Index)
		case ScopeFree:
			c.emit(node, parser.OpGetFree, symbol.Index)
		}
	case *parser.ArrayLit:
		splat := 0
		for _, elem := range node.Elements {
			if _, ok := elem.(*parser.SplatExpr); ok {
				splat = 1
			}
			if err := c.Compile(elem); err != nil {
				return err
			}
		}
		c.emit(node, parser.OpArray, len(node.Elements), splat)
	case *parser.MapLit:
		for _, elt := range node.Elements {
			if err := c.Compile(elt.Key); err != nil {
				return err
			}
			if err := c.Compile(elt.Value); err != nil {
				return err
			}
		}
		c.emit(node, parser.OpMap, len(node.Elements))
	case *parser.SelectorExpr: // selector on RHS side
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		if err := c.Compile(node.Sel); err != nil {
			return err
		}
		c.emit(node, parser.OpField)
	case *parser.IndexExpr:
		if err := c.compileIndexExpr(node, false); err != nil {
			return err
		}
	case *parser.SliceExpr:
		var opSliceOperand int
		if node.Low != nil {
			if err := c.Compile(node.Low); err != nil {
				return err
			}
			opSliceOperand |= 0x1
		}
		if node.High != nil {
			if err := c.Compile(node.High); err != nil {
				return err
			}
			opSliceOperand |= 0x2
		}
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		c.emit(node, parser.OpSliceIndex, opSliceOperand)
	case *parser.FuncLit:
		c.enterScope()

		for _, p := range node.Type.Params.List {
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
					c.emit(node, parser.OpNull)
					c.emit(node, parser.OpDefineLocal, s.Index)
					s.LocalAssigned = true
				}
				c.emit(node, parser.OpGetLocalPtr, s.Index)
			case ScopeFree:
				c.emit(node, parser.OpGetFreePtr, s.Index)
			}
		}

		compiledFunction := &CompiledFunction{
			instructions:  scope.instructions,
			numLocals:     numLocals,
			numParameters: len(node.Type.Params.List),
			varArgs:       node.Type.Params.VarArgs,
			sourceMap:     scope.sourceMap,
			deferMap:      scope.deferMap,
		}

		if len(freeSymbols) > 0 {
			c.emit(node, parser.OpClosure,
				c.addConstant(compiledFunction), len(freeSymbols))
		} else {
			c.emit(node, parser.OpConstant, c.addConstant(compiledFunction))
		}
	case *parser.ReturnStmt:
		if c.symbolTable.Parent(true) == nil {
			// outside the function
			return c.errorf(node, "return not allowed outside function")
		}
		for _, result := range node.Results {
			if err := c.Compile(result); err != nil {
				return err
			}
		}
		if len(node.Results) > 1 {
			c.emit(node, parser.OpTuple, len(node.Results), 0)
		}
		var opReturnOperand int
		if len(node.Results) != 0 {
			opReturnOperand = 1
		}
		if len(c.currentDeferMap()) != 0 {
			c.emitDeferLoop()
		}
		c.emit(node, parser.OpReturn, opReturnOperand)
	case *parser.DeferStmt:
		if err := c.Compile(node.CallExpr.Func); err != nil {
			return err
		}
		splat := 0
		for _, arg := range node.CallExpr.Args {
			if _, ok := arg.(*parser.SplatExpr); ok {
				splat = 1
			}
			if err := c.Compile(arg); err != nil {
				return err
			}
		}
		deferIdx := c.addDeferPos(node.CallExpr.Pos())
		c.emit(node, parser.OpSaveDefer, len(node.CallExpr.Args), splat, deferIdx)
	case *parser.SplatExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		c.emit(node, parser.OpSplat)
	case *parser.CallExpr:
		if err := c.Compile(node.Func); err != nil {
			return err
		}
		splat := 0
		for _, arg := range node.Args {
			if _, ok := arg.(*parser.SplatExpr); ok {
				splat = 1
			}
			if err := c.Compile(arg); err != nil {
				return err
			}
		}
		c.emit(node, parser.OpCall, len(node.Args), splat, 0)
	case *parser.ImportExpr:
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
				c.emit(node, parser.OpConstant, c.addConstant(compiled))
				c.emit(node, parser.OpCall, 0, 0, 0)
			case *BuiltinModule:
				c.emit(node, parser.OpConstant, c.addConstant(mod))
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

			c.emit(node, parser.OpConstant, c.addConstant(compiled))
			c.emit(node, parser.OpCall, 0, 0, 0)
		} else {
			return c.errorf(node, "module '%s' not found", node.ModuleName)
		}
	case *parser.ExportStmt:
		// export statement must be in top-level scope
		if c.scopeIndex != 0 {
			return c.errorf(node, "export not allowed inside function")
		}
		// export statement is simply ignore when compiling non-module code
		if c.parent == nil {
			break
		}
		if err := c.Compile(node.Result); err != nil {
			return err
		}
		c.emit(node, parser.OpImmutable)
		if len(c.currentDeferMap()) != 0 {
			c.emitDeferLoop()
		}
		c.emit(node, parser.OpReturn, 1)
	case *parser.ImmutableExpr:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}
		c.emit(node, parser.OpImmutable)
	case *parser.TupleLit:
		splat := 0
		for _, elem := range node.Elements {
			if _, ok := elem.(*parser.SplatExpr); ok {
				splat = 1
			}
			if err := c.Compile(elem); err != nil {
				return err
			}
		}
		c.emit(node, parser.OpTuple, len(node.Elements), splat)
	case *parser.CondExpr:
		if err := c.Compile(node.Cond); err != nil {
			return err
		}

		// first jump placeholder
		jumpPos1 := c.emit(node, parser.OpJumpFalsy, 0)
		if err := c.Compile(node.True); err != nil {
			return err
		}

		// second jump placeholder
		jumpPos2 := c.emit(node, parser.OpJump, 0)

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
	insts := c.currentInstructions()
	if len(c.currentDeferMap()) != 0 {
		curPos := len(insts)
		insts = append(insts, MakeInstruction(parser.OpPushDefer)...)
		jumpPos := len(insts)
		insts = append(insts, MakeInstruction(parser.OpJumpFalsy, 0)...)
		insts = append(insts, MakeInstruction(parser.OpCall, 0, 0, 1)...)
		insts = append(insts, MakeInstruction(parser.OpJump, curPos)...)
		copy(insts[jumpPos:], MakeInstruction(parser.OpJumpFalsy, len(insts)))
	}
	insts = append(insts, MakeInstruction(parser.OpSuspend)...)
	return &Bytecode{
		FileSet: c.file.Set(),
		MainFunction: &CompiledFunction{
			instructions: insts,
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

func (c *Compiler) compileIndexExpr(node *parser.IndexExpr, withOk bool) error {
	if err := c.Compile(node.Expr); err != nil {
		return err
	}
	if err := c.Compile(node.Index); err != nil {
		return err
	}
	opIndexOperand := 0
	if withOk {
		opIndexOperand = 1
	}
	c.emit(node, parser.OpIndex, opIndexOperand)
	return nil
}

func (c *Compiler) compileAssign(
	node parser.Node,
	lhs, rhs []parser.Expr,
	op token.Token,
) error {
	if op == token.Assign || op == token.Define {
		return c.compileAssignDefine(node, lhs, rhs, op)
	}

	ident, selectors := resolveAssignLHS(lhs[0])
	numSel := len(selectors)

	symbol, _, exists := c.symbolTable.Resolve(ident, false)
	if !exists {
		return c.errorf(node, "unresolved reference '%s'", ident)
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
		c.emit(node, parser.OpBinaryOp, int(token.Add))
	case token.SubAssign:
		c.emit(node, parser.OpBinaryOp, int(token.Sub))
	case token.MulAssign:
		c.emit(node, parser.OpBinaryOp, int(token.Mul))
	case token.QuoAssign:
		c.emit(node, parser.OpBinaryOp, int(token.Quo))
	case token.RemAssign:
		c.emit(node, parser.OpBinaryOp, int(token.Rem))
	case token.AndAssign:
		c.emit(node, parser.OpBinaryOp, int(token.And))
	case token.OrAssign:
		c.emit(node, parser.OpBinaryOp, int(token.Or))
	case token.AndNotAssign:
		c.emit(node, parser.OpBinaryOp, int(token.AndNot))
	case token.XorAssign:
		c.emit(node, parser.OpBinaryOp, int(token.Xor))
	case token.ShlAssign:
		c.emit(node, parser.OpBinaryOp, int(token.Shl))
	case token.ShrAssign:
		c.emit(node, parser.OpBinaryOp, int(token.Shr))
	}

	// compile selector expressions (right to left)
	for i := numSel - 1; i >= 0; i-- {
		if err := c.Compile(selectors[i]); err != nil {
			return err
		}
	}

	switch symbol.Scope {
	case ScopeGlobal:
		if numSel > 0 {
			c.emit(node, parser.OpSetSelGlobal, symbol.Index, numSel)
		} else {
			c.emit(node, parser.OpSetGlobal, symbol.Index)
		}
	case ScopeLocal:
		if numSel > 0 {
			c.emit(node, parser.OpSetSelLocal, symbol.Index, numSel)
		} else {
			if op == token.Define && !symbol.LocalAssigned {
				c.emit(node, parser.OpDefineLocal, symbol.Index)
			} else {
				c.emit(node, parser.OpSetLocal, symbol.Index)
			}
		}
		// mark the symbol as local-assigned
		symbol.LocalAssigned = true
	case ScopeFree:
		if numSel > 0 {
			c.emit(node, parser.OpSetSelFree, symbol.Index, numSel)
		} else {
			c.emit(node, parser.OpSetFree, symbol.Index)
		}
	default:
		panic(fmt.Errorf("invalid assignment variable scope: %s", symbol.Scope))
	}

	return nil
}

// compileAssignDefine only handles = and := operations.
func (c *Compiler) compileAssignDefine(
	node parser.Node,
	lhs, rhs []parser.Expr,
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
		ident     string
		selectors []parser.Expr
		symbol    *Symbol
		exists    bool
		isFunc    bool
	}

	var redecl int
	resolved := make([]*lhsResolved, 0, len(lhs))

	// we have to resolve everything first
	for j := range lhs {
		ident, selectors := resolveAssignLHS(lhs[j])
		numSel := len(selectors)

		if op == token.Define && numSel > 0 {
			// using selector on new variable does not make sense
			return c.errorf(node, "operator := not allowed with selector")
		}

		var isFunc bool
		if !unpacking {
			_, isFunc = rhs[j].(*parser.FuncLit)
		}

		symbol, depth, exists := c.symbolTable.Resolve(ident, false)
		if op == token.Define {
			if depth == 0 && exists {
				redecl += 1 // increment the number of variable redeclarations
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

		resolved = append(resolved, &lhsResolved{
			ident:     ident,
			selectors: selectors,
			symbol:    symbol,
			exists:    exists,
			isFunc:    isFunc,
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
		if idx, ok := rhs[0].(*parser.IndexExpr); ok {
			// x, ok := xs[0]
			if err := c.compileIndexExpr(idx, true); err != nil {
				return err
			}
		} else {
			if err := c.Compile(rhs[0]); err != nil {
				return err
			}
		}
		// since we can't check how many values an indexable contains at compile time
		// we have to check it at the runtime
		c.emit(node, parser.OpIdxAssignAssert, len(lhs))
	}

	for j, lr := range resolved {
		numSel := len(lr.selectors)

		if op == token.Define && !lr.exists && !lr.isFunc {
			lr.symbol = c.symbolTable.Define(lr.ident)
		}

		// compile selector expressions (right to left)
		for i := numSel - 1; i >= 0; i-- {
			if err := c.Compile(lr.selectors[i]); err != nil {
				return err
			}
		}

		if unpacking {
			// push value from the indexable to the top of the stack
			c.emit(node, parser.OpIdxElem, j)
		}

		switch lr.symbol.Scope {
		case ScopeGlobal:
			if numSel > 0 {
				c.emit(node, parser.OpSetSelGlobal, lr.symbol.Index, numSel)
			} else {
				c.emit(node, parser.OpSetGlobal, lr.symbol.Index)
			}
		case ScopeLocal:
			if numSel > 0 {
				c.emit(node, parser.OpSetSelLocal, lr.symbol.Index, numSel)
			} else {
				if op == token.Define && !lr.symbol.LocalAssigned {
					c.emit(node, parser.OpDefineLocal, lr.symbol.Index)
				} else {
					c.emit(node, parser.OpSetLocal, lr.symbol.Index)
				}
			}
			// mark the symbol as local-assigned
			lr.symbol.LocalAssigned = true
		case ScopeFree:
			if numSel > 0 {
				c.emit(node, parser.OpSetSelFree, lr.symbol.Index, numSel)
			} else {
				c.emit(node, parser.OpSetFree, lr.symbol.Index)
			}
		default:
			panic(fmt.Errorf("invalid assignment variable scope: %s", lr.symbol.Scope))
		}
	}

	if unpacking {
		// there will be indexable left, so we pop it
		c.emit(node, parser.OpPop)
	}

	return nil
}

func (c *Compiler) compileLogical(node *parser.BinaryExpr) error {
	// left side term
	if err := c.Compile(node.LHS); err != nil {
		return err
	}

	// jump position
	var jumpPos int
	if node.Token == token.LAnd {
		jumpPos = c.emit(node, parser.OpAndJump, 0)
	} else {
		jumpPos = c.emit(node, parser.OpOrJump, 0)
	}

	// right side term
	if err := c.Compile(node.RHS); err != nil {
		return err
	}

	c.changeOperand(jumpPos, len(c.currentInstructions()))
	return nil
}

func (c *Compiler) compileForStmt(stmt *parser.ForStmt, label string) error {
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
		postCondPos = c.emit(stmt, parser.OpJumpFalsy, 0)
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
	c.emit(stmt, parser.OpJump, preCondPos)

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

func (c *Compiler) compileForInStmt(stmt *parser.ForInStmt, label string) error {
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
	c.emit(stmt, parser.OpIteratorInit)
	if itSymbol.Scope == ScopeGlobal {
		c.emit(stmt, parser.OpSetGlobal, itSymbol.Index)
	} else {
		c.emit(stmt, parser.OpDefineLocal, itSymbol.Index)
	}

	var opNextOperand int

	// define key variable
	var keySymbol *Symbol
	if stmt.Key.Name != "_" {
		keySymbol = c.symbolTable.Define(stmt.Key.Name)
		opNextOperand |= 0x1
	}

	// define value variable
	var valueSymbol *Symbol
	if stmt.Value.Name != "_" {
		valueSymbol = c.symbolTable.Define(stmt.Value.Name)
		opNextOperand |= 0x2
	}

	// pre-condition position
	preCondPos := len(c.currentInstructions())

	// enter loop
	loop := c.enterLoop(label)

	// get next entry
	// k, v, ok := :it.next()
	if itSymbol.Scope == ScopeGlobal {
		c.emit(stmt, parser.OpGetGlobal, itSymbol.Index)
	} else {
		c.emit(stmt, parser.OpGetLocal, itSymbol.Index)
	}
	c.emit(stmt, parser.OpIteratorNext, opNextOperand)

	// condition jump position
	postCondPos := c.emit(stmt, parser.OpJumpFalsy, 0)

	// assign value variable
	if valueSymbol != nil {
		if valueSymbol.Scope == ScopeGlobal {
			c.emit(stmt, parser.OpSetGlobal, valueSymbol.Index)
		} else {
			valueSymbol.LocalAssigned = true
			c.emit(stmt, parser.OpDefineLocal, valueSymbol.Index)
		}
	}

	// assign key variable
	if keySymbol != nil {
		if keySymbol.Scope == ScopeGlobal {
			c.emit(stmt, parser.OpSetGlobal, keySymbol.Index)
		} else {
			keySymbol.LocalAssigned = true
			c.emit(stmt, parser.OpDefineLocal, keySymbol.Index)
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
	c.emit(stmt, parser.OpJump, preCondPos)

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
		c.emit(stmt, parser.OpGetGlobal, itSymbol.Index)
	} else {
		c.emit(stmt, parser.OpGetLocal, itSymbol.Index)
	}
	c.emit(stmt, parser.OpIteratorClose)

	return nil
}

func (c *Compiler) checkCyclicImports(
	node parser.Node,
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
	node parser.Node,
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
	moduleCompiler.optimizeFunc(node)
	compiledFunc := moduleCompiler.Bytecode().MainFunction
	compiledFunc.numLocals = symbolTable.MaxSymbols()
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

func (c *Compiler) currentLoop() *loop {
	if c.loopIndex >= 0 {
		return c.loops[c.loopIndex]
	}
	return nil
}

func (c *Compiler) currentInstructions() []byte {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) currentSourceMap() map[int]parser.Pos {
	return c.scopes[c.scopeIndex].sourceMap
}

func (c *Compiler) currentDeferMap() []parser.Pos {
	return c.scopes[c.scopeIndex].deferMap
}

func (c *Compiler) addDeferPos(pos parser.Pos) int {
	idx := len(c.scopes[c.scopeIndex].deferMap)
	c.scopes[c.scopeIndex].deferMap = append(c.scopes[c.scopeIndex].deferMap, pos)
	return idx
}

func (c *Compiler) enterScope() {
	scope := compilationScope{
		sourceMap: make(map[int]parser.Pos),
		labels:    make(map[string]parser.Pos),
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
	file *parser.SourceFile,
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

func (c *Compiler) error(node parser.Node, err error) error {
	return &CompilerError{
		FileSet: c.file.Set(),
		Node:    node,
		Err:     err,
	}
}

func (c *Compiler) errorf(node parser.Node, format string, args ...any) error {
	return &CompilerError{
		FileSet: c.file.Set(),
		Node:    node,
		Err:     fmt.Errorf(format, args...),
	}
}

func (c *Compiler) addConstant(o Object) int {
	if c.parent != nil {
		// module compilers will use their parent's constants array
		return c.parent.addConstant(o)
	}
	c.constants = append(c.constants, o)
	if c.trace != nil {
		c.printTrace(fmt.Sprintf("CONST %04d %s", len(c.constants)-1, o))
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
			FormatInstructions(c.scopes[c.scopeIndex].instructions[pos:], pos)[0]))
	}
}

func (c *Compiler) changeOperand(opPos int, operand ...int) {
	op := c.currentInstructions()[opPos]
	inst := MakeInstruction(op, operand...)
	c.replaceInstruction(opPos, inst)
}

// optimizeFunc performs some code-level optimization for the current function
// instructions. It also removes unreachable (dead code) instructions and adds
// "return" instruction if needed.
// Returns the number of removed local variables.
func (c *Compiler) optimizeFunc(node parser.Node) int {
	// any instructions between RETURN and the function end
	// or instructions between RETURN and jump target position
	// are considered as unreachable.

	// pass 1. identify all jump destinations
	dsts := make(map[int]bool)
	iterateInstructions(c.scopes[c.scopeIndex].instructions,
		func(pos int, opcode parser.Opcode, operands []int) bool {
			switch opcode {
			case parser.OpJump, parser.OpJumpFalsy,
				parser.OpAndJump, parser.OpOrJump:
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
		func(pos int, opcode parser.Opcode, operands []int) bool {
			if dsts[pos] {
				dstIdx++
				deadCode = false
			}
			switch {
			case opcode == parser.OpDefineLocal:
				if deadCode {
					deadLocals[operands[0]] = struct{}{}
					return true
				}
				delete(deadLocals, operands[0])
			case deadCode:
				return true
			case opcode == parser.OpReturn:
				deadCode = true
			}
			posMap[pos] = len(newInsts)
			newInsts = append(newInsts, MakeInstruction(opcode, operands...)...)
			return true
		})

	// pass 3. update jump positions
	var lastOp parser.Opcode
	var appendReturn bool
	endPos := len(c.scopes[c.scopeIndex].instructions)
	newEndPost := len(newInsts)
	iterateInstructions(newInsts,
		func(pos int, opcode parser.Opcode, operands []int) bool {
			switch opcode {
			case parser.OpJump, parser.OpJumpFalsy, parser.OpAndJump,
				parser.OpOrJump:
				newDst, ok := posMap[operands[0]]
				if ok {
					copy(newInsts[pos:], MakeInstruction(opcode, newDst))
				} else if endPos == operands[0] {
					// there's a jump instruction that jumps to the end of
					// function compiler should append "return".
					copy(newInsts[pos:], MakeInstruction(opcode, newEndPost))
					appendReturn = true
				} else {
					panic(fmt.Errorf("invalid jump position: %d", newDst))
				}
			}
			lastOp = opcode
			return true
		})
	if lastOp != parser.OpReturn {
		appendReturn = true
	}

	// pass 4. update source map
	newSourceMap := make(map[int]parser.Pos)
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
		if len(c.currentDeferMap()) != 0 {
			c.emitDeferLoop()
		}
		c.emit(node, parser.OpReturn, 0)
	}

	return len(deadLocals)
}

func (c *Compiler) emit(node parser.Node, opcode parser.Opcode, operands ...int) int {
	inst := MakeInstruction(opcode, operands...)
	pos := c.addInstruction(inst)
	if node != nil {
		c.scopes[c.scopeIndex].sourceMap[pos] = node.Pos()
	}
	if c.trace != nil {
		c.printTrace(fmt.Sprintf("EMIT: %s",
			FormatInstructions(c.scopes[c.scopeIndex].instructions[pos:], pos)[0]))
	}
	return pos
}

func (c *Compiler) emitDeferLoop() {
	curPos := len(c.currentInstructions())
	c.emit(nil, parser.OpPushDefer)
	jumpPos := c.emit(nil, parser.OpJumpFalsy, 0)
	c.emit(nil, parser.OpCall, 0, 0, 1)
	c.emit(nil, parser.OpPop)
	c.emit(nil, parser.OpJump, curPos)
	c.changeOperand(jumpPos, len(c.currentInstructions()))
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

func resolveAssignLHS(expr parser.Expr) (name string, selectors []parser.Expr) {
	switch term := expr.(type) {
	case *parser.SelectorExpr:
		name, selectors = resolveAssignLHS(term.Expr)
		selectors = append(selectors, term.Sel)
	case *parser.IndexExpr:
		name, selectors = resolveAssignLHS(term.Expr)
		selectors = append(selectors, term.Index)
	case *parser.Ident:
		name = term.Name
	}
	return name, selectors
}

func iterateInstructions(b []byte, fn func(pos int, opcode parser.Opcode, operands []int) bool) {
	for i := 0; i < len(b); i++ {
		numOperands := parser.OpcodeOperands[b[i]]
		operands, read := parser.ReadOperands(numOperands, b[i+1:])
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
