package stdlib

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"

	"github.com/d5/tengo/v2"
)

var osModule = map[string]tengo.Object{
	"platform":            &tengo.String{Value: runtime.GOOS},
	"arch":                &tengo.String{Value: runtime.GOARCH},
	"o_rdonly":            &tengo.Int{Value: int64(os.O_RDONLY)},
	"o_wronly":            &tengo.Int{Value: int64(os.O_WRONLY)},
	"o_rdwr":              &tengo.Int{Value: int64(os.O_RDWR)},
	"o_append":            &tengo.Int{Value: int64(os.O_APPEND)},
	"o_create":            &tengo.Int{Value: int64(os.O_CREATE)},
	"o_excl":              &tengo.Int{Value: int64(os.O_EXCL)},
	"o_sync":              &tengo.Int{Value: int64(os.O_SYNC)},
	"o_trunc":             &tengo.Int{Value: int64(os.O_TRUNC)},
	"mode_dir":            &tengo.Int{Value: int64(os.ModeDir)},
	"mode_append":         &tengo.Int{Value: int64(os.ModeAppend)},
	"mode_exclusive":      &tengo.Int{Value: int64(os.ModeExclusive)},
	"mode_temporary":      &tengo.Int{Value: int64(os.ModeTemporary)},
	"mode_symlink":        &tengo.Int{Value: int64(os.ModeSymlink)},
	"mode_device":         &tengo.Int{Value: int64(os.ModeDevice)},
	"mode_named_pipe":     &tengo.Int{Value: int64(os.ModeNamedPipe)},
	"mode_socket":         &tengo.Int{Value: int64(os.ModeSocket)},
	"mode_setuid":         &tengo.Int{Value: int64(os.ModeSetuid)},
	"mode_setgui":         &tengo.Int{Value: int64(os.ModeSetgid)},
	"mode_char_device":    &tengo.Int{Value: int64(os.ModeCharDevice)},
	"mode_sticky":         &tengo.Int{Value: int64(os.ModeSticky)},
	"mode_type":           &tengo.Int{Value: int64(os.ModeType)},
	"mode_perm":           &tengo.Int{Value: int64(os.ModePerm)},
	"path_separator":      &tengo.Char{Value: os.PathSeparator},
	"path_list_separator": &tengo.Char{Value: os.PathListSeparator},
	"dev_null":            &tengo.String{Value: os.DevNull},
	"seek_set":            &tengo.Int{Value: int64(io.SeekStart)},
	"seek_cur":            &tengo.Int{Value: int64(io.SeekCurrent)},
	"seek_end":            &tengo.Int{Value: int64(io.SeekEnd)},
	"args": &tengo.UserFunction{
		Name: "args",
		Func: osArgs,
	}, // args() => array(string)
	"chdir": &tengo.UserFunction{
		Name: "chdir",
		Func: FuncASRE(os.Chdir),
	}, // chdir(dir string) => error
	"chmod": osFuncASFmRE("chmod", os.Chmod), // chmod(name string, mode int) => error
	"chown": &tengo.UserFunction{
		Name: "chown",
		Func: FuncASIIRE(os.Chown),
	}, // chown(name string, uid int, gid int) => error
	"clearenv": &tengo.UserFunction{
		Name: "clearenv",
		Func: FuncAR(os.Clearenv),
	}, // clearenv()
	"environ": &tengo.UserFunction{
		Name: "environ",
		Func: FuncARSs(os.Environ),
	}, // environ() => array(string)
	"exit": &tengo.UserFunction{
		Name: "exit",
		Func: FuncAIR(os.Exit),
	}, // exit(code int)
	"expand_env": &tengo.UserFunction{
		Name: "expand_env",
		Func: osExpandEnv,
	}, // expand_env(s string) => string
	"getegid": &tengo.UserFunction{
		Name: "getegid",
		Func: FuncARI(os.Getegid),
	}, // getegid() => int
	"getenv": &tengo.UserFunction{
		Name: "getenv",
		Func: FuncASRS(os.Getenv),
	}, // getenv(s string) => string
	"geteuid": &tengo.UserFunction{
		Name: "geteuid",
		Func: FuncARI(os.Geteuid),
	}, // geteuid() => int
	"getgid": &tengo.UserFunction{
		Name: "getgid",
		Func: FuncARI(os.Getgid),
	}, // getgid() => int
	"getgroups": &tengo.UserFunction{
		Name: "getgroups",
		Func: FuncARIsE(os.Getgroups),
	}, // getgroups() => array(string)/error
	"getpagesize": &tengo.UserFunction{
		Name: "getpagesize",
		Func: FuncARI(os.Getpagesize),
	}, // getpagesize() => int
	"getpid": &tengo.UserFunction{
		Name: "getpid",
		Func: FuncARI(os.Getpid),
	}, // getpid() => int
	"getppid": &tengo.UserFunction{
		Name: "getppid",
		Func: FuncARI(os.Getppid),
	}, // getppid() => int
	"getuid": &tengo.UserFunction{
		Name: "getuid",
		Func: FuncARI(os.Getuid),
	}, // getuid() => int
	"getwd": &tengo.UserFunction{
		Name: "getwd",
		Func: FuncARSE(os.Getwd),
	}, // getwd() => string/error
	"hostname": &tengo.UserFunction{
		Name: "hostname",
		Func: FuncARSE(os.Hostname),
	}, // hostname() => string/error
	"lchown": &tengo.UserFunction{
		Name: "lchown",
		Func: FuncASIIRE(os.Lchown),
	}, // lchown(name string, uid int, gid int) => error
	"link": &tengo.UserFunction{
		Name: "link",
		Func: FuncASSRE(os.Link),
	}, // link(oldname string, newname string) => error
	"lookup_env": &tengo.UserFunction{
		Name: "lookup_env",
		Func: osLookupEnv,
	}, // lookup_env(key string) => string/false
	"mkdir":     osFuncASFmRE("mkdir", os.Mkdir),        // mkdir(name string, perm int) => error
	"mkdir_all": osFuncASFmRE("mkdir_all", os.MkdirAll), // mkdir_all(name string, perm int) => error
	"readlink": &tengo.UserFunction{
		Name: "readlink",
		Func: FuncASRSE(os.Readlink),
	}, // readlink(name string) => string/error
	"remove": &tengo.UserFunction{
		Name: "remove",
		Func: FuncASRE(os.Remove),
	}, // remove(name string) => error
	"remove_all": &tengo.UserFunction{
		Name: "remove_all",
		Func: FuncASRE(os.RemoveAll),
	}, // remove_all(name string) => error
	"rename": &tengo.UserFunction{
		Name: "rename",
		Func: FuncASSRE(os.Rename),
	}, // rename(oldpath string, newpath string) => error
	"setenv": &tengo.UserFunction{
		Name: "setenv",
		Func: FuncASSRE(os.Setenv),
	}, // setenv(key string, value string) => error
	"symlink": &tengo.UserFunction{
		Name: "symlink",
		Func: FuncASSRE(os.Symlink),
	}, // symlink(oldname string newname string) => error
	"temp_dir": &tengo.UserFunction{
		Name: "temp_dir",
		Func: FuncARS(os.TempDir),
	}, // temp_dir() => string
	"truncate": &tengo.UserFunction{
		Name: "truncate",
		Func: FuncASI64RE(os.Truncate),
	}, // truncate(name string, size int) => error
	"unsetenv": &tengo.UserFunction{
		Name: "unsetenv",
		Func: FuncASRE(os.Unsetenv),
	}, // unsetenv(key string) => error
	"create": &tengo.UserFunction{
		Name: "create",
		Func: osCreate,
	}, // create(name string) => imap(file)/error
	"open": &tengo.UserFunction{
		Name: "open",
		Func: osOpen,
	}, // open(name string) => imap(file)/error
	"open_file": &tengo.UserFunction{
		Name: "open_file",
		Func: osOpenFile,
	}, // open_file(name string, flag int, perm int) => imap(file)/error
	"find_process": &tengo.UserFunction{
		Name: "find_process",
		Func: osFindProcess,
	}, // find_process(pid int) => imap(process)/error
	"start_process": &tengo.UserFunction{
		Name: "start_process",
		Func: osStartProcess,
	}, // start_process(name string, argv array(string), dir string, env array(string)) => imap(process)/error
	"exec_look_path": &tengo.UserFunction{
		Name: "exec_look_path",
		Func: FuncASRSE(exec.LookPath),
	}, // exec_look_path(file) => string/error
	"exec": &tengo.UserFunction{
		Name: "exec",
		Func: osExec,
	}, // exec(name, args...) => command
	"stat": &tengo.UserFunction{
		Name: "stat",
		Func: osStat,
	}, // stat(name) => imap(fileinfo)/error
	"read_file": &tengo.UserFunction{
		Name: "read_file",
		Func: osReadFile,
	}, // readfile(name) => array(byte)/error
}

func osReadFile(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	fname, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}
	bytes, err := ioutil.ReadFile(fname)
	if err != nil {
		return wrapError(err), nil
	}
	if len(bytes) > tengo.MaxBytesLen {
		return nil, tengo.ErrBytesLimit
	}
	return &tengo.Bytes{Value: bytes}, nil
}

func osStat(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	fname, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}
	stat, err := os.Stat(fname)
	if err != nil {
		return wrapError(err), nil
	}
	fstat := &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"name":  &tengo.String{Value: stat.Name()},
			"mtime": &tengo.Time{Value: stat.ModTime()},
			"size":  &tengo.Int{Value: stat.Size()},
			"mode":  &tengo.Int{Value: int64(stat.Mode())},
		},
	}
	if stat.IsDir() {
		fstat.Value["directory"] = tengo.True
	} else {
		fstat.Value["directory"] = tengo.False
	}
	return fstat, nil
}

func osCreate(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	s1, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}
	res, err := os.Create(s1)
	if err != nil {
		return wrapError(err), nil
	}
	return makeOSFile(res), nil
}

func osOpen(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	s1, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}
	res, err := os.Open(s1)
	if err != nil {
		return wrapError(err), nil
	}
	return makeOSFile(res), nil
}

func osOpenFile(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 3 {
		return nil, tengo.ErrWrongNumArguments
	}
	s1, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}
	i2, ok := tengo.ToInt(args[1])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "int(compatible)",
			Found:    args[1].TypeName(),
		}
	}
	i3, ok := tengo.ToInt(args[2])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "third",
			Expected: "int(compatible)",
			Found:    args[2].TypeName(),
		}
	}
	res, err := os.OpenFile(s1, i2, os.FileMode(i3))
	if err != nil {
		return wrapError(err), nil
	}
	return makeOSFile(res), nil
}

func osArgs(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}
	arr := &tengo.Array{}
	for _, osArg := range os.Args {
		if len(osArg) > tengo.MaxStringLen {
			return nil, tengo.ErrStringLimit
		}
		arr.Value = append(arr.Value, &tengo.String{Value: osArg})
	}
	return arr, nil
}

func osFuncASFmRE(
	name string,
	fn func(string, os.FileMode) error,
) *tengo.UserFunction {
	return &tengo.UserFunction{
		Name: name,
		Func: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) != 2 {
				return nil, tengo.ErrWrongNumArguments
			}
			s1, ok := tengo.ToString(args[0])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "first",
					Expected: "string(compatible)",
					Found:    args[0].TypeName(),
				}
			}
			i2, ok := tengo.ToInt64(args[1])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "second",
					Expected: "int(compatible)",
					Found:    args[1].TypeName(),
				}
			}
			return wrapError(fn(s1, os.FileMode(i2))), nil
		},
	}
}

func osLookupEnv(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	s1, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}
	res, ok := os.LookupEnv(s1)
	if !ok {
		return tengo.False, nil
	}
	if len(res) > tengo.MaxStringLen {
		return nil, tengo.ErrStringLimit
	}
	return &tengo.String{Value: res}, nil
}

func osExpandEnv(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	s1, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}
	var vlen int
	var failed bool
	s := os.Expand(s1, func(k string) string {
		if failed {
			return ""
		}
		v := os.Getenv(k)

		// this does not count the other texts that are not being replaced
		// but the code checks the final length at the end
		vlen += len(v)
		if vlen > tengo.MaxStringLen {
			failed = true
			return ""
		}
		return v
	})
	if failed || len(s) > tengo.MaxStringLen {
		return nil, tengo.ErrStringLimit
	}
	return &tengo.String{Value: s}, nil
}

func osExec(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return nil, tengo.ErrWrongNumArguments
	}
	name, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}
	var execArgs []string
	for idx, arg := range args[1:] {
		execArg, ok := tengo.ToString(arg)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     fmt.Sprintf("args[%d]", idx),
				Expected: "string(compatible)",
				Found:    args[1+idx].TypeName(),
			}
		}
		execArgs = append(execArgs, execArg)
	}
	return makeOSExecCommand(exec.Command(name, execArgs...)), nil
}

func osFindProcess(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	i1, ok := tengo.ToInt(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "int(compatible)",
			Found:    args[0].TypeName(),
		}
	}
	proc, err := os.FindProcess(i1)
	if err != nil {
		return wrapError(err), nil
	}
	return makeOSProcess(proc), nil
}

func osStartProcess(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 4 {
		return nil, tengo.ErrWrongNumArguments
	}
	name, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}
	var argv []string
	var err error
	switch arg1 := args[1].(type) {
	case *tengo.Array:
		argv, err = stringArray(arg1.Value, "second")
		if err != nil {
			return nil, err
		}
	case *tengo.ImmutableArray:
		argv, err = stringArray(arg1.Value, "second")
		if err != nil {
			return nil, err
		}
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "array",
			Found:    arg1.TypeName(),
		}
	}

	dir, ok := tengo.ToString(args[2])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "third",
			Expected: "string(compatible)",
			Found:    args[2].TypeName(),
		}
	}

	var env []string
	switch arg3 := args[3].(type) {
	case *tengo.Array:
		env, err = stringArray(arg3.Value, "fourth")
		if err != nil {
			return nil, err
		}
	case *tengo.ImmutableArray:
		env, err = stringArray(arg3.Value, "fourth")
		if err != nil {
			return nil, err
		}
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "fourth",
			Expected: "array",
			Found:    arg3.TypeName(),
		}
	}

	proc, err := os.StartProcess(name, argv, &os.ProcAttr{
		Dir: dir,
		Env: env,
	})
	if err != nil {
		return wrapError(err), nil
	}
	return makeOSProcess(proc), nil
}

func stringArray(arr []tengo.Object, argName string) ([]string, error) {
	var sarr []string
	for idx, elem := range arr {
		str, ok := elem.(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     fmt.Sprintf("%s[%d]", argName, idx),
				Expected: "string",
				Found:    elem.TypeName(),
			}
		}
		sarr = append(sarr, str.Value)
	}
	return sarr, nil
}
