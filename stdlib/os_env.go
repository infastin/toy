package stdlib

import (
	"os"

	"github.com/infastin/toy"
)

var OSEnvModule = &toy.BuiltinModule{
	Name: "env",
	Members: map[string]toy.Object{
		"expand": &toy.BuiltinFunction{Name: "expand", Func: makeASRS("s", os.ExpandEnv)},
		"clear":  &toy.BuiltinFunction{Name: "clear", Func: osEnvClear},
		"get":    &toy.BuiltinFunction{Name: "get", Func: makeASRS("key", os.Getenv)},
		"set":    &toy.BuiltinFunction{Name: "set", Func: makeASSRE("key", "value", os.Setenv)},
		"unset":  &toy.BuiltinFunction{Name: "unset", Func: makeASRE("key", os.Unsetenv)},
		"lookup": &toy.BuiltinFunction{Name: "lookup", Func: osEnvLookup},
	},
}

func osEnvClear(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	os.Clearenv()
	return toy.Undefined, nil
}

func osEnvLookup(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var key string
	if err := toy.UnpackArgs(args, "key", &key); err != nil {
		return nil, err
	}
	value, ok := os.LookupEnv(key)
	if !ok {
		return toy.Undefined, nil
	}
	return toy.String(value), nil
}
