package toy

import (
	"fmt"
	"maps"
)

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
type ModuleMap map[string]Importable

// Add adds an import module.
func (m ModuleMap) Add(name string, module Importable) {
	m[name] = module
}

// AddBuiltinModule adds a builtin module.
func (m ModuleMap) AddBuiltinModule(name string, fields map[string]Object) {
	m[name] = &BuiltinModule{Name: name, Members: fields}
}

// AddSourceModule adds a source module.
func (m ModuleMap) AddSourceModule(name string, src []byte) {
	m[name] = SourceModule(src)
}

// Remove removes a named module.
func (m ModuleMap) Remove(name string) {
	delete(m, name)
}

// Get returns an import module identified by name. It returns if the name is
// not found.
func (m ModuleMap) Get(name string) Importable {
	return m[name]
}

// GetBuiltinModule returns a builtin module identified by name. It returns
// if the name is not found or the module is not a builtin module.
func (m ModuleMap) GetBuiltinModule(name string) *BuiltinModule {
	mod, _ := m[name].(*BuiltinModule)
	return mod
}

// GetSourceModule returns a source module identified by name. It returns if
// the name is not found or the module is not a source module.
func (m ModuleMap) GetSourceModule(name string) *SourceModule {
	mod, _ := m[name].(*SourceModule)
	return mod
}

// Copy creates a copy of the module map.
func (m ModuleMap) Copy() ModuleMap {
	return maps.Clone(m)
}

// Len returns the number of named modules.
func (m ModuleMap) Len() int {
	return len(m)
}

// AddMap adds named modules from another module map.
func (m ModuleMap) AddMap(o ModuleMap) {
	maps.Insert(m, maps.All(o))
}

// BuiltinModule is an importable module that's written in Go.
type BuiltinModule struct {
	Name    string
	Members map[string]Object
}

var BuiltinModuleType = NewType[*BuiltinModule]("module", nil)

func (m *BuiltinModule) importable() {}

func (m *BuiltinModule) Type() ObjectType { return BuiltinModuleType }
func (m *BuiltinModule) String() string   { return fmt.Sprintf("<module %q>", m.Name) }
func (m *BuiltinModule) IsFalsy() bool    { return false }

func (m *BuiltinModule) Clone() Object {
	fields := make(map[string]Object)
	for name, value := range m.Members {
		fields[name] = value.Clone()
	}
	return &BuiltinModule{Name: m.Name, Members: m.Members}
}

func (m *BuiltinModule) FieldGet(name string) (Object, error) {
	field, ok := m.Members[name]
	if !ok {
		return nil, ErrNoSuchField
	}
	return field, nil
}

// SourceModule is an importable module that's written in Toy.
type SourceModule []byte

// Import returns a module source code.
func (m SourceModule) importable() {}
