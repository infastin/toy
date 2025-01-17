package stdlib

import (
	"os"

	"github.com/infastin/toy"
)

var OSEnvModule = &toy.BuiltinModule{
	Name: "env",
	Members: map[string]toy.Object{
		"expand": toy.NewBuiltinFunction("env.expand", makeASRS("s", os.ExpandEnv)),
		"clear":  toy.NewBuiltinFunction("env.clear", osEnvClear),
		"get":    toy.NewBuiltinFunction("env.get", makeASRS("key", os.Getenv)),
		"set":    toy.NewBuiltinFunction("env.set", makeASSRE("key", "value", os.Setenv)),
		"unset":  toy.NewBuiltinFunction("env.unset", makeASRE("key", os.Unsetenv)),
		"lookup": toy.NewBuiltinFunction("env.lookup", osEnvLookup),
	},
}

func osEnvClear(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	os.Clearenv()
	return toy.Nil, nil
}

func osEnvLookup(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var key string
	if err := toy.UnpackArgs(args, "key", &key); err != nil {
		return nil, err
	}
	value, ok := os.LookupEnv(key)
	if !ok {
		return toy.Nil, nil
	}
	return toy.String(value), nil
}
