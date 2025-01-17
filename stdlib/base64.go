package stdlib

import (
	"encoding/base64"

	"github.com/infastin/toy"
)

var Base64Module = &toy.BuiltinModule{
	Name: "base64",
	Members: map[string]toy.Object{
		"encode":       toy.NewBuiltinFunction("base64.encode", makeBase64Encode(base64.StdEncoding)),
		"decode":       toy.NewBuiltinFunction("base64.decode", makeBase64Decode(base64.StdEncoding)),
		"rawEncode":    toy.NewBuiltinFunction("base64.encode", makeBase64Encode(base64.RawStdEncoding)),
		"rawDecode":    toy.NewBuiltinFunction("base64.decode", makeBase64Decode(base64.RawStdEncoding)),
		"urlEncode":    toy.NewBuiltinFunction("base64.encode", makeBase64Encode(base64.URLEncoding)),
		"urlDecode":    toy.NewBuiltinFunction("base64.decode", makeBase64Decode(base64.URLEncoding)),
		"rawURLEncode": toy.NewBuiltinFunction("base64.encode", makeBase64Encode(base64.RawURLEncoding)),
		"rawURLDecode": toy.NewBuiltinFunction("base64.decode", makeBase64Decode(base64.RawURLEncoding)),
	},
}

func makeBase64Encode(enc *base64.Encoding) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var data toy.StringOrBytes
		if err := toy.UnpackArgs(args, "data", &data); err != nil {
			return nil, err
		}
		return toy.String(enc.EncodeToString(data.Bytes())), nil
	}
}

func makeBase64Decode(enc *base64.Encoding) toy.CallableFunc {
	return func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
		var data toy.String
		if err := toy.UnpackArgs(args, "data", &data); err != nil {
			return nil, err
		}
		binary, err := enc.DecodeString(string(data))
		if err != nil {
			return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
		}
		return toy.Tuple{toy.Bytes(binary), toy.Nil}, nil
	}
}
