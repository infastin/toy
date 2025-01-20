package env

import (
	"os"

	"github.com/infastin/toy"
	"github.com/infastin/toy/internal/fndef"
)

var Module = &toy.BuiltinModule{
	Name: "env",
	Members: map[string]toy.Object{
		"expand": toy.NewBuiltinFunction("env.expand", fndef.ASRS("s", os.ExpandEnv)),
		"clear":  toy.NewBuiltinFunction("env.clear", fndef.AR(os.Clearenv)),
		"get":    toy.NewBuiltinFunction("env.get", fndef.ASRS("key", os.Getenv)),
		"set":    toy.NewBuiltinFunction("env.set", fndef.ASSRE("key", "value", os.Setenv)),
		"unset":  toy.NewBuiltinFunction("env.unset", fndef.ASRE("key", os.Unsetenv)),
		"lookup": toy.NewBuiltinFunction("env.lookup", fndef.ASRSB("key", os.LookupEnv)),
	},
}
