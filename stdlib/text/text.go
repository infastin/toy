package text

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/infastin/toy"
	"github.com/infastin/toy/internal/fndef"
	"github.com/infastin/toy/internal/xiter"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var Module = &toy.BuiltinModule{
	Name: "text",
	Members: map[string]toy.Value{
		"Builder": BuilderType,

		"contains":     toy.NewBuiltinFunction("text.contains", containsFn),
		"containsAny":  toy.NewBuiltinFunction("text.containsAny", fndef.ASSRB("s", "chars", strings.ContainsAny)),
		"hasPrefix":    toy.NewBuiltinFunction("text.hasPrefix", fndef.ASSRB("s", "prefix", strings.HasPrefix)),
		"hasSuffix":    toy.NewBuiltinFunction("text.hasSuffix", fndef.ASSRB("s", "suffix", strings.HasSuffix)),
		"trimLeft":     toy.NewBuiltinFunction("text.trimLeft", fndef.ASSRS("s", "cutset", strings.TrimLeft)),
		"trimRight":    toy.NewBuiltinFunction("text.trimRight", fndef.ASSRS("s", "cutset", strings.TrimRight)),
		"trimPrefix":   toy.NewBuiltinFunction("text.trimPrefix", fndef.ASSRS("s", "prefix", strings.TrimPrefix)),
		"trimSuffix":   toy.NewBuiltinFunction("text.trimSuffix", fndef.ASSRS("s", "suffix", strings.TrimSuffix)),
		"trimSpace":    toy.NewBuiltinFunction("text.trimSpace", fndef.ASRS("s", strings.TrimSpace)),
		"trim":         toy.NewBuiltinFunction("text.trim", fndef.ASSRS("s", "cutset", strings.Trim)),
		"toLower":      toy.NewBuiltinFunction("text.toLower", fndef.ASRS("s", strings.ToLower)),
		"toUpper":      toy.NewBuiltinFunction("text.toUpper", fndef.ASRS("s", strings.ToUpper)),
		"toTitle":      toy.NewBuiltinFunction("text.toTitle", toTitleFn),
		"join":         toy.NewBuiltinFunction("text.join", joinFn),
		"split":        toy.NewBuiltinFunction("text.split", splitFn),
		"splitAfter":   toy.NewBuiltinFunction("text.splitAfter", splitAfterFn),
		"fields":       toy.NewBuiltinFunction("text.fields", fndef.ASRSs("s", strings.Fields)),
		"replace":      toy.NewBuiltinFunction("text.replace", replaceFn),
		"cut":          toy.NewBuiltinFunction("text.cut", cutFn),
		"cutPrefix":    toy.NewBuiltinFunction("text.cutPrefix", fndef.ASSRSB("s", "prefix", strings.CutPrefix)),
		"cutSuffix":    toy.NewBuiltinFunction("text.cutSuffix", fndef.ASSRSB("s", "suffix", strings.CutSuffix)),
		"index":        toy.NewBuiltinFunction("text.index", indexFn),
		"indexAny":     toy.NewBuiltinFunction("text.indexAny", indexAnyFn),
		"lastIndex":    toy.NewBuiltinFunction("text.lastIndex", lastIndexFn),
		"lastIndexAny": toy.NewBuiltinFunction("text.lastIndexAny", lastIndexAnyFn),

		"quote":          toy.NewBuiltinFunction("text.quote", fndef.ASRS("s", strconv.Quote)),
		"quoteToASCII":   toy.NewBuiltinFunction("text.quoteToASCII", fndef.ASRS("s", strconv.QuoteToASCII)),
		"quoteToGraphic": toy.NewBuiltinFunction("text.quoteToGraphic", fndef.ASRS("s", strconv.QuoteToGraphic)),
		"unquote":        toy.NewBuiltinFunction("text.unquote", fndef.ASRSE("s", strconv.Unquote)),

		"parseInt":   toy.NewBuiltinFunction("text.parseInt", parseInt),
		"parseFloat": toy.NewBuiltinFunction("text.parseFloat", parseFloat),
		"parseBool":  toy.NewBuiltinFunction("text.parseBool", parseBool),
	},
}

func containsFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		s      string
		subset toy.StringOrChar
	)
	if err := toy.UnpackArgs(args, "s", &s, "subset", &subset); err != nil {
		return nil, err
	}
	return toy.Bool(strings.Contains(string(s), string(subset))), nil
}

func toTitleFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var s string
	if err := toy.UnpackArgs(args, "s", &s); err != nil {
		return nil, err
	}
	caser := cases.Title(language.Und, cases.NoLower)
	return toy.String(caser.String(s)), nil
}

func joinFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		elems toy.Sequence
		sep   string
	)
	if err := toy.UnpackArgs(args, "elems", &elems, "sep", &sep); err != nil {
		return nil, err
	}
	var strs []string
	for i, elem := range xiter.Enum(elems.Elements()) {
		s, ok := elem.(toy.String)
		if !ok {
			return nil, &toy.InvalidArgumentTypeError{
				Name: toy.TypeName(elems),
				Sel:  fmt.Sprintf("[%d]", i),
				Want: "string",
				Got:  toy.TypeName(elem),
			}
		}
		strs = append(strs, string(s))
	}
	return toy.String(strings.Join(strs, sep)), nil
}

func splitFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		s, sep string
		n      *int
	)
	if err := toy.UnpackArgs(args, "s", &s, "sep", &sep, "n?", &n); err != nil {
		return nil, err
	}
	var (
		elems []toy.Value
		parts []string
	)
	if n != nil {
		parts = strings.SplitN(s, sep, *n)
		elems = make([]toy.Value, 0, len(parts))
	} else {
		parts = strings.Split(s, sep)
		elems = make([]toy.Value, 0, len(parts))
	}
	for _, part := range parts {
		elems = append(elems, toy.String(part))
	}
	return toy.NewArray(elems), nil
}

func splitAfterFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		s, sep string
		n      *int
	)
	if err := toy.UnpackArgs(args, "s", &s, "sep", &sep, "n?", &n); err != nil {
		return nil, err
	}
	var (
		elems []toy.Value
		parts []string
	)
	if n != nil {
		parts = strings.SplitAfterN(s, sep, *n)
		elems = make([]toy.Value, 0, len(parts))
	} else {
		parts = strings.SplitAfterN(s, sep, *n)
		elems = make([]toy.Value, 0, len(parts))
	}
	for _, part := range parts {
		elems = append(elems, toy.String(part))
	}
	return toy.NewArray(elems), nil
}

func replaceFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		s, old, new string
		n           *int
	)
	if err := toy.UnpackArgs(args, "s", &s, "old", &old, "new", &new, "n?", &n); err != nil {
		return nil, err
	}
	if n != nil {
		return toy.String(strings.Replace(s, old, new, *n)), nil
	} else {
		return toy.String(strings.ReplaceAll(s, old, new)), nil
	}
}

func cutFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var s, sep string
	if err := toy.UnpackArgs(args, "s", &s, "sep", &sep); err != nil {
		return nil, err
	}
	before, after, found := strings.Cut(s, sep)
	return toy.Tuple{toy.String(before), toy.String(after), toy.Bool(found)}, nil
}

func indexFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		s      string
		subset toy.StringOrChar
	)
	if err := toy.UnpackArgs(args, "s", &s, "subset", &subset); err != nil {
		return nil, err
	}
	idx := strings.Index(string(s), string(subset))
	if idx <= 0 {
		return toy.Int(idx), nil
	}
	return toy.Int(utf8.RuneCountInString(string(s)) - utf8.RuneCountInString(string(s)[idx:])), nil
}

func indexAnyFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var s, chars string
	if err := toy.UnpackArgs(args, "s", &s, "chars", &chars); err != nil {
		return nil, err
	}
	idx := strings.IndexAny(s, chars)
	if idx <= 0 {
		return toy.Int(idx), nil
	}
	return toy.Int(utf8.RuneCountInString(string(s)) - utf8.RuneCountInString(string(s)[idx:])), nil
}

func lastIndexFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		s      string
		subset toy.StringOrChar
	)
	if err := toy.UnpackArgs(args, "s", &s, "subset", &subset); err != nil {
		return nil, err
	}
	idx := strings.LastIndex(string(s), string(subset))
	if idx <= 0 {
		return toy.Int(idx), nil
	}
	return toy.Int(utf8.RuneCountInString(string(s)) - utf8.RuneCountInString(string(s)[idx:])), nil
}

func lastIndexAnyFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var s, chars string
	if err := toy.UnpackArgs(args, "s", &s, "chars", &chars); err != nil {
		return nil, err
	}
	idx := strings.LastIndexAny(s, chars)
	if idx <= 0 {
		return toy.Int(idx), nil
	}
	return toy.Int(utf8.RuneCountInString(string(s)) - utf8.RuneCountInString(string(s)[idx:])), nil
}

func parseInt(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		s    string
		base = 10
	)
	if err := toy.UnpackArgs(args, "s", &s, "base?", &base); err != nil {
		return nil, err
	}
	i, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		return nil, errors.Unwrap(err)
	}
	return toy.Int(i), nil
}

func parseFloat(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var s string
	if err := toy.UnpackArgs(args, "s", &s); err != nil {
		return nil, err
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, errors.Unwrap(err)
	}
	return toy.Float(f), nil
}

func parseBool(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var s string
	if err := toy.UnpackArgs(args, "s", &s); err != nil {
		return nil, err
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return nil, errors.Unwrap(err)
	}
	return toy.Bool(b), nil
}

type Builder strings.Builder

var BuilderType = toy.NewType[*Builder]("text.Builder", func(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return new(Builder), nil
})

func (b *Builder) Type() toy.ValueType { return BuilderType }
func (b *Builder) String() string      { return "<text.Builder>" }
func (b *Builder) IsFalsy() bool       { return (*strings.Builder)(b).Len() == 0 }

func (b *Builder) Clone() toy.Value {
	c := new(strings.Builder)
	c.WriteString(b.String())
	return (*Builder)(c)
}

func (b *Builder) Len() int { return (*strings.Builder)(b).Len() }

func (b *Builder) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String((*strings.Builder)(b).String())
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func (b *Builder) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	method, ok := builderMethods[string(keyStr)]
	if !ok {
		return toy.Nil, false, nil
	}
	return method.WithReceiver(b), true, nil
}

var builderMethods = map[string]*toy.BuiltinFunction{
	"write": toy.NewBuiltinFunction("text.write", builderWriteMd),
	"reset": toy.NewBuiltinFunction("text.reset", builderResetMd),
}

func builderWriteMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	recv := args[0].(*Builder)
	args = args[1:]
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	switch x := args[0].(type) {
	case toy.String:
		(*strings.Builder)(recv).WriteString(string(x))
	case toy.Bytes:
		(*strings.Builder)(recv).Write(x)
	case toy.Char:
		(*strings.Builder)(recv).WriteRune(rune(x))
	default:
		return nil, &toy.InvalidArgumentTypeError{
			Name: "x",
			Want: "string, bytes or char",
			Got:  toy.TypeName(x),
		}
	}
	return toy.Nil, nil
}

func builderResetMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	recv := args[0].(*Builder)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	(*strings.Builder)(recv).Reset()
	return toy.Nil, nil
}
