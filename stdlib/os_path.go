package stdlib

import (
	"fmt"
	"os/user"
	"path/filepath"

	"github.com/infastin/toy"
)

var OSPathModule = &toy.BuiltinModule{
	Name: "path",
	Members: map[string]toy.Object{
		"join":      &toy.BuiltinFunction{Name: "join", Func: osPathJoin},
		"base":      &toy.BuiltinFunction{Name: "base", Func: makeASRS("path", filepath.Base)},
		"dir":       &toy.BuiltinFunction{Name: "dir", Func: makeASRS("path", filepath.Dir)},
		"ext":       &toy.BuiltinFunction{Name: "ext", Func: makeASRS("path", filepath.Ext)},
		"clean":     &toy.BuiltinFunction{Name: "clean", Func: makeASRS("path", filepath.Clean)},
		"split":     &toy.BuiltinFunction{Name: "split", Func: osPathSplit},
		"splitList": &toy.BuiltinFunction{Name: "splitList", Func: osPathSplitList},
		"glob":      &toy.BuiltinFunction{Name: "glob", Func: osPathGlob},
		"expand":    &toy.BuiltinFunction{Name: "expand", Func: osPathExpand},
	},
}

func osPathJoin(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
	return toy.String(filepath.Join(elems...)), nil
}

func osPathSplit(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s string
	if err := toy.UnpackArgs(args, "path", &s); err != nil {
		return nil, err
	}
	dir, file := filepath.Split(s)
	return toy.Tuple{toy.String(dir), toy.String(file)}, nil
}

func osPathSplitList(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s string
	if err := toy.UnpackArgs(args, "path", &s); err != nil {
		return nil, err
	}
	parts := filepath.SplitList(s)
	elems := make([]toy.Object, 0, len(parts))
	for _, part := range parts {
		elems = append(elems, toy.String(part))
	}
	return toy.NewArray(elems), nil
}

func osPathGlob(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var pattern string
	if err := toy.UnpackArgs(args, "pattern", &pattern); err != nil {
		return nil, err
	}
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return toy.NewError(err.Error()), nil
	}
	elems := make([]toy.Object, 0, len(matches))
	for _, match := range matches {
		elems = append(elems, toy.String(match))
	}
	return toy.NewArray(elems), nil
}

func osPathExpand(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s string
	if err := toy.UnpackArgs(args, "path", &s); err != nil {
		return nil, err
	}
	if len(s) == 0 || s[0] != '~' {
		return toy.String(s), nil
	}
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	return toy.String(filepath.Join(usr.HomeDir, s[1:])), nil
}
