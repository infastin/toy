package stdlib

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/infastin/toy"

	"github.com/go-faster/jx"
)

var JSONModule = &toy.BuiltinModule{
	Name: "json",
	Members: map[string]toy.Object{
		"encode": &toy.BuiltinFunction{Name: "encode", Func: jsonEncode},
		"decode": &toy.BuiltinFunction{Name: "decode", Func: jsonDecode},
	},
}

func sequenceToJSON(enc *jx.Encoder, seq toy.Sequence) (err error) {
	enc.ArrStart()
	for elem := range toy.Elements(seq) {
		if err := objectToJSON(enc, elem); err != nil {
			return err
		}
	}
	enc.ArrEnd()
	return err
}

func mappingToJSON(enc *jx.Encoder, mapping toy.Mapping) (err error) {
	enc.ObjStart()
	for key, value := range toy.Entries(mapping) {
		key, ok := key.(toy.String)
		if !ok {
			continue
		}
		enc.FieldStart(string(key))
		if err := objectToJSON(enc, value); err != nil {
			return fmt.Errorf("%s: %w", string(key), err)
		}
	}
	enc.ObjEnd()
	return nil
}

func objectToJSON(enc *jx.Encoder, o toy.Object) (err error) {
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
	case Time:
		enc.Str(time.Time(x).Format(time.RFC3339Nano))
		return nil
	case Duration:
		enc.Str(x.String())
		return nil
	case toy.Bool:
		enc.Bool(bool(x))
		return nil
	case toy.NilType:
		enc.Null()
		return nil
	case toy.Mapping:
		return mappingToJSON(enc, x)
	case toy.Sequence:
		return sequenceToJSON(enc, x)
	default:
		return fmt.Errorf("'%s' can't be encoded in json", x.TypeName())
	}
}

func jsonEncode(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		x      toy.Object
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
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}

	return toy.Tuple{toy.Bytes(enc.Bytes()), toy.Nil}, nil
}

func jsonArrayToArray(dec *jx.Decoder) (*toy.Array, error) {
	var elems []toy.Object
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

func jsonObjectToMap(dec *jx.Decoder) (*toy.Map, error) {
	m := new(toy.Map)
	if err := dec.Obj(func(d *jx.Decoder, key string) error {
		value, err := jsonToObject(d)
		if err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
		m.IndexSet(toy.String(key), value)
		return nil
	}); err != nil {
		return nil, err
	}
	return m, nil
}

func jsonToObject(dec *jx.Decoder) (toy.Object, error) {
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
		return jsonObjectToMap(dec)
	}
	return nil, dec.Skip()
}

func jsonDecode(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var data toy.StringOrBytes
	if err := toy.UnpackArgs(args, "data", &data); err != nil {
		return nil, err
	}

	dec := jx.GetDecoder()
	defer jx.PutDecoder(dec)

	dec.ResetBytes(data.Bytes())

	obj, err := jsonToObject(dec)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, err
	}

	return toy.Tuple{obj, toy.Nil}, nil
}
