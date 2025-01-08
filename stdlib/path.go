package stdlib

import (
	"fmt"
	"path"

	"github.com/infastin/toy"
)

var PathModule = &toy.BuiltinModule{
	Name: "path",
	Members: map[string]toy.Object{
		"join":  &toy.BuiltinFunction{Name: "join", Func: pathJoin},
		"base":  &toy.BuiltinFunction{Name: "base", Func: makeASRS("path", path.Base)},
		"dir":   &toy.BuiltinFunction{Name: "dir", Func: makeASRS("path", path.Dir)},
		"ext":   &toy.BuiltinFunction{Name: "ext", Func: makeASRS("path", path.Ext)},
		"clean": &toy.BuiltinFunction{Name: "clean", Func: makeASRS("path", path.Clean)},
		"split": &toy.BuiltinFunction{Name: "split", Func: pathSplit},
	},
}

func pathJoin(args ...toy.Object) (toy.Object, error) {
	var elems []string
	for i, arg := range args {
		str, ok := arg.(toy.String)
		if !ok {
			return nil, &toy.InvalidArgumentTypeError{
				Name: fmt.Sprintf("elems[%d]", i),
				Want: "string",
				Got:  arg.TypeName(),
			}
		}
		elems = append(elems, string(str))
	}
	return toy.String(path.Join(elems...)), nil
}

func pathSplit(args ...toy.Object) (toy.Object, error) {
	var s string
	if err := toy.UnpackArgs(args, "path", &s); err != nil {
		return nil, err
	}
	dir, file := path.Split(s)
	return toy.Tuple{toy.String(dir), toy.String(file)}, nil
}
