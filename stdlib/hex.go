package stdlib

import (
	"encoding/hex"

	"github.com/infastin/toy"
)

var HexModule = &toy.BuiltinModule{
	Name: "hex",
	Members: map[string]toy.Object{
		"encode": toy.NewBuiltinFunction("hex.encode", hexEncode),
		"decode": toy.NewBuiltinFunction("hex.decode", hexDecode),
	},
}

func hexEncode(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var src toy.StringOrBytes
	if err := toy.UnpackArgs(args, "src", &src); err != nil {
		return nil, err
	}
	return toy.String(hex.EncodeToString(src.Bytes())), nil
}

func hexDecode(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
