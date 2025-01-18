package stdlib

import "github.com/infastin/toy"

func makeAR(fn func()) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		if len(args) != 0 {
			return nil, &toy.WrongNumArgumentsError{Got: len(args)}
		}
		fn()
		return toy.Nil, nil
	}
}

func makeASRS(name string, fn func(string) string) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var s string
		if err := toy.UnpackArgs(args, name, &s); err != nil {
			return nil, err
		}
		return toy.String(fn(s)), nil
	}
}

func makeASRSS(name string, fn func(string) (string, string)) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var s string
		if err := toy.UnpackArgs(args, name, &s); err != nil {
			return nil, err
		}
		r1, r2 := fn(s)
		return toy.Tuple{toy.String(r1), toy.String(r2)}, nil
	}
}

func makeASRSs(name string, fn func(string) []string) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var s string
		if err := toy.UnpackArgs(args, name, &s); err != nil {
			return nil, err
		}
		strs := fn(s)
		elems := make([]toy.Object, 0, len(strs))
		for _, str := range strs {
			elems = append(elems, toy.String(str))
		}
		return toy.NewArray(elems), nil
	}
}

func makeASRSB(name string, fn func(string) (string, bool)) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var s string
		if err := toy.UnpackArgs(args, name, &s); err != nil {
			return nil, err
		}
		value, ok := fn(s)
		return toy.Tuple{toy.String(value), toy.Bool(ok)}, nil
	}
}

func makeASSRB(n1, n2 string, fn func(string, string) bool) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var s1, s2 string
		if err := toy.UnpackArgs(args, n1, &s1, n2, &s2); err != nil {
			return nil, err
		}
		return toy.Bool(fn(s1, s2)), nil
	}
}

func makeASSRSB(n1, n2 string, fn func(string, string) (string, bool)) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var s1, s2 string
		if err := toy.UnpackArgs(args, n1, &s1, n2, &s2); err != nil {
			return nil, err
		}
		value, ok := fn(s1, s2)
		return toy.Tuple{toy.String(value), toy.Bool(ok)}, nil
	}
}

func makeASSRS(n1, n2 string, fn func(string, string) string) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var s1, s2 string
		if err := toy.UnpackArgs(args, n1, &s1, n2, &s2); err != nil {
			return nil, err
		}
		return toy.String(fn(s1, s2)), nil
	}
}

func makeAFRF(name string, fn func(float64) float64) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var f float64
		if err := toy.UnpackArgs(args, name, &f); err != nil {
			return nil, err
		}
		return toy.Float(fn(f)), nil
	}
}

func makeAFFRF(n1, n2 string, fn func(float64, float64) float64) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var f1, f2 float64
		if err := toy.UnpackArgs(args, n1, &f1, n2, &f2); err != nil {
			return nil, err
		}
		return toy.Float(fn(f1, f2)), nil
	}
}

func makeAIFRF(n1, n2 string, fn func(int, float64) float64) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var (
			i1 int
			f2 float64
		)
		if err := toy.UnpackArgs(args, n1, &i1, n2, &f2); err != nil {
			return nil, err
		}
		return toy.Float(fn(i1, f2)), nil
	}
}

func makeAFIRF(n1, n2 string, fn func(float64, int) float64) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var (
			f1 float64
			i2 int
		)
		if err := toy.UnpackArgs(args, n1, &f1, n2, &i2); err != nil {
			return nil, err
		}
		return toy.Float(fn(f1, i2)), nil
	}
}

func makeAFRI(name string, fn func(float64) int) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var f float64
		if err := toy.UnpackArgs(args, name, &f); err != nil {
			return nil, err
		}
		return toy.Int(fn(f)), nil
	}
}

func makeAIRF(name string, fn func(int) float64) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var i int
		if err := toy.UnpackArgs(args, name, &i); err != nil {
			return nil, err
		}
		return toy.Float(fn(i)), nil
	}
}

func makeAFRB(name string, fn func(float64) bool) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var f float64
		if err := toy.UnpackArgs(args, name, &f); err != nil {
			return nil, err
		}
		return toy.Bool(fn(f)), nil
	}
}

func makeARI(fn func() int) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		if len(args) != 0 {
			return nil, &toy.WrongNumArgumentsError{Got: len(args)}
		}
		return toy.Int(fn()), nil
	}
}

func makeASSRE(n1, n2 string, fn func(string, string) error) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var s1, s2 string
		if err := toy.UnpackArgs(args, n1, &s1, n2, &s2); err != nil {
			return nil, err
		}
		if err := fn(s1, s2); err != nil {
			return toy.NewError(err.Error()), nil
		}
		return toy.Nil, nil
	}
}

func makeASRE(name string, fn func(string) error) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var s string
		if err := toy.UnpackArgs(args, name, &s); err != nil {
			return nil, err
		}
		if err := fn(s); err != nil {
			return toy.NewError(err.Error()), nil
		}
		return toy.Nil, nil
	}
}

func makeARS(fn func() string) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		if len(args) != 0 {
			return nil, &toy.WrongNumArgumentsError{Got: len(args)}
		}
		return toy.String(fn()), nil
	}
}

func makeARSE(fn func() (string, error)) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		if len(args) != 0 {
			return nil, &toy.WrongNumArgumentsError{Got: len(args)}
		}
		s, err := fn()
		if err != nil {
			return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
		}
		return toy.Tuple{toy.String(s), toy.Nil}, nil
	}
}

func makeASRSE(name string, fn func(string) (string, error)) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var s string
		if err := toy.UnpackArgs(args, name, &s); err != nil {
			return nil, err
		}
		res, err := fn(s)
		if err != nil {
			return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
		}
		return toy.Tuple{toy.String(res), toy.Nil}, nil
	}
}

func makeASRSSE(name string, fn func(string) (string, string, error)) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var s string
		if err := toy.UnpackArgs(args, name, &s); err != nil {
			return nil, err
		}
		r1, r2, err := fn(s)
		if err != nil {
			return toy.Tuple{toy.Nil, toy.Nil, toy.NewError(err.Error())}, nil
		}
		return toy.Tuple{toy.String(r1), toy.String(r2), toy.Nil}, nil
	}
}

func makeASRSsE(name string, fn func(string) ([]string, error)) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var s string
		if err := toy.UnpackArgs(args, name, &s); err != nil {
			return nil, err
		}
		strs, err := fn(s)
		if err != nil {
			return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
		}
		elems := make([]toy.Object, 0, len(strs))
		for _, str := range strs {
			elems = append(elems, toy.String(str))
		}
		return toy.Tuple{toy.NewArray(elems), toy.Nil}, nil
	}
}
