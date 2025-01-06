package stdlib

import (
	"os"
	"strings"

	"github.com/infastin/toy"
)

var FmtModule = &toy.BuiltinModule{
	Name: "fmt",
	Members: map[string]toy.Object{
		"print":   &toy.BuiltinFunction{Name: "print", Func: fmtPrint},
		"println": &toy.BuiltinFunction{Name: "println", Func: fmtPrintln},
		"printf":  &toy.BuiltinFunction{Name: "printf", Func: fmtPrintf},
		"printfn": &toy.BuiltinFunction{Name: "printfn", Func: fmtPrintfn},
	},
}

func fmtPrint(args ...toy.Object) (toy.Object, error) {
	var b strings.Builder
	for _, arg := range args {
		b.WriteString(arg.String())
	}
	if b.Len() != 0 {
		os.Stdout.WriteString(b.String())
	}
	return toy.Undefined, nil
}

func fmtPrintln(args ...toy.Object) (toy.Object, error) {
	var b strings.Builder
	for _, arg := range args {
		b.WriteString(arg.String())
	}
	b.WriteByte('\n')
	os.Stdout.WriteString(b.String())
	return toy.Undefined, nil
}

func fmtPrintf(args ...toy.Object) (toy.Object, error) {
	var (
		format string
		rest   []toy.Object
	)
	if err := toy.UnpackArgs(args, "format", &format, "...", &rest); err != nil {
		return nil, err
	}
	s, err := toy.Format(format, rest...)
	if err != nil {
		return nil, err
	}
	if len(s) != 0 {
		os.Stdout.WriteString(s)
	}
	return toy.Undefined, nil
}

func fmtPrintfn(args ...toy.Object) (toy.Object, error) {
	var (
		format string
		rest   []toy.Object
	)
	if err := toy.UnpackArgs(args, "format", &format, "...", &rest); err != nil {
		return nil, err
	}
	s, err := toy.Format(format, rest...)
	if err != nil {
		return nil, err
	}
	os.Stdout.WriteString(s + "\n")
	return toy.Undefined, nil
}
