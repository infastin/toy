package stdlib

import "github.com/infastin/toy"

var OSExecModule = &toy.BuiltinModule{
	Name: "exec",
	Members: map[string]toy.Object{
		"command":  &toy.BuiltinFunction{},
		"lookPath": &toy.BuiltinFunction{},
	},
}
