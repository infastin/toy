package base64

import (
	"encoding/base64"

	"github.com/infastin/toy"
)

var Module = &toy.BuiltinModule{
	Name: "base64",
	Members: map[string]toy.Value{
		"encode":       toy.NewBuiltinFunction("base64.encode", makeEncodeFn(base64.StdEncoding)),
		"decode":       toy.NewBuiltinFunction("base64.decode", makeDecodeFn(base64.StdEncoding)),
		"rawEncode":    toy.NewBuiltinFunction("base64.encode", makeEncodeFn(base64.RawStdEncoding)),
		"rawDecode":    toy.NewBuiltinFunction("base64.decode", makeDecodeFn(base64.RawStdEncoding)),
		"urlEncode":    toy.NewBuiltinFunction("base64.encode", makeEncodeFn(base64.URLEncoding)),
		"urlDecode":    toy.NewBuiltinFunction("base64.decode", makeDecodeFn(base64.URLEncoding)),
		"rawURLEncode": toy.NewBuiltinFunction("base64.encode", makeEncodeFn(base64.RawURLEncoding)),
		"rawURLDecode": toy.NewBuiltinFunction("base64.decode", makeDecodeFn(base64.RawURLEncoding)),
	},
}

func makeEncodeFn(enc *base64.Encoding) toy.CallableFunc {
	return func(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
		var data toy.StringOrBytes
		if err := toy.UnpackArgs(args, "data", &data); err != nil {
			return nil, err
		}
		return toy.String(enc.EncodeToString(data.Bytes())), nil
	}
}

func makeDecodeFn(enc *base64.Encoding) toy.CallableFunc {
	return func(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
		var data toy.String
		if err := toy.UnpackArgs(args, "data", &data); err != nil {
			return nil, err
		}
		binary, err := enc.DecodeString(string(data))
		if err != nil {
			return nil, err
		}
		return toy.Bytes(binary), nil
	}
}
