package fmt

import (
	"os"
	"strings"

	"github.com/infastin/toy"
)

var Module = &toy.BuiltinModule{
	Name: "fmt",
	Members: map[string]toy.Value{
		"print":   toy.NewBuiltinFunction("fmt.print", printFn),
		"println": toy.NewBuiltinFunction("fmt.println", printlnFn),
		"printf":  toy.NewBuiltinFunction("fmt.printf", printfFn),
		"printfn": toy.NewBuiltinFunction("fmt.printfn", printfnFn),
	},
}

func printFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var b strings.Builder
	for i, arg := range args {
		if i != 0 {
			b.WriteByte(' ')
		}
		b.WriteString(toy.AsString(arg))
	}
	if b.Len() != 0 {
		os.Stdout.WriteString(b.String())
	}
	return toy.Nil, nil
}

func printlnFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var b strings.Builder
	for i, arg := range args {
		if i != 0 {
			b.WriteByte(' ')
		}
		b.WriteString(toy.AsString(arg))
	}
	b.WriteByte('\n')
	os.Stdout.WriteString(b.String())
	return toy.Nil, nil
}

func printfFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		format string
		rest   []toy.Value
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

func printfnFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		format string
		rest   []toy.Value
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
