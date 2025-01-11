package stdlib

import "github.com/infastin/toy"

var OSProcModule = &toy.BuiltinModule{
	Name: "proc",
	Members: map[string]toy.Object{
		"find":  &toy.BuiltinFunction{},
		"start": &toy.BuiltinFunction{},
	},
}
