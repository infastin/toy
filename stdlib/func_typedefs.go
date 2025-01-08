package stdlib

import "github.com/infastin/toy"

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
