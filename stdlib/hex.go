package stdlib

import (
	"encoding/hex"

	"github.com/d5/tengo/v2"
)

var hexModule = map[string]tengo.Object{
	"encode": &tengo.UserFunction{Func: FuncAYRS(hex.EncodeToString)},
	"decode": &tengo.UserFunction{Func: FuncASRYE(hex.DecodeString)},
}
