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

		"join":  toy.NewBuiltinFunction("path.join", pathJoin),
		"base":  toy.NewBuiltinFunction("path.base", makeASRS("path", path.Base)),
		"dir":   toy.NewBuiltinFunction("path.dir", makeASRS("path", path.Dir)),
		"ext":   toy.NewBuiltinFunction("path.ext", makeASRS("path", path.Ext)),
		"clean": toy.NewBuiltinFunction("path.clean", makeASRS("path", path.Clean)),
		"split": toy.NewBuiltinFunction("path.split", pathSplit),
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
				Got:  toy.TypeName(arg),
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
