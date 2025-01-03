package tengo

import (
	"bytes"
	"fmt"
	"iter"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/d5/tengo/v2/token"
)

// Object represents an object in the VM.
type Object interface {
	// TypeName should return the name of the type.
	TypeName() string

	// String should return a string representation of the type's value.
	String() string

	// IsFalsy should return true if the value of the type should be considered as falsy.
	IsFalsy() bool

	// Copy should return a copy of the type (and its value). Copy function
	// will be used for copy() builtin function which is expected to deep-copy
	// the values generally.
	Copy() Object
}

// Hashable represents an object that is hashable (can be used in map).
type Hashable interface {
	Object
	// Hash returns a function of x such that Equals(x, y) => Hash(x) == Hash(y).
	// The hash is used only by maps and is not exposed to the Tengo program.
	Hash() uint32
}

// Freezable represents an object that can create immutable copies.
type Freezable interface {
	Object
	// Copy should return an immutable copy of the type (and its value).
	// AsImmutable function will be used for immutable() builtin keyword
	// which is expected to deep-copy the values generally.
	AsImmutable() Object
}

// A Comparable is a value that defines its own equivalence relation and
// perhaps ordered comparisons.
type Comparable interface {
	Object
	// CompareSameType compares one value to another.
	// The comparison operation must be one of Equal, NotEqual, Less, LessEq, Greater, or GreaterEq.
	// CompareSameType returns an error if an ordered comparison was
	// requested for a type that does not support it.
	// If CompareSameType returns an error, the VM will treat it as a run-time error.
	Compare(op token.Token, rhs Object) (bool, error)
}

// HasBinaryOp represents an object that supports binary operations (excluding comparison operators).
type HasBinaryOp interface {
	Object
	// BinaryOp should return another object that is the result of a given
	// binary operator and a right-hand side object.
	// If BinaryOp returns an error, the VM will treat it as a run-time error.
	BinaryOp(op token.Token, rhs Object) (Object, error)
}

// HasBinaryOp represents an object that supports unary operations.
type HasUnaryOp interface {
	Object
	// UnaryOp should return another object that is the result of a given unary operator.
	// If UnaryOp returns an error, the VM will treat it as a run-time error.
	UnaryOp(op token.Token) (Object, error)
}

// HasIndexGet represents an object that supports indexed access.
type HasIndexGet interface {
	Object
	// IndexGet should take an index Object and return a result Object or an
	// error for indexable objects. Indexable is an object that can take an
	// index and return an object. If error is returned, the runtime will treat
	// it as a run-time error and ignore returned value.
	// If nil is returned as value, it will be converted to Undefined value by the runtime.
	IndexGet(index Object) (value Object, err error)
}

// HasIndexSet represents an object that supports indexed assignment.
type HasIndexSet interface {
	Object
	// IndexSet should take an index Object and a value Object for index
	// assignable objects. Index assignable is an object that can take an index
	// and a value on the left-hand side of the assignment statement.
	// If an error is returned, it will be treated as a run-time error.
	IndexSet(index, value Object) error
}

// HasFieldGet represents an object that supports field access.
type HasFieldGet interface {
	Object
	// FieldGet should take a name of the field and return its value.
	// If error is returned, the runtime will treat
	// it as a run-time error and ignore returned value.
	FieldGet(name string) (Object, error)
}

// HasFieldGet represents an object that supports field assignment.
type HasFieldSet interface {
	Object
	// FieldSet should take the name of a field and its new value Object
	// and return an error if the field cannot be set.
	// If error is returned, the runtime will treat it as a run-time error.
	FieldSet(name string, value Object) error
}

// HasLen represents an object that can report its length or size.
type HasLen interface {
	Object
	// Len should return object's length or size.
	Len() int
}

// An Indexable is a sequence of known length that supports efficient random access.
// It is not necessarily iterable.
type Indexable interface {
	HasLen
	// The caller must ensure that 0 <= i < Len().
	At(i int) Object
}

// A Sliceable is a sequence that can be cut into pieces with the slice operator.
type Sliceable interface {
	Indexable
	// The caller must ensure that the low and high indices are valid.
	Slice(low, high int) Object
}

// Convertible represents an object that can be converted to another type.
type Convertible interface {
	Object
	// Convert should take a pointer to an Object and try to convert the Convertible object
	// to the type of the provided Object, and return an error if the conversion fails.
	Convert(p any) error
}

// Callable represents an object that can be called.
type Callable interface {
	Object
	// Call should take an arbitrary number of arguments and return a return
	// value and/or an error, which the VM will consider as a run-time error.
	Call(args ...Object) (ret Object, err error)
}

// Iterable represents an object that can be iterated.
type Iterable interface {
	Object
	// Iterate should return an Iterator for the type.
	Iterate() Iterator
}

// Sequence represents an iterable sequence of objects of known length.
type Sequence interface {
	Iterable
	HasLen
	// Items should return a slice containing all the elements in the Sequence.
	Items() []Object
}

// Mapping represents an iterable object that maps one Object to another.
type Mapping interface {
	Iterable
	// Items should return a slice containing all the entries in the Mapping.
	Items() []Tuple
}

// Iterator represents an iterator for underlying data type.
type Iterator interface {
	Object
	// If the iterator is exhausted, Next returns false.
	// Otherwise, returns true and sets *key and *value
	// to key/index and value of the current element, respectivly.
	// *key and *value can be nil to avoid being set.
	Next(key, value *Object) bool
}

// CloseableIterator represents an iterator
// that must be closed after iteration is completed.
type CloseableIterator interface {
	Iterator
	// Close releases associated resources.
	// This method must be called.
	Close()
}

func Hash(x Object) (uint32, error) {
	xh, ok := x.(Hashable)
	if !ok {
		return 0, ErrNotHashable
	}
	return xh.Hash(), nil
}

func AsImmutable(x Object) Object {
	xf, ok := x.(Freezable)
	if !ok {
		return x.Copy()
	}
	return xf.AsImmutable()
}

func Equals(x, y Object) (bool, error) {
	return Compare(token.Equal, x, y)
}

func Compare(op token.Token, x, y Object) (bool, error) {
	if x == y {
		return op == token.Equal, nil
	}
	xc, ok := x.(Comparable)
	if !ok {
		return false, ErrInvalidOperator
	}
	return xc.Compare(op, y)
}

func BinaryOp(op token.Token, x, y Object) (Object, error) {
	xb, ok := x.(HasBinaryOp)
	if !ok {
		return nil, ErrInvalidOperator
	}
	return xb.BinaryOp(op, y)
}

func UnaryOp(op token.Token, x Object) (res Object, err error) {
	if op == token.Not {
		return Bool(x.IsFalsy()), nil
	}
	xu, ok := x.(HasUnaryOp)
	if !ok {
		return nil, ErrInvalidOperator
	}
	return xu.UnaryOp(op)
}

func IndexGet(x, index Object) (Object, error) {
	if i, ok := index.(Int); ok {
		if xi, ok := x.(Indexable); ok {
			if i < 0 || int64(i) >= int64(xi.Len()) {
				return Undefined, nil
			}
			return xi.At(int(i)), nil
		}
	}
	xi, ok := x.(HasIndexGet)
	if !ok {
		return nil, ErrNotIndexable
	}
	return xi.IndexGet(index)
}

func IndexSet(x, index, value Object) error {
	xi, ok := x.(HasIndexSet)
	if !ok {
		return ErrNotIndexable
	}
	return xi.IndexSet(index, value)
}

func FieldGet(x Object, name string) (Object, error) {
	xf, ok := x.(HasFieldGet)
	if ok {
		return xf.FieldGet(name)
	}
	xi, ok := x.(HasIndexGet)
	if !ok {
		return nil, ErrNoFields
	}
	return xi.IndexGet(String(name))
}

func FieldSet(x Object, name string, value Object) error {
	xf, ok := x.(HasFieldSet)
	if ok {
		return xf.FieldSet(name, value)
	}
	xi, ok := x.(HasIndexSet)
	if !ok {
		return ErrNoFields
	}
	return xi.IndexSet(String(name), value)
}

func Slice(x Object, low, high int) (Object, error) {
	xs, ok := x.(Sliceable)
	if !ok {
		return nil, ErrNotSliceable
	}
	n := xs.Len()
	if low > high {
		return nil, fmt.Errorf("invalid slice indices: %d > %d", low, high)
	}
	if low < 0 || low > n {
		return nil, fmt.Errorf("slice bounds out of range [%d:%d]", low, n)
	}
	if high < 0 || high > n {
		return nil, fmt.Errorf("slice bounds out of range [%d:%d] with len %d", low, high, n)
	}
	return xs.Slice(low, high), nil
}

func Convert[T Object](p *T, o Object) error {
	if o == Undefined {
		return ErrNotConvertible
	}
	switch p := any(p).(type) {
	case *String:
		if x, ok := o.(String); ok {
			*p = x
		} else {
			*p = String(o.String())
		}
	case *Bool:
		*p = Bool(!o.IsFalsy())
	case *T:
		t, ok := o.(T)
		if ok {
			*p = t
			return nil
		}
		c, ok := o.(Convertible)
		if !ok {
			return ErrNotConvertible
		}
		return c.Convert(p)
	}
	return nil
}

func Elements(iterable Iterable) iter.Seq[Object] {
	type hasElements interface {
		Elements() iter.Seq[Object]
	}
	if iterable, ok := iterable.(hasElements); ok {
		return iterable.Elements()
	}
	it := iterable.Iterate()
	return func(yield func(Object) bool) {
		if c, ok := it.(CloseableIterator); ok {
			defer c.Close()
		}
		var value Object
		for it.Next(nil, &value) {
			if !yield(value) {
				break
			}
		}
	}
}

func Entries(iterable Iterable) iter.Seq2[Object, Object] {
	type hasEntries interface {
		Entries() iter.Seq2[Object, Object]
	}
	if iterable, ok := iterable.(hasEntries); ok {
		return iterable.Entries()
	}
	it := iterable.Iterate()
	return func(yield func(Object, Object) bool) {
		if c, ok := it.(CloseableIterator); ok {
			defer c.Close()
		}
		var key, value Object
		for it.Next(&key, &value) {
			if !yield(key, value) {
				break
			}
		}
	}
}

// UndefinedType represents an undefined value.
type UndefinedType byte

const Undefined = UndefinedType(0)

func (o UndefinedType) TypeName() string { return "undefined" }
func (o UndefinedType) String() string   { return "<undefined>" }
func (o UndefinedType) IsFalsy() bool    { return true }
func (o UndefinedType) Copy() Object     { return o }

// Bool represents a boolean value.
type Bool bool

const (
	True  = Bool(true)
	False = Bool(false)
)

func (o Bool) String() string {
	if o {
		return "true"
	}
	return "false"
}

func (o Bool) TypeName() string { return "bool" }
func (o Bool) IsFalsy() bool    { return !bool(o) }
func (o Bool) Copy() Object     { return o }

func (o Bool) Hash() uint32 {
	if o {
		return 1
	}
	return 0
}

func (o Bool) Compare(op token.Token, rhs Object) (bool, error) {
	y, ok := rhs.(Bool)
	if !ok {
		return false, ErrInvalidOperator
	}
	switch op {
	case token.Equal:
		return o == y, nil
	case token.NotEqual:
		return o != y, nil
	}
	return false, ErrInvalidOperator
}

// Float represents a floating point number value.
type Float float64

func (o Float) String() string   { return strconv.FormatFloat(float64(o), 'g', -1, 64) }
func (o Float) TypeName() string { return "float" }
func (o Float) IsFalsy() bool    { return math.IsNaN(float64(o)) }
func (o Float) Copy() Object     { return o }

func (o Float) Hash() uint32 {
	if float64(o) == math.Trunc(float64(o)) && float64(o) <= float64(math.MaxInt64) {
		return Int(o).Hash()
	}
	return murmur32Float64(float64(o))
}

func (o Float) Convert(p any) error {
	i, ok := p.(*Int)
	if !ok {
		return ErrNotConvertible
	}
	*i = Int(o)
	return nil
}

func (o Float) Compare(op token.Token, rhs Object) (bool, error) {
	switch y := rhs.(type) {
	case Float:
		switch op {
		case token.Equal:
			return o == y, nil
		case token.NotEqual:
			return o != y, nil
		case token.Less:
			return o < y, nil
		case token.Greater:
			return o > y, nil
		case token.LessEq:
			return o <= y, nil
		case token.GreaterEq:
			return o >= y, nil
		}
	case Int:
		switch op {
		case token.Equal:
			return o == Float(y), nil
		case token.NotEqual:
			return o != Float(y), nil
		case token.Less:
			return o < Float(y), nil
		case token.Greater:
			return o > Float(y), nil
		case token.LessEq:
			return o <= Float(y), nil
		case token.GreaterEq:
			return o >= Float(y), nil
		}
	}
	return false, ErrInvalidOperator
}

func (o Float) BinaryOp(op token.Token, rhs Object) (Object, error) {
	switch y := rhs.(type) {
	case Float:
		switch op {
		case token.Add:
			return o + y, nil
		case token.Sub:
			return o - y, nil
		case token.Mul:
			return o * y, nil
		case token.Quo:
			return o / y, nil
		}
	case Int:
		switch op {
		case token.Add:
			return o + Float(y), nil
		case token.Sub:
			return o - Float(y), nil
		case token.Mul:
			return o * Float(y), nil
		case token.Quo:
			return o / Float(y), nil
		}
	}
	return nil, ErrInvalidOperator
}

func (o Float) UnaryOp(op token.Token) (Object, error) {
	switch op {
	case token.Add:
		return o, nil
	case token.Sub:
		return -o, nil
	}
	return nil, ErrInvalidOperator
}

// Int represents an integer value.
type Int int64

func (o Int) String() string   { return strconv.FormatInt(int64(o), 10) }
func (o Int) TypeName() string { return "int" }
func (o Int) IsFalsy() bool    { return o == 0 }
func (o Int) Copy() Object     { return o }
func (o Int) Hash() uint32     { return murmur32Int64(int64(o)) }

func (o Int) Convert(p any) error {
	switch p := p.(type) {
	case *Float:
		*p = Float(o)
		return nil
	case *Char:
		*p = Char(o)
		return nil
	}
	return ErrNotConvertible
}

func (o Int) Compare(op token.Token, rhs Object) (bool, error) {
	switch y := rhs.(type) {
	case Int:
		switch op {
		case token.Equal:
			return o == y, nil
		case token.NotEqual:
			return o != y, nil
		case token.Less:
			return o < y, nil
		case token.Greater:
			return o > y, nil
		case token.LessEq:
			return o <= y, nil
		case token.GreaterEq:
			return o >= y, nil
		}
	case Float:
		switch op {
		case token.Equal:
			return o == Int(y), nil
		case token.NotEqual:
			return o != Int(y), nil
		case token.Less:
			return o < Int(y), nil
		case token.Greater:
			return o > Int(y), nil
		case token.LessEq:
			return o <= Int(y), nil
		case token.GreaterEq:
			return o >= Int(y), nil
		}
	case Char:
		switch op {
		case token.Equal:
			return o == Int(y), nil
		case token.NotEqual:
			return o != Int(y), nil
		case token.Less:
			return o < Int(y), nil
		case token.Greater:
			return o > Int(y), nil
		case token.LessEq:
			return o <= Int(y), nil
		case token.GreaterEq:
			return o >= Int(y), nil
		}
	}
	return false, ErrInvalidOperator
}

func (o Int) BinaryOp(op token.Token, rhs Object) (Object, error) {
	switch y := rhs.(type) {
	case Int:
		switch op {
		case token.Add:
			return o + y, nil
		case token.Sub:
			return o - y, nil
		case token.Mul:
			return o * y, nil
		case token.Quo:
			return o / y, nil
		case token.Rem:
			return o % y, nil
		case token.And:
			return o & y, nil
		case token.Or:
			return o | y, nil
		case token.Xor:
			return o ^ y, nil
		case token.AndNot:
			return o &^ y, nil
		case token.Shl:
			return o << y, nil
		case token.Shr:
			return o >> y, nil
		}
	case Float:
		switch op {
		case token.Add:
			return Float(o) + y, nil
		case token.Sub:
			return Float(o) - y, nil
		case token.Mul:
			return Float(o) * y, nil
		case token.Quo:
			return Float(o) / y, nil
		}
	case Char:
		switch op {
		case token.Add:
			return Char(o) + y, nil
		case token.Sub:
			return Char(o) - y, nil
		}
	}
	return nil, ErrInvalidOperator
}

func (o Int) UnaryOp(op token.Token) (Object, error) {
	switch op {
	case token.Add:
		return o, nil
	case token.Sub:
		return -o, nil
	case token.Xor:
		return ^o, nil
	}
	return nil, ErrInvalidOperator
}

// String represents a string value.
type String string

func (o String) TypeName() string { return "string" }
func (o String) String() string   { return strconv.Quote(string(o)) }
func (o String) IsFalsy() bool    { return len(o) == 0 }
func (o String) Copy() Object     { return o }
func (o String) Hash() uint32     { return murmur32String(string(o)) }

func (o String) Len() int { return utf8.RuneCountInString(string(o)) }

func (o String) At(idx int) Object {
	for i, r := range o {
		if int64(i) == int64(idx) {
			return Char(r)
		}
	}
	return Undefined // should not happend
}

func (o String) Slice(low, high int) Object {
	rs := []rune(o)
	return String(rs[low:high])
}

func (o String) Convert(p any) error {
	b, ok := p.(*Bytes)
	if !ok {
		return ErrNotConvertible
	}
	*b = []byte(o)
	return nil
}

func (o String) Compare(op token.Token, rhs Object) (bool, error) {
	y, ok := rhs.(String)
	if !ok {
		return false, ErrInvalidOperator
	}
	switch op {
	case token.Equal:
		return o == y, nil
	case token.NotEqual:
		return o != y, nil
	case token.Less:
		return o < y, nil
	case token.Greater:
		return o > y, nil
	case token.LessEq:
		return o <= y, nil
	case token.GreaterEq:
		return o >= y, nil
	}
	return false, ErrInvalidOperator
}

func (o String) BinaryOp(op token.Token, rhs Object) (Object, error) {
	y, ok := rhs.(String)
	if !ok {
		return nil, ErrInvalidOperator
	}
	switch op {
	case token.Add:
		return o + y, nil
	}
	return nil, ErrInvalidOperator
}

func (o String) IndexGet(index Object) (res Object, err error) {
	intIdx, ok := index.(Int)
	if !ok {
		return nil, ErrInvalidIndexType
	}
	if intIdx < 0 {
		return Undefined, nil
	}
	for i, r := range o {
		if int64(i) == int64(intIdx) {
			return Char(r), nil
		}
	}
	return Undefined, nil
}

func (o String) Iterate() Iterator { return &stringIterator{s: []rune(o), i: 0} }

type stringIterator struct {
	s []rune
	i int
}

func (it *stringIterator) TypeName() string { return "string-iterator" }
func (it *stringIterator) String() string   { return "<string-iterator>" }
func (it *stringIterator) IsFalsy() bool    { return true }
func (it *stringIterator) Copy() Object     { return &stringIterator{s: it.s, i: it.i} }

func (it *stringIterator) Next(key, value *Object) bool {
	if it.i < len(it.s) {
		if key != nil {
			*key = Int(it.i)
		}
		if value != nil {
			*value = Char(it.s[it.i])
		}
		it.i++
		return true
	}
	return false
}

// Bytes represents a byte array.
type Bytes []byte

func (o Bytes) String() string   { return string(o) }
func (o Bytes) TypeName() string { return "bytes" }
func (o Bytes) IsFalsy() bool    { return len(o) == 0 }
func (o Bytes) Copy() Object     { return slices.Clone(o) }
func (o Bytes) Hash() uint32     { return murmur32(o) }

func (o Bytes) Len() int                   { return len(o) }
func (o Bytes) At(i int) Object            { return Int(o[i]) }
func (o Bytes) Slice(low, high int) Object { return o[low:high] }

func (o Bytes) Convert(p any) error {
	s, ok := p.(*String)
	if !ok {
		return ErrNotConvertible
	}
	*s = String(o)
	return nil
}

func (o Bytes) Compare(op token.Token, rhs Object) (bool, error) {
	y, ok := rhs.(Bytes)
	if !ok {
		return false, ErrInvalidOperator
	}
	switch op {
	case token.Equal:
		return bytes.Equal(o, y), nil
	case token.NotEqual:
		return !bytes.Equal(o, y), nil
	case token.Less:
		return bytes.Compare(o, y) < 0, nil
	case token.Greater:
		return bytes.Compare(o, y) > 0, nil
	case token.LessEq:
		return bytes.Compare(o, y) <= 0, nil
	case token.GreaterEq:
		return bytes.Compare(o, y) >= 0, nil
	}
	return false, ErrInvalidOperator
}

func (o Bytes) BinaryOp(op token.Token, rhs Object) (Object, error) {
	y, ok := rhs.(Bytes)
	if !ok {
		return nil, ErrInvalidOperator
	}
	switch op {
	case token.Add:
		if len(o)+len(y) > MaxBytesLen {
			return nil, ErrBytesLimit
		}
		return slices.Concat(o, y), nil
	}
	return nil, ErrInvalidOperator
}

func (o Bytes) IndexGet(index Object) (res Object, err error) {
	intIdx, ok := index.(Int)
	if !ok {
		return nil, ErrInvalidIndexType
	}
	if intIdx < 0 || int64(intIdx) >= int64(len(o)) {
		return Undefined, nil
	}
	return Int(o[intIdx]), nil
}

func (o Bytes) Iterate() Iterator { return &bytesIterator{b: o, i: 0} }

type bytesIterator struct {
	b Bytes
	i int
}

func (it *bytesIterator) TypeName() string { return "bytes-iterator" }
func (it *bytesIterator) String() string   { return "<bytes-iterator>" }
func (it *bytesIterator) IsFalsy() bool    { return true }
func (it *bytesIterator) Copy() Object     { return &bytesIterator{b: it.b, i: it.i} }

func (it *bytesIterator) Next(key, value *Object) bool {
	if it.i < len(it.b) {
		if key != nil {
			*key = Int(it.i)
		}
		if value != nil {
			*value = Int(it.b[it.i])
		}
		it.i++
		return true
	}
	return false
}

// Char represents a character value.
type Char rune

func (o Char) String() string   { return string(o) }
func (o Char) TypeName() string { return "char" }
func (o Char) IsFalsy() bool    { return o == 0 }
func (o Char) Copy() Object     { return o }
func (o Char) Hash() uint32     { return murmur32Int32(int32(o)) }

func (o Char) Convert(p any) error {
	i, ok := p.(*Int)
	if !ok {
		return ErrNotConvertible
	}
	*i = Int(o)
	return nil
}

func (o Char) Compare(op token.Token, rhs Object) (bool, error) {
	switch y := rhs.(type) {
	case Char:
		switch op {
		case token.Equal:
			return o == y, nil
		case token.NotEqual:
			return o != y, nil
		case token.Less:
			return o < y, nil
		case token.Greater:
			return o > y, nil
		case token.LessEq:
			return o <= y, nil
		case token.GreaterEq:
			return o >= y, nil
		}
	case Int:
		switch op {
		case token.Equal:
			return Int(o) == y, nil
		case token.NotEqual:
			return Int(o) != y, nil
		case token.Less:
			return Int(o) < y, nil
		case token.Greater:
			return Int(o) > y, nil
		case token.LessEq:
			return Int(o) <= y, nil
		case token.GreaterEq:
			return Int(o) >= y, nil
		}
	}
	return false, ErrInvalidOperator
}

func (o Char) BinaryOp(op token.Token, rhs Object) (Object, error) {
	switch y := rhs.(type) {
	case Char:
		switch op {
		case token.Add:
			return o + y, nil
		case token.Sub:
			return o - y, nil
		}
	case Int:
		switch op {
		case token.Add:
			return o + Char(y), nil
		case token.Sub:
			return o - Char(y), nil
		}
	}
	return nil, ErrInvalidOperator
}

type Array struct {
	elems     []Object
	immutable bool
	itercount uint32 // number of active iterators (ignored if frozen)
}

func NewArray(elems []Object) *Array {
	return &Array{
		elems:     elems,
		immutable: false,
		itercount: 0,
	}
}

func (o *Array) TypeName() string { return "array" }

func (o *Array) String() string {
	var b strings.Builder
	b.WriteByte('[')
	for i, v := range o.elems {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(v.String())
	}
	b.WriteByte(']')
	return b.String()
}

func (o *Array) IsFalsy() bool { return len(o.elems) == 0 }

func (o *Array) Copy() Object {
	if o.immutable {
		return o
	}
	elems := make([]Object, 0, len(o.elems))
	for _, elem := range o.elems {
		elems = append(elems, elem.Copy())
	}
	return &Array{elems: elems, immutable: false}
}

func (o *Array) AsImmutable() Object {
	if o.immutable {
		return o
	}
	elems := make([]Object, 0, len(o.elems))
	for _, elem := range o.elems {
		elems = append(elems, AsImmutable(elem))
	}
	return &Array{elems: elems, immutable: true}
}

func (o *Array) Len() int        { return len(o.elems) }
func (o *Array) At(i int) Object { return o.elems[i] }
func (o *Array) Items() []Object { return o.elems }

func (o *Array) Slice(low, high int) Object {
	return &Array{
		elems:     o.elems[low:high],
		immutable: o.immutable,
		itercount: 0,
	}
}

func (o *Array) Append(xs ...Object) error {
	if err := o.checkMutable("append to"); err != nil {
		return err
	}
	o.elems = append(o.elems, xs...)
	return nil
}

func (o *Array) Clear() error {
	if err := o.checkMutable("clear"); err != nil {
		return err
	}
	clear(o.elems)
	o.elems = o.elems[:0]
	return nil
}

func (o *Array) Compare(op token.Token, rhs Object) (bool, error) {
	y, ok := rhs.(*Array)
	if !ok {
		return false, ErrInvalidOperator
	}
	switch op {
	case token.Equal:
		if len(o.elems) != len(y.elems) {
			return false, nil
		}
		for i := range len(o.elems) {
			if eq, err := Equals(o.elems[i], y.elems[i]); err != nil {
				return false, err
			} else if !eq {
				return false, nil
			}
		}
	case token.NotEqual:
		if len(o.elems) != len(y.elems) {
			return true, nil
		}
		for i := range len(o.elems) {
			if eq, err := Equals(o.elems[i], y.elems[i]); err != nil {
				return false, err
			} else if !eq {
				return true, nil
			}
		}
	}
	return false, ErrInvalidOperator
}

func (o *Array) BinaryOp(op token.Token, rhs Object) (Object, error) {
	y, ok := rhs.(*Array)
	if !ok {
		return nil, ErrInvalidOperator
	}
	switch op {
	case token.Add:
		return &Array{
			elems:     append(o.elems, y.elems...),
			immutable: o.immutable,
			itercount: 0,
		}, nil
	}
	return nil, ErrInvalidOperator
}

func (o *Array) IndexGet(index Object) (res Object, err error) {
	intIdx, ok := index.(Int)
	if !ok {
		return nil, ErrInvalidIndexType
	}
	if intIdx < 0 || int64(intIdx) >= int64(len(o.elems)) {
		return Undefined, nil
	}
	return o.elems[intIdx], nil
}

func (o *Array) IndexSet(index, value Object) (err error) {
	if err := o.checkMutable("assign to element of"); err != nil {
		return err
	}
	intIdx, ok := index.(Int)
	if !ok {
		return ErrInvalidIndexType
	}
	n := len(o.elems)
	if intIdx < 0 || int64(intIdx) >= int64(n) {
		return fmt.Errorf("index %d out of range [:%d]", intIdx, n)
	}
	o.elems[intIdx] = value
	return nil
}

func (o *Array) Iterate() Iterator {
	if !o.immutable {
		o.itercount++
	}
	return &arrayIterator{a: o, i: 0}
}

func (o *Array) Elements() iter.Seq[Object] {
	return func(yield func(Object) bool) {
		if !o.immutable {
			o.itercount++
			defer func() { o.itercount-- }()
		}
		for _, x := range o.elems {
			if !yield(x) {
				break
			}
		}
	}
}

func (o *Array) Entries() iter.Seq2[Object, Object] {
	return func(yield func(Object, Object) bool) {
		if !o.immutable {
			o.itercount++
			defer func() { o.itercount-- }()
		}
		for i, x := range o.elems {
			if !yield(Int(i), x) {
				break
			}
		}
	}
}

// checkMutable reports an error if the array should not be mutated.
// verb+" dict" should describe the operation.
func (o *Array) checkMutable(verb string) error {
	if o.immutable {
		return fmt.Errorf("cannot %s immutable array", verb)
	}
	if o.itercount > 0 {
		return fmt.Errorf("cannot %s hash table during iteration", verb)
	}
	return nil
}

type arrayIterator struct {
	a *Array
	i int
}

func (it *arrayIterator) TypeName() string { return "array-iterator" }
func (it *arrayIterator) String() string   { return "<array-iterator>" }
func (it *arrayIterator) IsFalsy() bool    { return true }
func (it *arrayIterator) Copy() Object     { return &arrayIterator{a: it.a, i: it.i} }

func (it *arrayIterator) Next(key, value *Object) bool {
	if it.i < len(it.a.elems) {
		if key != nil {
			*key = Int(it.i)
		}
		if value != nil {
			*value = it.a.elems[it.i]
		}
		it.i++
		return true
	}
	return false
}

func (it *arrayIterator) Close() {
	if !it.a.immutable {
		it.a.itercount--
	}
}

// Map represents a map of objects.
type Map struct {
	ht hashtable
}

func NewMap(size int) *Map {
	m := new(Map)
	m.ht.init(size)
	return m
}

func (o *Map) TypeName() string { return "map" }

func (o *Map) String() string {
	var b strings.Builder
	b.WriteByte('{')
	for key, value := range o.ht.entries() {
		if b.Len() != 1 {
			b.WriteString(", ")
		}
		b.WriteString(key.String())
		b.WriteString(": ")
		b.WriteString(value.String())
	}
	b.WriteByte('}')
	return b.String()
}

func (o *Map) IsFalsy() bool { return o.ht.len == 0 }

func (o *Map) Copy() Object {
	m := new(Map)
	m.ht.init(o.Len())
	m.ht.copyAll(&o.ht)
	return m
}

func (o *Map) AsImmutable() Object {
	if o.ht.immutable {
		return o
	}
	m := new(Map)
	m.ht.init(o.Len())
	m.ht.copyAllImmutable(&o.ht)
	m.ht.immutable = true
	return m
}

func (o *Map) Len() int { return int(o.ht.len) }

func (o *Map) Compare(op token.Token, rhs Object) (bool, error) {
	y, ok := rhs.(*Map)
	if !ok {
		return false, ErrInvalidOperator
	}
	switch op {
	case token.Equal:
		eq, err := o.ht.equals(&y.ht)
		if err != nil {
			return false, err
		}
		return eq, nil
	case token.NotEqual:
		eq, err := o.ht.equals(&y.ht)
		if err != nil {
			return false, err
		}
		return !eq, nil
	}
	return false, ErrInvalidOperator
}

func (o *Map) BinaryOp(op token.Token, rhs Object) (Object, error) {
	y, ok := rhs.(*Map)
	if !ok {
		return nil, ErrInvalidOperator
	}
	switch op {
	case token.Or:
		return o.Union(y), nil
	}
	return nil, ErrInvalidOperator
}

func (o *Map) IndexGet(index Object) (res Object, err error) { return o.ht.lookup(index) }
func (o *Map) IndexSet(index, value Object) (err error)      { return o.ht.insert(index, value) }
func (o *Map) Iterate() Iterator                             { return o.ht.iterate() }
func (o *Map) Elements() iter.Seq[Object]                    { return o.ht.elements() }
func (o *Map) Entries() iter.Seq2[Object, Object]            { return o.ht.entries() }

func (o *Map) Delete(key Object) (Object, error) { return o.ht.delete(key) }
func (o *Map) Clear() error                      { return o.ht.clear() }
func (o *Map) Keys() []Object                    { return o.ht.keys() }
func (o *Map) Values() []Object                  { return o.ht.values() }
func (o *Map) Items() []Tuple                    { return o.ht.items() }

func (o *Map) Union(y *Map) *Map {
	z := new(Map)
	z.ht.init(o.Len())
	z.ht.addAll(&o.ht)
	z.ht.addAll(&y.ht)
	return z
}

type Tuple []Object

func (o Tuple) TypeName() string { return "tuple" }

func (o Tuple) String() string {
	var b strings.Builder
	for i, v := range o {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(v.String())
	}
	return b.String()
}

func (o Tuple) IsFalsy() bool { return len(o) == 0 }

func (o Tuple) Copy() Object {
	t := make(Tuple, 0, len(o))
	for _, elem := range o {
		t = append(t, elem.Copy())
	}
	return t
}

func (o Tuple) Len() int                   { return len(o) }
func (o Tuple) At(i int) Object            { return o[i] }
func (o Tuple) Slice(low, high int) Object { return o[low:high] }
func (o Tuple) Items() []Object            { return o }

func (o Tuple) Compare(op token.Token, rhs Object) (bool, error) {
	y, ok := rhs.(Tuple)
	if !ok {
		return false, ErrInvalidOperator
	}
	switch op {
	case token.Equal:
		if len(o) != len(y) {
			return false, nil
		}
		for i := range o.Len() {
			if eq, err := Equals(o[i], y[i]); err != nil {
				return false, err
			} else if !eq {
				return false, nil
			}
		}
	case token.NotEqual:
		if len(o) != len(y) {
			return true, nil
		}
		for i := range len(o) {
			if eq, err := Equals(o[i], y[i]); err != nil {
				return false, err
			} else if !eq {
				return true, nil
			}
		}
	}
	return false, ErrInvalidOperator
}

func (o Tuple) IndexGet(index Object) (res Object, err error) {
	intIdx, ok := index.(Int)
	if !ok {
		return nil, ErrInvalidIndexType
	}
	if intIdx < 0 || int64(intIdx) >= int64(len(o)) {
		return Undefined, nil
	}
	return o[intIdx], nil
}

func (o Tuple) Iterate() Iterator { return &tupleIterator{t: o, i: 0} }

func (o Tuple) Elements() iter.Seq[Object] {
	return func(yield func(Object) bool) {
		for _, x := range o {
			if !yield(x) {
				break
			}
		}
	}
}

func (o Tuple) Entries() iter.Seq2[Object, Object] {
	return func(yield func(Object, Object) bool) {
		for i, x := range o {
			if !yield(Int(i), x) {
				break
			}
		}
	}
}

type tupleIterator struct {
	t Tuple
	i int
}

func (it *tupleIterator) TypeName() string { return "tuple-iterator" }
func (it *tupleIterator) String() string   { return "<tuple-iterator>" }
func (it *tupleIterator) IsFalsy() bool    { return true }
func (it *tupleIterator) Copy() Object     { return &tupleIterator{t: it.t, i: it.i} }

func (it *tupleIterator) Next(key, value *Object) bool {
	if it.i < len(it.t) {
		if key != nil {
			*key = Int(it.i)
		}
		if value != nil {
			*value = it.t[it.i]
		}
		it.i++
		return true
	}
	return false
}

type Error struct {
	message string
	cause   *Error
}

func (e *Error) Message() string { return e.message }
func (e *Error) Cause() *Error   { return e.cause }

func (e *Error) TypeName() string { return "error" }

func (e *Error) String() string {
	var b strings.Builder
	b.WriteString(e.message)
	if e.cause != nil {
		b.WriteString(": ")
		b.WriteString(e.cause.String())
	}
	return b.String()
}

func (e *Error) IsFalsy() bool { return true }

func (e *Error) Copy() Object {
	var cause *Error
	if e.cause != nil {
		cause = e.cause.Copy().(*Error)
	}
	return &Error{message: e.message, cause: cause}
}

func (e *Error) Hash() uint32 { return murmur32String(e.String()) }

func (e *Error) Compare(op token.Token, rhs Object) (bool, error) {
	y, ok := rhs.(*Error)
	if !ok {
		return false, ErrInvalidOperator
	}
	switch op {
	case token.Equal:
		for x := e; x != nil; x = x.cause {
			if x == y {
				return true, nil
			}
		}
		return false, nil
	case token.NotEqual:
		for x := e; x != nil; x = x.cause {
			if x == y {
				return false, nil
			}
		}
		return true, nil
	}
	return false, ErrInvalidOperator
}

func (e *Error) FieldGet(name string) (res Object, err error) {
	switch name {
	case "message":
		return String(e.message), nil
	case "cause":
		if e.cause != nil {
			return e.cause, nil
		}
		return Undefined, nil
	}
	return nil, ErrNoSuchField
}

// objectPtr represents a free variable.
type objectPtr struct {
	p *Object
}

func (o *objectPtr) String() string   { return "free-var" }
func (o *objectPtr) TypeName() string { return "<free-var>" }
func (o *objectPtr) IsFalsy() bool    { return o.p == nil }
func (o *objectPtr) Copy() Object     { return o }
