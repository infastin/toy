package math

import (
	"math"

	"github.com/infastin/toy"
	"github.com/infastin/toy/internal/fndef"
)

var Module = &toy.BuiltinModule{
	Name: "math",
	Members: map[string]toy.Value{
		"e":                    toy.Float(math.E),
		"pi":                   toy.Float(math.Pi),
		"phi":                  toy.Float(math.Phi),
		"sqrt2":                toy.Float(math.Sqrt2),
		"sqrtE":                toy.Float(math.SqrtE),
		"sqrtPi":               toy.Float(math.SqrtPi),
		"sqrtPhi":              toy.Float(math.SqrtPhi),
		"ln2":                  toy.Float(math.Ln2),
		"log2E":                toy.Float(math.Log2E),
		"ln10":                 toy.Float(math.Ln10),
		"log10E":               toy.Float(math.Log10E),
		"maxFloat":             toy.Float(math.MaxFloat64),
		"smallestNonzeroFloat": toy.Float(math.SmallestNonzeroFloat64),
		"maxInt":               toy.Int(math.MaxInt64),
		"minInt":               toy.Int(math.MinInt64),
		"nan":                  toy.Float(math.NaN()),
		"inf":                  toy.Float(math.Inf(1)),
		"negInf":               toy.Float(math.Inf(-1)),

		"isInf":    toy.NewBuiltinFunction("math.isInf", isInfFn),
		"isNegInf": toy.NewBuiltinFunction("math.isInf", isNegInfFn),
		"isNaN":    toy.NewBuiltinFunction("math.isNaN", fndef.AFRB("f", math.IsNaN)),

		"abs":       toy.NewBuiltinFunction("math.abs", fndef.AFRF("x", math.Abs)),
		"acos":      toy.NewBuiltinFunction("math.acos", fndef.AFRF("x", math.Acos)),
		"acosh":     toy.NewBuiltinFunction("math.acosh", fndef.AFRF("x", math.Acosh)),
		"asin":      toy.NewBuiltinFunction("math.asin", fndef.AFRF("x", math.Asin)),
		"asinh":     toy.NewBuiltinFunction("math.asinh", fndef.AFRF("x", math.Asinh)),
		"atan":      toy.NewBuiltinFunction("math.atan", fndef.AFRF("x", math.Atan)),
		"atan2":     toy.NewBuiltinFunction("math.atan2", fndef.AFFRF("y", "x", math.Atan2)),
		"atanh":     toy.NewBuiltinFunction("math.atanh", fndef.AFRF("x", math.Atanh)),
		"cbrt":      toy.NewBuiltinFunction("math.cbrt", fndef.AFRF("x", math.Cbrt)),
		"ceil":      toy.NewBuiltinFunction("math.ceil", fndef.AFRF("x", math.Ceil)),
		"copysign":  toy.NewBuiltinFunction("math.copysign", fndef.AFFRF("f", "sign", math.Copysign)),
		"cos":       toy.NewBuiltinFunction("math.cos", fndef.AFRF("x", math.Cos)),
		"cosh":      toy.NewBuiltinFunction("math.cosh", fndef.AFRF("x", math.Cosh)),
		"dim":       toy.NewBuiltinFunction("math.dim", fndef.AFFRF("x", "y", math.Dim)),
		"erf":       toy.NewBuiltinFunction("math.erf", fndef.AFRF("x", math.Erf)),
		"erfc":      toy.NewBuiltinFunction("math.erfc", fndef.AFRF("x", math.Erfc)),
		"exp":       toy.NewBuiltinFunction("math.exp", fndef.AFRF("x", math.Exp)),
		"exp2":      toy.NewBuiltinFunction("math.exp2", fndef.AFRF("x", math.Exp2)),
		"expm1":     toy.NewBuiltinFunction("math.expm1", fndef.AFRF("x", math.Expm1)),
		"floor":     toy.NewBuiltinFunction("math.floor", fndef.AFRF("x", math.Floor)),
		"gamma":     toy.NewBuiltinFunction("math.gamma", fndef.AFRF("x", math.Gamma)),
		"hypot":     toy.NewBuiltinFunction("math.hypot", fndef.AFFRF("p", "q", math.Hypot)),
		"ilogb":     toy.NewBuiltinFunction("math.ilogb", fndef.AFRI("x", math.Ilogb)),
		"j0":        toy.NewBuiltinFunction("math.j0", fndef.AFRF("x", math.J0)),
		"j1":        toy.NewBuiltinFunction("math.j1", fndef.AFRF("x", math.J1)),
		"jn":        toy.NewBuiltinFunction("math.jn", fndef.AIFRF("n", "x", math.Jn)),
		"ldexp":     toy.NewBuiltinFunction("math.ldexp", fndef.AFIRF("frac", "exp", math.Ldexp)),
		"log":       toy.NewBuiltinFunction("math.log", fndef.AFRF("x", math.Log)),
		"log10":     toy.NewBuiltinFunction("math.log10", fndef.AFRF("x", math.Log10)),
		"log1p":     toy.NewBuiltinFunction("math.log1p", fndef.AFRF("x", math.Log1p)),
		"log2":      toy.NewBuiltinFunction("math.log2", fndef.AFRF("x", math.Log2)),
		"logb":      toy.NewBuiltinFunction("math.logb", fndef.AFRF("x", math.Logb)),
		"max":       toy.NewBuiltinFunction("math.max", fndef.AFFRF("x", "y", math.Max)),
		"min":       toy.NewBuiltinFunction("math.min", fndef.AFFRF("x", "y", math.Min)),
		"mod":       toy.NewBuiltinFunction("math.mod", fndef.AFFRF("x", "y", math.Mod)),
		"nextafter": toy.NewBuiltinFunction("math.nextafter", fndef.AFFRF("x", "y", math.Nextafter)),
		"pow":       toy.NewBuiltinFunction("math.pow", fndef.AFFRF("x", "y", math.Pow)),
		"pow10":     toy.NewBuiltinFunction("math.pow10", fndef.AIRF("n", math.Pow10)),
		"remainder": toy.NewBuiltinFunction("math.remainder", fndef.AFFRF("x", "y", math.Remainder)),
		"signbit":   toy.NewBuiltinFunction("math.signbit", fndef.AFRB("x", math.Signbit)),
		"sin":       toy.NewBuiltinFunction("math.sin", fndef.AFRF("x", math.Sin)),
		"sinh":      toy.NewBuiltinFunction("math.sinh", fndef.AFRF("x", math.Sinh)),
		"sqrt":      toy.NewBuiltinFunction("math.sqrt", fndef.AFRF("x", math.Sqrt)),
		"tan":       toy.NewBuiltinFunction("math.tan", fndef.AFRF("x", math.Tan)),
		"tanh":      toy.NewBuiltinFunction("math.tanh", fndef.AFRF("x", math.Tanh)),
		"trunc":     toy.NewBuiltinFunction("math.trunc", fndef.AFRF("x", math.Trunc)),
		"y0":        toy.NewBuiltinFunction("math.y0", fndef.AFRF("x", math.Y0)),
		"y1":        toy.NewBuiltinFunction("math.y1", fndef.AFRF("x", math.Y1)),
		"yn":        toy.NewBuiltinFunction("math.yn", fndef.AIFRF("n", "x", math.Yn)),
	},
}

func isInfFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var f float64
	if err := toy.UnpackArgs(args, "f", &f); err != nil {
		return nil, err
	}
	return toy.Bool(math.IsInf(f, 1)), nil
}

func isNegInfFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var f float64
	if err := toy.UnpackArgs(args, "f", &f); err != nil {
		return nil, err
	}
	return toy.Bool(math.IsInf(f, -1)), nil
}
