package stdlib

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/infastin/toy"
)

var OSPathModule = &toy.BuiltinModule{
	Name: "path",
	Members: map[string]toy.Object{
		"join":         toy.NewBuiltinFunction("path.join", osPathJoin),
		"base":         toy.NewBuiltinFunction("path.base", makeASRS("path", filepath.Base)),
		"dir":          toy.NewBuiltinFunction("path.dir", makeASRS("path", filepath.Dir)),
		"ext":          toy.NewBuiltinFunction("path.ext", makeASRS("path", filepath.Ext)),
		"clean":        toy.NewBuiltinFunction("path.clean", makeASRS("path", filepath.Clean)),
		"split":        toy.NewBuiltinFunction("path.split", osPathSplit),
		"splitList":    toy.NewBuiltinFunction("path.splitList", osPathSplitList),
		"glob":         toy.NewBuiltinFunction("path.glob", osPathGlob),
		"expand":       toy.NewBuiltinFunction("path.expand", osPathExpand),
		"exists":       toy.NewBuiltinFunction("path.exists", osPathExists),
		"isRegular":    toy.NewBuiltinFunction("path.isRegular", makeOSPathIs(os.ModeType)),
		"isDir":        toy.NewBuiltinFunction("path.isDir", makeOSPathIs(os.ModeDir)),
		"isSymlink":    toy.NewBuiltinFunction("path.isSymlink", makeOSPathIs(os.ModeSymlink)),
		"isNamedPipe":  toy.NewBuiltinFunction("path.isNamedPipe", makeOSPathIs(os.ModeNamedPipe)),
		"isSocket":     toy.NewBuiltinFunction("path.isSocket", makeOSPathIs(os.ModeSocket)),
		"isDevice":     toy.NewBuiltinFunction("path.isDevice", makeOSPathIs(os.ModeDevice)),
		"isCharDevice": toy.NewBuiltinFunction("path.isCharDevice", makeOSPathIs(os.ModeCharDevice)),
		"isIrregular":  toy.NewBuiltinFunction("path.isIrregular", makeOSPathIs(os.ModeIrregular)),
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
				Got:  toy.TypeName(arg),
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
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	elems := make([]toy.Object, 0, len(matches))
	for _, match := range matches {
		elems = append(elems, toy.String(match))
	}
	return toy.Tuple{toy.NewArray(elems), toy.Nil}, nil
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

func osPathExists(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var name string
	if err := toy.UnpackArgs(args, "name", &name); err != nil {
		return nil, err
	}
	if _, err := os.Stat(name); err != nil {
		return toy.False, nil
	}
	return toy.True, nil
}

func makeOSPathIs(typ os.FileMode) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var name string
		if err := toy.UnpackArgs(args, "name", &name); err != nil {
			return nil, err
		}
		stat, err := os.Stat(name)
		if err != nil {
			return toy.False, nil
		}
		return toy.Bool(stat.Mode().Type() == typ), nil
	}
}
