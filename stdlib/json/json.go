package json

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/infastin/toy"
	toytime "github.com/infastin/toy/stdlib/time"

	"github.com/go-faster/jx"
)

var Module = &toy.BuiltinModule{
	Name: "json",
	Members: map[string]toy.Value{
		"encode": toy.NewBuiltinFunction("json.encode", encodeFn),
		"decode": toy.NewBuiltinFunction("json.decode", decodeFn),
	},
}

func sequenceToJSON(enc *jx.Encoder, seq toy.Sequence) (err error) {
	enc.ArrStart()
	for elem := range seq.Elements() {
		if err := objectToJSON(enc, elem); err != nil {
			return err
		}
	}
	enc.ArrEnd()
	return err
}

func mappingToJSON(enc *jx.Encoder, mapping toy.Mapping) (err error) {
	enc.ObjStart()
	for key, value := range mapping.Entries() {
		keyStr, ok := key.(toy.String)
		if !ok {
			return fmt.Errorf("unsupported key type: %s", toy.TypeName(key))
		}
		enc.FieldStart(string(keyStr))
		if err := objectToJSON(enc, value); err != nil {
			return fmt.Errorf("%s: %w", string(keyStr), err)
		}
	}
	enc.ObjEnd()
	return nil
}

func objectToJSON(enc *jx.Encoder, o toy.Value) (err error) {
	switch x := o.(type) {
	case json.Marshaler:
		data, err := x.MarshalJSON()
		if err != nil {
			return err
		}
		enc.Raw(data)
		return nil
	case toy.Int:
		enc.Int64(int64(x))
		return nil
	case toy.Float:
		enc.Float64(float64(x))
		return nil
	case toy.String:
		enc.Str(string(x))
		return nil
	case toy.Bytes:
		enc.Base64(x)
		return nil
	case toytime.Time:
		enc.Str(time.Time(x).Format(time.RFC3339Nano))
		return nil
	case toytime.Duration:
		enc.Str(time.Duration(x).String())
		return nil
	case toy.Bool:
		enc.Bool(bool(x))
		return nil
	case toy.NilValue:
		enc.Null()
		return nil
	case toy.Mapping:
		return mappingToJSON(enc, x)
	case toy.Sequence:
		return sequenceToJSON(enc, x)
	default:
		enc.Str(toy.AsString(x))
		return nil
	}
}

func encodeFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		x      toy.Value
		indent *int
	)
	if err := toy.UnpackArgs(args, "x", &x, "indent?", &indent); err != nil {
		return nil, err
	}

	enc := jx.GetEncoder()
	defer jx.PutEncoder(enc)

	enc.Reset()
	if indent != nil {
		enc.SetIdent(*indent)
	}

	if err := objectToJSON(enc, x); err != nil {
		return nil, err
	}

	return toy.Bytes(enc.Bytes()), nil
}

func jsonArrayToArray(dec *jx.Decoder) (*toy.Array, error) {
	var elems []toy.Value
	if err := dec.Arr(func(d *jx.Decoder) error {
		obj, err := jsonToObject(d)
		if err != nil {
			return err
		}
		elems = append(elems, obj)
		return nil
	}); err != nil {
		return nil, err
	}
	return toy.NewArray(elems), nil
}

func jsonObjectToTable(dec *jx.Decoder) (*toy.Table, error) {
	t := new(toy.Table)
	if err := dec.Obj(func(d *jx.Decoder, key string) error {
		value, err := jsonToObject(d)
		if err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
		t.SetProperty(toy.String(key), value)
		return nil
	}); err != nil {
		return nil, err
	}
	return t, nil
}

func jsonToObject(dec *jx.Decoder) (toy.Value, error) {
	switch dec.Next() {
	case jx.Number:
		num, err := dec.Num()
		if err != nil {
			return nil, err
		}
		if num.IsInt() {
			i, _ := num.Int64()
			return toy.Int(i), nil
		}
		f, _ := num.Float64()
		return toy.Float(f), nil
	case jx.String:
		str, err := dec.Str()
		if err != nil {
			return nil, err
		}
		return toy.String(str), nil
	case jx.Bool:
		b, err := dec.Bool()
		if err != nil {
			return nil, err
		}
		return toy.Bool(b), nil
	case jx.Null:
		if err := dec.Null(); err != nil {
			return nil, err
		}
		return toy.Nil, nil
	case jx.Array:
		return jsonArrayToArray(dec)
	case jx.Object:
		return jsonObjectToTable(dec)
	}
	return nil, dec.Skip()
}

func decodeFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var data toy.StringOrBytes
	if err := toy.UnpackArgs(args, "data", &data); err != nil {
		return nil, err
	}

	dec := jx.GetDecoder()
	defer jx.PutDecoder(dec)

	dec.ResetBytes(data.Bytes())

	obj, err := jsonToObject(dec)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
