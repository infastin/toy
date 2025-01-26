package os

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/infastin/toy"
	"github.com/infastin/toy/internal/fndef"
	"github.com/infastin/toy/stdlib/enum"
	"github.com/infastin/toy/stdlib/time"
	"github.com/infastin/toy/token"
)

var Module = &toy.BuiltinModule{
	Name: "os",
	Members: map[string]toy.Object{
		"platform": toy.String(runtime.GOOS),
		"arch":     toy.String(runtime.GOARCH),
		"devnull":  toy.String(os.DevNull),

		"stdin":  (*File)(os.Stdin),
		"stdout": (*File)(os.Stdout),
		"stderr": (*File)(os.Stderr),

		"O":    O,
		"Seek": Seek,

		"FileMode": FileModeType,
		"FileInfo": FileInfoType,
		"DirEntry": DirEntryType,
		"File":     FileType,

		"args":       toy.NewBuiltinFunction("os.args", argsFn),
		"environ":    toy.NewBuiltinFunction("os.environ", environFn),
		"hostname":   toy.NewBuiltinFunction("os.hostname", fndef.ARSE(os.Hostname)),
		"tempDir":    toy.NewBuiltinFunction("os.tempDir", fndef.ARS(os.TempDir)),
		"executable": toy.NewBuiltinFunction("os.executable", fndef.ARSE(os.Executable)),

		"readfile":   toy.NewBuiltinFunction("os.readfile", readFileFn),
		"writefile":  toy.NewBuiltinFunction("os.writefile", writeFileFn),
		"readdir":    toy.NewBuiltinFunction("os.readdir", readDirFn),
		"mkdir":      toy.NewBuiltinFunction("os.mkdir", mkdirFn),
		"mkdirTemp":  toy.NewBuiltinFunction("os.mkdirTemp", mkdirTempFn),
		"remove":     toy.NewBuiltinFunction("os.remove", removeFn),
		"rename":     toy.NewBuiltinFunction("os.rename", fndef.ASSRE("oldpath", "newpath", os.Rename)),
		"link":       toy.NewBuiltinFunction("os.link", fndef.ASSRE("oldname", "newname", os.Link)),
		"readlink":   toy.NewBuiltinFunction("os.readlink", fndef.ASRSE("name", os.Readlink)),
		"symlink":    toy.NewBuiltinFunction("os.symlink", fndef.ASSRE("oldname", "newname", os.Symlink)),
		"chdir":      toy.NewBuiltinFunction("os.chdir", fndef.ASRE("dir", os.Chdir)),
		"chmod":      toy.NewBuiltinFunction("os.chmod", chmodFn),
		"chown":      toy.NewBuiltinFunction("os.chown", chownFn),
		"lchown":     toy.NewBuiltinFunction("os.lchown", lchownFn),
		"open":       toy.NewBuiltinFunction("os.open", openFn),
		"create":     toy.NewBuiltinFunction("os.create", createFn),
		"createTemp": toy.NewBuiltinFunction("os.createTemp", createTempFn),
		"stat":       toy.NewBuiltinFunction("os.stat", statFn),
		"lstat":      toy.NewBuiltinFunction("os.lstat", lstatFn),
		"truncate":   toy.NewBuiltinFunction("os.truncate", truncateFn),
		"getwd":      toy.NewBuiltinFunction("os.getwd", fndef.ARSE(os.Getwd)),

		"getuid":  toy.NewBuiltinFunction("os.getuid", fndef.ARI(os.Getuid)),
		"getgid":  toy.NewBuiltinFunction("os.getgid", fndef.ARI(os.Getgid)),
		"geteuid": toy.NewBuiltinFunction("os.geteuid", fndef.ARI(os.Geteuid)),
		"getegid": toy.NewBuiltinFunction("os.getegid", fndef.ARI(os.Getegid)),
		"getpid":  toy.NewBuiltinFunction("os.getpid", fndef.ARI(os.Getpid)),
		"getppid": toy.NewBuiltinFunction("os.getppid", fndef.ARI(os.Getppid)),
	},
}

var O = enum.New("os.O", map[string]toy.Int{
	"RDONLY": toy.Int(os.O_RDONLY),
	"WRONLY": toy.Int(os.O_WRONLY),
	"RDWR":   toy.Int(os.O_RDWR),
	"APPEND": toy.Int(os.O_APPEND),
	"CREATE": toy.Int(os.O_CREATE),
	"EXCL":   toy.Int(os.O_EXCL),
	"SYNC":   toy.Int(os.O_SYNC),
	"TRUNC":  toy.Int(os.O_TRUNC),
}, nil)

var Seek = enum.New("os.Seek", map[string]toy.Int{
	"SET": toy.Int(io.SeekStart),
	"END": toy.Int(io.SeekEnd),
	"CUR": toy.Int(io.SeekCurrent),
}, nil)

type FileMode os.FileMode

var FileModeType = enum.New("os.FileMode", map[string]FileMode{
	"DIR":         FileMode(os.ModeDir),
	"APPEND":      FileMode(os.ModeAppend),
	"EXCLUSIVE":   FileMode(os.ModeExclusive),
	"TEMPORARY":   FileMode(os.ModeTemporary),
	"SYMLINK":     FileMode(os.ModeSymlink),
	"DEVICE":      FileMode(os.ModeDevice),
	"NAMED_PIPE":  FileMode(os.ModeNamedPipe),
	"SOCKET":      FileMode(os.ModeSocket),
	"SETUID":      FileMode(os.ModeSetuid),
	"SETGID":      FileMode(os.ModeSetgid),
	"CHAR_DEVICE": FileMode(os.ModeCharDevice),
	"STICKY":      FileMode(os.ModeSticky),
	"IRREGULAR":   FileMode(os.ModeIrregular),
	"TYPE":        FileMode(os.ModeType),
	"PERM":        FileMode(os.ModePerm),
}, func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	switch x := args[0].(type) {
	case toy.Int:
		return FileMode(x), nil
	default:
		var m FileMode
		if err := toy.Convert(&m, x); err != nil {
			return nil, err
		}
		return m, nil
	}
})

func (m FileMode) Type() toy.ObjectType { return FileModeType }
func (m FileMode) String() string       { return fmt.Sprintf("os.FileMode(%q)", os.FileMode(m).String()) }
func (m FileMode) IsFalsy() bool        { return false }
func (m FileMode) Clone() toy.Object    { return m }

func (m FileMode) Convert(p any) error {
	switch p := p.(type) {
	case *toy.Int:
		*p = toy.Int(m)
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func (m FileMode) FieldGet(name string) (toy.Object, error) {
	switch name {
	case "type":
		return FileMode(os.FileMode(m).Type()), nil
	case "perm":
		return FileMode(os.FileMode(m).Perm()), nil
	case "isDir":
		return toy.Bool(os.FileMode(m)&os.ModeDir != 0), nil
	case "isRegular":
		return toy.Bool(os.FileMode(m)&os.ModeType == 0), nil
	case "isSymlink":
		return toy.Bool(os.FileMode(m)&os.ModeSymlink != 0), nil
	case "isNamedPipe":
		return toy.Bool(os.FileMode(m)&os.ModeNamedPipe != 0), nil
	case "isSocket":
		return toy.Bool(os.FileMode(m)&os.ModeSocket != 0), nil
	case "isDevice":
		return toy.Bool(os.FileMode(m)&os.ModeDevice != 0), nil
	case "isCharDevice":
		return toy.Bool(os.FileMode(m)&os.ModeCharDevice != 0), nil
	case "isIrregular":
		return toy.Bool(os.FileMode(m)&os.ModeIrregular != 0), nil
	}
	return nil, toy.ErrNoSuchField
}

func (m FileMode) BinaryOp(op token.Token, other toy.Object, right bool) (toy.Object, error) {
	switch y := other.(type) {
	case FileMode:
		switch op {
		case token.And:
			return m & y, nil
		case token.Or:
			return m | y, nil
		case token.Xor:
			return m ^ y, nil
		case token.AndNot:
			return m &^ y, nil
		}
	case toy.Int:
		switch op {
		case token.And:
			return m & FileMode(y), nil
		case token.Or:
			return m | FileMode(y), nil
		case token.Xor:
			return m ^ FileMode(y), nil
		case token.AndNot:
			return m &^ FileMode(y), nil
		case token.Shl:
			if !right {
				return m << y, nil
			}
		case token.Shr:
			if !right {
				return m >> y, nil
			}
		}
	}
	return nil, toy.ErrInvalidOperation
}

func (m FileMode) UnaryOp(op token.Token) (toy.Object, error) {
	switch op {
	case token.Xor:
		return ^m, nil
	}
	return nil, toy.ErrInvalidOperation
}

func readFileFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var name string
	if err := toy.UnpackArgs(args, "name", &name); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{toy.Bytes(data), toy.Nil}, nil
}

func writeFileFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		name string
		data toy.StringOrBytes
		perm FileMode = 0644
	)
	if err := toy.UnpackArgs(args, "name", &name, "data", &data, "perm?", &perm); err != nil {
		return nil, err
	}
	if err := os.WriteFile(name, data.Bytes(), os.FileMode(perm)); err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Nil, nil
}

func readDirFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var name string
	if err := toy.UnpackArgs(args, "name", &name); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(name)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error()), toy.Nil}, nil
	}
	elems := make([]toy.Object, 0, len(entries))
	for _, entry := range entries {
		elems = append(elems, &DirEntry{entry: entry})
	}
	return toy.Tuple{toy.NewArray(elems), toy.Nil}, nil
}

func mkdirFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		name string
		perm FileMode = 0755
		all           = false
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
	return toy.Nil, nil
}

func mkdirTempFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var dir, pattern string
	if err := toy.UnpackArgs(args, "dir", &dir, "pattern", &pattern); err != nil {
		return nil, err
	}
	res, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{toy.String(res), toy.Nil}, nil
}

func removeFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
	return toy.Nil, nil
}

func argsFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	elems := make([]toy.Object, 0, len(os.Args))
	for _, arg := range os.Args {
		elems = append(elems, toy.String(arg))
	}
	return toy.NewArray(elems), nil
}

func environFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func chmodFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		name string
		mode FileMode
	)
	if err := toy.UnpackArgs(args, "name", &name, "mode", &mode); err != nil {
		return nil, err
	}
	if err := os.Chmod(name, os.FileMode(mode)); err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Nil, nil
}

func chownFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		name     string
		uid, gid int
	)
	if err := toy.UnpackArgs(args, "name", &name, "uid", &uid, "gid", &gid); err != nil {
		return nil, err
	}
	if err := os.Chown(name, uid, gid); err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Nil, nil
}

func lchownFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		name     string
		uid, gid int
	)
	if err := toy.UnpackArgs(args, "name", &name, "uid", &uid, "gid", &gid); err != nil {
		return nil, err
	}
	if err := os.Lchown(name, uid, gid); err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Nil, nil
}

type FileInfo struct {
	info os.FileInfo
}

var FileInfoType = toy.NewType[*FileInfo]("os.FileInfo", nil)

func (f *FileInfo) Type() toy.ObjectType { return FileInfoType }
func (f *FileInfo) String() string       { return fmt.Sprintf("os.FileInfo(%q)", f.info.Name()) }
func (f *FileInfo) IsFalsy() bool        { return false }
func (f *FileInfo) Clone() toy.Object    { return &FileInfo{info: f.info} }

func (f *FileInfo) FieldGet(name string) (toy.Object, error) {
	switch name {
	case "name":
		return toy.String(f.info.Name()), nil
	case "size":
		return toy.Int(f.info.Size()), nil
	case "mode":
		return FileMode(f.info.Mode()), nil
	case "modTime":
		return time.Time(f.info.ModTime()), nil
	case "type":
		return FileMode(f.info.Mode().Type()), nil
	case "perm":
		return FileMode(f.info.Mode().Perm()), nil
	case "isDir":
		return toy.Bool(f.info.Mode()&os.ModeDir != 0), nil
	case "isRegular":
		return toy.Bool(f.info.Mode()&os.ModeType == 0), nil
	case "isSymlink":
		return toy.Bool(f.info.Mode()&os.ModeSymlink != 0), nil
	case "isNamedPipe":
		return toy.Bool(f.info.Mode()&os.ModeNamedPipe != 0), nil
	case "isSocket":
		return toy.Bool(f.info.Mode()&os.ModeSocket != 0), nil
	case "isDevice":
		return toy.Bool(f.info.Mode()&os.ModeDevice != 0), nil
	case "isCharDevice":
		return toy.Bool(f.info.Mode()&os.ModeCharDevice != 0), nil
	case "isIrregular":
		return toy.Bool(f.info.Mode()&os.ModeIrregular != 0), nil
	}
	return nil, toy.ErrNoSuchField
}

func statFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var name string
	if err := toy.UnpackArgs(args, "name", &name); err != nil {
		return nil, err
	}
	info, err := os.Stat(name)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{&FileInfo{info: info}, toy.Nil}, nil
}

func lstatFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var name string
	if err := toy.UnpackArgs(args, "name", &name); err != nil {
		return nil, err
	}
	info, err := os.Lstat(name)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{&FileInfo{info: info}, toy.Nil}, nil
}

func truncateFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		name string
		size int64
	)
	if err := toy.UnpackArgs(args, "name", &name, "size", &size); err != nil {
		return nil, err
	}
	if err := os.Truncate(name, size); err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Nil, nil
}

type DirEntry struct {
	entry os.DirEntry
}

var DirEntryType = toy.NewType[*DirEntry]("os.DirEntry", nil)

func (e *DirEntry) Type() toy.ObjectType { return DirEntryType }
func (e *DirEntry) String() string       { return fmt.Sprintf("os.DirEntry(%q)", e.entry.Name()) }
func (e *DirEntry) IsFalsy() bool        { return false }
func (e *DirEntry) Clone() toy.Object    { return &DirEntry{entry: e.entry} }

func (e *DirEntry) FieldGet(name string) (toy.Object, error) {
	switch name {
	case "name":
		return toy.String(e.entry.Name()), nil
	case "type":
		return FileMode(e.entry.Type()), nil
	case "isDir":
		return toy.Bool(e.entry.Type()&os.ModeDir != 0), nil
	case "isRegular":
		return toy.Bool(e.entry.Type()&os.ModeType == 0), nil
	case "isSymlink":
		return toy.Bool(e.entry.Type()&os.ModeSymlink != 0), nil
	case "isNamedPipe":
		return toy.Bool(e.entry.Type()&os.ModeNamedPipe != 0), nil
	case "isSocket":
		return toy.Bool(e.entry.Type()&os.ModeSocket != 0), nil
	case "isDevice":
		return toy.Bool(e.entry.Type()&os.ModeDevice != 0), nil
	case "isCharDevice":
		return toy.Bool(e.entry.Type()&os.ModeCharDevice != 0), nil
	case "isIrregular":
		return toy.Bool(e.entry.Type()&os.ModeIrregular != 0), nil
	}
	method, ok := dirEntryMethods[name]
	if ok {
		return method.WithReceiver(e), nil
	}
	return nil, toy.ErrNoSuchField
}

var dirEntryMethods = map[string]*toy.BuiltinFunction{
	"info": toy.NewBuiltinFunction("info", dirEntryInfoMd),
}

func dirEntryInfoMd(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(*DirEntry)
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args[1:])}
	}
	info, err := recv.entry.Info()
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{&FileInfo{info: info}, toy.Nil}, nil
}

type File os.File

var FileType = toy.NewType[*File]("os.File", nil)

func (f *File) Type() toy.ObjectType { return FileType }
func (f *File) String() string       { return fmt.Sprintf("os.File(%q)", (*os.File)(f).Name()) }
func (f *File) IsFalsy() bool        { return false }

func (f *File) Clone() toy.Object {
	c := new(File)
	*c = *f
	return c
}

func (f *File) FieldGet(name string) (toy.Object, error) {
	switch name {
	case "name":
		return toy.String((*os.File)(f).Name()), nil
	}
	method, ok := fileMethods[name]
	if ok {
		return method.WithReceiver(f), nil
	}
	return nil, toy.ErrNoSuchField
}

var fileMethods = map[string]*toy.BuiltinFunction{
	"write":    toy.NewBuiltinFunction("write", fileWriteMd),
	"read":     toy.NewBuiltinFunction("read", fileReadMd),
	"close":    toy.NewBuiltinFunction("close", fileCloseMd),
	"stat":     toy.NewBuiltinFunction("stat", fileStatMd),
	"sync":     toy.NewBuiltinFunction("sync", fileSyncMd),
	"truncate": toy.NewBuiltinFunction("truncate", fileTruncateMd),
	"chown":    toy.NewBuiltinFunction("chown", fileChownMd),
	"chmod":    toy.NewBuiltinFunction("chmod", fileChmodMd),
	"chdir":    toy.NewBuiltinFunction("chdir", fileChdirMd),
	"seek":     toy.NewBuiltinFunction("seek", fileSeekMd),
	"readdir":  toy.NewBuiltinFunction("readdir", fileReaddirMd),
}

func fileWriteMd(_ *toy.VM, args ...toy.Object) (_ toy.Object, err error) {
	var (
		recv = args[0].(*File)
		data toy.StringOrBytes
		off  *int64
	)
	if err := toy.UnpackArgs(args[1:], "data", &data, "off?", &off); err != nil {
		return nil, err
	}
	var n int
	if off != nil {
		n, err = (*os.File)(recv).WriteAt(data.Bytes(), *off)
		if err != nil {
			return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
		}
	} else {
		n, err = (*os.File)(recv).Write(data.Bytes())
		if err != nil {
			return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
		}
	}
	return toy.Tuple{toy.Int(n), toy.Nil}, nil
}

func fileReadMd(_ *toy.VM, args ...toy.Object) (_ toy.Object, err error) {
	var (
		recv = args[0].(*File)
		buf  toy.Bytes
		off  *int64
	)
	if err := toy.UnpackArgs(args[1:], "buf", &buf, "off?", &off); err != nil {
		return nil, err
	}
	var n int
	if off != nil {
		n, err = (*os.File)(recv).ReadAt(buf, *off)
		if err != nil {
			return toy.Tuple{toy.Nil, toy.Nil, toy.NewError(err.Error())}, nil
		}
	} else {
		n, err = (*os.File)(recv).Read(buf)
		if err != nil {
			return toy.Tuple{toy.Nil, toy.Nil, toy.NewError(err.Error())}, nil
		}
	}
	return toy.Tuple{buf[:n], toy.Int(n), toy.Nil}, nil
}

func fileCloseMd(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(*File)
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args[1:])}
	}
	if err := (*os.File)(recv).Close(); err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Nil, nil
}

func fileStatMd(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(*File)
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args[1:])}
	}
	info, err := (*os.File)(recv).Stat()
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{&FileInfo{info: info}, toy.Nil}, nil
}

func fileSyncMd(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(*File)
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args[1:])}
	}
	if err := (*os.File)(recv).Sync(); err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Nil, nil
}

func fileTruncateMd(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv = args[0].(*File)
		size int64
	)
	if err := toy.UnpackArgs(args[1:], "size", &size); err != nil {
		return nil, err
	}
	if err := (*os.File)(recv).Truncate(size); err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Nil, nil
}

func fileChownMd(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv     = args[0].(*File)
		uid, gid int
	)
	if err := toy.UnpackArgs(args[1:], "uid", &uid, "gid", &gid); err != nil {
		return nil, err
	}
	if err := (*os.File)(recv).Chown(uid, gid); err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Nil, nil
}

func fileChmodMd(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv = args[0].(*File)
		mode FileMode
	)
	if err := toy.UnpackArgs(args[1:], "mode", &mode); err != nil {
		return nil, err
	}
	if err := (*os.File)(recv).Chmod(os.FileMode(mode)); err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Nil, nil
}

func fileChdirMd(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(*File)
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args[1:])}
	}
	if err := (*os.File)(recv).Chdir(); err != nil {
		return toy.NewError(err.Error()), nil
	}
	return toy.Nil, nil
}

func fileSeekMd(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv   = args[0].(*File)
		offset int64
		whence int
	)
	if err := toy.UnpackArgs(args[1:], "offset", &offset, "whence", &whence); err != nil {
		return nil, err
	}
	ret, err := (*os.File)(recv).Seek(offset, whence)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{toy.Int(ret), toy.Nil}, nil
}

func fileReaddirMd(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv = args[0].(*File)
		n    = -1
	)
	if err := toy.UnpackArgs(args[1:], "n?", &n); err != nil {
		return nil, err
	}
	entries, err := (*os.File)(recv).ReadDir(n)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	elems := make([]toy.Object, 0, len(entries))
	for _, entry := range entries {
		elems = append(elems, &DirEntry{entry: entry})
	}
	return toy.Tuple{toy.NewArray(elems), toy.Nil}, nil
}

func openFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		name string
		flag = os.O_RDONLY
		perm FileMode
	)
	if err := toy.UnpackArgs(args, "name", &name, "flag?", &flag, "perm?", &perm); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(name, flag, os.FileMode(perm))
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{(*File)(file), toy.Nil}, nil
}

func createFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var name string
	if err := toy.UnpackArgs(args, "name", &name); err != nil {
		return nil, err
	}
	file, err := os.Create(name)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{(*File)(file), toy.Nil}, nil
}

func createTempFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var dir, pattern string
	if err := toy.UnpackArgs(args, "dir", &dir, "pattern", &pattern); err != nil {
		return nil, err
	}
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{(*File)(file), toy.Nil}, nil
}
