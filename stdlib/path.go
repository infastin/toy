package stdlib

import (
	"fmt"
	"os"
	"path"

	"github.com/infastin/toy"
)

var PathModule = &toy.BuiltinModule{
	Name: "path",
	Members: map[string]toy.Object{
		"separator":     toy.Char(os.PathSeparator),
		"listSeparator": toy.Char(os.PathListSeparator),

		"join":  &toy.BuiltinFunction{Name: "path.join", Func: pathJoin},
		"base":  &toy.BuiltinFunction{Name: "path.base", Func: makeASRS("path", path.Base)},
		"dir":   &toy.BuiltinFunction{Name: "path.dir", Func: makeASRS("path", path.Dir)},
		"ext":   &toy.BuiltinFunction{Name: "path.ext", Func: makeASRS("path", path.Ext)},
		"clean": &toy.BuiltinFunction{Name: "path.clean", Func: makeASRS("path", path.Clean)},
		"split": &toy.BuiltinFunction{Name: "path.split", Func: pathSplit},
	},
}

func pathJoin(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func pathSplit(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s string
	if err := toy.UnpackArgs(args, "path", &s); err != nil {
		return nil, err
	}
	dir, file := path.Split(s)
	return toy.Tuple{toy.String(dir), toy.String(file)}, nil
}
