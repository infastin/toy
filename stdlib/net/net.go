package net

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"iter"
	"net"
	"slices"

	"github.com/infastin/toy"
	"github.com/infastin/toy/hash"
	"github.com/infastin/toy/internal/fndef"
	"github.com/infastin/toy/internal/xiter"
	"github.com/infastin/toy/stdlib/enum"
	"github.com/infastin/toy/token"
)

var Module = &toy.BuiltinModule{
	Name: "net",
	Members: map[string]toy.Value{
		"IP":       IPType,
		"parseIP":  toy.NewBuiltinFunction("parseIP", parseIPFn),
		"ipv4":     toy.NewBuiltinFunction("ipv4", ipv4Fn),
		"lookupIP": toy.NewBuiltinFunction("lookupIP", lookupIPFn),

		"IPMask":   IPMaskType,
		"cidrMask": toy.NewBuiltinFunction("cidrMask", cidrMaskFn),
		"ipv4Mask": toy.NewBuiltinFunction("ipv4Mask", ipv4MaskFn),

		"IPNet":     IPNetType,
		"parseCIDR": toy.NewBuiltinFunction("parseCIDR", parseCIDRFn),

		"MAC":      MACType,
		"parseMAC": toy.NewBuiltinFunction("parseMAC", parseMACFn),

		"Addr":           AddrType,
		"interfaceAddrs": toy.NewBuiltinFunction("interfaceAddrs", interfaceAddrsFn),

		"IPAddr":        IPAddrType,
		"resolveIPAddr": toy.NewBuiltinFunction("resolveIPAddr", resolveIPAddrFn),

		"Interface":       InterfaceType,
		"InterfaceFlags":  InterfaceFlagsType,
		"lookupInterface": toy.NewBuiltinFunction("lookupInterface", lookupInterfaceFn),
		"interfaces":      toy.NewBuiltinFunction("interfaces", interfacesFn),

		"joinHostPort":  toy.NewBuiltinFunction("joinHostPort", fndef.ASSRS("host", "port", net.JoinHostPort)),
		"splitHostPort": toy.NewBuiltinFunction("splitHostPort", fndef.ASRSSE("hostport", net.SplitHostPort)),
		"lookupAddr":    toy.NewBuiltinFunction("lookupAddr", fndef.ASRSsE("addr", net.LookupAddr)),
		"lookupHost":    toy.NewBuiltinFunction("lookupHost", fndef.ASRSsE("host", net.LookupHost)),
	},
}

type IP net.IP

var IPType = toy.NewType[IP]("net.IP", func(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	switch x := args[0].(type) {
	case toy.String:
		ip := net.ParseIP(string(x))
		if ip == nil {
			return nil, errors.New("invalid ip")
		}
		return IP(ip), nil
	case toy.Bytes:
		if len(x) != 4 && len(x) != 16 {
			return nil, &toy.InvalidArgumentTypeError{
				Name: "value",
				Want: "bytes[4] or bytes[16]",
				Got:  fmt.Sprintf("bytes[%d]", len(x)),
			}
		}
		return IP(x), nil
	case toy.Sequence:
		if x.Len() != 4 && x.Len() != 16 {
			return nil, &toy.InvalidArgumentTypeError{
				Name: "value",
				Want: "sequence[int, 4] or sequence[int, 16]",
				Got:  fmt.Sprintf("sequence[object, %d]", x.Len()),
			}
		}
		ip := make(IP, x.Len())
		for i, v := range xiter.Enum(x.Elements()) {
			b, ok := v.(toy.Int)
			if !ok {
				return nil, &toy.InvalidArgumentTypeError{
					Name: "value",
					Sel:  fmt.Sprintf("[%d]", i),
					Want: "int",
					Got:  toy.TypeName(v),
				}
			}
			ip[i] = byte(b)
		}
		return ip, nil
	default:
		var ip IP
		if err := toy.Convert(&ip, x); err != nil {
			return nil, err
		}
		return ip, nil
	}
})

func (ip IP) Type() toy.ValueType { return IPType }
func (ip IP) String() string      { return fmt.Sprintf("net.IP(%q)", net.IP(ip).String()) }
func (ip IP) IsFalsy() bool       { return net.IP(ip).IsUnspecified() }
func (ip IP) Clone() toy.Value    { return slices.Clone(ip) }
func (ip IP) Hash() uint64        { return hash.Bytes(ip[:]) }

func (ip IP) Compare(op token.Token, rhs toy.Value) (bool, error) {
	y, ok := rhs.(IP)
	if !ok {
		return false, toy.ErrInvalidOperation
	}
	switch op {
	case token.Equal:
		return net.IP(ip).Equal(net.IP(y)), nil
	case token.NotEqual:
		return !net.IP(ip).Equal(net.IP(y)), nil
	}
	return false, toy.ErrInvalidOperation
}

func (ip IP) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	method, ok := ipMethods[string(keyStr)]
	if !ok {
		return toy.Nil, false, nil
	}
	return method.WithReceiver(ip), true, nil
}

func (ip IP) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String(net.IP(ip).String())
	case *toy.Bytes:
		*p = toy.Bytes(ip)
	case **toy.Array:
		elems := make([]toy.Value, len(ip))
		for i, b := range ip {
			elems[i] = toy.Int(b)
		}
		*p = toy.NewArray(elems)
	case *toy.Tuple:
		tup := make(toy.Tuple, len(ip))
		for i, b := range ip {
			tup[i] = toy.Int(b)
		}
		*p = tup
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func (ip IP) Len() int           { return len(ip) }
func (ip IP) At(i int) toy.Value { return toy.Int(ip[i]) }

func (ip IP) Items() []toy.Value {
	elems := make([]toy.Value, len(ip))
	for i, b := range ip {
		elems[i] = toy.Int(b)
	}
	return elems
}

func (ip IP) Iterate() iter.Seq[toy.Value] {
	return func(yield func(toy.Value) bool) {
		for _, b := range ip {
			if !yield(toy.Int(b)) {
				break
			}
		}
	}
}

var ipMethods = map[string]*toy.BuiltinFunction{
	"isUnspecified":        toy.NewBuiltinFunction("isUnspecified", ipIsUnspecifiedMd),
	"isLoopback":           toy.NewBuiltinFunction("isLoopback", ipIsLoopbackMd),
	"isPrivate":            toy.NewBuiltinFunction("isPrivate", ipIsPrivateMd),
	"isMulticast":          toy.NewBuiltinFunction("isMulticast", ipIsMulticastMd),
	"isLinkLocalMulticast": toy.NewBuiltinFunction("isLinkLocalMulticast", ipIsLinkLocalMulticastMd),
	"isLinkLocalUnicast":   toy.NewBuiltinFunction("isLinkLocalUnicast", ipIsLinkLocalUnicastMd),
	"isGlobalUnicast":      toy.NewBuiltinFunction("isGlobalUnicast", ipIsGlobalUnicastMd),
	"to4":                  toy.NewBuiltinFunction("to4", ipTo4Md),
	"to16":                 toy.NewBuiltinFunction("to16", ipTo16Md),
	"mask":                 toy.NewBuiltinFunction("mask", ipMaskMd),
	"defaultMask":          toy.NewBuiltinFunction("defaultMask", ipDefaultMaskMd),
}

func ipIsUnspecifiedMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsUnspecified()), nil
}

func ipIsLoopbackMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsLoopback()), nil
}

func ipIsPrivateMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsPrivate()), nil
}

func ipIsMulticastMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsMulticast()), nil
}

func ipIsLinkLocalMulticastMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsLinkLocalMulticast()), nil
}

func ipIsLinkLocalUnicastMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsLinkLocalUnicast()), nil
}

func ipIsGlobalUnicastMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsGlobalUnicast()), nil
}

func ipTo4Md(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return IP(net.IP(recv).To4()), nil
}

func ipTo16Md(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return IP(net.IP(recv).To16()), nil
}

func ipMaskMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var (
		recv = args[0].(IP)
		mask IPMask
	)
	if err := toy.UnpackArgs(args[1:], "mask", &mask); err != nil {
		return nil, err
	}
	return IP(net.IP(recv).Mask(net.IPMask(mask))), nil
}

func ipDefaultMaskMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return IP(net.IP(recv).DefaultMask()), nil
}

func parseIPFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var s string
	if err := toy.UnpackArgs(args, "s", &s); err != nil {
		return nil, err
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, errors.New("invalid ip")
	}
	return IP(ip), nil
}

func ipv4Fn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var a, b, c, d byte
	if err := toy.UnpackArgs(args, "a", &a, "b", &b, "c", &c, "d", &d); err != nil {
		return nil, err
	}
	return IP(net.IPv4(a, b, c, d)), nil
}

func lookupIPFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var host string
	if err := toy.UnpackArgs(args, "host", &host); err != nil {
		return nil, err
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	elems := make([]toy.Value, 0, len(ips))
	for _, ip := range ips {
		elems = append(elems, IP(ip))
	}
	return toy.NewArray(elems), nil
}

type IPMask net.IPMask

var IPMaskType = toy.NewType[IPMask]("net.IPMask", func(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) < 1 && len(args) > 2 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 2,
			Got:     len(args),
		}
	}
	if len(args) == 1 {
		switch x := args[0].(type) {
		case toy.String:
			if len(x) != 8 && len(x) != 32 {
				return nil, &toy.InvalidArgumentTypeError{
					Name: "value",
					Want: "string[8] or string[32]",
					Got:  fmt.Sprintf("string[%d]", len(x)),
				}
			}
			b, err := hex.DecodeString(string(x))
			if err != nil {
				return nil, err
			}
			return IPMask(b), nil
		case toy.Bytes:
			if len(x) != 4 && len(x) != 16 {
				return nil, &toy.InvalidArgumentTypeError{
					Name: "value",
					Want: "bytes[4] or bytes[16]",
					Got:  fmt.Sprintf("bytes[%d]", len(x)),
				}
			}
			return IPMask(x), nil
		case toy.Sequence:
			if x.Len() != 4 && x.Len() != 16 {
				return nil, &toy.InvalidArgumentTypeError{
					Name: "value",
					Want: "sequence[int, 4] or sequence[int, 16]",
					Got:  fmt.Sprintf("sequence[object, %d]", x.Len()),
				}
			}
			m := make(IPMask, x.Len())
			for i, v := range xiter.Enum(x.Elements()) {
				b, ok := v.(toy.Int)
				if !ok {
					return nil, &toy.InvalidArgumentTypeError{
						Name: "value",
						Sel:  fmt.Sprintf("[%d]", i),
						Want: "int",
						Got:  toy.TypeName(v),
					}
				}
				m[i] = byte(b)
			}
			return m, nil
		default:
			var m IPMask
			if err := toy.Convert(&m, x); err != nil {
				return nil, err
			}
			return m, nil
		}
	}
	var ones, bits int
	if err := toy.UnpackArgs(args, "ones", &ones, "bits", &bits); err != nil {
		return nil, err
	}
	mask := net.CIDRMask(ones, bits)
	if mask == nil {
		return nil, errors.New("invalid number of ones and bits")
	}
	return IPMask(mask), nil
})

func (m IPMask) Type() toy.ValueType { return IPMaskType }
func (m IPMask) String() string      { return fmt.Sprintf("net.IPMask(%q)", net.IPMask(m).String()) }
func (m IPMask) IsFalsy() bool       { return len(m) == 0 }
func (m IPMask) Clone() toy.Value    { return slices.Clone(m) }
func (m IPMask) Hash() uint64        { return hash.Bytes(m[:]) }

func (m IPMask) Compare(op token.Token, rhs toy.Value) (bool, error) {
	y, ok := rhs.(IPMask)
	if !ok {
		return false, toy.ErrInvalidOperation
	}
	switch op {
	case token.Equal:
		return bytes.Equal(m, y), nil
	case token.NotEqual:
		return !bytes.Equal(m, y), nil
	}
	return false, toy.ErrInvalidOperation
}

func (m IPMask) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	method, ok := ipMaskMethods[string(keyStr)]
	if !ok {
		return toy.Nil, false, nil
	}
	return method.WithReceiver(m), true, nil
}

func (m IPMask) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String(net.IPMask(m).String())
	case *toy.Bytes:
		*p = toy.Bytes(m)
	case **toy.Array:
		elems := make([]toy.Value, len(m))
		for i, b := range m {
			elems[i] = toy.Int(b)
		}
		*p = toy.NewArray(elems)
	case *toy.Tuple:
		tup := make(toy.Tuple, len(m))
		for i, b := range m {
			tup[i] = toy.Int(b)
		}
		*p = tup
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func (m IPMask) Len() int           { return len(m) }
func (m IPMask) At(i int) toy.Value { return toy.Int(m[i]) }

func (m IPMask) Items() []toy.Value {
	elems := make([]toy.Value, len(m))
	for i, b := range m {
		elems[i] = toy.Int(b)
	}
	return elems
}

func (m IPMask) Iterate() iter.Seq[toy.Value] {
	return func(yield func(toy.Value) bool) {
		for _, b := range m {
			if !yield(toy.Int(b)) {
				break
			}
		}
	}
}

var ipMaskMethods = map[string]*toy.BuiltinFunction{
	"size": toy.NewBuiltinFunction("size", ipMaskSizeMd),
}

func ipMaskSizeMd(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	recv := args[0].(IPMask)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	ones, bits := net.IPMask(recv).Size()
	return toy.Tuple{toy.Int(ones), toy.Int(bits)}, nil
}

func cidrMaskFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var ones, bits int
	if err := toy.UnpackArgs(args, "ones", &ones, "bits", &bits); err != nil {
		return nil, err
	}
	mask := net.CIDRMask(ones, bits)
	if mask == nil {
		return nil, errors.New("invalid number of ones and bits")
	}
	return IPMask(mask), nil
}

func ipv4MaskFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var a, b, c, d byte
	if err := toy.UnpackArgs(args, "a", &a, "b", &b, "c", &c, "d", &d); err != nil {
		return nil, err
	}
	return IPMask(net.IPv4Mask(a, b, c, d)), nil
}

func parseCIDRFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var s string
	if err := toy.UnpackArgs(args, "s", &s); err != nil {
		return nil, err
	}
	ip, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	return toy.Tuple{IP(ip), (*IPNet)(ipNet)}, nil
}

type IPNet net.IPNet

var IPNetType = toy.NewType[*IPNet]("net.IPNet", func(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) < 1 && len(args) > 2 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 2,
			Got:     len(args),
		}
	}
	if len(args) == 1 {
		switch x := args[0].(type) {
		case toy.String:
			_, ipNet, err := net.ParseCIDR(string(x))
			if err != nil {
				return nil, err
			}
			return (*IPNet)(ipNet), nil
		default:
			var n *IPNet
			if err := toy.Convert(&n, x); err != nil {
				return nil, err
			}
			return n, nil
		}
	}
	var (
		ip   IP
		mask IPMask
	)
	if err := toy.UnpackArgs(args, "ip", &ip, "mask", &mask); err != nil {
		return nil, err
	}
	return &IPNet{IP: net.IP(ip), Mask: net.IPMask(mask)}, nil
})

func (n *IPNet) Type() toy.ValueType { return IPNetType }
func (n *IPNet) String() string      { return fmt.Sprintf("net.IPNet(%q)", (*net.IPNet)(n).String()) }
func (n *IPNet) IsFalsy() bool       { return false }
func (n *IPNet) Clone() toy.Value    { return &IPNet{IP: n.IP, Mask: n.Mask} }

func (n *IPNet) Compare(op token.Token, rhs toy.Value) (bool, error) {
	y, ok := rhs.(*IPNet)
	if !ok {
		return false, toy.ErrInvalidOperation
	}
	switch op {
	case token.Equal:
		return net.IP(n.IP).Equal(net.IP(y.IP)) && bytes.Equal(n.Mask, y.Mask), nil
	case token.NotEqual:
		return !net.IP(n.IP).Equal(net.IP(y.IP)) || !bytes.Equal(n.Mask, y.Mask), nil
	}
	return false, toy.ErrInvalidOperation
}

func (n *IPNet) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	switch string(keyStr) {
	case "ip":
		return IP(n.IP), true, nil
	case "mask":
		return IPMask(n.Mask), true, nil
	}
	return toy.Nil, false, nil
}

func (n *IPNet) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String((*net.IPNet)(n).String())
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func (n *IPNet) Contains(value toy.Value) (bool, error) {
	ip, ok := value.(IP)
	if !ok {
		return false, &toy.InvalidValueTypeError{
			Want: "net.IP",
			Got:  toy.TypeName(value),
		}
	}
	return (*net.IPNet)(n).Contains(net.IP(ip)), nil
}

type MAC net.HardwareAddr

var MACType = toy.NewType[MAC]("net.MAC", func(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	switch x := args[0].(type) {
	case toy.String:
		m, err := net.ParseMAC(string(x))
		if err != nil {
			return nil, err
		}
		return MAC(m), nil
	case toy.Bytes:
		if len(x) != 6 && len(x) != 8 && len(x) != 20 {
			return nil, &toy.InvalidArgumentTypeError{
				Name: "value",
				Want: "bytes[6], bytes[8] or bytes[20]",
				Got:  fmt.Sprintf("bytes[%d]", len(x)),
			}
		}
		return MAC(x), nil
	case toy.Sequence:
		if x.Len() != 6 && x.Len() != 8 && x.Len() != 20 {
			return nil, &toy.InvalidArgumentTypeError{
				Name: "value",
				Want: "sequence[int, 6], sequence[int, 8] or sequence[int, 20]",
				Got:  fmt.Sprintf("sequence[object, %d]", x.Len()),
			}
		}
		m := make(MAC, x.Len())
		for i, v := range xiter.Enum(x.Elements()) {
			b, ok := v.(toy.Int)
			if !ok {
				return nil, &toy.InvalidArgumentTypeError{
					Name: "value",
					Sel:  fmt.Sprintf("[%d]", i),
					Want: "int",
					Got:  toy.TypeName(v),
				}
			}
			m[i] = byte(b)
		}
		return m, nil
	default:
		var m MAC
		if err := toy.Convert(&m, x); err != nil {
			return nil, err
		}
		return m, nil
	}
})

func (m MAC) Type() toy.ValueType { return MACType }
func (m MAC) String() string      { return fmt.Sprintf("net.MAC(%q)", net.HardwareAddr(m).String()) }
func (m MAC) IsFalsy() bool       { return len(m) == 0 }
func (m MAC) Clone() toy.Value    { return slices.Clone(m) }
func (m MAC) Hash() uint64        { return hash.Bytes(m[:]) }

func (m MAC) Compare(op token.Token, rhs toy.Value) (bool, error) {
	y, ok := rhs.(MAC)
	if !ok {
		return false, toy.ErrInvalidOperation
	}
	switch op {
	case token.Equal:
		return bytes.Equal(m, y), nil
	case token.NotEqual:
		return !bytes.Equal(m, y), nil
	}
	return false, toy.ErrInvalidOperation
}

func (m MAC) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String(net.IP(m).String())
	case *toy.Bytes:
		*p = toy.Bytes(m)
	case **toy.Array:
		elems := make([]toy.Value, len(m))
		for i, b := range m {
			elems[i] = toy.Int(b)
		}
		*p = toy.NewArray(elems)
	case *toy.Tuple:
		tup := make(toy.Tuple, len(m))
		for i, b := range m {
			tup[i] = toy.Int(b)
		}
		*p = tup
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func (m MAC) Len() int           { return len(m) }
func (m MAC) At(i int) toy.Value { return toy.Int(m[i]) }

func (m MAC) Items() []toy.Value {
	elems := make([]toy.Value, len(m))
	for i, b := range m {
		elems[i] = toy.Int(b)
	}
	return elems
}

func (m MAC) Iterate() iter.Seq[toy.Value] {
	return func(yield func(toy.Value) bool) {
		for _, b := range m {
			if !yield(toy.Int(b)) {
				break
			}
		}
	}
}

func parseMACFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var s string
	if err := toy.UnpackArgs(args, "s", &s); err != nil {
		return nil, err
	}
	mac, err := net.ParseMAC(s)
	if err != nil {
		return nil, err
	}
	return MAC(mac), nil
}

type Addr struct {
	a net.Addr
}

var AddrType = toy.NewType[*Addr]("net.Addr", nil)

func (a *Addr) Type() toy.ValueType { return AddrType }
func (a *Addr) String() string      { return fmt.Sprintf("net.Addr(%q, %q)", a.a.Network(), a.a.String()) }
func (a *Addr) IsFalsy() bool       { return false }
func (a *Addr) Clone() toy.Value    { return &Addr{a: a.a} }

func (a *Addr) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	switch string(keyStr) {
	case "network":
		return toy.String(a.a.Network()), true, nil
	case "address":
		return toy.String(a.a.String()), true, nil
	}
	return toy.Nil, false, nil
}

func interfaceAddrsFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	elems := make([]toy.Value, 0, len(addrs))
	for _, addr := range addrs {
		elems = append(elems, &Addr{a: addr})
	}
	return toy.NewArray(elems), nil
}

type IPAddr net.IPAddr

var IPAddrType = toy.NewType[*IPAddr]("net.IPAddr", nil)

func (a *IPAddr) Type() toy.ValueType { return IPAddrType }
func (a *IPAddr) String() string      { return fmt.Sprintf("net.IPAddr(%q)", (*net.IPAddr)(a).String()) }
func (a *IPAddr) IsFalsy() bool       { return false }
func (a *IPAddr) Clone() toy.Value    { return &IPAddr{IP: a.IP, Zone: a.Zone} }

func (a *IPAddr) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	switch string(keyStr) {
	case "ip":
		return IP(a.IP), true, nil
	case "zone":
		return toy.String(a.Zone), true, nil
	}
	return toy.Nil, false, nil
}

func resolveIPAddrFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	var network, address string
	if err := toy.UnpackArgs(args, "network", &network, "address", &address); err != nil {
		return nil, err
	}
	addr, err := net.ResolveIPAddr(network, address)
	if err != nil {
		return nil, err
	}
	return (*IPAddr)(addr), nil
}

type Interface net.Interface

var InterfaceType = toy.NewType[*Interface]("net.Interface", nil)

func (i *Interface) Type() toy.ValueType { return InterfaceType }

func (i *Interface) String() string {
	return fmt.Sprintf("net.Interface(%q)", (*net.Interface)(i).Name)
}

func (i *Interface) IsFalsy() bool { return false }

func (i *Interface) Clone() toy.Value {
	c := new(Interface)
	*c = *i
	return c
}

func (i *Interface) Property(key toy.Value) (value toy.Value, found bool, err error) {
	keyStr, ok := key.(toy.String)
	if !ok {
		return nil, false, &toy.InvalidKeyTypeError{
			Want: "string",
			Got:  toy.TypeName(key),
		}
	}
	switch string(keyStr) {
	case "index":
		return toy.Int(i.Index), true, nil
	case "mtu":
		return toy.Int(i.MTU), true, nil
	case "name":
		return toy.String(i.Name), true, nil
	case "mac":
		return MAC(i.HardwareAddr), true, nil
	case "flags":
		return InterfaceFlags(i.Flags), true, nil
	}
	return toy.Nil, false, nil
}

type InterfaceFlags net.Flags

var InterfaceFlagsType = enum.New("net.InterfaceFlags", map[string]InterfaceFlags{
	"UP":             InterfaceFlags(net.FlagUp),
	"BROADCAST":      InterfaceFlags(net.FlagBroadcast),
	"LOOPBACK":       InterfaceFlags(net.FlagLoopback),
	"POINT_TO_POINT": InterfaceFlags(net.FlagPointToPoint),
	"MULTICAST":      InterfaceFlags(net.FlagMulticast),
	"RUNNING":        InterfaceFlags(net.FlagRunning),
}, func(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	switch x := args[0].(type) {
	case toy.Int:
		return InterfaceFlags(x), nil
	default:
		var f InterfaceFlags
		if err := toy.Convert(&f, x); err != nil {
			return nil, err
		}
		return f, nil
	}
})

func (f InterfaceFlags) Type() toy.ValueType { return InterfaceFlagsType }

func (f InterfaceFlags) String() string {
	return fmt.Sprintf("net.InterfaceFlags(%q)", net.Flags(f).String())
}

func (f InterfaceFlags) IsFalsy() bool    { return false }
func (f InterfaceFlags) Clone() toy.Value { return f }

func (f InterfaceFlags) Convert(p any) error {
	switch p := p.(type) {
	case *toy.Int:
		*p = toy.Int(f)
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func (f InterfaceFlags) BinaryOp(op token.Token, other toy.Value, right bool) (toy.Value, error) {
	switch y := other.(type) {
	case InterfaceFlags:
		switch op {
		case token.And:
			return f & y, nil
		case token.Or:
			return f | y, nil
		case token.Xor:
			return f ^ y, nil
		case token.AndNot:
			return f &^ y, nil
		}
	case toy.Int:
		switch op {
		case token.And:
			return f & InterfaceFlags(y), nil
		case token.Or:
			return f | InterfaceFlags(y), nil
		case token.Xor:
			return f ^ InterfaceFlags(y), nil
		case token.AndNot:
			return f &^ InterfaceFlags(y), nil
		case token.Shl:
			if !right {
				return f << y, nil
			}
		case token.Shr:
			if !right {
				return f >> y, nil
			}
		}
	}
	return nil, toy.ErrInvalidOperation
}

func (f InterfaceFlags) UnaryOp(op token.Token) (toy.Value, error) {
	switch op {
	case token.Xor:
		return ^f, nil
	}
	return nil, toy.ErrInvalidOperation
}

func lookupInterfaceFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 1 {
		return nil, &toy.WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	switch x := args[0].(type) {
	case toy.String:
		iface, err := net.InterfaceByName(string(x))
		if err != nil {
			return nil, err
		}
		return (*Interface)(iface), nil
	case toy.Int:
		iface, err := net.InterfaceByIndex(int(x))
		if err != nil {
			return nil, err
		}
		return (*Interface)(iface), nil
	default:
		return nil, &toy.InvalidArgumentTypeError{
			Name: "id",
			Want: "string or int",
			Got:  toy.TypeName(x),
		}
	}
}

func interfacesFn(_ *toy.Runtime, args ...toy.Value) (toy.Value, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	elems := make([]toy.Value, 0, len(ifaces))
	for i := range ifaces {
		iface := (*Interface)(&ifaces[i])
		elems = append(elems, iface)
	}
	return toy.NewArray(elems), nil
}
