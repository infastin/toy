package stdlib

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"slices"

	"github.com/infastin/toy"
	"github.com/infastin/toy/hash"
	"github.com/infastin/toy/token"
)

var NetModule = &toy.BuiltinModule{
	Name: "net",
	Members: map[string]toy.Object{
		"IP":       IPType,
		"parseIP":  toy.NewBuiltinFunction("parseIP", netParseIP),
		"ipv4":     toy.NewBuiltinFunction("ipv4", netIPv4),
		"lookupIP": toy.NewBuiltinFunction("lookupIP", netLookupIP),

		"IPMask":   IPMaskType,
		"cidrMask": toy.NewBuiltinFunction("cidrMask", netCIDRMask),
		"ipv4Mask": toy.NewBuiltinFunction("ipv4Mask", netIPv4Mask),

		"IPNet":     IPNetType,
		"parseCIDR": toy.NewBuiltinFunction("parseCIDR", netParseCIDR),

		"MAC":      MACType,
		"parseMAC": toy.NewBuiltinFunction("parseMAC", netParseMAC),

		"Addr":           AddrType,
		"interfaceAddrs": toy.NewBuiltinFunction("interfaceAddrs", netInterfaceAddrs),

		"IPAddr":        IPAddrType,
		"resolveIPAddr": toy.NewBuiltinFunction("resolveIPAddr", netResolveIPAddr),

		"Interface":       InterfaceType,
		"InterfaceFlags":  InterfaceFlagsType,
		"lookupInterface": toy.NewBuiltinFunction("lookupInterface", netLookupInterface),
		"interfaces":      toy.NewBuiltinFunction("interfaces", netInterfaces),

		"joinHostPort":  toy.NewBuiltinFunction("joinHostPort", makeASSRS("host", "port", net.JoinHostPort)),
		"splitHostPort": toy.NewBuiltinFunction("splitHostPort", makeASRSSE("hostport", net.SplitHostPort)),
		"lookupAddr":    toy.NewBuiltinFunction("lookupAddr", makeASRSsE("addr", net.LookupAddr)),
		"lookupHost":    toy.NewBuiltinFunction("lookupHost", makeASRSsE("host", net.LookupHost)),
	},
}

type IP net.IP

var IPType = toy.NewType[IP]("net.IP", func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
		i := 0
		for v := range toy.Elements(x) {
			b, ok := v.(toy.Int)
			if !ok {
				return nil, fmt.Errorf("value[%d]: want 'int', got '%s'", i, toy.TypeName(v))
			}
			ip[i] = byte(b)
			i++
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

func (ip IP) Type() toy.ObjectType { return IPType }
func (ip IP) String() string       { return fmt.Sprintf("net.IP(%q)", net.IP(ip).String()) }
func (ip IP) IsFalsy() bool        { return net.IP(ip).IsUnspecified() }
func (ip IP) Clone() toy.Object    { return slices.Clone(ip) }
func (ip IP) Hash() uint64         { return hash.Bytes(ip[:]) }

func (ip IP) Compare(op token.Token, rhs toy.Object) (bool, error) {
	y, ok := rhs.(IP)
	if !ok {
		return false, toy.ErrInvalidOperator
	}
	switch op {
	case token.Equal:
		return net.IP(ip).Equal(net.IP(y)), nil
	case token.NotEqual:
		return !net.IP(ip).Equal(net.IP(y)), nil
	}
	return false, toy.ErrInvalidOperator
}

func (ip IP) FieldGet(name string) (toy.Object, error) {
	method, ok := netIPMethods[name]
	if !ok {
		return nil, toy.ErrNoSuchField
	}
	return method.WithReceiver(ip), nil
}

func (ip IP) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String(net.IP(ip).String())
	case *toy.Bytes:
		*p = toy.Bytes(ip)
	case **toy.Array:
		elems := make([]toy.Object, len(ip))
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

func (ip IP) Len() int            { return len(ip) }
func (ip IP) At(i int) toy.Object { return toy.Int(ip[i]) }

func (ip IP) Items() []toy.Object {
	elems := make([]toy.Object, len(ip))
	for i, b := range ip {
		elems[i] = toy.Int(b)
	}
	return elems
}

func (ip IP) Iterate() toy.Iterator { return &ipIterator{ip: ip, i: 0} }

type ipIterator struct {
	ip IP
	i  int
}

var ipIteratorType = toy.NewType[*ipIterator]("net.IP-iterator", nil)

func (it *ipIterator) Type() toy.ObjectType { return ipIteratorType }
func (it *ipIterator) String() string       { return "<net.IP-iterator>" }
func (it *ipIterator) IsFalsy() bool        { return true }
func (it *ipIterator) Clone() toy.Object    { return &ipIterator{ip: it.ip, i: it.i} }

func (it *ipIterator) Next(key, value *toy.Object) bool {
	if it.i < len(it.ip) {
		if key != nil {
			*key = toy.Int(it.i)
		}
		if value != nil {
			*value = toy.Int(it.ip[it.i])
		}
		it.i++
		return true
	}
	return false
}

var netIPMethods = map[string]*toy.BuiltinFunction{
	"isUnspecified":        toy.NewBuiltinFunction("isUnspecified", netIPIsUnspecified),
	"isLoopback":           toy.NewBuiltinFunction("isLoopback", netIPIsLoopback),
	"isPrivate":            toy.NewBuiltinFunction("isPrivate", netIPIsPrivate),
	"isMulticast":          toy.NewBuiltinFunction("isMulticast", netIPIsMulticast),
	"isLinkLocalMulticast": toy.NewBuiltinFunction("isLinkLocalMulticast", netIPIsLinkLocalMulticast),
	"isLinkLocalUnicast":   toy.NewBuiltinFunction("isLinkLocalUnicast", netIPIsLinkLocalUnicast),
	"isGlobalUnicast":      toy.NewBuiltinFunction("isGlobalUnicast", netIPIsGlobalUnicast),
	"to4":                  toy.NewBuiltinFunction("to4", netIPTo4),
	"to16":                 toy.NewBuiltinFunction("to16", netIPTo16),
	"mask":                 toy.NewBuiltinFunction("mask", netIPMask),
	"defaultMask":          toy.NewBuiltinFunction("defaultMask", netIPDefaultMask),
}

func netIPIsUnspecified(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsUnspecified()), nil
}

func netIPIsLoopback(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsLoopback()), nil
}

func netIPIsPrivate(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsPrivate()), nil
}

func netIPIsMulticast(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsMulticast()), nil
}

func netIPIsLinkLocalMulticast(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsLinkLocalMulticast()), nil
}

func netIPIsLinkLocalUnicast(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsLinkLocalUnicast()), nil
}

func netIPIsGlobalUnicast(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return toy.Bool(net.IP(recv).IsGlobalUnicast()), nil
}

func netIPTo4(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return IP(net.IP(recv).To4()), nil
}

func netIPTo16(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return IP(net.IP(recv).To16()), nil
}

func netIPMask(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var (
		recv = args[0].(IP)
		mask IPMask
	)
	if err := toy.UnpackArgs(args[1:], "mask", &mask); err != nil {
		return nil, err
	}
	return IP(net.IP(recv).Mask(net.IPMask(mask))), nil
}

func netIPDefaultMask(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(IP)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	return IP(net.IP(recv).DefaultMask()), nil
}

func netParseIP(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s string
	if err := toy.UnpackArgs(args, "s", &s); err != nil {
		return nil, err
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return toy.Tuple{toy.Nil, toy.NewError("invalid ip")}, nil
	}
	return toy.Tuple{IP(ip), toy.Nil}, nil
}

func netIPv4(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var a, b, c, d byte
	if err := toy.UnpackArgs(args, "a", &a, "b", &b, "c", &c, "d", &d); err != nil {
		return nil, err
	}
	return IP(net.IPv4(a, b, c, d)), nil
}

func netLookupIP(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var host string
	if err := toy.UnpackArgs(args, "host", &host); err != nil {
		return nil, err
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	elems := make([]toy.Object, 0, len(ips))
	for _, ip := range ips {
		elems = append(elems, IP(ip))
	}
	return toy.Tuple{toy.NewArray(elems), toy.Nil}, nil
}

type IPMask net.IPMask

var IPMaskType = toy.NewType[IPMask]("net.IPMask", func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
			i := 0
			for v := range toy.Elements(x) {
				b, ok := v.(toy.Int)
				if !ok {
					return nil, fmt.Errorf("value[%d]: want 'int', got '%s'", i, toy.TypeName(v))
				}
				m[i] = byte(b)
				i++
			}
			return m, nil
		default:
			var m IPMask
			if err := toy.Convert(&m, args[0]); err != nil {
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

func (m IPMask) Type() toy.ObjectType { return IPMaskType }
func (m IPMask) String() string       { return fmt.Sprintf("net.IPMask(%q)", net.IPMask(m).String()) }
func (m IPMask) IsFalsy() bool        { return len(m) == 0 }
func (m IPMask) Clone() toy.Object    { return slices.Clone(m) }
func (m IPMask) Hash() uint64         { return hash.Bytes(m[:]) }

func (m IPMask) Compare(op token.Token, rhs toy.Object) (bool, error) {
	y, ok := rhs.(IPMask)
	if !ok {
		return false, toy.ErrInvalidOperator
	}
	switch op {
	case token.Equal:
		return bytes.Equal(m, y), nil
	case token.NotEqual:
		return !bytes.Equal(m, y), nil
	}
	return false, toy.ErrInvalidOperator
}

func (m IPMask) FieldGet(name string) (toy.Object, error) {
	method, ok := netIPMaskMethods[name]
	if !ok {
		return nil, toy.ErrNoSuchField
	}
	return method.WithReceiver(m), nil
}

func (m IPMask) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String(net.IPMask(m).String())
	case *toy.Bytes:
		*p = toy.Bytes(m)
	case **toy.Array:
		elems := make([]toy.Object, len(m))
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

func (m IPMask) Len() int            { return len(m) }
func (m IPMask) At(i int) toy.Object { return toy.Int(m[i]) }

func (m IPMask) Items() []toy.Object {
	elems := make([]toy.Object, len(m))
	for i, b := range m {
		elems[i] = toy.Int(b)
	}
	return elems
}

func (m IPMask) Iterate() toy.Iterator { return &ipMaskIterator{m: m, i: 0} }

type ipMaskIterator struct {
	m IPMask
	i int
}

var ipMaskIteratorType = toy.NewType[*ipMaskIterator]("net.IPMask-iterator", nil)

func (it *ipMaskIterator) Type() toy.ObjectType { return ipMaskIteratorType }
func (it *ipMaskIterator) String() string       { return "<net.IPMask-iterator>" }
func (it *ipMaskIterator) IsFalsy() bool        { return true }
func (it *ipMaskIterator) Clone() toy.Object    { return &ipMaskIterator{m: it.m, i: it.i} }

func (it *ipMaskIterator) Next(key, value *toy.Object) bool {
	if it.i < len(it.m) {
		if key != nil {
			*key = toy.Int(it.i)
		}
		if value != nil {
			*value = toy.Int(it.m[it.i])
		}
		it.i++
		return true
	}
	return false
}

var netIPMaskMethods = map[string]*toy.BuiltinFunction{
	"size": toy.NewBuiltinFunction("size", netIPMaskSize),
}

func netIPMaskSize(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	recv := args[0].(IPMask)
	args = args[1:]
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	ones, bits := net.IPMask(recv).Size()
	return toy.Tuple{toy.Int(ones), toy.Int(bits)}, nil
}

func netCIDRMask(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var ones, bits int
	if err := toy.UnpackArgs(args, "ones", &ones, "bits", &bits); err != nil {
		return nil, err
	}
	mask := net.CIDRMask(ones, bits)
	if mask == nil {
		return toy.Tuple{toy.Nil, toy.NewError("invalid number of ones and bits")}, nil
	}
	return toy.Tuple{IPMask(mask), toy.Nil}, nil
}

func netIPv4Mask(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var a, b, c, d byte
	if err := toy.UnpackArgs(args, "a", &a, "b", &b, "c", &c, "d", &d); err != nil {
		return nil, err
	}
	return IPMask(net.IPv4Mask(a, b, c, d)), nil
}

func netParseCIDR(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s string
	if err := toy.UnpackArgs(args, "s", &s); err != nil {
		return nil, err
	}
	ip, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{IP(ip), (*IPNet)(ipNet), toy.Nil}, nil
}

type IPNet net.IPNet

var IPNetType = toy.NewType[*IPNet]("net.IPNet", func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
			if err := toy.Convert(&n, args[0]); err != nil {
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

func (n *IPNet) Type() toy.ObjectType { return IPNetType }
func (n *IPNet) String() string       { return fmt.Sprintf("net.IPNet(%q)", (*net.IPNet)(n).String()) }
func (n *IPNet) IsFalsy() bool        { return false }
func (n *IPNet) Clone() toy.Object    { return &IPNet{IP: n.IP, Mask: n.Mask} }

func (n *IPNet) Compare(op token.Token, rhs toy.Object) (bool, error) {
	y, ok := rhs.(*IPNet)
	if !ok {
		return false, toy.ErrInvalidOperator
	}
	switch op {
	case token.Equal:
		return net.IP(n.IP).Equal(net.IP(y.IP)) && bytes.Equal(n.Mask, y.Mask), nil
	case token.NotEqual:
		return !net.IP(n.IP).Equal(net.IP(y.IP)) || !bytes.Equal(n.Mask, y.Mask), nil
	}
	return false, toy.ErrInvalidOperator
}

func (n *IPNet) FieldGet(name string) (toy.Object, error) {
	switch name {
	case "ip":
		return IP(n.IP), nil
	case "mask":
		return IPMask(n.Mask), nil
	}
	return nil, toy.ErrNoSuchField
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

func (n *IPNet) Contains(value toy.Object) (bool, error) {
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

var MACType = toy.NewType[MAC]("net.MAC", func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
		i := 0
		for v := range toy.Elements(x) {
			b, ok := v.(toy.Int)
			if !ok {
				return nil, fmt.Errorf("value[%d]: want 'int', got '%s'", i, toy.TypeName(v))
			}
			m[i] = byte(b)
			i++
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

func (m MAC) Type() toy.ObjectType { return MACType }
func (m MAC) String() string       { return fmt.Sprintf("net.MAC(%q)", net.HardwareAddr(m).String()) }
func (m MAC) IsFalsy() bool        { return len(m) == 0 }
func (m MAC) Clone() toy.Object    { return slices.Clone(m) }
func (m MAC) Hash() uint64         { return hash.Bytes(m[:]) }

func (m MAC) Compare(op token.Token, rhs toy.Object) (bool, error) {
	y, ok := rhs.(MAC)
	if !ok {
		return false, toy.ErrInvalidOperator
	}
	switch op {
	case token.Equal:
		return bytes.Equal(m, y), nil
	case token.NotEqual:
		return !bytes.Equal(m, y), nil
	}
	return false, toy.ErrInvalidOperator
}

func (m MAC) Convert(p any) error {
	switch p := p.(type) {
	case *toy.String:
		*p = toy.String(net.IP(m).String())
	case *toy.Bytes:
		*p = toy.Bytes(m)
	case **toy.Array:
		elems := make([]toy.Object, len(m))
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

func (m MAC) Len() int            { return len(m) }
func (m MAC) At(i int) toy.Object { return toy.Int(m[i]) }

func (m MAC) Items() []toy.Object {
	elems := make([]toy.Object, len(m))
	for i, b := range m {
		elems[i] = toy.Int(b)
	}
	return elems
}

func (m MAC) Iterate() toy.Iterator { return &macIterator{m: m, i: 0} }

type macIterator struct {
	m MAC
	i int
}

var macIteratorType = toy.NewType[*macIterator]("net.MAC-iterator", nil)

func (it *macIterator) Type() toy.ObjectType { return macIteratorType }
func (it *macIterator) String() string       { return "<net.MAC-iterator>" }
func (it *macIterator) IsFalsy() bool        { return true }
func (it *macIterator) Clone() toy.Object    { return &macIterator{m: it.m, i: it.i} }

func (it *macIterator) Next(key, value *toy.Object) bool {
	if it.i < len(it.m) {
		if key != nil {
			*key = toy.Int(it.i)
		}
		if value != nil {
			*value = toy.Int(it.m[it.i])
		}
		it.i++
		return true
	}
	return false
}

func netParseMAC(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var s string
	if err := toy.UnpackArgs(args, "s", &s); err != nil {
		return nil, err
	}
	mac, err := net.ParseMAC(s)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return MAC(mac), nil
}

type Addr struct {
	a net.Addr
}

var AddrType = toy.NewType[*Addr]("net.Addr", nil)

func (a *Addr) Type() toy.ObjectType { return AddrType }
func (a *Addr) String() string       { return fmt.Sprintf("net.Addr(%q, %q)", a.a.Network(), a.a.String()) }
func (a *Addr) IsFalsy() bool        { return false }
func (a *Addr) Clone() toy.Object    { return &Addr{a: a.a} }

func (a *Addr) FieldGet(name string) (toy.Object, error) {
	switch name {
	case "network":
		return toy.String(a.a.Network()), nil
	case "address":
		return toy.String(a.a.String()), nil
	}
	return nil, toy.ErrNoSuchField
}

func netInterfaceAddrs(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	elems := make([]toy.Object, 0, len(addrs))
	for _, addr := range addrs {
		elems = append(elems, &Addr{a: addr})
	}
	return toy.Tuple{toy.NewArray(elems), toy.Nil}, nil
}

type IPAddr net.IPAddr

var IPAddrType = toy.NewType[*IPAddr]("net.IPAddr", nil)

func (a *IPAddr) Type() toy.ObjectType { return IPAddrType }
func (a *IPAddr) String() string       { return fmt.Sprintf("net.IPAddr(%q)", (*net.IPAddr)(a).String()) }
func (a *IPAddr) IsFalsy() bool        { return false }
func (a *IPAddr) Clone() toy.Object    { return &IPAddr{IP: a.IP, Zone: a.Zone} }

func (a *IPAddr) FieldGet(name string) (toy.Object, error) {
	switch name {
	case "ip":
		return IP(a.IP), nil
	case "zone":
		return toy.String(a.Zone), nil
	}
	return nil, toy.ErrNoSuchField
}

func netResolveIPAddr(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	var network, address string
	if err := toy.UnpackArgs(args, "network", &network, "address", &address); err != nil {
		return nil, err
	}
	addr, err := net.ResolveIPAddr(network, address)
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	return toy.Tuple{(*IPAddr)(addr), toy.Nil}, nil
}

type Interface net.Interface

var InterfaceType = toy.NewType[*Interface]("net.Interface", nil)

func (i *Interface) Type() toy.ObjectType { return InterfaceType }

func (i *Interface) String() string {
	return fmt.Sprintf("net.Interface(%q)", (*net.Interface)(i).Name)
}

func (i *Interface) IsFalsy() bool { return false }

func (i *Interface) Clone() toy.Object {
	c := new(Interface)
	*c = *i
	return c
}

func (i *Interface) FieldGet(name string) (toy.Object, error) {
	switch name {
	case "index":
		return toy.Int(i.Index), nil
	case "mtu":
		return toy.Int(i.MTU), nil
	case "name":
		return toy.String(i.Name), nil
	case "mac":
		return MAC(i.HardwareAddr), nil
	case "flags":
		return InterfaceFlags(i.Flags), nil
	}
	return nil, toy.ErrNoSuchField
}

type InterfaceFlags net.Flags

var InterfaceFlagsType = NewEnum("net.InterfaceFlags", map[string]InterfaceFlags{
	"UP":             InterfaceFlags(net.FlagUp),
	"BROADCAST":      InterfaceFlags(net.FlagBroadcast),
	"LOOPBACK":       InterfaceFlags(net.FlagLoopback),
	"POINT_TO_POINT": InterfaceFlags(net.FlagPointToPoint),
	"MULTICAST":      InterfaceFlags(net.FlagMulticast),
	"RUNNING":        InterfaceFlags(net.FlagRunning),
}, func(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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

func (f InterfaceFlags) Type() toy.ObjectType { return InterfaceFlagsType }

func (f InterfaceFlags) String() string {
	return fmt.Sprintf("net.InterfaceFlags(%q)", net.Flags(f).String())
}

func (f InterfaceFlags) IsFalsy() bool     { return false }
func (f InterfaceFlags) Clone() toy.Object { return f }

func (f InterfaceFlags) Convert(p any) error {
	switch p := p.(type) {
	case *toy.Int:
		*p = toy.Int(f)
	default:
		return toy.ErrNotConvertible
	}
	return nil
}

func (f InterfaceFlags) BinaryOp(op token.Token, other toy.Object, right bool) (toy.Object, error) {
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
	return nil, toy.ErrInvalidOperator
}

func (f InterfaceFlags) UnaryOp(op token.Token) (toy.Object, error) {
	switch op {
	case token.Xor:
		return ^f, nil
	}
	return nil, toy.ErrInvalidOperator
}

func netLookupInterface(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
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
			return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
		}
		return toy.Tuple{(*Interface)(iface), toy.Nil}, nil
	case toy.Int:
		iface, err := net.InterfaceByIndex(int(x))
		if err != nil {
			return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
		}
		return toy.Tuple{(*Interface)(iface), toy.Nil}, nil
	default:
		return nil, &toy.InvalidArgumentTypeError{
			Name: "id",
			Want: "string or int",
			Got:  toy.TypeName(x),
		}
	}
}

func netInterfaces(_ *toy.VM, args ...toy.Object) (toy.Object, error) {
	if len(args) != 0 {
		return nil, &toy.WrongNumArgumentsError{Got: len(args)}
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return toy.Tuple{toy.Nil, toy.NewError(err.Error())}, nil
	}
	elems := make([]toy.Object, 0, len(ifaces))
	for i := range ifaces {
		iface := (*Interface)(&ifaces[i])
		elems = append(elems, iface)
	}
	return toy.Tuple{toy.NewArray(elems), toy.Nil}, nil
}
