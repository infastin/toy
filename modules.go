package toy

// Importable interface represents importable module instance.
type Importable interface {
	importable()
}

// ModuleGetter enables implementing dynamic module loading.
type ModuleGetter interface {
	Get(name string) Importable
}

// ModuleMap represents a set of named modules. Use NewModuleMap to create a
// new module map.
type ModuleMap struct {
	m map[string]Importable
}

// NewModuleMap creates a new module map.
func NewModuleMap() *ModuleMap {
	return &ModuleMap{
		m: make(map[string]Importable),
	}
}

// Add adds an import module.
func (m *ModuleMap) Add(name string, module Importable) {
	m.m[name] = module
}

// AddBuiltinModule adds a builtin module.
func (m *ModuleMap) AddBuiltinModule(name string, fields map[string]Object) {
	m.m[name] = &BuiltinModule{Name: name, Fields: fields}
}

// AddSourceModule adds a source module.
func (m *ModuleMap) AddSourceModule(name string, src []byte) {
	m.m[name] = SourceModule(src)
}

// Remove removes a named module.
func (m *ModuleMap) Remove(name string) {
	delete(m.m, name)
}

// Get returns an import module identified by name. It returns if the name is
// not found.
func (m *ModuleMap) Get(name string) Importable {
	return m.m[name]
}

// GetBuiltinModule returns a builtin module identified by name. It returns
// if the name is not found or the module is not a builtin module.
func (m *ModuleMap) GetBuiltinModule(name string) *BuiltinModule {
	mod, _ := m.m[name].(*BuiltinModule)
	return mod
}

// GetSourceModule returns a source module identified by name. It returns if
// the name is not found or the module is not a source module.
func (m *ModuleMap) GetSourceModule(name string) *SourceModule {
	mod, _ := m.m[name].(*SourceModule)
	return mod
}

// Copy creates a copy of the module map.
func (m *ModuleMap) Copy() *ModuleMap {
	c := &ModuleMap{
		m: make(map[string]Importable),
	}
	for name, mod := range m.m {
		c.m[name] = mod
	}
	return c
}

// Len returns the number of named modules.
func (m *ModuleMap) Len() int {
	return len(m.m)
}

// AddMap adds named modules from another module map.
func (m *ModuleMap) AddMap(o *ModuleMap) {
	for name, mod := range o.m {
		m.m[name] = mod
	}
}

// BuiltinModule is an importable module that's written in Go.
type BuiltinModule struct {
	Name   string
	Fields map[string]Object
}

func (m *BuiltinModule) importable() {}

func (m *BuiltinModule) TypeName() string { return "module" }
func (m *BuiltinModule) String() string   { return "<module>" }
func (m *BuiltinModule) IsFalsy() bool    { return false }

func (m *BuiltinModule) Copy() Object {
	fields := make(map[string]Object)
	for name, value := range m.Fields {
		fields[name] = value.Copy()
	}
	return &BuiltinModule{Name: m.Name, Fields: m.Fields}
}

func (m *BuiltinModule) FieldGet(name string) (Object, error) {
	field, ok := m.Fields[name]
	if !ok {
		return nil, ErrNoSuchField
	}
	return field, nil
}

// SourceModule is an importable module that's written in Toy.
type SourceModule []byte

// Import returns a module source code.
func (m SourceModule) importable() {}
