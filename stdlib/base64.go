package stdlib

import (
	"encoding/base64"

	"github.com/d5/tengo/v2"
)

var base64Module = map[string]tengo.Object{
	"encode": &tengo.UserFunction{
		Func: FuncAYRS(base64.StdEncoding.EncodeToString),
	},
	"decode": &tengo.UserFunction{
		Func: FuncASRYE(base64.StdEncoding.DecodeString),
	},
	"raw_encode": &tengo.UserFunction{
		Func: FuncAYRS(base64.RawStdEncoding.EncodeToString),
	},
	"raw_decode": &tengo.UserFunction{
		Func: FuncASRYE(base64.RawStdEncoding.DecodeString),
	},
	"url_encode": &tengo.UserFunction{
		Func: FuncAYRS(base64.URLEncoding.EncodeToString),
	},
	"url_decode": &tengo.UserFunction{
		Func: FuncASRYE(base64.URLEncoding.DecodeString),
	},
	"raw_url_encode": &tengo.UserFunction{
		Func: FuncAYRS(base64.RawURLEncoding.EncodeToString),
	},
	"raw_url_decode": &tengo.UserFunction{
		Func: FuncASRYE(base64.RawURLEncoding.DecodeString),
	},
}
