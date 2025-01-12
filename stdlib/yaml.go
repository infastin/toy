package stdlib

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/infastin/toy"

	"gopkg.in/yaml.v3"
)

var YAMLModule = &toy.BuiltinModule{
	Name: "yaml",
	Members: map[string]toy.Object{
		"encode": &toy.BuiltinFunction{Name: "encode", Func: yamlEncode},
		"decode": &toy.BuiltinFunction{Name: "decode", Func: yamlDecode},
	},
}

func sequenceToYAML(seq toy.Sequence) (*yaml.Node, error) {
	nodes := make([]*yaml.Node, 0, seq.Len())
	for elem := range toy.Elements(seq) {
		node, err := objectToYAML(elem)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return &yaml.Node{
		Kind:    yaml.SequenceNode,
		Content: nodes,
	}, nil
}

func mappingToYAML(mapping toy.Mapping) (_ *yaml.Node, err error) {
	var (
		nodeStyle yaml.Style
		nodes     []*yaml.Node
	)
	for key, value := range toy.Entries(mapping) {
		keyStr, ok := key.(toy.String)
		if !ok {
			return nil, fmt.Errorf("unsupported key type: %s", key.TypeName())
		}
		if keyStr == "_yaml" {
			switch x := value.(type) {
			case toy.String:
				nodeStyle, err = parseStyle(string(x))
				if err != nil {
					return nil, fmt.Errorf("%s: %w", string(keyStr), err)
				}
			case toy.Sequence:
				for i, elem := range toy.Entries(x) {
					style, ok := elem.(toy.String)
					if !ok {
						return nil, fmt.Errorf("%s[%d]: want 'string', got '%s'", string(keyStr), i, elem.TypeName())
					}
					tmp, err := parseStyle(string(style))
					if err != nil {
						return nil, fmt.Errorf("%s[%d]: %w", string(keyStr), i, err)
					}
					nodeStyle |= tmp
				}
			default:
				return nil, fmt.Errorf("%s: want 'string or sequence', got '%s'", string(keyStr), value.TypeName())
			}
			continue
		}
		node, err := objectToYAML(value)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", string(keyStr), err)
		}
		nodes = append(nodes,
			&yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: string(keyStr),
			},
			node,
		)
	}
	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Style:   nodeStyle,
		Content: nodes,
	}, nil
}

func parseStyle(s string) (yaml.Style, error) {
	switch s {
	case "tagged":
		return yaml.TaggedStyle, nil
	case "double-quoted":
		return yaml.DoubleQuotedStyle, nil
	case "single-quoted":
		return yaml.SingleQuotedStyle, nil
	case "literal":
		return yaml.LiteralStyle, nil
	case "folded":
		return yaml.FoldedStyle, nil
	case "flow":
		return yaml.FlowStyle, nil
	}
	return 0, fmt.Errorf("invalid yaml style: %s", s)
}

func objectToYAML(o toy.Object) (*yaml.Node, error) {
	switch x := o.(type) {
	case yaml.Marshaler:
		data, err := x.MarshalYAML()
		if err != nil {
			return nil, err
		}
		node, ok := data.(*yaml.Node)
		if !ok {
			return nil, fmt.Errorf("%s.MarshalYAML: got %s, want *yaml.Node",
				reflect.TypeOf(o), reflect.TypeOf(data))
		}
		return node, nil
	case toy.Int:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!int",
			Value: strconv.FormatInt(int64(x), 10),
		}, nil
	case toy.Float:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!float",
			Value: strconv.FormatFloat(float64(x), 'g', -1, 64),
		}, nil
	case toy.String:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: string(x),
		}, nil
	case toy.Bytes:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!binary",
			Value: base64.StdEncoding.EncodeToString(x),
		}, nil
	case Time:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!timestamp",
			Value: time.Time(x).Format(time.RFC3339Nano),
		}, nil
	case Duration:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: x.String(),
		}, nil
	case toy.Bool:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!bool",
			Value: strconv.FormatBool(bool(x)),
		}, nil
	case toy.NilType:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!null",
			Value: "null",
		}, nil
	case toy.Mapping:
		return mappingToYAML(x)
	case toy.Sequence:
		return sequenceToYAML(x)
	default:
		return nil, fmt.Errorf("'%s' can't be encoded in yaml", o.TypeName())
	}
}

func yamlEncode(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		x      toy.Object
		indent = 2
	)
	if err := toy.UnpackArgs(args, "x", &x, "indent?", &indent); err != nil {
		return nil, err
	}

	node, err := objectToYAML(x)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, err
	}

	var buf bytes.Buffer

	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(indent)

	if err := encoder.Encode(node); err != nil {
		return nil, err
	}

	return toy.Tuple{toy.Bytes(buf.Bytes()), toy.Nil}, nil
}

func yamlSequenceToArray(seq *yaml.Node) (*toy.Array, error) {
	elems := make([]toy.Object, 0, len(seq.Content))
	for _, node := range seq.Content {
		value, err := yamlToObject(node)
		if err != nil {
			return nil, err
		}
		elems = append(elems, value)
	}
	return toy.NewArray(elems), nil
}

func yamlMappingToMap(mapping *yaml.Node) (*toy.Map, error) {
	m := toy.NewMap(len(mapping.Content) / 2)
	for i := 0; i < len(mapping.Content); i += 2 {
		value, err := yamlToObject(mapping.Content[i+1])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", mapping.Content[i].Value, err)
		}
		m.IndexSet(toy.String(mapping.Content[i].Value), value)
	}
	return m, nil
}

func yamlToObject(node *yaml.Node) (toy.Object, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		switch node.Tag {
		case "!!int":
			i, _ := strconv.ParseInt(node.Value, 10, 64)
			return toy.Int(i), nil
		case "!!float":
			f, _ := strconv.ParseFloat(node.Value, 64)
			return toy.Float(f), nil
		case "!!str":
			return toy.String(node.Value), nil
		case "!!binary":
			data, err := base64.StdEncoding.DecodeString(node.Value)
			if err != nil {
				return nil, fmt.Errorf("!!binary value contains invalid base64 data: %w", err)
			}
			return toy.Bytes(data), nil
		case "!!timestamp":
			t, _ := time.Parse(time.RFC3339Nano, node.Value)
			return Time(t), nil
		case "!!bool":
			b, _ := strconv.ParseBool(node.Value)
			return toy.Bool(b), nil
		case "!!null":
			return toy.Nil, nil
		default:
			return nil, fmt.Errorf("value with tag %q can't be decoded", node.Tag)
		}
	case yaml.SequenceNode:
		return yamlSequenceToArray(node)
	case yaml.MappingNode:
		return yamlMappingToMap(node)
	}
	return nil, fmt.Errorf("value with kind %d can't be decoded", node.Kind)
}

func yamlDecode(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var data toy.StringOrBytes
	if err := toy.UnpackArgs(args, "data", &data); err != nil {
		return nil, err
	}

	node := new(yaml.Node)
	if err := yaml.Unmarshal(data.Bytes(), node); err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, err
	}

	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) != 1 {
			return toy.Tuple{toy.Nil, toy.NewError("invalid yaml document")}, nil
		}
		node = node.Content[0]
	case yaml.AliasNode:
		node = node.Alias
	}

	obj, err := yamlToObject(node)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, err
	}

	return toy.Tuple{obj, toy.Nil}, nil
}
