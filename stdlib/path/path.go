package path

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/infastin/toy"
	"github.com/infastin/toy/internal/fndef"
)

var Module = &toy.BuiltinModule{
	Name: "path",
	Members: map[string]toy.Value{
		"separator":     toy.Char(os.PathSeparator),
		"listSeparator": toy.Char(os.PathListSeparator),

		"join":  toy.NewBuiltinFunction("path.join", joinFn),
		"base":  toy.NewBuiltinFunction("path.base", fndef.ASRS("path", path.Base)),
		"dir":   toy.NewBuiltinFunction("path.dir", fndef.ASRS("path", path.Dir)),
		"ext":   toy.NewBuiltinFunction("path.ext", fndef.ASRS("path", path.Ext)),
		"noext": toy.NewBuiltinFunction("path.noext", noextFn),
		"stem":  toy.NewBuiltinFunction("path.stem", stemFn),
		"clean": toy.NewBuiltinFunction("path.clean", fndef.ASRS("path", path.Clean)),
		"split": toy.NewBuiltinFunction("path.split", fndef.ASRSS("path", path.Split)),
	},
}

func joinFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
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

func noextFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var s string
	if err := toy.UnpackArgs(args, "path", &s); err != nil {
		return nil, err
	}
	return toy.String(strings.TrimSuffix(s, path.Ext(s))), nil
}

func stemFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var s string
	if err := toy.UnpackArgs(args, "path", &s); err != nil {
		return nil, err
	}
	s = path.Base(s)
	return toy.String(strings.TrimSuffix(s, path.Ext(s))), nil
}
