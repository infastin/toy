package stdlib

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/infastin/toy"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var TextModule = &toy.BuiltinModule{
	Name: "text",
	Members: map[string]toy.Object{
		"contains":     &toy.BuiltinFunction{Name: "contains", Func: textContains},
		"containsAny":  &toy.BuiltinFunction{Name: "containsAny", Func: makeASSRB("s", "chars", strings.ContainsAny)},
		"hasPrefix":    &toy.BuiltinFunction{Name: "hasPrefix", Func: makeASSRB("s", "prefix", strings.HasPrefix)},
		"hasSuffix":    &toy.BuiltinFunction{Name: "hasSuffix", Func: makeASSRB("s", "suffix", strings.HasSuffix)},
		"trimLeft":     &toy.BuiltinFunction{Name: "trimLeft", Func: makeASSRS("s", "cutset", strings.TrimLeft)},
		"trimRight":    &toy.BuiltinFunction{Name: "trimRight", Func: makeASSRS("s", "cutset", strings.TrimRight)},
		"trimPrefix":   &toy.BuiltinFunction{Name: "trimPrefix", Func: makeASSRS("s", "prefix", strings.TrimPrefix)},
		"trimSuffix":   &toy.BuiltinFunction{Name: "trimSuffix", Func: makeASSRS("s", "suffix", strings.TrimSuffix)},
		"trimSpace":    &toy.BuiltinFunction{Name: "trimSpace", Func: makeASRS("s", strings.TrimSpace)},
		"trim":         &toy.BuiltinFunction{Name: "trim", Func: makeASSRS("s", "cutset", strings.Trim)},
		"toLower":      &toy.BuiltinFunction{Name: "toLower", Func: makeASRS("s", strings.ToLower)},
		"toUpper":      &toy.BuiltinFunction{Name: "toUpper", Func: makeASRS("s", strings.ToUpper)},
		"toTitle":      &toy.BuiltinFunction{Name: "toTitle", Func: textToTitle},
		"join":         &toy.BuiltinFunction{Name: "join", Func: textJoin},
		"split":        &toy.BuiltinFunction{Name: "split", Func: textSplit},
		"splitAfter":   &toy.BuiltinFunction{Name: "splitAfter", Func: textSplitAfter},
		"fields":       &toy.BuiltinFunction{Name: "fields", Func: textFields},
		"replace":      &toy.BuiltinFunction{Name: "replace", Func: textReplace},
		"cut":          &toy.BuiltinFunction{Name: "cut", Func: textCut},
		"cutPrefix":    &toy.BuiltinFunction{Name: "cutPrefix", Func: textCutPrefix},
		"cutSuffix":    &toy.BuiltinFunction{Name: "cutSuffix", Func: textCutSuffix},
		"index":        &toy.BuiltinFunction{Name: "index", Func: textIndex},
		"indexAny":     &toy.BuiltinFunction{Name: "indexAny", Func: textIndexAny},
		"lastIndex":    &toy.BuiltinFunction{Name: "lastIndex", Func: textLastIndex},
		"lastIndexAny": &toy.BuiltinFunction{Name: "lastIndexAny", Func: textLastIndexAny},
	},
}

func textContains(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
			Got:  args[0].TypeName(),
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
			Got:  a2.TypeName(),
		}
	}
}

func textToTitle(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s string
	if err := toy.UnpackArgs(args, "s", &s); err != nil {
		return nil, err
	}
	caser := cases.Title(language.Und, cases.NoLower)
	return toy.String(caser.String(s)), nil
}

func textJoin(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
			return nil, fmt.Errorf("%s[%d]: want string, got %s", elems.TypeName(), i, elem.TypeName())
		}
		strs = append(strs, string(s))
	}
	return toy.String(strings.Join(strs, sep)), nil
}

func textSplit(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func textSplitAfter(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func textFields(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s string
	if err := toy.UnpackArgs(args, "s", &s); err != nil {
		return nil, err
	}
	var elems []toy.Object
	for _, field := range strings.Fields(s) {
		elems = append(elems, toy.String(field))
	}
	return toy.NewArray(elems), nil
}

func textReplace(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func textCut(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s, sep string
	if err := toy.UnpackArgs(args, "s", &s, "sep", &sep); err != nil {
		return nil, err
	}
	before, after, found := strings.Cut(s, sep)
	return toy.Tuple{toy.String(before), toy.String(after), toy.Bool(found)}, nil
}

func textCutPrefix(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s, prefix string
	if err := toy.UnpackArgs(args, "s", &s, "prefix", &prefix); err != nil {
		return nil, err
	}
	after, found := strings.CutPrefix(s, prefix)
	return toy.Tuple{toy.String(after), toy.Bool(found)}, nil
}

func textCutSuffix(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s, suffix string
	if err := toy.UnpackArgs(args, "s", &s, "suffix", &suffix); err != nil {
		return nil, err
	}
	after, found := strings.CutSuffix(s, suffix)
	return toy.Tuple{toy.String(after), toy.Bool(found)}, nil
}

func textIndex(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
			Got:  args[0].TypeName(),
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
			Got:  a.TypeName(),
		}
	}
	if idx <= 0 {
		return toy.Int(idx), nil
	}
	return toy.Int(utf8.RuneCountInString(string(s)) - utf8.RuneCountInString(string(s)[idx:])), nil
}

func textIndexAny(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func textLastIndex(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
			Got:  args[0].TypeName(),
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
			Got:  a.TypeName(),
		}
	}
	if idx <= 0 {
		return toy.Int(idx), nil
	}
	return toy.Int(utf8.RuneCountInString(string(s)) - utf8.RuneCountInString(string(s)[idx:])), nil
}

func textLastIndexAny(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
