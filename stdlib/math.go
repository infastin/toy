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

		"isInf":    &toy.BuiltinFunction{Name: "math.isInf", Func: mathIsInf},
		"isNegInf": &toy.BuiltinFunction{Name: "math.isInf", Func: mathIsNegInf},
		"isNaN":    &toy.BuiltinFunction{Name: "math.isNaN", Func: makeAFRB("f", math.IsNaN)},

		"abs":       &toy.BuiltinFunction{Name: "math.abs", Func: makeAFRF("x", math.Abs)},
		"acos":      &toy.BuiltinFunction{Name: "math.acos", Func: makeAFRF("x", math.Acos)},
		"acosh":     &toy.BuiltinFunction{Name: "math.acosh", Func: makeAFRF("x", math.Acosh)},
		"asin":      &toy.BuiltinFunction{Name: "math.asin", Func: makeAFRF("x", math.Asin)},
		"asinh":     &toy.BuiltinFunction{Name: "math.asinh", Func: makeAFRF("x", math.Asinh)},
		"atan":      &toy.BuiltinFunction{Name: "math.atan", Func: makeAFRF("x", math.Atan)},
		"atan2":     &toy.BuiltinFunction{Name: "math.atan2", Func: makeAFFRF("y", "x", math.Atan2)},
		"atanh":     &toy.BuiltinFunction{Name: "math.atanh", Func: makeAFRF("x", math.Atanh)},
		"cbrt":      &toy.BuiltinFunction{Name: "math.cbrt", Func: makeAFRF("x", math.Cbrt)},
		"ceil":      &toy.BuiltinFunction{Name: "math.ceil", Func: makeAFRF("x", math.Ceil)},
		"copysign":  &toy.BuiltinFunction{Name: "math.copysign", Func: makeAFFRF("f", "sign", math.Copysign)},
		"cos":       &toy.BuiltinFunction{Name: "math.cos", Func: makeAFRF("x", math.Cos)},
		"cosh":      &toy.BuiltinFunction{Name: "math.cosh", Func: makeAFRF("x", math.Cosh)},
		"dim":       &toy.BuiltinFunction{Name: "math.dim", Func: makeAFFRF("x", "y", math.Dim)},
		"erf":       &toy.BuiltinFunction{Name: "math.erf", Func: makeAFRF("x", math.Erf)},
		"erfc":      &toy.BuiltinFunction{Name: "math.erfc", Func: makeAFRF("x", math.Erfc)},
		"exp":       &toy.BuiltinFunction{Name: "math.exp", Func: makeAFRF("x", math.Exp)},
		"exp2":      &toy.BuiltinFunction{Name: "math.exp2", Func: makeAFRF("x", math.Exp2)},
		"expm1":     &toy.BuiltinFunction{Name: "math.expm1", Func: makeAFRF("x", math.Expm1)},
		"floor":     &toy.BuiltinFunction{Name: "math.floor", Func: makeAFRF("x", math.Floor)},
		"gamma":     &toy.BuiltinFunction{Name: "math.gamma", Func: makeAFRF("x", math.Gamma)},
		"hypot":     &toy.BuiltinFunction{Name: "math.hypot", Func: makeAFFRF("p", "q", math.Hypot)},
		"ilogb":     &toy.BuiltinFunction{Name: "math.ilogb", Func: makeAFRI("x", math.Ilogb)},
		"j0":        &toy.BuiltinFunction{Name: "math.j0", Func: makeAFRF("x", math.J0)},
		"j1":        &toy.BuiltinFunction{Name: "math.j1", Func: makeAFRF("x", math.J1)},
		"jn":        &toy.BuiltinFunction{Name: "math.jn", Func: makeAIFRF("n", "x", math.Jn)},
		"ldexp":     &toy.BuiltinFunction{Name: "math.ldexp", Func: makeAFIRF("frac", "exp", math.Ldexp)},
		"log":       &toy.BuiltinFunction{Name: "math.log", Func: makeAFRF("x", math.Log)},
		"log10":     &toy.BuiltinFunction{Name: "math.log10", Func: makeAFRF("x", math.Log10)},
		"log1p":     &toy.BuiltinFunction{Name: "math.log1p", Func: makeAFRF("x", math.Log1p)},
		"log2":      &toy.BuiltinFunction{Name: "math.log2", Func: makeAFRF("x", math.Log2)},
		"logb":      &toy.BuiltinFunction{Name: "math.logb", Func: makeAFRF("x", math.Logb)},
		"max":       &toy.BuiltinFunction{Name: "math.max", Func: makeAFFRF("x", "y", math.Max)},
		"min":       &toy.BuiltinFunction{Name: "math.min", Func: makeAFFRF("x", "y", math.Min)},
		"mod":       &toy.BuiltinFunction{Name: "math.mod", Func: makeAFFRF("x", "y", math.Mod)},
		"nextafter": &toy.BuiltinFunction{Name: "math.nextafter", Func: makeAFFRF("x", "y", math.Nextafter)},
		"pow":       &toy.BuiltinFunction{Name: "math.pow", Func: makeAFFRF("x", "y", math.Pow)},
		"pow10":     &toy.BuiltinFunction{Name: "math.pow10", Func: makeAIRF("n", math.Pow10)},
		"remainder": &toy.BuiltinFunction{Name: "math.remainder", Func: makeAFFRF("x", "y", math.Remainder)},
		"signbit":   &toy.BuiltinFunction{Name: "math.signbit", Func: makeAFRB("x", math.Signbit)},
		"sin":       &toy.BuiltinFunction{Name: "math.sin", Func: makeAFRF("x", math.Sin)},
		"sinh":      &toy.BuiltinFunction{Name: "math.sinh", Func: makeAFRF("x", math.Sinh)},
		"sqrt":      &toy.BuiltinFunction{Name: "math.sqrt", Func: makeAFRF("x", math.Sqrt)},
		"tan":       &toy.BuiltinFunction{Name: "math.tan", Func: makeAFRF("x", math.Tan)},
		"tanh":      &toy.BuiltinFunction{Name: "math.tanh", Func: makeAFRF("x", math.Tanh)},
		"trunc":     &toy.BuiltinFunction{Name: "math.trunc", Func: makeAFRF("x", math.Trunc)},
		"y0":        &toy.BuiltinFunction{Name: "math.y0", Func: makeAFRF("x", math.Y0)},
		"y1":        &toy.BuiltinFunction{Name: "math.y1", Func: makeAFRF("x", math.Y1)},
		"yn":        &toy.BuiltinFunction{Name: "math.yn", Func: makeAIFRF("n", "x", math.Yn)},
	},
}

func mathIsInf(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var f float64
	if err := toy.UnpackArgs(args, "f", &f); err != nil {
		return nil, err
	}
	return toy.Bool(math.IsInf(f, 1)), nil
}

func mathIsNegInf(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var f float64
	if err := toy.UnpackArgs(args, "f", &f); err != nil {
		return nil, err
	}
	return toy.Bool(math.IsInf(f, -1)), nil
}
