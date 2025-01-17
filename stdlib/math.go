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

		"isInf":    toy.NewBuiltinFunction("math.isInf", mathIsInf),
		"isNegInf": toy.NewBuiltinFunction("math.isInf", mathIsNegInf),
		"isNaN":    toy.NewBuiltinFunction("math.isNaN", makeAFRB("f", math.IsNaN)),

		"abs":       toy.NewBuiltinFunction("math.abs", makeAFRF("x", math.Abs)),
		"acos":      toy.NewBuiltinFunction("math.acos", makeAFRF("x", math.Acos)),
		"acosh":     toy.NewBuiltinFunction("math.acosh", makeAFRF("x", math.Acosh)),
		"asin":      toy.NewBuiltinFunction("math.asin", makeAFRF("x", math.Asin)),
		"asinh":     toy.NewBuiltinFunction("math.asinh", makeAFRF("x", math.Asinh)),
		"atan":      toy.NewBuiltinFunction("math.atan", makeAFRF("x", math.Atan)),
		"atan2":     toy.NewBuiltinFunction("math.atan2", makeAFFRF("y", "x", math.Atan2)),
		"atanh":     toy.NewBuiltinFunction("math.atanh", makeAFRF("x", math.Atanh)),
		"cbrt":      toy.NewBuiltinFunction("math.cbrt", makeAFRF("x", math.Cbrt)),
		"ceil":      toy.NewBuiltinFunction("math.ceil", makeAFRF("x", math.Ceil)),
		"copysign":  toy.NewBuiltinFunction("math.copysign", makeAFFRF("f", "sign", math.Copysign)),
		"cos":       toy.NewBuiltinFunction("math.cos", makeAFRF("x", math.Cos)),
		"cosh":      toy.NewBuiltinFunction("math.cosh", makeAFRF("x", math.Cosh)),
		"dim":       toy.NewBuiltinFunction("math.dim", makeAFFRF("x", "y", math.Dim)),
		"erf":       toy.NewBuiltinFunction("math.erf", makeAFRF("x", math.Erf)),
		"erfc":      toy.NewBuiltinFunction("math.erfc", makeAFRF("x", math.Erfc)),
		"exp":       toy.NewBuiltinFunction("math.exp", makeAFRF("x", math.Exp)),
		"exp2":      toy.NewBuiltinFunction("math.exp2", makeAFRF("x", math.Exp2)),
		"expm1":     toy.NewBuiltinFunction("math.expm1", makeAFRF("x", math.Expm1)),
		"floor":     toy.NewBuiltinFunction("math.floor", makeAFRF("x", math.Floor)),
		"gamma":     toy.NewBuiltinFunction("math.gamma", makeAFRF("x", math.Gamma)),
		"hypot":     toy.NewBuiltinFunction("math.hypot", makeAFFRF("p", "q", math.Hypot)),
		"ilogb":     toy.NewBuiltinFunction("math.ilogb", makeAFRI("x", math.Ilogb)),
		"j0":        toy.NewBuiltinFunction("math.j0", makeAFRF("x", math.J0)),
		"j1":        toy.NewBuiltinFunction("math.j1", makeAFRF("x", math.J1)),
		"jn":        toy.NewBuiltinFunction("math.jn", makeAIFRF("n", "x", math.Jn)),
		"ldexp":     toy.NewBuiltinFunction("math.ldexp", makeAFIRF("frac", "exp", math.Ldexp)),
		"log":       toy.NewBuiltinFunction("math.log", makeAFRF("x", math.Log)),
		"log10":     toy.NewBuiltinFunction("math.log10", makeAFRF("x", math.Log10)),
		"log1p":     toy.NewBuiltinFunction("math.log1p", makeAFRF("x", math.Log1p)),
		"log2":      toy.NewBuiltinFunction("math.log2", makeAFRF("x", math.Log2)),
		"logb":      toy.NewBuiltinFunction("math.logb", makeAFRF("x", math.Logb)),
		"max":       toy.NewBuiltinFunction("math.max", makeAFFRF("x", "y", math.Max)),
		"min":       toy.NewBuiltinFunction("math.min", makeAFFRF("x", "y", math.Min)),
		"mod":       toy.NewBuiltinFunction("math.mod", makeAFFRF("x", "y", math.Mod)),
		"nextafter": toy.NewBuiltinFunction("math.nextafter", makeAFFRF("x", "y", math.Nextafter)),
		"pow":       toy.NewBuiltinFunction("math.pow", makeAFFRF("x", "y", math.Pow)),
		"pow10":     toy.NewBuiltinFunction("math.pow10", makeAIRF("n", math.Pow10)),
		"remainder": toy.NewBuiltinFunction("math.remainder", makeAFFRF("x", "y", math.Remainder)),
		"signbit":   toy.NewBuiltinFunction("math.signbit", makeAFRB("x", math.Signbit)),
		"sin":       toy.NewBuiltinFunction("math.sin", makeAFRF("x", math.Sin)),
		"sinh":      toy.NewBuiltinFunction("math.sinh", makeAFRF("x", math.Sinh)),
		"sqrt":      toy.NewBuiltinFunction("math.sqrt", makeAFRF("x", math.Sqrt)),
		"tan":       toy.NewBuiltinFunction("math.tan", makeAFRF("x", math.Tan)),
		"tanh":      toy.NewBuiltinFunction("math.tanh", makeAFRF("x", math.Tanh)),
		"trunc":     toy.NewBuiltinFunction("math.trunc", makeAFRF("x", math.Trunc)),
		"y0":        toy.NewBuiltinFunction("math.y0", makeAFRF("x", math.Y0)),
		"y1":        toy.NewBuiltinFunction("math.y1", makeAFRF("x", math.Y1)),
		"yn":        toy.NewBuiltinFunction("math.yn", makeAIFRF("n", "x", math.Yn)),
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
