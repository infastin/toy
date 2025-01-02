package stdlib

import (
	"math"

	"github.com/d5/tengo/v2"
)

var mathModule = map[string]tengo.Object{
	"e":                      &tengo.Float{Value: math.E},
	"pi":                     &tengo.Float{Value: math.Pi},
	"phi":                    &tengo.Float{Value: math.Phi},
	"sqrt2":                  &tengo.Float{Value: math.Sqrt2},
	"sqrtE":                  &tengo.Float{Value: math.SqrtE},
	"sqrtPi":                 &tengo.Float{Value: math.SqrtPi},
	"sqrtPhi":                &tengo.Float{Value: math.SqrtPhi},
	"ln2":                    &tengo.Float{Value: math.Ln2},
	"log2E":                  &tengo.Float{Value: math.Log2E},
	"ln10":                   &tengo.Float{Value: math.Ln10},
	"log10E":                 &tengo.Float{Value: math.Log10E},
	"maxFloat32":             &tengo.Float{Value: math.MaxFloat32},
	"smallestNonzeroFloat32": &tengo.Float{Value: math.SmallestNonzeroFloat32},
	"maxFloat64":             &tengo.Float{Value: math.MaxFloat64},
	"smallestNonzeroFloat64": &tengo.Float{Value: math.SmallestNonzeroFloat64},
	"maxInt":                 &tengo.Int{Value: math.MaxInt},
	"minInt":                 &tengo.Int{Value: math.MinInt},
	"maxInt8":                &tengo.Int{Value: math.MaxInt8},
	"minInt8":                &tengo.Int{Value: math.MinInt8},
	"maxInt16":               &tengo.Int{Value: math.MaxInt16},
	"minInt16":               &tengo.Int{Value: math.MinInt16},
	"maxInt32":               &tengo.Int{Value: math.MaxInt32},
	"minInt32":               &tengo.Int{Value: math.MinInt32},
	"maxInt64":               &tengo.Int{Value: math.MaxInt64},
	"minInt64":               &tengo.Int{Value: math.MinInt64},
	"abs": &tengo.UserFunction{
		Name: "abs",
		Func: FuncAFRF(math.Abs),
	},
	"acos": &tengo.UserFunction{
		Name: "acos",
		Func: FuncAFRF(math.Acos),
	},
	"acosh": &tengo.UserFunction{
		Name: "acosh",
		Func: FuncAFRF(math.Acosh),
	},
	"asin": &tengo.UserFunction{
		Name: "asin",
		Func: FuncAFRF(math.Asin),
	},
	"asinh": &tengo.UserFunction{
		Name: "asinh",
		Func: FuncAFRF(math.Asinh),
	},
	"atan": &tengo.UserFunction{
		Name: "atan",
		Func: FuncAFRF(math.Atan),
	},
	"atan2": &tengo.UserFunction{
		Name: "atan2",
		Func: FuncAFFRF(math.Atan2),
	},
	"atanh": &tengo.UserFunction{
		Name: "atanh",
		Func: FuncAFRF(math.Atanh),
	},
	"cbrt": &tengo.UserFunction{
		Name: "cbrt",
		Func: FuncAFRF(math.Cbrt),
	},
	"ceil": &tengo.UserFunction{
		Name: "ceil",
		Func: FuncAFRF(math.Ceil),
	},
	"copysign": &tengo.UserFunction{
		Name: "copysign",
		Func: FuncAFFRF(math.Copysign),
	},
	"cos": &tengo.UserFunction{
		Name: "cos",
		Func: FuncAFRF(math.Cos),
	},
	"cosh": &tengo.UserFunction{
		Name: "cosh",
		Func: FuncAFRF(math.Cosh),
	},
	"dim": &tengo.UserFunction{
		Name: "dim",
		Func: FuncAFFRF(math.Dim),
	},
	"erf": &tengo.UserFunction{
		Name: "erf",
		Func: FuncAFRF(math.Erf),
	},
	"erfc": &tengo.UserFunction{
		Name: "erfc",
		Func: FuncAFRF(math.Erfc),
	},
	"exp": &tengo.UserFunction{
		Name: "exp",
		Func: FuncAFRF(math.Exp),
	},
	"exp2": &tengo.UserFunction{
		Name: "exp2",
		Func: FuncAFRF(math.Exp2),
	},
	"expm1": &tengo.UserFunction{
		Name: "expm1",
		Func: FuncAFRF(math.Expm1),
	},
	"floor": &tengo.UserFunction{
		Name: "floor",
		Func: FuncAFRF(math.Floor),
	},
	"gamma": &tengo.UserFunction{
		Name: "gamma",
		Func: FuncAFRF(math.Gamma),
	},
	"hypot": &tengo.UserFunction{
		Name: "hypot",
		Func: FuncAFFRF(math.Hypot),
	},
	"ilogb": &tengo.UserFunction{
		Name: "ilogb",
		Func: FuncAFRI(math.Ilogb),
	},
	"inf": &tengo.UserFunction{
		Name: "inf",
		Func: FuncAIRF(math.Inf),
	},
	"is_inf": &tengo.UserFunction{
		Name: "is_inf",
		Func: FuncAFIRB(math.IsInf),
	},
	"is_nan": &tengo.UserFunction{
		Name: "is_nan",
		Func: FuncAFRB(math.IsNaN),
	},
	"j0": &tengo.UserFunction{
		Name: "j0",
		Func: FuncAFRF(math.J0),
	},
	"j1": &tengo.UserFunction{
		Name: "j1",
		Func: FuncAFRF(math.J1),
	},
	"jn": &tengo.UserFunction{
		Name: "jn",
		Func: FuncAIFRF(math.Jn),
	},
	"ldexp": &tengo.UserFunction{
		Name: "ldexp",
		Func: FuncAFIRF(math.Ldexp),
	},
	"log": &tengo.UserFunction{
		Name: "log",
		Func: FuncAFRF(math.Log),
	},
	"log10": &tengo.UserFunction{
		Name: "log10",
		Func: FuncAFRF(math.Log10),
	},
	"log1p": &tengo.UserFunction{
		Name: "log1p",
		Func: FuncAFRF(math.Log1p),
	},
	"log2": &tengo.UserFunction{
		Name: "log2",
		Func: FuncAFRF(math.Log2),
	},
	"logb": &tengo.UserFunction{
		Name: "logb",
		Func: FuncAFRF(math.Logb),
	},
	"max": &tengo.UserFunction{
		Name: "max",
		Func: FuncAFFRF(math.Max),
	},
	"min": &tengo.UserFunction{
		Name: "min",
		Func: FuncAFFRF(math.Min),
	},
	"mod": &tengo.UserFunction{
		Name: "mod",
		Func: FuncAFFRF(math.Mod),
	},
	"nan": &tengo.UserFunction{
		Name: "nan",
		Func: FuncARF(math.NaN),
	},
	"nextafter": &tengo.UserFunction{
		Name: "nextafter",
		Func: FuncAFFRF(math.Nextafter),
	},
	"pow": &tengo.UserFunction{
		Name: "pow",
		Func: FuncAFFRF(math.Pow),
	},
	"pow10": &tengo.UserFunction{
		Name: "pow10",
		Func: FuncAIRF(math.Pow10),
	},
	"remainder": &tengo.UserFunction{
		Name: "remainder",
		Func: FuncAFFRF(math.Remainder),
	},
	"signbit": &tengo.UserFunction{
		Name: "signbit",
		Func: FuncAFRB(math.Signbit),
	},
	"sin": &tengo.UserFunction{
		Name: "sin",
		Func: FuncAFRF(math.Sin),
	},
	"sinh": &tengo.UserFunction{
		Name: "sinh",
		Func: FuncAFRF(math.Sinh),
	},
	"sqrt": &tengo.UserFunction{
		Name: "sqrt",
		Func: FuncAFRF(math.Sqrt),
	},
	"tan": &tengo.UserFunction{
		Name: "tan",
		Func: FuncAFRF(math.Tan),
	},
	"tanh": &tengo.UserFunction{
		Name: "tanh",
		Func: FuncAFRF(math.Tanh),
	},
	"trunc": &tengo.UserFunction{
		Name: "trunc",
		Func: FuncAFRF(math.Trunc),
	},
	"y0": &tengo.UserFunction{
		Name: "y0",
		Func: FuncAFRF(math.Y0),
	},
	"y1": &tengo.UserFunction{
		Name: "y1",
		Func: FuncAFRF(math.Y1),
	},
	"yn": &tengo.UserFunction{
		Name: "yn",
		Func: FuncAIFRF(math.Yn),
	},
}
