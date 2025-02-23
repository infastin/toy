package toy

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/infastin/toy/parser"
	"github.com/infastin/toy/token"
)

// Script can simplify compilation and execution of embedded scripts.
type Script struct {
	variables        map[string]*Variable
	modules          ModuleGetter
	input            []byte
	enableFileImport bool
	importDir        string
}

// NewScript creates a Script instance with an input script.
func NewScript(input []byte) *Script {
	return &Script{
		variables: make(map[string]*Variable),
		input:     input,
	}
}

// Add adds a new variable or updates an existing variable to the script.
func (s *Script) Add(name string, value Value) {
	s.variables[name] = &Variable{name, value}
}

// AddAll adds or updates multiple variables to the script.
func (s *Script) AddAll(variables []*Variable) {
	for _, v := range variables {
		s.variables[v.name] = v
	}
}

// Remove removes (undefines) an existing variable for the script. It returns
// false if the variable name is not defined.
func (s *Script) Remove(name string) bool {
	if _, ok := s.variables[name]; !ok {
		return false
	}
	delete(s.variables, name)
	return true
}

// SetImports sets import modules.
func (s *Script) SetImports(modules ModuleGetter) {
	s.modules = modules
}

// SetImportDir sets the initial import directory for script files.
func (s *Script) SetImportDir(dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	s.importDir = dir
	return nil
}

// EnableFileImport enables or disables module loading from local files. Local
// file modules are disabled by default.
func (s *Script) EnableFileImport(enable bool) {
	s.enableFileImport = enable
}

// Compile compiles the script with all the defined variables,
// and returns Compiled object.
func (s *Script) Compile() (*Compiled, error) {
	symbolTable, globals := s.prepCompile()

	fileSet := token.NewFileSet()
	srcFile := fileSet.AddFile("(main)", -1, len(s.input))
	p := parser.NewParser(srcFile, s.input, nil)
	file, err := p.ParseFile()
	if err != nil {
		return nil, err
	}

	c := NewCompiler(srcFile, symbolTable, nil, s.modules, nil)
	c.EnableFileImport(s.enableFileImport)
	c.SetImportDir(s.importDir)
	if err := c.Compile(file); err != nil {
		return nil, err
	}

	// reduce globals size
	globals = globals[:symbolTable.MaxSymbols()+1]

	// global symbol names to indexes
	globalIndexes := make(map[string]int, len(globals))
	for _, name := range symbolTable.Names() {
		symbol, _, _ := symbolTable.Resolve(name, false)
		if symbol.Scope == ScopeGlobal {
			globalIndexes[name] = symbol.Index
		}
	}

	// remove duplicates from constants
	bytecode := c.Bytecode()
	bytecode.RemoveDuplicates()
	bytecode.RemoveUnused()

	return &Compiled{
		globalIndexes: globalIndexes,
		bytecode:      bytecode,
		globals:       globals,
	}, nil
}

// Run compiles and runs the scripts. Use returned compiled object to access
// global variables.
func (s *Script) Run() (compiled *Compiled, err error) {
	compiled, err = s.Compile()
	if err != nil {
		return
	}
	if err := compiled.Run(); err != nil {
		return nil, err
	}
	return compiled, nil
}

// RunContext is like Run but includes a context.
func (s *Script) RunContext(ctx context.Context) (compiled *Compiled, err error) {
	compiled, err = s.Compile()
	if err != nil {
		return
	}
	if err := compiled.RunContext(ctx); err != nil {
		return nil, err
	}
	return compiled, nil
}

func (s *Script) prepCompile() (symbolTable *SymbolTable, globals []Value) {
	var names []string
	for name := range s.variables {
		names = append(names, name)
	}

	symbolTable = NewSymbolTable()
	for i, v := range Universe {
		symbolTable.DefineBuiltin(i, v.name)
	}

	globals = make([]Value, GlobalsSize)

	for idx, name := range names {
		symbol := symbolTable.Define(name)
		if symbol.Index != idx {
			panic(fmt.Errorf("wrong symbol index: %d != %d", idx, symbol.Index))
		}
		globals[symbol.Index] = s.variables[name].Value()
	}

	return symbolTable, globals
}

// Compiled is a compiled instance of the user script.
// Use Script.Compile() to create Compiled object.
type Compiled struct {
	globalIndexes map[string]int // global symbol name to index
	bytecode      *Bytecode
	globals       []Value
	lock          sync.RWMutex
}

// Run executes the compiled script in the virtual machine.
func (c *Compiled) Run() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	r := NewRuntime(c.bytecode, c.globals)
	return r.Run()
}

// RunContext is like Run but includes a context.
func (c *Compiled) RunContext(ctx context.Context) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	r := NewRuntime(c.bytecode, c.globals)
	ch := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				switch e := r.(type) {
				case string:
					ch <- errors.New(e)
				case error:
					ch <- e
				default:
					ch <- fmt.Errorf("unknown panic: %v", e)
				}
			}
		}()
		ch <- r.Run()
	}()

	select {
	case <-ctx.Done():
		r.Abort()
		<-ch
		err = ctx.Err()
	case err = <-ch:
	}

	return err
}

// Bytecode returns a compiled bytecode.
func (c *Compiled) Bytecode() *Bytecode {
	return c.bytecode
}

// Clone creates a new copy of Compiled. Cloned copies are safe for concurrent
// use by multiple goroutines.
func (c *Compiled) Clone() *Compiled {
	c.lock.RLock()
	defer c.lock.RUnlock()

	clone := &Compiled{
		globalIndexes: c.globalIndexes,
		bytecode:      c.bytecode,
		globals:       make([]Value, len(c.globals)),
	}

	// copy global objects
	for idx, g := range c.globals {
		if g != nil {
			clone.globals[idx] = g.Clone()
		}
	}

	return clone
}

// IsDefined returns true if the variable name is defined (has value) before or
// after the execution.
func (c *Compiled) IsDefined(name string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	idx, ok := c.globalIndexes[name]
	if !ok {
		return false
	}

	v := c.globals[idx]
	if v == nil {
		return false
	}

	return v != Nil
}

// Get returns a variable identified by the name.
func (c *Compiled) Get(name string) *Variable {
	c.lock.RLock()
	defer c.lock.RUnlock()

	value := Value(Nil)
	if idx, ok := c.globalIndexes[name]; ok {
		value = c.globals[idx]
		if value == nil {
			value = Nil
		}
	}

	return &Variable{
		name:  name,
		value: value,
	}
}

// GetAll returns all the variables that are defined by the compiled script.
func (c *Compiled) GetAll() []*Variable {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var vars []*Variable
	for name, idx := range c.globalIndexes {
		value := c.globals[idx]
		if value == nil {
			value = Nil
		}
		vars = append(vars, &Variable{
			name:  name,
			value: value,
		})
	}

	return vars
}

// Set replaces the value of a global variable identified by the name. An error
// will be returned if the name was not defined during compilation.
func (c *Compiled) Set(name string, value Value) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	idx, ok := c.globalIndexes[name]
	if !ok {
		return fmt.Errorf("'%s' is not defined", name)
	}
	c.globals[idx] = value

	return nil
}
