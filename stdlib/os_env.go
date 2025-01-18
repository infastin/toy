package stdlib

import (
	"os"

	"github.com/infastin/toy"
)

var OSEnvModule = &toy.BuiltinModule{
	Name: "env",
	Members: map[string]toy.Object{
		"expand": toy.NewBuiltinFunction("env.expand", makeASRS("s", os.ExpandEnv)),
		"clear":  toy.NewBuiltinFunction("env.clear", makeAR(os.Clearenv)),
		"get":    toy.NewBuiltinFunction("env.get", makeASRS("key", os.Getenv)),
		"set":    toy.NewBuiltinFunction("env.set", makeASSRE("key", "value", os.Setenv)),
		"unset":  toy.NewBuiltinFunction("env.unset", makeASRE("key", os.Unsetenv)),
		"lookup": toy.NewBuiltinFunction("env.lookup", makeASRSB("key", os.LookupEnv)),
	},
}
