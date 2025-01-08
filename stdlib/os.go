package stdlib

import "github.com/infastin/toy"

var OSModule = &toy.BuiltinModule{
	Name: "os",
	Members: map[string]toy.Object{
		"path": OSPathModule,
	},
}
