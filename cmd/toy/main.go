package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/urfave/cli/v2"

	"github.com/infastin/toy"
	"github.com/infastin/toy/parser"
)

const (
	sourceFileExt = ".toy"
	replPrompt    = ">>> "
)

func main() {
	app := &cli.App{
		Name:      "toy",
		Usage:     "Toy language interpreter",
		Version:   "dev",
		ArgsUsage: "[FILE]",
		Action:    mainAction,
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func mainAction(ctx *cli.Context) error {
	var inputFile string
	if args := ctx.Args(); args.Len() > 0 {
		inputFile = args.First()
	}
	if inputFile == "" {
		return RunREPL(os.Stderr, os.Stdout)
	}
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}
	if len(inputData) > 1 && string(inputData[:2]) == "#!" {
		copy(inputData, "//")
	}
	if err := CompileAndRun(inputData, inputFile); err != nil {
		return fmt.Errorf("failed to compile and run: %w", err)
	}
	return nil
}

// CompileAndRun compiles the source code and executes it.
func CompileAndRun(inputData []byte, inputFile string) error {
	script := toy.NewScript(inputData)
	script.EnableFileImport(true)
	if err := script.SetImportDir(filepath.Dir(inputFile)); err != nil {
		return err
	}
	if _, err := script.Run(); err != nil {
		return err
	}
	return nil
}

// RunREPL starts REPL.
func RunREPL(in io.ReadCloser, out io.Writer) error {
	rl, err := readline.NewEx(&readline.Config{
		Prompt: replPrompt,
		Stdin:  in,
		Stdout: out,
	})
	if err != nil {
		return err
	}
	defer rl.Clean()

	replPrint := func(args ...toy.Object) (ret toy.Object, err error) {
		var printArgs []string
		for _, arg := range args {
			if arg == toy.Undefined {
				printArgs = append(printArgs, "<undefined>")
			} else {
				printArgs = append(printArgs, arg.String())
			}
		}
		fmt.Fprintln(rl, strings.Join(printArgs, " "))
		return toy.Undefined, nil
	}

	toy.BuiltinFuncs = append(toy.BuiltinFuncs,
		&toy.BuiltinFunction{Name: "print", Func: replPrint})

	fileSet := parser.NewFileSet()
	globals := make([]toy.Object, toy.GlobalsSize)
	symbolTable := toy.NewSymbolTable()
	for idx, fn := range toy.BuiltinFuncs {
		symbolTable.DefineBuiltin(idx, fn.Name)
	}

	var constants []toy.Object
	for {
		line, err := rl.Readline()
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		srcFile := fileSet.AddFile("repl", -1, len(line))
		p := parser.NewParser(srcFile, []byte(line), nil)

		file, err := p.ParseFile()
		if err != nil {
			fmt.Fprint(rl, err.Error())
			continue
		}
		file = addPrints(file)

		c := toy.NewCompiler(srcFile, symbolTable, constants, nil, nil)
		if err := c.Compile(file); err != nil {
			fmt.Fprint(rl, err.Error())
			continue
		}

		bytecode := c.Bytecode()
		bytecode.RemoveDuplicates()

		machine := toy.NewVM(bytecode, globals)
		if err := machine.Run(); err != nil {
			fmt.Fprint(rl, err.Error())
			continue
		}
		constants = bytecode.Constants
	}
}

func addPrints(file *parser.File) *parser.File {
	var stmts []parser.Stmt
	for _, s := range file.Stmts {
		switch s := s.(type) {
		case *parser.ExprStmt:
			stmts = append(stmts, &parser.ExprStmt{
				Expr: &parser.CallExpr{
					Func: &parser.Ident{Name: "print"},
					Args: []parser.Expr{s.Expr},
				},
			})
		case *parser.AssignStmt:
			stmts = append(stmts, s, &parser.ExprStmt{
				Expr: &parser.CallExpr{
					Func: &parser.Ident{
						Name: "print",
					},
					Args: s.LHS,
				},
			})
		default:
			stmts = append(stmts, s)
		}
	}
	return &parser.File{
		InputFile: file.InputFile,
		Stmts:     stmts,
	}
}
