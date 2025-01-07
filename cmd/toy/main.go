package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/infastin/toy"
	"github.com/infastin/toy/parser"
	"github.com/infastin/toy/stdlib"
)

var (
	version         = "dev"
	compilationDate = "2006-01-02 15:04:05"
)

func main() {
	app := &cli.App{
		Name:      "toy",
		Usage:     "Toy language interpreter",
		Version:   fmt.Sprintf("%s (%s)", version, compilationDate),
		ArgsUsage: "[FILE]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "trace",
				Usage:   "compile and show trace",
				Aliases: []string{"t"},
			},
		},
		Action: mainAction,
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
	if ctx.Bool("trace") {
		if err := PrintTrace(inputData, inputFile); err != nil {
			return fmt.Errorf("failed to compile: %w", err)
		}
	} else {
		if err := CompileAndRun(inputData, inputFile); err != nil {
			return fmt.Errorf("failed to compile and run: %w", err)
		}
	}
	return nil
}

type compileTracer struct {
	Out []string
}

func (o *compileTracer) Write(p []byte) (n int, err error) {
	o.Out = append(o.Out, string(p))
	return len(p), nil
}

// PrintTrace compiles the source code and prints compiler trace.
func PrintTrace(inputData []byte, inputFile string) error {
	fileSet := parser.NewFileSet()
	file := fileSet.AddFile(inputFile, -1, len(inputData))

	symTable := toy.NewSymbolTable()
	for idx, fn := range toy.BuiltinFuncs {
		symTable.DefineBuiltin(idx, fn.Name)
	}

	p := parser.NewParser(file, []byte(inputData), nil)
	parsed, err := p.ParseFile()
	if err != nil {
		return err
	}

	tr := &compileTracer{}

	c := toy.NewCompiler(file, symTable, nil, nil, tr)
	if err := c.Compile(parsed); err != nil {
		return err
	}

	bytecode := c.Bytecode()
	bytecode.RemoveDuplicates()
	bytecode.RemoveUnused()

	var b strings.Builder

	b.WriteString(`
######################
##  Compiler Trace  ##
######################

`)

	for _, line := range tr.Out {
		b.WriteString(line)
	}

	b.WriteString(`
##########################
##  Compiler Constants  ##
##########################

`)

	for i, line := range bytecode.FormatConstants() {
		if i != 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line)
	}

	b.WriteString(`

#############################
##  Compiler Instructions  ##
#############################

`)

	for i, line := range bytecode.FormatInstructions() {
		if i != 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line)
	}

	fmt.Println(b.String())

	return nil
}

// CompileAndRun compiles the source code and executes it.
func CompileAndRun(inputData []byte, inputFile string) error {
	script := toy.NewScript(inputData)

	mods := toy.NewModuleMap()
	mods.Add("fmt", stdlib.FmtModule)
	mods.Add("text", stdlib.TextModule)
	mods.Add("regexp", stdlib.RegexpModule)
	script.SetImports(mods)

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
	model := newModel()
	if err := model.Run(in, out); err != nil {
		return err
	}
	return nil
}
