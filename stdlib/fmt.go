package stdlib

import (
	"os"
	"strings"

	"github.com/infastin/toy"
)

var FmtModule = &toy.BuiltinModule{
	Name: "fmt",
	Members: map[string]toy.Object{
		"print":   &toy.BuiltinFunction{Name: "fmt.print", Func: fmtPrint},
		"println": &toy.BuiltinFunction{Name: "fmt.println", Func: fmtPrintln},
		"printf":  &toy.BuiltinFunction{Name: "fmt.printf", Func: fmtPrintf},
		"printfn": &toy.BuiltinFunction{Name: "fmt.printfn", Func: fmtPrintfn},
	},
}

func fmtPrint(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var b strings.Builder
	for i, arg := range args {
		var s toy.String
		if err := toy.Convert(&s, arg); err != nil {
			return nil, err
		}
		if i != 0 {
			b.WriteByte(' ')
		}
		b.WriteString(string(s))
	}
	if b.Len() != 0 {
		os.Stdout.WriteString(b.String())
	}
	return toy.Nil, nil
}

func fmtPrintln(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var b strings.Builder
	for i, arg := range args {
		var s toy.String
		if err := toy.Convert(&s, arg); err != nil {
			return nil, err
		}
		if i != 0 {
			b.WriteByte(' ')
		}
		b.WriteString(string(s))
	}
	b.WriteByte('\n')
	os.Stdout.WriteString(b.String())
	return toy.Nil, nil
}

func fmtPrintf(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
	return toy.Nil, nil
}

func fmtPrintfn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
	return toy.Nil, nil
}
