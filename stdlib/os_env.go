package stdlib

import (
	"os"

	"github.com/infastin/toy"
)

var OSEnvModule = &toy.BuiltinModule{
	Name: "env",
	Members: map[string]toy.Object{
		"expand": &toy.BuiltinFunction{Name: "env.expand", Func: makeASRS("s", os.ExpandEnv)},
		"clear":  &toy.BuiltinFunction{Name: "env.clear", Func: osEnvClear},
		"get":    &toy.BuiltinFunction{Name: "env.get", Func: makeASRS("key", os.Getenv)},
		"set":    &toy.BuiltinFunction{Name: "env.set", Func: makeASSRE("key", "value", os.Setenv)},
		"unset":  &toy.BuiltinFunction{Name: "env.unset", Func: makeASRE("key", os.Unsetenv)},
		"lookup": &toy.BuiltinFunction{Name: "env.lookup", Func: osEnvLookup},
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
