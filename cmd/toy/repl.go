package main

import (
	"bytes"
	"io"
	"slices"
	"strings"
	"unicode"

	"github.com/infastin/toy"
	"github.com/infastin/toy/parser"
	"github.com/infastin/toy/stdlib"
	"github.com/infastin/toy/token"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type compiler struct {
	symbolTable *toy.SymbolTable
	globals     []toy.Object
	modules     toy.ModuleMap
	constants   []toy.Object
	output      []string
}

func newCompiler() *compiler {
	s := new(compiler)

	replPrintFunc := func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		if len(args) == 1 && args[0] == toy.Nil {
			return toy.Nil, nil
		}
		var b strings.Builder
		for i, arg := range args {
			if i != 0 {
				b.WriteByte(' ')
			}
			b.WriteString(arg.String())
		}
		if b.Len() != 0 {
			s.output = append(s.output, b.String())
		}
		return toy.Nil, nil
	}

	printFunc := func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
			s.output = append(s.output, b.String())
		}
		return toy.Nil, nil
	}

	printfFunc := func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var (
			format string
			rest   []toy.Object
		)
		if err := toy.UnpackArgs(args, "format", &format, "...", &rest); err != nil {
			return nil, err
		}
		str, err := toy.Format(format, rest...)
		if err != nil {
			return nil, err
		}
		if len(str) != 0 {
			s.output = append(s.output, str)
		}
		return toy.Nil, nil
	}

	toy.Universe = append(toy.Universe,
		toy.NewVariable("__replPrint__", toy.NewBuiltinFunction("__replPrint__", replPrintFunc)),
		toy.NewVariable("print", toy.NewBuiltinFunction("print", printFunc)),
		toy.NewVariable("printf", toy.NewBuiltinFunction("printf", printfFunc)),
	)

	s.globals = make([]toy.Object, toy.GlobalsSize)

	s.symbolTable = toy.NewSymbolTable()
	for i, v := range toy.Universe {
		s.symbolTable.DefineBuiltin(i, v.Name())
	}

	s.modules = stdlib.StdLib.Copy()
	s.modules.Remove("fmt")

	return s
}

func (s *compiler) compileAndRun(input []byte) (string, error) {
	fileSet := parser.NewFileSet()
	srcFile := fileSet.AddFile("(repl)", -1, len(input))

	p := parser.NewParser(srcFile, input, nil)
	file, err := p.ParseFile()
	if err != nil {
		return "", err
	}

	file = addPrints(file)
	symbolTable := s.symbolTable.Copy()

	c := toy.NewCompiler(srcFile, symbolTable, s.constants, s.modules, nil)
	if err := c.Compile(file); err != nil {
		return "", err
	}

	bytecode := c.Bytecode()
	bytecode.RemoveDuplicates()

	machine := toy.NewVM(bytecode, s.globals)
	if err := machine.Run(); err != nil {
		return "", err
	}

	s.constants = bytecode.Constants
	s.symbolTable = symbolTable

	output := strings.Join(s.output, "\n")
	s.output = s.output[:0]

	return output, nil
}

func addPrints(file *parser.File) *parser.File {
	var stmts []parser.Stmt
	for _, s := range file.Stmts {
		switch s := s.(type) {
		case *parser.ExprStmt:
			stmts = append(stmts, &parser.ExprStmt{
				Expr: &parser.CallExpr{
					Func: &parser.Ident{Name: "__replPrint__"},
					Args: []parser.Expr{s.Expr},
				},
			})
		case *parser.AssignStmt:
			stmts = append(stmts, s, &parser.ExprStmt{
				Expr: &parser.CallExpr{
					Func: &parser.Ident{
						Name: "__replPrint__",
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

type model struct {
	input         [][]rune
	line          int
	col           int
	compiler      *compiler
	quitting      bool
	err           error
	history       [][][]rune
	uncommited    [][][]rune
	uncommitedIdx int
	textStyle     lipgloss.Style
	cursorStyle   lipgloss.Style
}

func newModel() model {
	return model{
		input:         make([][]rune, 1),
		line:          0,
		col:           0,
		quitting:      false,
		compiler:      newCompiler(),
		err:           nil,
		history:       nil,
		uncommited:    make([][][]rune, 1),
		uncommitedIdx: 0,
		textStyle:     lipgloss.NewStyle().Inline(true),
		cursorStyle:   lipgloss.NewStyle().Inline(true).Reverse(true),
	}
}

func (m *model) reset() {
	clear(m.uncommited)
	m.uncommitedIdx = len(m.uncommited) - 1
	m.input = m.input[:1]
	m.input[0] = m.input[0][:0]
	m.line = 0
	m.col = 0
}

func (m *model) upHistory() {
	if m.uncommitedIdx > 0 {
		m.uncommited[m.uncommitedIdx] = m.input
		m.uncommitedIdx--
		if m.uncommited[m.uncommitedIdx] == nil {
			histItem := slices.Clone(m.history[m.uncommitedIdx])
			for i := range histItem {
				histItem[i] = slices.Clone(histItem[i])
			}
			m.uncommited[m.uncommitedIdx] = histItem
		}
		m.input = m.uncommited[m.uncommitedIdx]
		m.line = len(m.input) - 1
		m.col = len(m.input[m.line])
	}
}

func (m *model) downHistory() {
	if m.uncommitedIdx+1 < len(m.uncommited) {
		m.uncommited[m.uncommitedIdx] = m.input
		m.uncommitedIdx++
		m.input = m.uncommited[m.uncommitedIdx]
		m.line = len(m.input) - 1
		m.col = len(m.input[m.line])
	}
}

func (m *model) prevLineOrUpHistory() {
	if m.line > 0 {
		m.line--
		if m.col >= len(m.input[m.line]) {
			m.col = len(m.input[m.line])
		}
	} else if len(m.input) == 1 {
		m.upHistory()
	}
}

func (m *model) nextLineOrDownHistory() {
	if m.line+1 < len(m.input) {
		m.line++
		if m.col >= len(m.input[m.line]) {
			m.col = len(m.input[m.line])
		}
	} else if len(m.input) == 1 && m.uncommitedIdx+1 < len(m.uncommited) {
		m.downHistory()
	}
}

func (m *model) charForward() {
	if m.col > 0 {
		m.col--
	} else if m.line > 0 {
		m.line--
		m.col = len(m.input[m.line])
	}
}

func (m *model) charBackward() {
	if m.col < len(m.input[m.line]) {
		m.col++
	} else if m.line+1 < len(m.input) {
		m.line++
		m.col = 0
	}
}

func (m *model) deleteCharBefore() {
	if m.col > 0 {
		m.input[m.line] = slices.Delete(m.input[m.line], m.col-1, m.col)
		m.col -= 1
	} else if m.line > 0 {
		m.col = len(m.input[m.line-1])
		m.input[m.line-1] = append(m.input[m.line-1], m.input[m.line]...)
		m.input = slices.Delete(m.input, m.line, m.line+1)
		m.line--
	}
}

func (m *model) deleteCharAfter() {
	if m.col < len(m.input[m.line]) {
		m.input[m.line] = slices.Delete(m.input[m.line], m.col, m.col+1)
	} else if m.line+1 < len(m.input) {
		m.input[m.line] = append(m.input[m.line], m.input[m.line+1]...)
		m.input = slices.Delete(m.input, m.line+1, m.line+2)
	}
}

func (m *model) lineStart() {
	m.col = 0
}

func (m *model) lineEnd() {
	m.col = len(m.input[m.line])
}

// Unicode character ranges that
// are considered to be a part of a word.
var wordRange = []*unicode.RangeTable{
	unicode.L,
	unicode.Nd,
	unicode.Pc,
}

// Looks forward, find a first non-word character
// (while ignoring leading spaces) and moves the cursor
// to that character.
func (m *model) wordForward() {
	skipping := true
	i := m.line
	j := m.col
	for ; i < len(m.input); i++ {
		for ; j < len(m.input[i]); j++ {
			r := m.input[i][j]
			if skipping {
				// skip spaces
				if unicode.IsSpace(r) {
					continue
				}
				skipping = false
			} else if !unicode.In(r, wordRange...) {
				// encountered non-word character
				break
			}
		}
		if !skipping {
			// set cursor to be at the non-word character
			m.line = i
			m.col = j
			return
		}
		// we go to the next line,
		// so we have to adjust the column index
		j = 0
	}
	// everything after the cursor is just spaces,
	// so we set the cursor to the end of the input
	m.line = len(m.input) - 1
	m.col = len(m.input[m.line])
}

// Looks backward, find a first non-space character
// (while ignoring leading word characters) and moves the cursor
// to the character after that character.
func (m *model) wordBackward() {
	skipping := true
	i := m.line
	j := m.col - 1
	for ; i >= 0; i-- {
		for ; j >= 0; j-- {
			r := m.input[i][j]
			if skipping {
				// skip word
				if unicode.In(r, wordRange...) {
					continue
				}
				skipping = false
			} else if !unicode.IsSpace(r) {
				// encountered space
				break
			}
		}
		if !skipping {
			// set cursor to be at the non-word character
			m.line = i
			m.col = j + 1
			return
		}
		if i > 0 {
			// set the column index to the last character
			j = len(m.input[i-1]) - 1
		}
		// we go the next line,
		// which means that we skipped a word
		skipping = false
	}
	// everything before the cursor is just spaces,
	// so we set the cursor to the beginning of the input
	m.line = 0
	m.col = 0
}

func (m *model) deleteWordBackward() {
	oldLine := m.line
	oldCol := m.col
	m.wordBackward()
	switch {
	case m.line == oldLine && m.col == oldCol:
		return
	case m.line == oldLine:
		m.input[m.line] = slices.Delete(m.input[m.line], m.col, oldCol)
	default:
		m.input[m.line] = append(m.input[m.line][:m.col], m.input[oldLine][oldCol:]...)
		m.input = slices.Delete(m.input, m.line+1, oldLine+1)
	}
}

func (m *model) deleteWordForward() {
	oldLine := m.line
	oldCol := m.col
	m.wordForward()
	switch {
	case m.line == oldLine && m.col == oldCol:
		return
	case m.line == oldLine:
		m.input[oldLine] = slices.Delete(m.input[oldLine], oldCol, m.col)
		m.col = oldCol
	default:
		m.input[oldLine] = append(m.input[oldLine][:oldCol], m.input[m.line][m.col:]...)
		m.input = slices.Delete(m.input, oldLine+1, m.line+1)
		m.line = oldLine
		m.col = oldCol
	}
}

func (m *model) deleteAfterCursor() {
	if m.col != len(m.input[m.line]) {
		m.input[m.line] = m.input[m.line][:m.col]
	} else if m.line+1 < len(m.input) {
		m.input[m.line] = append(m.input[m.line], m.input[m.line+1]...)
		m.input = slices.Delete(m.input, m.line+1, m.line+2)
	}
}

func (m *model) deleteBeforeCursor() {
	if m.col != 0 {
		m.input[m.line] = slices.Delete(m.input[m.line], 0, m.col)
		m.col = 0
	} else if m.line > 0 {
		m.col = len(m.input[m.line-1])
		m.input[m.line-1] = append(m.input[m.line-1], m.input[m.line]...)
		m.input = slices.Delete(m.input, m.line, m.line+1)
		m.line--
	}
}

func (m *model) newLine() {
	m.handleUserInput([]rune("\n"))
}

func (m *model) onEnter() (tea.Model, tea.Cmd) {
	var buf bytes.Buffer
	for i, line := range m.input {
		if i != 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(string(line))
	}

	input := bytes.TrimSpace(buf.Bytes())
	if len(input) == 0 {
		return m, nil
	}

	nl, err := checkNewLine(input)
	if err != nil {
		m.err = err
		return m, nil
	}
	if nl {
		m.newLine()
		return m, nil
	}

	output, err := m.compiler.compileAndRun(input)
	if err != nil {
		m.err = err
		return m, nil
	}

	cmds := []tea.Cmd{tea.Println(m.view(true))}
	if output != "" {
		cmds = append(cmds, tea.Println(output))
	}

	m.history = append(m.history, m.input)
	clear(m.uncommited)
	m.uncommited = append(m.uncommited, nil)
	m.uncommitedIdx = len(m.uncommited) - 1

	m.input = make([][]rune, 1)
	m.line = 0
	m.col = 0

	return m, tea.Sequence(cmds...)
}

func checkNewLine(input []byte) (bool, error) {
	var errors parser.ErrorList

	fileSet := parser.NewFileSet()
	srcFile := fileSet.AddFile("(repl)", -1, len(input))
	scanner := parser.NewScanner(srcFile, input, errors.Add, 0)

	var parens, braces, brackets int
loop:
	for {
		tok, _, _ := scanner.Scan()
		switch tok {
		case token.LParen:
			parens += 1
		case token.RParen:
			parens -= 1
		case token.LBrace:
			braces += 1
		case token.RBrace:
			braces -= 1
		case token.LBrack:
			brackets += 1
		case token.RBrack:
			brackets -= 1
		case token.EOF:
			break loop
		}
	}

	if err := errors.Err(); err != nil {
		return false, err
	}

	if parens != 0 || braces != 0 || brackets != 0 {
		return true, nil
	}

	return false, nil
}

func (m *model) handleUserInput(runes []rune) {
	var buf, rem []rune
	for _, r := range runes {
		switch {
		case r == '\r' || r == '\n':
			rem = append(rem, m.input[m.line][m.col:]...)
			m.input[m.line] = append(m.input[m.line][:m.col], buf...)
			buf = buf[:0]
			m.col = 0
			m.line++
			if m.line == len(m.input) {
				m.input = append(m.input, nil)
			}
		case r == '\t':
			buf = append(buf, ' ', ' ')
		case unicode.IsPrint(r):
			buf = append(buf, r)
		}
	}
	if len(buf) != 0 || len(rem) != 0 {
		m.input[m.line] = slices.Concat(m.input[m.line][:m.col], buf, rem, m.input[m.line][m.col:])
		m.col += len(buf)
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Printf("Toy Language %s (%s)", version, compilationDate)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		m.err = nil
	}
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+q":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+d":
			if len(m.input) == 1 && len(m.input[m.line]) == 0 {
				// quit if input is empty
				m.quitting = true
				return m, tea.Quit
			}
			m.deleteCharAfter()
		case "ctrl+l":
			return m, tea.ClearScreen
		case "ctrl+c":
			m.reset()
		case "up":
			m.prevLineOrUpHistory()
		case "down":
			m.nextLineOrDownHistory()
		case "ctrl+p":
			m.upHistory()
		case "ctrl+n":
			m.downHistory()
		case "left":
			m.charForward()
		case "right", "ctrl+f":
			m.charBackward()
		case "backspace", "ctrl+h":
			m.deleteCharBefore()
		case "delete":
			m.deleteCharAfter()
		case "home", "ctrl+a":
			m.lineStart()
		case "end", "ctrl+e":
			m.lineEnd()
		case "alt+right", "ctrl+right", "alt+f":
			m.wordForward()
		case "alt+left", "ctrl+left", "alt+b":
			m.wordBackward()
		case "alt+backspace", "ctrl+w":
			m.deleteWordBackward()
		case "alt+delete", "alt+d":
			m.deleteWordForward()
		case "ctrl+k":
			m.deleteAfterCursor()
		case "ctrl+u":
			m.deleteBeforeCursor()
		case "enter":
			return m.onEnter()
		case "tab":
			m.handleUserInput([]rune{' ', ' '})
		default:
			m.handleUserInput(msg.Runes)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *model) view(persist bool) string {
	if persist || m.quitting {
		cursorStyle := m.cursorStyle
		m.cursorStyle = m.textStyle
		defer func() { m.cursorStyle = cursorStyle }()
	}
	var b strings.Builder
	for i, line := range m.input {
		if i == 0 {
			b.WriteString(">>> ")
		} else {
			b.WriteString("\n... ")
		}
		if m.line != i {
			b.WriteString(m.textStyle.Render(string(line)))
			continue
		}
		b.WriteString(m.textStyle.Render(string(line[:m.col])))
		if m.col < len(line) {
			b.WriteString(m.cursorStyle.Render(string(line[m.col])))
			b.WriteString(m.textStyle.Render(string(line[m.col+1:])))
		} else {
			b.WriteString(m.cursorStyle.Render(" "))
		}
	}
	if !persist {
		b.WriteByte('\n')
		if m.err != nil {
			b.WriteString(m.err.Error())
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (m *model) View() string {
	return m.view(false)
}

func (m *model) Run(in io.Reader, out io.Writer) error {
	p := tea.NewProgram(m, tea.WithInput(in), tea.WithOutput(out))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
