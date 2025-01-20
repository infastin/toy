package text

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/infastin/toy"
	"github.com/infastin/toy/internal/fndef"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var Module = &toy.BuiltinModule{
	Name: "text",
	Members: map[string]toy.Object{
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
	},
}

func containsFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 2 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 2,
			WantMax: 2,
			Got:     len(args),
		}
	}
	s1, ok := args[0].(toy.String)
	if !ok {
		return nil, &toy.InvalidArgumentTypeError{
			Name: "s",
			Want: "string",
			Got:  toy.TypeName(args[0]),
		}
	}
	switch a2 := args[1].(type) {
	case toy.String:
		return toy.Bool(strings.Contains(string(s1), string(a2))), nil
	case toy.Char:
		return toy.Bool(strings.ContainsRune(string(s1), rune(a2))), nil
	default:
		return nil, &toy.InvalidArgumentTypeError{
			Name: "subset",
			Want: "string or char",
			Got:  toy.TypeName(a2),
		}
	}
}

func toTitleFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s string
	if err := toy.UnpackArgs(args, "s", &s); err != nil {
		return nil, err
	}
	caser := cases.Title(language.Und, cases.NoLower)
	return toy.String(caser.String(s)), nil
}

func joinFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		elems toy.Sequence
		sep   string
	)
	if err := toy.UnpackArgs(args, "elems", &elems, "sep", &sep); err != nil {
		return nil, err
	}
	var strs []string
	for i, elem := range toy.Entries(elems) {
		s, ok := elem.(toy.String)
		if !ok {
			return nil, fmt.Errorf("%s[%d]: want string, got %s", toy.TypeName(elems), i, toy.TypeName(elem))
		}
		strs = append(strs, string(s))
	}
	return toy.String(strings.Join(strs, sep)), nil
}

func splitFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		s, sep string
		n      *int
	)
	if err := toy.UnpackArgs(args, "s", &s, "sep", &sep, "n?", &n); err != nil {
		return nil, err
	}
	var (
		elems []toy.Object
		parts []string
	)
	if n != nil {
		parts = strings.SplitN(s, sep, *n)
		elems = make([]toy.Object, 0, len(parts))
	} else {
		parts = strings.Split(s, sep)
		elems = make([]toy.Object, 0, len(parts))
	}
	for _, part := range parts {
		elems = append(elems, toy.String(part))
	}
	return toy.NewArray(elems), nil
}

func splitAfterFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		s, sep string
		n      *int
	)
	if err := toy.UnpackArgs(args, "s", &s, "sep", &sep, "n?", &n); err != nil {
		return nil, err
	}
	var (
		elems []toy.Object
		parts []string
	)
	if n != nil {
		parts = strings.SplitAfterN(s, sep, *n)
		elems = make([]toy.Object, 0, len(parts))
	} else {
		parts = strings.SplitAfterN(s, sep, *n)
		elems = make([]toy.Object, 0, len(parts))
	}
	for _, part := range parts {
		elems = append(elems, toy.String(part))
	}
	return toy.NewArray(elems), nil
}

func replaceFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func cutFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s, sep string
	if err := toy.UnpackArgs(args, "s", &s, "sep", &sep); err != nil {
		return nil, err
	}
	before, after, found := strings.Cut(s, sep)
	return toy.Tuple{toy.String(before), toy.String(after), toy.Bool(found)}, nil
}

func indexFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 2 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 2,
			WantMax: 2,
			Got:     len(args),
		}
	}
	s, ok := args[0].(toy.String)
	if !ok {
		return nil, &toy.InvalidArgumentTypeError{
			Name: "s",
			Want: "string",
			Got:  toy.TypeName(args[0]),
		}
	}
	var idx int
	switch a := args[1].(type) {
	case toy.String:
		idx = strings.Index(string(s), string(a))
	case toy.Char:
		idx = strings.IndexRune(string(s), rune(a))
	default:
		return nil, &toy.InvalidArgumentTypeError{
			Name: "subset",
			Want: "string or char",
			Got:  toy.TypeName(a),
		}
	}
	if idx <= 0 {
		return toy.Int(idx), nil
	}
	return toy.Int(utf8.RuneCountInString(string(s)) - utf8.RuneCountInString(string(s)[idx:])), nil
}

func indexAnyFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func lastIndexFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 2 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 2,
			WantMax: 2,
			Got:     len(args),
		}
	}
	s, ok := args[0].(toy.String)
	if !ok {
		return nil, &toy.InvalidArgumentTypeError{
			Name: "s",
			Want: "string",
			Got:  toy.TypeName(args[0]),
		}
	}
	var idx int
	switch a := args[1].(type) {
	case toy.String:
		idx = strings.LastIndex(string(s), string(a))
	case toy.Char:
		idx = strings.LastIndex(string(s), string(a))
	default:
		return nil, &toy.InvalidArgumentTypeError{
			Name: "subset",
			Want: "string or char",
			Got:  toy.TypeName(a),
		}
	}
	if idx <= 0 {
		return toy.Int(idx), nil
	}
	return toy.Int(utf8.RuneCountInString(string(s)) - utf8.RuneCountInString(string(s)[idx:])), nil
}

func lastIndexAnyFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
