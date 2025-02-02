package hex

import (
	"encoding/hex"

	"github.com/infastin/toy"
)

var Module = &toy.BuiltinModule{
	Name: "hex",
	Members: map[string]toy.Value{
		"encode": toy.NewBuiltinFunction("hex.encode", encodeFn),
		"decode": toy.NewBuiltinFunction("hex.decode", decodeFn),
	},
}

func encodeFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var src toy.StringOrBytes
	if err := toy.UnpackArgs(args, "src", &src); err != nil {
		return nil, err
	}
	return toy.String(hex.EncodeToString(src.Bytes())), nil
}

func decodeFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var data toy.StringOrBytes
	if err := toy.UnpackArgs(args, "data", &data); err != nil {
		return nil, err
	}
	res, err := hex.DecodeString(data.String())
	if err != nil {
		return nil, err
	}
	return toy.Bytes(res), nil
}
