package stdlib

import (
	"fmt"
	"regexp"

	"github.com/infastin/toy"
)

var RegexpModule = &toy.BuiltinModule{
	Name: "regexp",
	Members: map[string]toy.Object{
		"compile": &toy.BuiltinFunction{Name: "compile", Func: regexpCompile},
		"match":   &toy.BuiltinFunction{Name: "match", Func: regexpMatch},
		"find":    &toy.BuiltinFunction{Name: "find", Func: regexpFind},
		"replace": &toy.BuiltinFunction{Name: "replace", Func: regexpReplace},
	},
}

type RegexpMatch struct {
	text       string
	begin, end int
}

func (m RegexpMatch) TypeName() string { return "RegexpMatch" }
func (m RegexpMatch) String() string   { return fmt.Sprintf("RegexpMatch(%q)", m.text) }
func (m RegexpMatch) IsFalsy() bool    { return len(m.text) == 0 }
func (m RegexpMatch) Copy() toy.Object { return m }

func (m RegexpMatch) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String(m.text)
	case *toy.Bytes:
		*p = toy.Bytes(m.text)
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func (m RegexpMatch) FieldGet(name string) (toy.Object, error) {
	switch name {
	case "text":
		return toy.String(m.text), nil
	case "begin":
		return toy.Int(m.begin), nil
	case "end":
		return toy.Int(m.end), nil
	}
	return nil, toy.ErrNoSuchField
}

type Regexp regexp.Regexp

func (r *Regexp) Unpack(o toy.Object) error {
	switch x := o.(type) {
	case *Regexp:
		*r = *x
	case toy.String:
		rx, err := regexp.Compile(string(x))
		if err != nil {
			return err
		}
		*r = Regexp(*rx)
	default:
		return &toy.InvalidArgumentTypeError{Want: "Regexp or string"}
	}
	return nil
}

func (r *Regexp) TypeName() string { return "Regexp" }
func (r *Regexp) String() string   { return fmt.Sprintf("/%s/", (*regexp.Regexp)(r).String()) }
func (r *Regexp) IsFalsy() bool    { return false }
func (r *Regexp) Copy() toy.Object { return r }

func (r *Regexp) FieldGet(name string) (toy.Object, error) {
	m, ok := regexpRegexpMethods[name]
	if !ok {
		return nil, toy.ErrNoSuchField
	}
	return m.WithReceiver(r), nil
}

var regexpRegexpMethods = map[string]*toy.BuiltinFunction{
	"find":    {Name: "find", Func: regexpRegexpFind},
	"replace": {Name: "replace", Func: regexpRegexpReplace},
}

func regexpRegexpFind(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv  = args[0].(*Regexp)
		input toy.StringOrBytes
		n     *int
	)
	if err := toy.UnpackArgs(args[1:], "input", &input, "n?", &n); err != nil {
		return nil, err
	}
	return regexpFindStringSubmatch((*regexp.Regexp)(recv), input.String(), n)
}

func regexpRegexpReplace(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv  = args[0].(*Regexp)
		input toy.StringOrBytes
		repl  toy.StringOrBytes
	)
	if err := toy.UnpackArgs(args[1:], "input", &input, "repl", &repl); err != nil {
		return nil, err
	}
	result := (*regexp.Regexp)(recv).ReplaceAllString(input.String(), repl.String())
	return toy.String(result), nil
}

func regexpCompile(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var expr string
	if err := toy.UnpackArgs(args, "expr", &expr); err != nil {
		return nil, err
	}
	r, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return (*Regexp)(r), nil
}

func regexpMatch(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		pattern string
		data    toy.StringOrBytes
	)
	if err := toy.UnpackArgs(args, "pattern", &pattern, "data", &data); err != nil {
		return nil, err
	}
	matched, err := regexp.Match(pattern, data.Bytes())
	if err != nil {
		return nil, err
	}
	return toy.Bool(matched), nil
}

func regexpFind(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		expr  string
		input toy.StringOrBytes
		n     *int
	)
	if err := toy.UnpackArgs(args, "expr", &expr, "input", &input, "n?", &n); err != nil {
		return nil, err
	}
	r, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return regexpFindStringSubmatch(r, input.String(), n)
}

func regexpReplace(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		expr  string
		input toy.StringOrBytes
		repl  toy.StringOrBytes
	)
	if err := toy.UnpackArgs(args, "expr", &expr, "input", &input, "repl", &repl); err != nil {
		return nil, err
	}
	r, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	result := r.ReplaceAllString(input.String(), repl.String())
	return toy.String(result), nil
}

func regexpFindStringSubmatch(r *regexp.Regexp, s string, n *int) (toy.Object, error) {
	if n != nil {
		matches := r.FindAllStringSubmatchIndex(s, *n)
		if matches == nil {
			return toy.Undefined, nil
		}
		results := make([]toy.Object, 0, len(matches))
		for _, match := range matches {
			result := make([]toy.Object, 0, len(match))
			for i := 0; i < len(match); i += 2 {
				begin, end := match[i], match[i+1]
				result = append(result, RegexpMatch{
					text:  s[begin:end],
					begin: begin,
					end:   end,
				})
			}
			results = append(results, toy.NewArray(result))
		}
		return toy.NewArray(results), nil
	}
	matches := r.FindStringSubmatchIndex(s)
	if matches == nil {
		return toy.Undefined, nil
	}
	results := make([]toy.Object, 0, len(matches)/2)
	for i := 0; i < len(matches); i += 2 {
		begin, end := matches[i], matches[i+1]
		results = append(results, RegexpMatch{
			text:  s[begin:end],
			begin: begin,
			end:   end,
		})
	}
	return toy.NewArray(results), nil
}
