package stdlib

import (
	"math"

	"github.com/infastin/toy"
)

var MathModule = &toy.BuiltinModule{
	Name: "math",
	Members: map[string]toy.Object{
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

		"isInf":    &toy.BuiltinFunction{Name: "isInf", Func: mathIsInf},
		"isNegInf": &toy.BuiltinFunction{Name: "isInf", Func: mathIsNegInf},
		"isNaN":    &toy.BuiltinFunction{Name: "isNaN", Func: makeAFRB("f", math.IsNaN)},

		"abs":       &toy.BuiltinFunction{Name: "abs", Func: makeAFRF("x", math.Abs)},
		"acos":      &toy.BuiltinFunction{Name: "acos", Func: makeAFRF("x", math.Acos)},
		"acosh":     &toy.BuiltinFunction{Name: "acosh", Func: makeAFRF("x", math.Acosh)},
		"asin":      &toy.BuiltinFunction{Name: "asin", Func: makeAFRF("x", math.Asin)},
		"asinh":     &toy.BuiltinFunction{Name: "asinh", Func: makeAFRF("x", math.Asinh)},
		"atan":      &toy.BuiltinFunction{Name: "atan", Func: makeAFRF("x", math.Atan)},
		"atan2":     &toy.BuiltinFunction{Name: "atan2", Func: makeAFFRF("y", "x", math.Atan2)},
		"atanh":     &toy.BuiltinFunction{Name: "atanh", Func: makeAFRF("x", math.Atanh)},
		"cbrt":      &toy.BuiltinFunction{Name: "cbrt", Func: makeAFRF("x", math.Cbrt)},
		"ceil":      &toy.BuiltinFunction{Name: "ceil", Func: makeAFRF("x", math.Ceil)},
		"copysign":  &toy.BuiltinFunction{Name: "copysign", Func: makeAFFRF("f", "sign", math.Copysign)},
		"cos":       &toy.BuiltinFunction{Name: "cos", Func: makeAFRF("x", math.Cos)},
		"cosh":      &toy.BuiltinFunction{Name: "cosh", Func: makeAFRF("x", math.Cosh)},
		"dim":       &toy.BuiltinFunction{Name: "dim", Func: makeAFFRF("x", "y", math.Dim)},
		"erf":       &toy.BuiltinFunction{Name: "erf", Func: makeAFRF("x", math.Erf)},
		"erfc":      &toy.BuiltinFunction{Name: "erfc", Func: makeAFRF("x", math.Erfc)},
		"exp":       &toy.BuiltinFunction{Name: "exp", Func: makeAFRF("x", math.Exp)},
		"exp2":      &toy.BuiltinFunction{Name: "exp2", Func: makeAFRF("x", math.Exp2)},
		"expm1":     &toy.BuiltinFunction{Name: "expm1", Func: makeAFRF("x", math.Expm1)},
		"floor":     &toy.BuiltinFunction{Name: "floor", Func: makeAFRF("x", math.Floor)},
		"gamma":     &toy.BuiltinFunction{Name: "gamma", Func: makeAFRF("x", math.Gamma)},
		"hypot":     &toy.BuiltinFunction{Name: "hypot", Func: makeAFFRF("p", "q", math.Hypot)},
		"ilogb":     &toy.BuiltinFunction{Name: "ilogb", Func: makeAFRI("x", math.Ilogb)},
		"j0":        &toy.BuiltinFunction{Name: "j0", Func: makeAFRF("x", math.J0)},
		"j1":        &toy.BuiltinFunction{Name: "j1", Func: makeAFRF("x", math.J1)},
		"jn":        &toy.BuiltinFunction{Name: "jn", Func: makeAIFRF("n", "x", math.Jn)},
		"ldexp":     &toy.BuiltinFunction{Name: "ldexp", Func: makeAFIRF("frac", "exp", math.Ldexp)},
		"log":       &toy.BuiltinFunction{Name: "log", Func: makeAFRF("x", math.Log)},
		"log10":     &toy.BuiltinFunction{Name: "log10", Func: makeAFRF("x", math.Log10)},
		"log1p":     &toy.BuiltinFunction{Name: "log1p", Func: makeAFRF("x", math.Log1p)},
		"log2":      &toy.BuiltinFunction{Name: "log2", Func: makeAFRF("x", math.Log2)},
		"logb":      &toy.BuiltinFunction{Name: "logb", Func: makeAFRF("x", math.Logb)},
		"max":       &toy.BuiltinFunction{Name: "max", Func: makeAFFRF("x", "y", math.Max)},
		"min":       &toy.BuiltinFunction{Name: "min", Func: makeAFFRF("x", "y", math.Min)},
		"mod":       &toy.BuiltinFunction{Name: "mod", Func: makeAFFRF("x", "y", math.Mod)},
		"nextafter": &toy.BuiltinFunction{Name: "nextafter", Func: makeAFFRF("x", "y", math.Nextafter)},
		"pow":       &toy.BuiltinFunction{Name: "pow", Func: makeAFFRF("x", "y", math.Pow)},
		"pow10":     &toy.BuiltinFunction{Name: "pow10", Func: makeAIRF("n", math.Pow10)},
		"remainder": &toy.BuiltinFunction{Name: "remainder", Func: makeAFFRF("x", "y", math.Remainder)},
		"signbit":   &toy.BuiltinFunction{Name: "signbit", Func: makeAFRB("x", math.Signbit)},
		"sin":       &toy.BuiltinFunction{Name: "sin", Func: makeAFRF("x", math.Sin)},
		"sinh":      &toy.BuiltinFunction{Name: "sinh", Func: makeAFRF("x", math.Sinh)},
		"sqrt":      &toy.BuiltinFunction{Name: "sqrt", Func: makeAFRF("x", math.Sqrt)},
		"tan":       &toy.BuiltinFunction{Name: "tan", Func: makeAFRF("x", math.Tan)},
		"tanh":      &toy.BuiltinFunction{Name: "tanh", Func: makeAFRF("x", math.Tanh)},
		"trunc":     &toy.BuiltinFunction{Name: "trunc", Func: makeAFRF("x", math.Trunc)},
		"y0":        &toy.BuiltinFunction{Name: "y0", Func: makeAFRF("x", math.Y0)},
		"y1":        &toy.BuiltinFunction{Name: "y1", Func: makeAFRF("x", math.Y1)},
		"yn":        &toy.BuiltinFunction{Name: "yn", Func: makeAIFRF("n", "x", math.Yn)},
	},
}

func mathIsInf(args ...toy.Object) (toy.Object, error) {
	var f float64
	if err := toy.UnpackArgs(args, "f", &f); err != nil {
		return nil, err
	}
	return toy.Bool(math.IsInf(f, 1)), nil
}

func mathIsNegInf(args ...toy.Object) (toy.Object, error) {

	var f float64
	if err := toy.UnpackArgs(args, "f", &f); err != nil {
		return nil, err
	}
	return toy.Bool(math.IsInf(f, -1)), nil
}
