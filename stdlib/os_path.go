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
		"split":        toy.NewBuiltinFunction("path.split", makeASRSS("path", filepath.Split)),
		"splitList":    toy.NewBuiltinFunction("path.splitList", makeASRSs("path", filepath.SplitList)),
		"glob":         toy.NewBuiltinFunction("path.glob", makeASRSsE("pattern", filepath.Glob)),
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
