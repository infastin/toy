package stdlib

import (
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/infastin/toy"
)

var OSModule = &toy.BuiltinModule{
	Name: "os",
	Members: map[string]toy.Object{
		"path": OSPathModule,
		"env":  OSEnvModule,
		"exec": OSExecModule,
		"proc": OSProcModule,
		"user": OSUserModule,

		"platform": toy.String(runtime.GOOS),
		"arch":     toy.String(runtime.GOARCH),

		"FileMode": &Enum{
			Name: "FileMode",
			Variants: []EnumVariant{
				{"DIR", toy.Int(os.ModeDir)},
				{"APPEND", toy.Int(os.ModeAppend)},
				{"EXCLUSIVE", toy.Int(os.ModeExclusive)},
				{"TEMPORARY", toy.Int(os.ModeTemporary)},
				{"SYMLINK", toy.Int(os.ModeSymlink)},
				{"DEVICE", toy.Int(os.ModeDevice)},
				{"NAMED_PIPE", toy.Int(os.ModeNamedPipe)},
				{"SOCKET", toy.Int(os.ModeSocket)},
				{"SETUID", toy.Int(os.ModeSetuid)},
				{"SETGID", toy.Int(os.ModeSetgid)},
				{"CHAR_DEVICE", toy.Int(os.ModeCharDevice)},
				{"STICKY", toy.Int(os.ModeSticky)},
				{"IRREGULAR", toy.Int(os.ModeIrregular)},
			},
		},

		"O": &Enum{
			Name: "O",
			Variants: []EnumVariant{
				{"RDONLY", toy.Int(os.O_RDONLY)},
				{"WRONLY", toy.Int(os.O_WRONLY)},
				{"RDWR", toy.Int(os.O_RDWR)},
				{"APPEND", toy.Int(os.O_APPEND)},
				{"CREATE", toy.Int(os.O_CREATE)},
				{"EXCL", toy.Int(os.O_EXCL)},
				{"SYNC", toy.Int(os.O_SYNC)},
				{"TRUNC", toy.Int(os.O_TRUNC)},
			},
		},

		"Seek": &Enum{
			Name: "Seek",
			Variants: []EnumVariant{
				{"SET", toy.Int(io.SeekStart)},
				{"END", toy.Int(io.SeekEnd)},
				{"CUR", toy.Int(io.SeekCurrent)},
			},
		},

		"args":       &toy.BuiltinFunction{Name: "args", Func: osArgs},
		"environ":    &toy.BuiltinFunction{Name: "environ", Func: osEnviron},
		"hostname":   &toy.BuiltinFunction{Name: "hostname", Func: makeARSE(os.Hostname)},
		"tempDir":    &toy.BuiltinFunction{Name: "tempDir", Func: makeARS(os.TempDir)},
		"executable": &toy.BuiltinFunction{Name: "executable", Func: makeARSE(os.Executable)},

		"readfile":   &toy.BuiltinFunction{Name: "readfile", Func: osReadFile},
		"writefile":  &toy.BuiltinFunction{Name: "writefile", Func: osWriteFile},
		"mkdir":      &toy.BuiltinFunction{Name: "mkdir", Func: osMkdir},
		"mkdirTemp":  &toy.BuiltinFunction{Name: "mkdirTemp", Func: osMkdirTemp},
		"remove":     &toy.BuiltinFunction{Name: "remove", Func: osRemove},
		"rename":     &toy.BuiltinFunction{Name: "rename", Func: makeASSRE("oldpath", "newpath", os.Rename)},
		"link":       &toy.BuiltinFunction{Name: "link", Func: makeASSRE("oldname", "newname", os.Link)},
		"readlink":   &toy.BuiltinFunction{Name: "readlink", Func: makeASRSE("name", os.Readlink)},
		"symlink":    &toy.BuiltinFunction{Name: "symlink", Func: makeASSRE("oldname", "newname", os.Symlink)},
		"chdir":      &toy.BuiltinFunction{Name: "chdir", Func: makeASRE("dir", os.Chdir)},
		"chmod":      &toy.BuiltinFunction{Name: "chmod", Func: osChmod},
		"chown":      &toy.BuiltinFunction{Name: "chown", Func: nil},
		"lchown":     &toy.BuiltinFunction{},
		"open":       &toy.BuiltinFunction{},
		"create":     &toy.BuiltinFunction{},
		"createTemp": &toy.BuiltinFunction{},
		"stat":       &toy.BuiltinFunction{},
		"lstat":      &toy.BuiltinFunction{},
		"truncate":   &toy.BuiltinFunction{Name: "truncate", Func: nil},
		"getwd":      &toy.BuiltinFunction{Name: "getwd", Func: makeARSE(os.Getwd)},

		"getuid":  &toy.BuiltinFunction{Name: "getuid", Func: makeARI(os.Getuid)},
		"getgid":  &toy.BuiltinFunction{Name: "getgid", Func: makeARI(os.Getgid)},
		"geteuid": &toy.BuiltinFunction{Name: "geteuid", Func: makeARI(os.Geteuid)},
		"getegid": &toy.BuiltinFunction{Name: "getegid", Func: makeARI(os.Getegid)},
		"getpid":  &toy.BuiltinFunction{Name: "getpid", Func: makeARI(os.Getpid)},
		"getppid": &toy.BuiltinFunction{Name: "getppid", Func: makeARI(os.Getppid)},
	},
}

func osReadFile(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var name string
	if err := toy.UnpackArgs(args, "name", &name); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Bytes(data), nil
}

func osWriteFile(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		name string
		data toy.StringOrBytes
		perm = 0644
	)
	if err := toy.UnpackArgs(args, "name", &name, "data", &data, "perm?", &perm); err != nil {
		return nil, err
	}
	if err := os.WriteFile(name, data.Bytes(), os.FileMode(perm)); err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Undefined, nil
}

func osMkdir(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		name string
		perm = 0755
		all  = false
	)
	if err := toy.UnpackArgs(args, "name", &name, "perm?", &perm, "all?", &all); err != nil {
		return nil, err
	}
	if all {
		if err := os.MkdirAll(name, os.FileMode(perm)); err != nil {
			return toy.NewError(err.Error()), nil
		}
	} else {
		if err := os.Mkdir(name, os.FileMode(perm)); err != nil {
			return toy.NewError(err.Error()), nil
		}
	}
	return toy.Undefined, nil
}

func osMkdirTemp(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var dir, pattern string
	if err := toy.UnpackArgs(args, "dir", &dir, "pattern", &pattern); err != nil {
		return nil, err
	}
	res, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.String(res), nil
}

func osRemove(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		name string
		all  = false
	)
	if err := toy.UnpackArgs(args, "name", &name, "all?", &all); err != nil {
		return nil, err
	}
	if all {
		if err := os.RemoveAll(name); err != nil {
			return toy.NewError(err.Error()), nil
		}
	} else {
		if err := os.Remove(name); err != nil {
			return toy.NewError(err.Error()), nil
		}
	}
	return toy.Undefined, nil
}

func osArgs(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	elems := make([]toy.Object, 0, len(os.Args))
	for _, arg := range os.Args {
		elems = append(elems, toy.String(arg))
	}
	return toy.NewArray(elems), nil
}

func osEnviron(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	envs := os.Environ()
	m := toy.NewMap(len(envs))
	for _, env := range envs {
		parts := strings.SplitN(env, "=", 2)
		m.IndexSet(toy.String(parts[0]), toy.String(parts[1]))
	}
	return m, nil
}

func osChmod(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		name string
		mode int
	)
	if err := toy.UnpackArgs(args, "name", &name, "mode", &mode); err != nil {
		return nil, err
	}
	if err := os.Chmod(name, os.FileMode(mode)); err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Undefined, nil
}
