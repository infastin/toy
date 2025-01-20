package hex

import (
	"encoding/hex"

	"github.com/infastin/toy"
)

var Module = &toy.BuiltinModule{
	Name: "hex",
	Members: map[string]toy.Object{
		"encode": toy.NewBuiltinFunction("hex.encode", encodeFn),
		"decode": toy.NewBuiltinFunction("hex.decode", decodeFn),
	},
}

func encodeFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var src toy.StringOrBytes
	if err := toy.UnpackArgs(args, "src", &src); err != nil {
		return nil, err
	}
	return toy.String(hex.EncodeToString(src.Bytes())), nil
}

func decodeFn(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var data toy.StringOrBytes
	if err := toy.UnpackArgs(args, "data", &data); err != nil {
		return nil, err
	}
	res, err := hex.DecodeString(data.String())
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{toy.Bytes(res), toy.Nil}, nil
}
