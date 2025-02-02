package regexp

import (
	"fmt"
	"regexp"

	"github.com/infastin/toy"
)

var Module = &toy.BuiltinModule{
	Name: "regexp",
	Members: map[string]toy.Value{
		"Regexp": RegexpType,
		"Match":  RegexpMatchType,

		"match":   toy.NewBuiltinFunction("regexp.match", matchFn),
		"find":    toy.NewBuiltinFunction("regexp.find", findFn),
		"replace": toy.NewBuiltinFunction("regexp.replace", replaceFn),
	},
}

type Regexp regexp.Regexp

var RegexpType = toy.NewType[*Regexp]("regexp.Regexp", func(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	switch x := args[0].(type) {
	case toy.String:
		r, err := regexp.Compile(string(x))
		if err != nil {
			return nil, err
		}
		return (*Regexp)(r), nil
	default:
		var rx *Regexp
		if err := toy.Convert(&rx, x); err != nil {
			return rx, nil
		}
		return rx, nil
	}
})

func (r *Regexp) Type() toy.ValueType { return RegexpType }
func (r *Regexp) String() string      { return fmt.Sprintf("/%s/", (*regexp.Regexp)(r).String()) }
func (r *Regexp) IsFalsy() bool       { return false }
func (r *Regexp) Clone() toy.Value    { return r }

func (r *Regexp) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String((*regexp.Regexp)(r).String())
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func (r *Regexp) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	m, ok := regexpMethods[string(keyStr)]
	if !ok {
		return toy.Nil, false, nil
	}
	return m.WithReceiver(r), true, nil
}

var regexpMethods = map[string]*toy.BuiltinFunction{
	"find":    toy.NewBuiltinFunction("find", regexpFindMd),
	"replace": toy.NewBuiltinFunction("replace", regexpReplaceMd),
}

func regexpFindMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
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

func regexpReplaceMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
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

type RegexpMatch struct {
	text       string
	begin, end int
}

var RegexpMatchType = toy.NewType[RegexpMatch]("regexp.Match", nil)

func (m RegexpMatch) Type() toy.ValueType { return RegexpMatchType }
func (m RegexpMatch) String() string      { return fmt.Sprintf("regexp.Match(%q)", m.text) }
func (m RegexpMatch) IsFalsy() bool       { return len(m.text) == 0 }
func (m RegexpMatch) Clone() toy.Value    { return m }

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

func (m RegexpMatch) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	switch string(keyStr) {
	case "text":
		return toy.String(m.text), true, nil
	case "begin":
		return toy.Int(m.begin), true, nil
	case "end":
		return toy.Int(m.end), true, nil
	}
	return toy.Nil, false, nil
}

func matchFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
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

func findFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
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

func replaceFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
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

func regexpFindStringSubmatch(r *regexp.Regexp, s string, n *int) (toy.Value, error) {
	if n != nil {
		matches := r.FindAllStringSubmatchIndex(s, *n)
		if matches == nil {
			return toy.Nil, nil
		}
		results := make([]toy.Value, 0, len(matches))
		for _, match := range matches {
			result := make([]toy.Value, 0, len(match))
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
		return toy.Nil, nil
	}
	results := make([]toy.Value, 0, len(matches)/2)
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
