package stdlib

import (
	"math/rand"

	"github.com/d5/tengo/v2"
)

var randModule = map[string]tengo.Object{
	"int": &tengo.UserFunction{
		Name: "int",
		Func: FuncARI64(rand.Int63),
	},
	"float": &tengo.UserFunction{
		Name: "float",
		Func: FuncARF(rand.Float64),
	},
	"intn": &tengo.UserFunction{
		Name: "intn",
		Func: FuncAI64RI64(rand.Int63n),
	},
	"exp_float": &tengo.UserFunction{
		Name: "exp_float",
		Func: FuncARF(rand.ExpFloat64),
	},
	"norm_float": &tengo.UserFunction{
		Name: "norm_float",
		Func: FuncARF(rand.NormFloat64),
	},
	"perm": &tengo.UserFunction{
		Name: "perm",
		Func: FuncAIRIs(rand.Perm),
	},
	"seed": &tengo.UserFunction{
		Name: "seed",
		Func: FuncAI64R(rand.Seed),
	},
	"read": &tengo.UserFunction{
		Name: "read",
		Func: func(args ...tengo.Object) (ret tengo.Object, err error) {
			if len(args) != 1 {
				return nil, tengo.ErrWrongNumArguments
			}
			y1, ok := args[0].(*tengo.Bytes)
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "first",
					Expected: "bytes",
					Found:    args[0].TypeName(),
				}
			}
			res, err := rand.Read(y1.Value)
			if err != nil {
				ret = wrapError(err)
				return
			}
			return &tengo.Int{Value: int64(res)}, nil
		},
	},
	"rand": &tengo.UserFunction{
		Name: "rand",
		Func: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) != 1 {
				return nil, tengo.ErrWrongNumArguments
			}
			i1, ok := tengo.ToInt64(args[0])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "first",
					Expected: "int(compatible)",
					Found:    args[0].TypeName(),
				}
			}
			src := rand.NewSource(i1)
			return randRand(rand.New(src)), nil
		},
	},
}

func randRand(r *rand.Rand) *tengo.ImmutableMap {
	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"int": &tengo.UserFunction{
				Name: "int",
				Func: FuncARI64(r.Int63),
			},
			"float": &tengo.UserFunction{
				Name: "float",
				Func: FuncARF(r.Float64),
			},
			"intn": &tengo.UserFunction{
				Name: "intn",
				Func: FuncAI64RI64(r.Int63n),
			},
			"exp_float": &tengo.UserFunction{
				Name: "exp_float",
				Func: FuncARF(r.ExpFloat64),
			},
			"norm_float": &tengo.UserFunction{
				Name: "norm_float",
				Func: FuncARF(r.NormFloat64),
			},
			"perm": &tengo.UserFunction{
				Name: "perm",
				Func: FuncAIRIs(r.Perm),
			},
			"seed": &tengo.UserFunction{
				Name: "seed",
				Func: FuncAI64R(r.Seed),
			},
			"read": &tengo.UserFunction{
				Name: "read",
				Func: func(args ...tengo.Object) (
					ret tengo.Object,
					err error,
				) {
					if len(args) != 1 {
						return nil, tengo.ErrWrongNumArguments
					}
					y1, ok := args[0].(*tengo.Bytes)
					if !ok {
						return nil, tengo.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "bytes",
							Found:    args[0].TypeName(),
						}
					}
					res, err := r.Read(y1.Value)
					if err != nil {
						ret = wrapError(err)
						return
					}
					return &tengo.Int{Value: int64(res)}, nil
				},
			},
		},
	}
}
