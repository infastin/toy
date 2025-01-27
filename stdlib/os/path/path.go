package path

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/infastin/toy"
	"github.com/infastin/toy/internal/fndef"
)

var Module = &toy.BuiltinModule{
	Name: "path",
	Members: map[string]toy.Object{
		"abs":          toy.NewBuiltinFunction("path.abs", fndef.ASRSE("path", filepath.Abs)),
		"localize":     toy.NewBuiltinFunction("path.localize", fndef.ASRSE("path", filepath.Localize)),
		"base":         toy.NewBuiltinFunction("path.base", fndef.ASRS("path", filepath.Base)),
		"dir":          toy.NewBuiltinFunction("path.dir", fndef.ASRS("path", filepath.Dir)),
		"ext":          toy.NewBuiltinFunction("path.ext", fndef.ASRS("path", filepath.Ext)),
		"clean":        toy.NewBuiltinFunction("path.clean", fndef.ASRS("path", filepath.Clean)),
		"evalSymlinks": toy.NewBuiltinFunction("path.evalSymlinks", fndef.ASRSE("path", filepath.EvalSymlinks)),
		"split":        toy.NewBuiltinFunction("path.split", fndef.ASRSS("path", filepath.Split)),
		"splitList":    toy.NewBuiltinFunction("path.splitList", fndef.ASRSs("path", filepath.SplitList)),
		"match":        toy.NewBuiltinFunction("path.match", fndef.ASSRBE("pattern", "name", filepath.Match)),
		"glob":         toy.NewBuiltinFunction("path.glob", fndef.ASRSsE("pattern", filepath.Glob)),
		"rel":          toy.NewBuiltinFunction("path.rel", fndef.ASSRSE("basepath", "targetpath", filepath.Rel)),
		"fromSlash":    toy.NewBuiltinFunction("path.fromSlash", fndef.ASRS("path", filepath.FromSlash)),
		"toSlash":      toy.NewBuiltinFunction("path.toSlash", fndef.ASRS("path", filepath.ToSlash)),
		"isAbs":        toy.NewBuiltinFunction("path.isAbs", fndef.ASRB("path", filepath.IsAbs)),
		"isLocal":      toy.NewBuiltinFunction("path.isAbs", fndef.ASRB("path", filepath.IsLocal)),
		"volumeName":   toy.NewBuiltinFunction("path.volumeName", fndef.ASRS("path", filepath.VolumeName)),
		"join":         toy.NewBuiltinFunction("path.join", joinFn),
		"expand":       toy.NewBuiltinFunction("path.expand", expandFn),
		"exists":       toy.NewBuiltinFunction("path.exists", existsFn),
		"isRegular":    toy.NewBuiltinFunction("path.isRegular", makeIsFn(os.ModeType)),
		"isDir":        toy.NewBuiltinFunction("path.isDir", makeIsFn(os.ModeDir)),
		"isSymlink":    toy.NewBuiltinFunction("path.isSymlink", makeIsFn(os.ModeSymlink)),
		"isNamedPipe":  toy.NewBuiltinFunction("path.isNamedPipe", makeIsFn(os.ModeNamedPipe)),
		"isSocket":     toy.NewBuiltinFunction("path.isSocket", makeIsFn(os.ModeSocket)),
		"isDevice":     toy.NewBuiltinFunction("path.isDevice", makeIsFn(os.ModeDevice)),
		"isCharDevice": toy.NewBuiltinFunction("path.isCharDevice", makeIsFn(os.ModeCharDevice)),
		"isIrregular":  toy.NewBuiltinFunction("path.isIrregular", makeIsFn(os.ModeIrregular)),
	},
}

func joinFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func expandFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func existsFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var name string
	if err := toy.UnpackArgs(args, "name", &name); err != nil {
		return nil, err
	}
	if _, err := os.Stat(name); err != nil {
		return toy.False, nil
	}
	return toy.True, nil
}

func makeIsFn(typ os.FileMode) toy.CallableFunc {
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
