package stdlib

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/infastin/toy"
)

var TextModule = &toy.BuiltinModule{
	Name: "text",
	Members: map[string]toy.Object{
		"contains": &toy.BuiltinFunction{
			Name: "contains",
			Func: textContains,
		},
		"containsAny": &toy.BuiltinFunction{
			Name: "containsAny",
			Func: makeASSRB("s", "chars", strings.ContainsAny),
		},
		"hasPrefix": &toy.BuiltinFunction{
			Name: "hasPrefix",
			Func: makeASSRB("s", "prefix", strings.HasPrefix),
		},
		"hasSuffix": &toy.BuiltinFunction{
			Name: "hasSuffix",
			Func: makeASSRB("s", "suffix", strings.HasSuffix),
		},
		"trimLeft": &toy.BuiltinFunction{
			Name: "trimLeft",
			Func: makeASSRS("s", "cutset", strings.TrimLeft),
		},
		"trimRight": &toy.BuiltinFunction{
			Name: "trimRight",
			Func: makeASSRS("s", "cutset", strings.TrimRight),
		},
		"trimPrefix": &toy.BuiltinFunction{
			Name: "trimPrefix",
			Func: makeASSRS("s", "prefix", strings.TrimPrefix),
		},
		"trimSuffix": &toy.BuiltinFunction{
			Name: "trimSuffix",
			Func: makeASSRS("s", "suffix", strings.TrimSuffix),
		},
		"trimSpace": &toy.BuiltinFunction{
			Name: "trimSpace",
			Func: makeASRS("s", strings.TrimSpace),
		},
		"trim": &toy.BuiltinFunction{
			Name: "trim",
			Func: makeASSRS("s", "cutset", strings.Trim),
		},
		"toLower": &toy.BuiltinFunction{
			Name: "toLower",
			Func: makeASRS("s", strings.ToLower),
		},
		"toUpper": &toy.BuiltinFunction{
			Name: "toUpper",
			Func: makeASRS("s", strings.ToUpper),
		},
		"toTitle": &toy.BuiltinFunction{
			Name: "toTitle",
			Func: makeASRS("s", strings.ToTitle),
		},
		"join": &toy.BuiltinFunction{
			Name: "join",
			Func: textJoin,
		},
		"split": &toy.BuiltinFunction{
			Name: "split",
			Func: textSplit,
		},
		"splitAfter": &toy.BuiltinFunction{
			Name: "splitAfter",
			Func: textSplitAfter,
		},
		"fields": &toy.BuiltinFunction{
			Name: "fields",
			Func: textFields,
		},
		"replace": &toy.BuiltinFunction{
			Name: "replace",
			Func: textReplace,
		},
		"cut": &toy.BuiltinFunction{
			Name: "cut",
			Func: textCut,
		},
		"cutPrefix": &toy.BuiltinFunction{
			Name: "cutPrefix",
			Func: textCutPrefix,
		},
		"cutSuffix": &toy.BuiltinFunction{
			Name: "cutSuffix",
			Func: textCutSuffix,
		},
		"index": &toy.BuiltinFunction{
			Name: "index",
			Func: textIndex,
		},
		"indexAny": &toy.BuiltinFunction{
			Name: "indexAny",
			Func: textIndexAny,
		},
		"lastIndex": &toy.BuiltinFunction{
			Name: "lastIndex",
			Func: textLastIndex,
		},
		"lastIndexAny": &toy.BuiltinFunction{
			Name: "lastIndexAny",
			Func: textLastIndexAny,
		},
	},
}

func makeASRS(name string, fn func(string) string) toy.CallableFunc {
	return func(args ...toy.Object) (toy.Object, error) {
		var s string
		if err := toy.UnpackArgs(args, name, &s); err != nil {
			return nil, err
		}
		return toy.String(fn(s)), nil
	}
}

func makeASSRB(n1, n2 string, fn func(string, string) bool) toy.CallableFunc {
	return func(args ...toy.Object) (toy.Object, error) {
		var s1, s2 string
		if err := toy.UnpackArgs(args, n1, &s1, n2, &s2); err != nil {
			return nil, err
		}
		return toy.Bool(fn(s1, s2)), nil
	}
}

func makeASSRS(n1, n2 string, fn func(string, string) string) toy.CallableFunc {
	return func(args ...toy.Object) (toy.Object, error) {
		var s1, s2 string
		if err := toy.UnpackArgs(args, n1, &s1, n2, &s2); err != nil {
			return nil, err
		}
		return toy.String(fn(s1, s2)), nil
	}
}

func textContains(args ...toy.Object) (toy.Object, error) {
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

func textJoin(args ...toy.Object) (toy.Object, error) {
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

func textSplit(args ...toy.Object) (toy.Object, error) {
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

func textSplitAfter(args ...toy.Object) (toy.Object, error) {
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

func textFields(args ...toy.Object) (toy.Object, error) {
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

func textReplace(args ...toy.Object) (toy.Object, error) {
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

func textCut(args ...toy.Object) (toy.Object, error) {
	var s, sep string
	if err := toy.UnpackArgs(args, "s", &s, "sep", &sep); err != nil {
		return nil, err
	}
	before, after, found := strings.Cut(s, sep)
	return toy.Tuple{toy.String(before), toy.String(after), toy.Bool(found)}, nil
}

func textCutPrefix(args ...toy.Object) (toy.Object, error) {
	var s, prefix string
	if err := toy.UnpackArgs(args, "s", &s, "prefix", &prefix); err != nil {
		return nil, err
	}
	after, found := strings.CutPrefix(s, prefix)
	return toy.Tuple{toy.String(after), toy.Bool(found)}, nil
}

func textCutSuffix(args ...toy.Object) (toy.Object, error) {
	var s, suffix string
	if err := toy.UnpackArgs(args, "s", &s, "suffix", &suffix); err != nil {
		return nil, err
	}
	after, found := strings.CutSuffix(s, suffix)
	return toy.Tuple{toy.String(after), toy.Bool(found)}, nil
}

func textIndex(args ...toy.Object) (toy.Object, error) {
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

func textIndexAny(args ...toy.Object) (toy.Object, error) {
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

func textLastIndex(args ...toy.Object) (toy.Object, error) {
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

func textLastIndexAny(args ...toy.Object) (toy.Object, error) {
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
