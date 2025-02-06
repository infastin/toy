package toy

import (
	"bytes"
	"fmt"
	"iter"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/infastin/toy/hash"
	"github.com/infastin/toy/parser"
	"github.com/infastin/toy/token"
)

// Value represents a value in the Toy Language.
type Value interface {
	// Type returns the type of the value.
	// Types themselves should return nil.
	//
	// Client code should not call this method.
	// Instead, use the standalone Type function.
	Type() ValueType
	// String returns a string representation of the value.
	String() string
	// IsFalsy returns true if the value should be considered as falsy.
	IsFalsy() bool
	// Clone returns a deep-copy of a value.
	Clone() Value
}

// ValueType represents a value type in the Toy Language.
type ValueType interface {
	Value
	// Name returns the name of the type.
	Name() string
}

// Hashable represents a value that is hashable (can be used in Table).
type Hashable interface {
	Value
	// Hash returns a function of x such that Equal(x, y) => Hash(x) == Hash(y).
	Hash() uint64
}

// Freezable represents a value that can be made immutable.
// Types that do not implement Freezable are assumed to be immutable.
type Freezable interface {
	Value
	// Freeze returns an immutable value.
	// Typically it should just make the value immutable and return it,
	// but in some cases (arrays) it may create an immutable copy of the value.
	// If the value is already immutable, the same value must be returned.
	Freeze() Value
	// Immutable returns true if the value is immutable.
	Immutable() bool
}

// A Comparable is a value that defines its own equivalence relation
// and perhaps ordered comparisons.
type Comparable interface {
	Value
	// Compare compares one value to another.
	// The comparison operation must be one of Equal, NotEqual, Less, LessEq, Greater, or GreaterEq.
	// It returns the result of the operation or an error
	// if the operation is not supported or has failed.
	//
	// Client code should not call this method.
	// Instead, use the standalone Compare or Equals functions.
	Compare(op token.Token, rhs Value) (bool, error)
}

// HasBinaryOp represents a value that supports binary operations (excluding comparison operators).
type HasBinaryOp interface {
	Value
	// BinaryOp performs a binary operation on the current value with the provided value.
	// The right parameter indicates whether the current value is the right operand (true)
	// or the left operand (false) in the operation.
	// It returns the result of the operation or an error
	// if the operation is not supported or has failed.
	//
	// Client code should not call this method.
	// Instead, use the standalone BinaryOp function.
	BinaryOp(op token.Token, other Value, right bool) (Value, error)
}

// HasBinaryOp represents a value that supports unary operations.
type HasUnaryOp interface {
	Value
	// UnaryOp returns another value that is the result of a given unary operator.
	// It returns the result of the operation or an error
	// if the operation is not supported or has failed.
	//
	// Client code should not call this method.
	// Instead, use the standalone UnaryOp function.
	UnaryOp(op token.Token) (Value, error)
}

// PropertyAccessible represents a value that supports property access.
type PropertyAccessible interface {
	Value
	// Property returns a value associated with the given key.
	// It the value assigned to the specified key exist, returns it with found = true.
	// Otherwise, returns Nil with found = false.
	// It returns an error if the key is of invalid type or if the operation has failed.
	//
	// Client code should not call this method.
	// Instead, use the standalone Property function.
	Property(key Value) (value Value, found bool, err error)
}

// PropertyAssignable represents a value that supports property access and assignment.
type PropertyAssignable interface {
	PropertyAccessible
	// SetProperty assigns the specified value to the specified key.
	// It returns an error if the key is of invalid type or if the operation has failed.
	//
	// Client code should not call this method.
	// Instead, use the standalone SetProperty function.
	SetProperty(index, value Value) error
}

// Sized represents a value that can report its length or size.
type Sized interface {
	Value
	// Len returns the length or size of the value.
	Len() int
}

// IndexAccessible is a sequence of known length
// that supports efficient random access operation.
type IndexAccessible interface {
	Sized
	// At returns the value at the specified index.
	// The caller must ensure that 0 <= i < Len().
	At(i int) Value
}

// IndexAssignable is a sequence of known length
// that supports efficient random access and assignment operations.
type IndexAssignable interface {
	IndexAccessible
	// SetAt assigns the specified value to the specified index.
	// The caller must ensure that 0 <= i < Len().
	SetAt(i int, value Value) error
}

// A Sliceable is a sequence that can be cut into pieces with the slice operator.
type Sliceable interface {
	IndexAccessible
	// Slice returns a piece of the value.
	// The caller must ensure that the low and high indices are valid.
	Slice(low, high int) Value
}

// Convertible represents a value that can be converted to a value of another type.
type Convertible interface {
	Value
	// Convert takes a pointer to a Value and converts
	// the Convertible value to the type of the provided Value.
	// Returns an error if the conversion has failed.
	//
	// Client code should not call this method.
	// Instead, use the standalone Convert function.
	Convert(p any) error
}

// Callable represents an value that can be called.
type Callable interface {
	Value
	// Call should take an arbitrary number of arguments and return a return
	// value and/or an error, which the VM will consider as a runtime error.
	// If multiple values are to be returned, Call should return Tuple.
	//
	// Client code should not call this method.
	// Instead, use the standalone Call function.
	Call(r *Runtime, args ...Value) (Value, error)
}

// Container represents a value that contains some value(s).
type Container interface {
	Value
	// Contains checks if the specified value is containing within Container.
	// Returns an error if the value is of invalid type or if the operation has failed.
	Contains(value Value) (bool, error)
}

// Iterable represents a value that allows to iterate over it's values.
type Iterable interface {
	Value
	// Elements returns an iterator for traversing over Iterable's values.
	Elements() iter.Seq[Value]
}

// Sequence represents a sequence of values of known length that can be iterated.
type Sequence interface {
	Iterable
	Sized
	// Items returns a slice containing all the elements in the Sequence.
	Items() []Value
}

// KVIterable represents a value that allows to iterate over it's key-value pairs.
type KVIterable interface {
	Value
	// Entries returns an iterator for traversing over KVIterable's entries.
	Entries() iter.Seq2[Value, Value]
}

// Mapping represents an associative data structure of known length that can be iterated.
type Mapping interface {
	KVIterable
	Sized
	// Items returns a slice containing all the entries in the Mapping.
	Items() []Tuple
}

// TypeName returns the type of the value.
func Type(x Value) ValueType {
	if x == nil {
		return NilType
	}
	if typ := x.Type(); typ != nil {
		return typ
	}
	return rootTypeImpl{}
}

// TypeName returns the name of the value's type.
func TypeName(x Value) string {
	if x == nil {
		return "nil"
	}
	if typ := x.Type(); typ != nil {
		return typ.Name()
	}
	return "type"
}

// Hash tries to calculate hash of the given value.
func Hash(x Value) (uint64, error) {
	xh, ok := x.(Hashable)
	if !ok {
		return 0, fmt.Errorf("'%s' is not hashable", TypeName(x))
	}
	return xh.Hash(), nil
}

// Freeze makes the given value immutable.
// If the value type does not implement Freezable,
// Freeze does nothing.
func Freeze(x Value) Value {
	xf, ok := x.(Freezable)
	if ok {
		return xf.Freeze()
	}
	return x
}

// Immutable returns true if a value is immutable.
// All instances of types that do not implement Freezable
// are considered immutable.
func Immutable(x Value) bool {
	m, ok := x.(Freezable)
	if ok {
		return m.Immutable()
	}
	return true
}

// Equal returns whether two values are equal or not.
// It will return an error if the values aren't comparable
// or if the comparison has failed.
func Equal(x, y Value) (bool, error) {
	return Compare(token.Equal, x, y)
}

// Compare compares two values using the given comparison operator.
// If the comparsion between x and y has failed, it will try to compare y and x.
// It will return an error if the values can't be compared using the given operator.
// or if the comparison has failed.
//
// Equality comparsion with NilValue is defined implicitly,
// so types do not need to implement it themselves.
//
// Equality comparison between ValueType is defined implicitly,
// but types that implement ValueType can implement Comparable
// to override this behaviour.
//
// Equality comparsion for two Go-comparable values
// having the same value is defined implicitly.
func Compare(op token.Token, x, y Value) (res bool, err error) {
	if x == Nil || y == Nil {
		eq := (x != Nil) == (y != Nil)
		switch op {
		case token.Equal:
			return eq, nil
		case token.NotEqual:
			return !eq, nil
		}
	}
	if xt, ok := x.(ValueType); ok {
		if yt, ok := y.(ValueType); ok {
			xtc, ok := xt.(Comparable)
			if ok {
				res, err := xtc.Compare(op, yt)
				if err != nil {
					return false, err
				}
				return res, nil
			}
			switch op {
			case token.Equal:
				return xt == yt, nil
			case token.NotEqual:
				return xt != yt, nil
			}
		}
	}
	if xt := reflect.TypeOf(x); xt.Comparable() {
		if yt := reflect.TypeOf(y); xt == yt && x == y {
			switch op {
			case token.Equal:
				return true, nil
			case token.NotEqual:
				return false, nil
			}
		}
	}
	xc, ok := x.(Comparable)
	if ok {
		if res, err = xc.Compare(op, y); err == nil {
			return res, nil
		} else if x.Type() == y.Type() {
			return false, err
		}
	} else {
		err = ErrInvalidOperation
	}
	yc, ok := y.(Comparable)
	if !ok {
		return false, fmt.Errorf("operation '%s %s %s' has failed: %w",
			TypeName(x), op.String(), TypeName(y), err)
	}
	xOp := op
	switch op {
	case token.Less:
		op = token.Greater
	case token.Greater:
		op = token.Less
	case token.LessEq:
		op = token.GreaterEq
	case token.GreaterEq:
		op = token.LessEq
	}
	res, yErr := yc.Compare(op, x)
	if yErr != nil {
		return false, fmt.Errorf("operation '%s %s %s' has failed: %w",
			TypeName(x), xOp.String(), TypeName(y), err)
	}
	return res, nil
}

// BinaryOp performs a binary operation with the given operator.
// If the operation for x and y has failed,
// it will try to perform the same operation for y and x.
// It will return an error if the given binary operation
// can't be performed on the given values or if the operation has failed.
func BinaryOp(op token.Token, x, y Value) (res Value, err error) {
	xb, ok := x.(HasBinaryOp)
	if ok {
		if res, err = xb.BinaryOp(op, y, false); err == nil {
			return res, nil
		} else if x.Type() == y.Type() {
			return nil, err
		}
	} else {
		err = ErrInvalidOperation
	}
	yb, ok := y.(HasBinaryOp)
	if !ok {
		return nil, fmt.Errorf("operation '%s %s %s' has failed: %w",
			TypeName(x), op.String(), TypeName(y), err)
	}
	res, yErr := yb.BinaryOp(op, x, true)
	if yErr != nil {
		return nil, fmt.Errorf("operation '%s %s %s' has failed: %w",
			TypeName(x), op.String(), TypeName(y), err)
	}
	return res, nil
}

// UnaryOp performs an unary operation with the given operator.
// It will return an error if the given unary operation
// can't be performed on the given value or if the operation has failed.
func UnaryOp(op token.Token, x Value) (Value, error) {
	if op == token.Not {
		return Bool(x.IsFalsy()), nil
	}
	xu, ok := x.(HasUnaryOp)
	if !ok {
		return nil, fmt.Errorf("operation '%s%s' has failed: %w",
			op.String(), TypeName(x), ErrInvalidOperation)
	}
	res, err := xu.UnaryOp(op)
	if err != nil {
		return nil, fmt.Errorf("operation '%s%s' has failed: %w",
			op.String(), TypeName(x), err)
	}
	return res, nil
}

// Property retrieves the value associated
// with the specified key from the given value.
//
// If the specified key is Int and the provided value is IndexAccessible,
// it retrieves the value at the specified index using At() method.
//
// Otherwise it checks if the provided value is PropertyAccessible and
// retrieves the value associated with the specified key.
//
// Returns an error if the operation can't be performed
// on the given value or if the operation has failed.
func Property(x, key Value) (value Value, found bool, err error) {
	if i, ok := key.(Int); ok {
		if xi, ok := x.(IndexAccessible); ok {
			if i < 0 || int64(i) >= int64(xi.Len()) {
				return Nil, false, nil
			}
			return xi.At(int(i)), true, nil
		}
	}
	xi, ok := x.(PropertyAccessible)
	if !ok {
		return nil, false, fmt.Errorf("'%s' is not property accessible", TypeName(x))
	}
	res, found, err := xi.Property(key)
	if err != nil {
		return nil, false, fmt.Errorf("failed to retrieve value of type '%s' from '%s': %w",
			TypeName(key), TypeName(x), err)
	}
	return res, found, nil
}

// SetProperty assigns a value to the specified key in the value.
//
// If the specified key is Int and the provided value is IndexAssignable,
// it assigns the specified value to the specified index using SetAt() method.
//
// Otherwise it checks if the provided value is PropertyAssignable and
// assigns the specified value to the specified key.
//
// Returns an error if the operation can't be performed
// on the given value or if the operation has failed.
func SetProperty(x, key, value Value) error {
	if i, ok := key.(Int); ok {
		if xi, ok := x.(IndexAssignable); ok {
			if i < 0 {
				return fmt.Errorf("negative index: %d", i)
			}
			n := xi.Len()
			if int64(i) >= int64(n) {
				return fmt.Errorf("index %d out of range [:%d]", i, n)
			}
			if err := xi.SetAt(int(i), value); err != nil {
				return fmt.Errorf("failed to assign '%s' to '%s' in '%s': %w",
					TypeName(key), TypeName(value), TypeName(x), err)
			}
			return nil
		}
	}
	xi, ok := x.(PropertyAssignable)
	if !ok {
		return fmt.Errorf("'%s' is not property assignable", TypeName(x))
	}
	if err := xi.SetProperty(key, value); err != nil {
		return fmt.Errorf("failed to assign '%s' to '%s' in '%s': %w",
			TypeName(key), TypeName(value), TypeName(x), err)
	}
	return nil
}

// Len returns the size/length of the given value.
// Returns -1 if the value doesn't implement Sized interface.
func Len(x Value) int {
	xs, ok := x.(Sized)
	if !ok {
		return -1
	}
	return xs.Len()
}

// Slice performs a slice operation on a Sliceable value.
// Returns an error if the slice operation can't be performed
// on the given value or if the operation has failed.
func Slice(x Value, low, high int) (Value, error) {
	xs, ok := x.(Sliceable)
	if !ok {
		return nil, fmt.Errorf("'%s' is not sliceable", TypeName(x))
	}
	n := xs.Len()
	if low < 0 || high < 0 {
		neg := high
		if low < 0 {
			neg = low
		}
		return nil, fmt.Errorf("negative slice index: %d", neg)
	}
	if low > n {
		return nil, fmt.Errorf("slice bounds out of range [%d:%d]", low, n)
	}
	if high > n {
		return nil, fmt.Errorf("slice bounds out of range [%d:%d] with len %d", low, high, n)
	}
	if low > high {
		return nil, fmt.Errorf("invalid slice indices: %d > %d", low, high)
	}
	return xs.Slice(low, high), nil
}

// Convert converts the provided value to the one pointed by the pointer.
//
// For String conversions, it checks if the value is already a String,
// and if not, it attempts to convert it using Convertible interface.
// If the provided value doesn't implement Convertible it uses String() method
// to convert the value to String.
//
// For Bool conversions, it uses IsFalsy() to convert the value to Bool.
//
// If the value type is the same as of the pointer value,
// no conversion is performed and the value is assigned to the value instead.
//
// Otherwise, it attempts to convert it using the Convertible interface.
//
// Returns an error if the conversion is not possible
// or the conversion has failed.
func Convert[T Value](p *T, v Value) (err error) {
	switch p := any(p).(type) {
	case *String:
		if s, ok := v.(String); ok {
			*p = s
			return nil
		}
		if c, ok := v.(Convertible); ok {
			if err := c.Convert(p); err == nil {
				return nil
			}
		}
		*p = String(v.String())
	case *Bool:
		*p = Bool(!v.IsFalsy())
	case *T:
		if t, ok := v.(T); ok {
			*p = t
			return nil
		}
		c, ok := v.(Convertible)
		if !ok {
			return fmt.Errorf("'%s' is not convertible", TypeName(v))
		}
		if err := c.Convert(p); err != nil {
			// It should be safe to call TypeName on potentially nil value.
			return fmt.Errorf("failed to convert '%s' to '%s': %w",
				TypeName(v), TypeName(*p), err)
		}
	}
	return nil
}

// Call calls a Callable value.
// Returns an error if the value can't be called
// or if the call returned an error.
func Call(r *Runtime, fn Value, args ...Value) (Value, error) {
	callable, ok := fn.(Callable)
	if !ok {
		return nil, fmt.Errorf("'%s' is not callable", TypeName(fn))
	}
	ret, err := r.safeCall(callable, args)
	if err != nil {
		return nil, fmt.Errorf("error during call to '%s': %w",
			TypeName(callable), err)
	}
	return ret, nil
}

// rootTypeImpl represents the root Type.
type rootTypeImpl struct{}

func (rootTypeImpl) Type() ValueType { return rootTypeImpl{} }
func (rootTypeImpl) String() string  { return "<type>" }
func (rootTypeImpl) IsFalsy() bool   { return false }
func (rootTypeImpl) Clone() Value    { return rootTypeImpl{} }
func (rootTypeImpl) Name() string    { return "type" }

func (rootTypeImpl) Call(r *Runtime, args ...Value) (Value, error) {
	if len(args) != 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	return Type(args[0]), nil
}

// typeImpl represents a default ValueType implementation.
type typeImpl[T Value] struct {
	name string
	fn   CallableFunc
}

// NewType create a new type.
// If constructor is nil, the default constructor will be used.
func NewType[T Value](name string, constructor CallableFunc) ValueType {
	fn := constructor
	if fn == nil {
		fn = func(r *Runtime, args ...Value) (Value, error) {
			if len(args) != 1 {
				return nil, &WrongNumArgumentsError{
					WantMin: 1,
					WantMax: 1,
					Got:     len(args),
				}
			}
			var res T
			if err := Convert(&res, args[0]); err != nil {
				return nil, err
			}
			return res, nil
		}
	}
	return &typeImpl[T]{
		name: name,
		fn:   fn,
	}
}

func (t *typeImpl[T]) Type() ValueType                               { return nil }
func (t *typeImpl[T]) String() string                                { return fmt.Sprintf("<%s>", t.name) }
func (t *typeImpl[T]) IsFalsy() bool                                 { return false }
func (t *typeImpl[T]) Clone() Value                                  { return t }
func (t *typeImpl[T]) Name() string                                  { return t.name }
func (t *typeImpl[T]) Call(r *Runtime, args ...Value) (Value, error) { return t.fn(r, args...) }

// NilValue represents a nil value.
type NilValue byte

// NilType is the type of NilValue.
var NilType = NewType[NilValue]("nil", nil)

const Nil = NilValue(0)

func (v NilValue) Type() ValueType { return NilType }
func (v NilValue) String() string  { return "<nil>" }
func (v NilValue) IsFalsy() bool   { return true }
func (v NilValue) Clone() Value    { return v }

// Bool represents a boolean value.
type Bool bool

// BoolType is the type of Bool.
var BoolType = NewType[Bool]("bool", nil)

const (
	True  = Bool(true)
	False = Bool(false)
)

func (v Bool) String() string {
	if v {
		return "true"
	}
	return "false"
}

func (v Bool) Type() ValueType { return BoolType }
func (v Bool) IsFalsy() bool   { return !bool(v) }
func (v Bool) Clone() Value    { return v }

func (v Bool) Hash() uint64 {
	if v {
		return 1
	}
	return 0
}

func (v Bool) Compare(op token.Token, rhs Value) (bool, error) {
	y, ok := rhs.(Bool)
	if !ok {
		return false, ErrInvalidOperation
	}
	switch op {
	case token.Equal:
		return v == y, nil
	case token.NotEqual:
		return v != y, nil
	}
	return false, ErrInvalidOperation
}

func (v Bool) Convert(p any) error {
	switch p := p.(type) {
	case *Float:
		if v {
			*p = 1.0
		} else {
			*p = 0.0
		}
		return nil
	case *Int:
		if v {
			*p = 1
		} else {
			*p = 0
		}
		return nil
	}
	return ErrNotConvertible
}

// Float represents a floating point number value.
type Float float64

// FloatType is the type of Float.
var FloatType = NewType[Float]("float", nil)

func (v Float) Type() ValueType { return FloatType }
func (v Float) String() string  { return strconv.FormatFloat(float64(v), 'g', -1, 64) }
func (v Float) IsFalsy() bool   { return math.IsNaN(float64(v)) }
func (v Float) Clone() Value    { return v }

func (v Float) Hash() uint64 {
	if float64(v) == math.Trunc(float64(v)) && float64(v) <= float64(math.MaxInt64) {
		return hash.Int64(int64(v))
	}
	return hash.Float64(float64(v))
}

func (v Float) Convert(p any) error {
	i, ok := p.(*Int)
	if !ok {
		return ErrNotConvertible
	}
	*i = Int(v)
	return nil
}

func (v Float) Compare(op token.Token, rhs Value) (bool, error) {
	switch y := rhs.(type) {
	case Float:
		switch op {
		case token.Equal:
			return v == y, nil
		case token.NotEqual:
			return v != y, nil
		case token.Less:
			return v < y, nil
		case token.Greater:
			return v > y, nil
		case token.LessEq:
			return v <= y, nil
		case token.GreaterEq:
			return v >= y, nil
		}
	case Int:
		switch op {
		case token.Equal:
			return v == Float(y), nil
		case token.NotEqual:
			return v != Float(y), nil
		case token.Less:
			return v < Float(y), nil
		case token.Greater:
			return v > Float(y), nil
		case token.LessEq:
			return v <= Float(y), nil
		case token.GreaterEq:
			return v >= Float(y), nil
		}
	}
	return false, ErrInvalidOperation
}

func (v Float) BinaryOp(op token.Token, other Value, right bool) (Value, error) {
	switch y := other.(type) {
	case Float:
		switch op {
		case token.Add:
			return v + y, nil
		case token.Sub:
			return v - y, nil
		case token.Mul:
			return v * y, nil
		case token.Quo:
			return v / y, nil
		}
	case Int:
		switch op {
		case token.Add:
			return v + Float(y), nil
		case token.Sub:
			if right {
				return Float(y) - v, nil
			}
			return v - Float(y), nil
		case token.Mul:
			return v * Float(y), nil
		case token.Quo:
			if right {
				return Float(y) / v, nil
			}
			return v / Float(y), nil
		}
	}
	return nil, ErrInvalidOperation
}

func (v Float) UnaryOp(op token.Token) (Value, error) {
	switch op {
	case token.Add:
		return v, nil
	case token.Sub:
		return -v, nil
	}
	return nil, ErrInvalidOperation
}

// Int represents an integer value.
type Int int64

// IntType is the type of Int.
var IntType = NewType[Int]("int", nil)

func (v Int) Type() ValueType { return IntType }
func (v Int) String() string  { return strconv.FormatInt(int64(v), 10) }
func (v Int) IsFalsy() bool   { return v == 0 }
func (v Int) Clone() Value    { return v }
func (v Int) Hash() uint64    { return hash.Int64(int64(v)) }

func (v Int) Convert(p any) error {
	switch p := p.(type) {
	case *Float:
		*p = Float(v)
		return nil
	case *Char:
		*p = Char(v)
		return nil
	}
	return ErrNotConvertible
}

func (v Int) Compare(op token.Token, rhs Value) (bool, error) {
	switch y := rhs.(type) {
	case Int:
		switch op {
		case token.Equal:
			return v == y, nil
		case token.NotEqual:
			return v != y, nil
		case token.Less:
			return v < y, nil
		case token.Greater:
			return v > y, nil
		case token.LessEq:
			return v <= y, nil
		case token.GreaterEq:
			return v >= y, nil
		}
	case Float:
		switch op {
		case token.Equal:
			return Float(v) == y, nil
		case token.NotEqual:
			return Float(v) != y, nil
		case token.Less:
			return Float(v) < y, nil
		case token.Greater:
			return Float(v) > y, nil
		case token.LessEq:
			return Float(v) <= y, nil
		case token.GreaterEq:
			return Float(v) >= y, nil
		}
	case Char:
		switch op {
		case token.Equal:
			return v == Int(y), nil
		case token.NotEqual:
			return v != Int(y), nil
		case token.Less:
			return v < Int(y), nil
		case token.Greater:
			return v > Int(y), nil
		case token.LessEq:
			return v <= Int(y), nil
		case token.GreaterEq:
			return v >= Int(y), nil
		}
	}
	return false, ErrInvalidOperation
}

func (v Int) BinaryOp(op token.Token, other Value, right bool) (Value, error) {
	switch y := other.(type) {
	case Int:
		switch op {
		case token.Add:
			return v + y, nil
		case token.Sub:
			return v - y, nil
		case token.Mul:
			return v * y, nil
		case token.Quo:
			if y == 0 {
				return nil, ErrDivisionByZero
			}
			return v / y, nil
		case token.Rem:
			return v % y, nil
		case token.And:
			return v & y, nil
		case token.Or:
			return v | y, nil
		case token.Xor:
			return v ^ y, nil
		case token.AndNot:
			return v &^ y, nil
		case token.Shl:
			return v << y, nil
		case token.Shr:
			return v >> y, nil
		}
	case Float:
		switch op {
		case token.Add:
			return Float(v) + y, nil
		case token.Sub:
			if right {
				return y - Float(v), nil
			}
			return Float(v) - y, nil
		case token.Mul:
			return Float(v) * y, nil
		case token.Quo:
			if right {
				return y / Float(v), nil
			}
			return Float(v) / y, nil
		}
	case Char:
		switch op {
		case token.Add:
			return Char(v) + y, nil
		case token.Sub:
			if right {
				return y - Char(v), nil
			}
			return Char(v) - y, nil
		}
	}
	return nil, ErrInvalidOperation
}

func (v Int) UnaryOp(op token.Token) (Value, error) {
	switch op {
	case token.Add:
		return v, nil
	case token.Sub:
		return -v, nil
	case token.Xor:
		return ^v, nil
	}
	return nil, ErrInvalidOperation
}

// String represents a string value.
type String string

// StringType is the type of String.
var StringType = NewType[String]("string", nil)

func (v String) Type() ValueType { return StringType }
func (v String) String() string  { return strconv.Quote(string(v)) }
func (v String) IsFalsy() bool   { return len(v) == 0 }
func (v String) Clone() Value    { return v }
func (v String) Hash() uint64    { return hash.String(string(v)) }

func (v String) Len() int { return utf8.RuneCountInString(string(v)) }

func (v String) At(idx int) Value {
	for i, r := range v {
		if int64(i) == int64(idx) {
			return Char(r)
		}
	}
	return Nil // should not happen
}

func (v String) Slice(low, high int) Value {
	rs := []rune(v)
	return String(rs[low:high])
}

func (v String) Convert(p any) error {
	b, ok := p.(*Bytes)
	if !ok {
		return ErrNotConvertible
	}
	*b = []byte(v)
	return nil
}

func (v String) Compare(op token.Token, rhs Value) (bool, error) {
	y, ok := rhs.(String)
	if !ok {
		return false, ErrInvalidOperation
	}
	switch op {
	case token.Equal:
		return v == y, nil
	case token.NotEqual:
		return v != y, nil
	case token.Less:
		return v < y, nil
	case token.Greater:
		return v > y, nil
	case token.LessEq:
		return v <= y, nil
	case token.GreaterEq:
		return v >= y, nil
	}
	return false, ErrInvalidOperation
}

func (v String) BinaryOp(op token.Token, other Value, right bool) (Value, error) {
	switch op {
	case token.Add:
		switch y := other.(type) {
		case String:
			return v + y, nil
		case Char:
			if right {
				return String(y) + v, nil
			}
			return v + String(y), nil
		}
	}
	return nil, ErrInvalidOperation
}

func (v String) Contains(value Value) (bool, error) {
	switch x := value.(type) {
	case String:
		return strings.Contains(string(v), string(x)), nil
	case Char:
		return strings.ContainsRune(string(v), rune(x)), nil
	default:
		return false, &InvalidValueTypeError{
			Want: "string or char",
			Got:  TypeName(value),
		}
	}
}

func (v String) Elements() iter.Seq[Value] {
	return func(yield func(Value) bool) {
		for _, r := range v {
			if !yield(Char(r)) {
				return
			}
		}
	}
}

// Bytes represents a byte array.
type Bytes []byte

// Bytes is the type of Bytes.
var BytesType = NewType[Bytes]("bytes", func(_ *Runtime, args ...Value) (Value, error) {
	if len(args) != 1 {
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 1,
			Got:     len(args),
		}
	}
	if i, ok := args[0].(Int); ok {
		return make(Bytes, i), nil
	}
	var b Bytes
	if err := Convert(&b, args[0]); err != nil {
		return nil, err
	}
	return b, nil
})

func (v Bytes) Type() ValueType { return BytesType }
func (v Bytes) String() string  { return fmt.Sprintf("bytes(%q)", []byte(v)) }
func (v Bytes) IsFalsy() bool   { return len(v) == 0 }
func (v Bytes) Clone() Value    { return slices.Clone(v) }
func (v Bytes) Hash() uint64    { return hash.Bytes(v) }

func (v Bytes) Len() int                  { return len(v) }
func (v Bytes) At(i int) Value            { return Int(v[i]) }
func (v Bytes) Slice(low, high int) Value { return v[low:high] }

func (v Bytes) Convert(p any) error {
	s, ok := p.(*String)
	if !ok {
		return ErrNotConvertible
	}
	*s = String(v)
	return nil
}

func (v Bytes) Compare(op token.Token, rhs Value) (bool, error) {
	y, ok := rhs.(Bytes)
	if !ok {
		return false, ErrInvalidOperation
	}
	switch op {
	case token.Equal:
		return bytes.Equal(v, y), nil
	case token.NotEqual:
		return !bytes.Equal(v, y), nil
	case token.Less:
		return bytes.Compare(v, y) < 0, nil
	case token.Greater:
		return bytes.Compare(v, y) > 0, nil
	case token.LessEq:
		return bytes.Compare(v, y) <= 0, nil
	case token.GreaterEq:
		return bytes.Compare(v, y) >= 0, nil
	}
	return false, ErrInvalidOperation
}

func (v Bytes) BinaryOp(op token.Token, other Value, right bool) (Value, error) {
	switch op {
	case token.Add:
		switch y := other.(type) {
		case Bytes:
			return append(v, y...), nil
		}
	}
	return nil, ErrInvalidOperation
}

func (v Bytes) Contains(value Value) (bool, error) {
	switch x := value.(type) {
	case Bytes:
		return bytes.Contains(v, x), nil
	case Char:
		return bytes.ContainsRune(v, rune(x)), nil
	case Int:
		return bytes.IndexByte(v, byte(x)) != -1, nil
	default:
		return false, &InvalidValueTypeError{
			Want: "bytes, char or int",
			Got:  TypeName(value),
		}
	}
}

func (v Bytes) Elements() iter.Seq[Value] {
	return func(yield func(Value) bool) {
		for _, b := range v {
			if !yield(Int(b)) {
				return
			}
		}
	}
}

// Char represents a character value.
type Char rune

// CharType is the type of Char.
var CharType = NewType[Char]("char", nil)

func (v Char) Type() ValueType { return CharType }
func (v Char) String() string  { return strconv.QuoteRune(rune(v)) }
func (v Char) IsFalsy() bool   { return v == 0 }
func (v Char) Clone() Value    { return v }
func (v Char) Hash() uint64    { return hash.Int32(int32(v)) }

func (v Char) Convert(p any) error {
	switch p := p.(type) {
	case *Int:
		*p = Int(v)
	case *String:
		*p = String(v)
	default:
		return ErrNotConvertible
	}
	return nil
}

func (v Char) Compare(op token.Token, rhs Value) (bool, error) {
	switch y := rhs.(type) {
	case Char:
		switch op {
		case token.Equal:
			return v == y, nil
		case token.NotEqual:
			return v != y, nil
		case token.Less:
			return v < y, nil
		case token.Greater:
			return v > y, nil
		case token.LessEq:
			return v <= y, nil
		case token.GreaterEq:
			return v >= y, nil
		}
	case Int:
		switch op {
		case token.Equal:
			return Int(v) == y, nil
		case token.NotEqual:
			return Int(v) != y, nil
		case token.Less:
			return Int(v) < y, nil
		case token.Greater:
			return Int(v) > y, nil
		case token.LessEq:
			return Int(v) <= y, nil
		case token.GreaterEq:
			return Int(v) >= y, nil
		}
	}
	return false, ErrInvalidOperation
}

func (v Char) BinaryOp(op token.Token, other Value, right bool) (Value, error) {
	switch y := other.(type) {
	case Char:
		switch op {
		case token.Add:
			return v + y, nil
		case token.Sub:
			return v - y, nil
		}
	case Int:
		switch op {
		case token.Add:
			return v + Char(y), nil
		case token.Sub:
			if right {
				return Char(y) - v, nil
			}
			return v - Char(y), nil
		}
	}
	return nil, ErrInvalidOperation
}

// Array represents a array of values.
// Array is mutable by default.
// Freeze will create an immutable copy if the array is mutable,
// otherwise the same array will be returned.
type Array struct {
	elems     []Value
	immutable bool
	itercount uint64 // number of active iterators (ignored if frozen)
}

// ArrayType is the type of Array.
var ArrayType = NewType[*Array]("array", func(_ *Runtime, args ...Value) (Value, error) {
	switch len(args) {
	case 1:
		if _, isInt := args[0].(Int); !isInt {
			var arr *Array
			if err := Convert(&arr, args[0]); err != nil {
				return nil, err
			}
			return arr, nil
		}
		fallthrough
	case 2:
		var (
			size  int
			value Value = Nil
		)
		if err := UnpackArgs(args, "size", &size, "value?", &value); err != nil {
			return nil, err
		}
		if size < 0 {
			return nil, fmt.Errorf("negative array size: %d", size)
		}
		arr := make([]Value, int(size))
		for i := range arr {
			arr[i] = value
		}
		return NewArray(arr), nil
	default:
		return nil, &WrongNumArgumentsError{
			WantMin: 1,
			WantMax: 2,
			Got:     len(args),
		}
	}
})

// NewArray returns an array containing the specified elements.
// Callers should not subsequently modify elems.
func NewArray(elems []Value) *Array {
	return &Array{
		elems:     elems,
		immutable: false,
		itercount: 0,
	}
}

func (v *Array) Type() ValueType { return ArrayType }

func (v *Array) String() string {
	var b strings.Builder
	b.WriteByte('[')
	for i, v := range v.elems {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(v.String())
	}
	b.WriteByte(']')
	return b.String()
}

func (v *Array) IsFalsy() bool { return len(v.elems) == 0 }

func (v *Array) Clone() Value {
	if v.immutable {
		return v
	}
	elems := make([]Value, 0, len(v.elems))
	for _, elem := range v.elems {
		elems = append(elems, elem.Clone())
	}
	return &Array{elems: elems, immutable: false}
}

func (v *Array) Freeze() Value {
	if v.immutable {
		return v
	}
	elems := make([]Value, 0, len(v.elems))
	for _, elem := range v.elems {
		elems = append(elems, Freeze(elem))
	}
	return &Array{elems: elems, immutable: true}
}

func (v *Array) Immutable() bool { return v.immutable }
func (v *Array) Len() int        { return len(v.elems) }
func (v *Array) At(i int) Value  { return v.elems[i] }
func (v *Array) Items() []Value  { return v.elems }

func (v *Array) Slice(low, high int) Value {
	return &Array{
		elems:     v.elems[low:high],
		immutable: v.immutable,
		itercount: 0,
	}
}

func (v *Array) SetAt(i int, value Value) error {
	if err := v.checkMutable("assign to element of"); err != nil {
		return err
	}
	v.elems[i] = value
	return nil
}

func (v *Array) Append(xs ...Value) error {
	if err := v.checkMutable("append to"); err != nil {
		return err
	}
	v.elems = append(v.elems, xs...)
	return nil
}

func (v *Array) Clear() error {
	if err := v.checkMutable("clear"); err != nil {
		return err
	}
	clear(v.elems)
	v.elems = v.elems[:0]
	return nil
}

func (v *Array) Compare(op token.Token, rhs Value) (bool, error) {
	y, ok := rhs.(*Array)
	if !ok {
		return false, ErrInvalidOperation
	}
	switch op {
	case token.Equal:
		if len(v.elems) != len(y.elems) {
			return false, nil
		}
		for i := range len(v.elems) {
			if eq, err := Equal(v.elems[i], y.elems[i]); err != nil {
				return false, err
			} else if !eq {
				return false, nil
			}
		}
		return true, nil
	case token.NotEqual:
		if len(v.elems) != len(y.elems) {
			return true, nil
		}
		for i := range len(v.elems) {
			if eq, err := Equal(v.elems[i], y.elems[i]); err != nil {
				return false, err
			} else if !eq {
				return true, nil
			}
		}
		return false, nil
	}
	return false, ErrInvalidOperation
}

func (v *Array) BinaryOp(op token.Token, other Value, right bool) (Value, error) {
	switch op {
	case token.Add:
		switch y := other.(type) {
		case *Array:
			return &Array{
				elems:     slices.Concat(v.elems, y.elems),
				immutable: v.immutable,
				itercount: 0,
			}, nil
		}
	}
	return nil, ErrInvalidOperation
}

func (v *Array) Contains(value Value) (bool, error) {
	for _, obj := range v.elems {
		if eq, err := Equal(obj, value); err != nil {
			return false, err
		} else if eq {
			return true, nil
		}
	}
	return false, nil
}

func (v *Array) Elements() iter.Seq[Value] {
	return func(yield func(Value) bool) {
		if !v.immutable {
			v.itercount++
			defer func() { v.itercount-- }()
		}
		for _, x := range v.elems {
			if !yield(x) {
				break
			}
		}
	}
}

// checkMutable reports an error if the array should not be mutated.
// verb+" immutable array" should describe the operation.
func (v *Array) checkMutable(verb string) error {
	if v.immutable {
		return fmt.Errorf("cannot %s immutable array", verb)
	}
	if v.itercount > 0 {
		return fmt.Errorf("cannot %s array during iteration", verb)
	}
	return nil
}

// Table represents an associated data structure that maps keys to values.
// Table is mutable by default. Freeze will make it immutable.
type Table struct {
	ht hashtable
}

// MapType is the type of Map.
var TableType = NewType[*Table]("table", nil)

// NewTable returns a Table with initial space for
// at least size insertions before rehashing.
func NewTable(size int) *Table {
	t := new(Table)
	t.ht.init(size)
	return t
}

func (v *Table) Type() ValueType { return TableType }

func (v *Table) String() string {
	var b strings.Builder
	b.WriteByte('{')
	for key, value := range v.ht.entries() {
		if b.Len() != 1 {
			b.WriteString(", ")
		}
		if keyStr, ok := key.(String); ok && parser.IsIdent(string(keyStr)) {
			b.WriteString(string(keyStr))
		} else {
			b.WriteByte('[')
			b.WriteString(key.String())
			b.WriteByte(']')
		}
		b.WriteString(": ")
		b.WriteString(value.String())
	}
	b.WriteByte('}')
	return b.String()
}

func (v *Table) IsFalsy() bool { return v.ht.len == 0 }

func (v *Table) Clone() Value {
	t := new(Table)
	t.ht.init(v.Len())
	t.ht.cloneAll(&v.ht)
	return t
}

func (v *Table) Freeze() Value {
	v.ht.freeze()
	return v
}

func (v *Table) Immutable() bool { return v.ht.immutable }
func (v *Table) Len() int        { return int(v.ht.len) }

func (v *Table) Compare(op token.Token, rhs Value) (bool, error) {
	y, ok := rhs.(*Table)
	if !ok {
		return false, ErrInvalidOperation
	}
	switch op {
	case token.Equal:
		eq, err := v.ht.equal(&y.ht)
		if err != nil {
			return false, err
		}
		return eq, nil
	case token.NotEqual:
		eq, err := v.ht.equal(&y.ht)
		if err != nil {
			return false, err
		}
		return !eq, nil
	}
	return false, ErrInvalidOperation
}

func (v *Table) Property(key Value) (res Value, found bool, err error) { return v.ht.lookup(key) }
func (v *Table) SetProperty(key, value Value) (err error)              { return v.ht.insert(key, value) }
func (v *Table) Contains(key Value) (bool, error)                      { return v.ht.contains(key) }
func (v *Table) Elements() iter.Seq[Value]                             { return v.ht.elements() }
func (v *Table) Entries() iter.Seq2[Value, Value]                      { return v.ht.entries() }

func (v *Table) Delete(key Value) (Value, error) { return v.ht.delete(key) }
func (v *Table) Clear() error                    { return v.ht.clear() }
func (v *Table) Keys() []Value                   { return v.ht.keys() }
func (v *Table) Values() []Value                 { return v.ht.values() }
func (v *Table) Items() []Tuple                  { return v.ht.items() }

// Tuple represents a tuple of values.
type Tuple []Value

// TupleType is the type of Tuple.
var TupleType = NewType[Tuple]("tuple", func(_ *Runtime, args ...Value) (Value, error) {
	if len(args) == 1 {
		var t Tuple
		if err := Convert(&t, args[0]); err == nil {
			return t, nil
		}
	}
	return Tuple(args), nil
})

func (v Tuple) Type() ValueType { return TupleType }

func (v Tuple) String() string {
	var b strings.Builder
	b.WriteString("tuple(")
	for i, v := range v {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(v.String())
	}
	b.WriteByte(')')
	return b.String()
}

func (v Tuple) IsFalsy() bool { return len(v) == 0 }

func (v Tuple) Clone() Value {
	t := make(Tuple, 0, len(v))
	for _, elem := range v {
		t = append(t, elem.Clone())
	}
	return t
}

func (v Tuple) Len() int       { return len(v) }
func (v Tuple) At(i int) Value { return v[i] }

func (v Tuple) Slice(low, high int) Value { return v[low:high] }
func (v Tuple) Items() []Value            { return v }

func (v Tuple) SetAt(i int, value Value) error {
	v[i] = value
	return nil
}

func (v Tuple) Compare(op token.Token, rhs Value) (bool, error) {
	y, ok := rhs.(Tuple)
	if !ok {
		return false, ErrInvalidOperation
	}
	switch op {
	case token.Equal:
		if len(v) != len(y) {
			return false, nil
		}
		for i := range v.Len() {
			if eq, err := Equal(v[i], y[i]); err != nil {
				return false, err
			} else if !eq {
				return false, nil
			}
		}
	case token.NotEqual:
		if len(v) != len(y) {
			return true, nil
		}
		for i := range len(v) {
			if eq, err := Equal(v[i], y[i]); err != nil {
				return false, err
			} else if !eq {
				return true, nil
			}
		}
	}
	return false, ErrInvalidOperation
}

func (v Tuple) BinaryOp(op token.Token, other Value, right bool) (Value, error) {
	switch op {
	case token.Add:
		switch y := other.(type) {
		case Tuple:
			return slices.Concat(v, y), nil
		}
	}
	return nil, ErrInvalidOperation
}

func (v Tuple) Contains(value Value) (bool, error) {
	for _, obj := range v {
		if eq, err := Equal(obj, value); err != nil {
			return false, err
		} else if eq {
			return true, nil
		}
	}
	return false, nil
}

func (v Tuple) Elements() iter.Seq[Value] {
	return func(yield func(Value) bool) {
		for _, x := range v {
			if !yield(x) {
				break
			}
		}
	}
}

// Range represents a range value.
type Range struct {
	start, stop, step int
}

func NewRange(start, stop, step int) *Range {
	if step <= 0 {
		panic(fmt.Sprintf("invalid range step: must be > 0, got %d", step))
	}
	return &Range{
		start: start,
		stop:  stop,
		step:  step,
	}
}

// RangeType is the type of rangeValue.
var RangeType = NewType[*Range]("range", func(_ *Runtime, args ...Value) (Value, error) {
	var (
		start, stop int
		step        = 1
	)
	if err := UnpackArgs(args,
		"start", &start,
		"stop", &stop,
		"step?", &step,
	); err != nil {
		return nil, err
	}
	if step <= 0 {
		return nil, fmt.Errorf("invalid range step: must be > 0, got %d", step)
	}
	return &Range{
		start: start,
		stop:  stop,
		step:  step,
	}, nil
})

func (v *Range) Type() ValueType { return RangeType }

func (v *Range) String() string {
	var b strings.Builder
	b.WriteString("range(")
	b.WriteString(strconv.Itoa(v.start))
	b.WriteString(", ")
	b.WriteString(strconv.Itoa(v.stop))
	if v.step != 1 {
		b.WriteString(", ")
		b.WriteString(strconv.Itoa(v.step))
	}
	b.WriteByte(')')
	return b.String()
}

func (v *Range) IsFalsy() bool { return v.Len() == 0 }

func (v *Range) Clone() Value {
	return &Range{
		start: v.start,
		stop:  v.stop,
		step:  v.step,
	}
}

func (v *Range) Len() int {
	if v.start <= v.stop {
		return ((v.stop - v.start - 1) / v.step) + 1
	}
	return ((v.start - v.stop - 1) / v.step) + 1
}

func (v *Range) At(i int) Value {
	if v.start <= v.stop {
		return Int(v.start + i*v.step)
	}
	return Int(v.start - i*v.step)
}

func (v *Range) Slice(low, high int) Value {
	if v.start <= v.stop {
		return &Range{
			start: v.start + low*v.step,
			stop:  v.start + high*v.step,
			step:  v.step,
		}
	}
	return &Range{
		start: v.start - low*v.step,
		stop:  v.start - high*v.step,
		step:  v.step,
	}
}

func (v *Range) Items() []Value {
	var elems []Value
	if v.start <= v.stop {
		elems = make([]Value, 0, ((v.stop-v.start-1)/v.step)+1)
		for i := v.start; i < v.stop; i += v.step {
			elems = append(elems, Int(i))
		}
	} else {
		elems = make([]Value, 0, ((v.start-v.stop-1)/v.step)+1)
		for i := v.start; i > v.stop; i -= v.step {
			elems = append(elems, Int(i))
		}
	}
	return elems
}

func (v *Range) Contains(value Value) (bool, error) {
	other, ok := value.(Int)
	if !ok {
		return false, &InvalidArgumentTypeError{
			Want: "int",
			Got:  TypeName(value),
		}
	}
	if v.start <= v.stop {
		return v.start <= int(other) && v.stop > int(other), nil
	} else {
		return v.start > int(other) && v.stop <= int(other), nil
	}
}

func (v *Range) Start() int { return v.start }
func (v *Range) Stop() int  { return v.stop }
func (v *Range) Step() int  { return v.step }

func (v *Range) Elements() iter.Seq[Value] {
	step := v.step
	if v.start > v.stop {
		step = -step
	}
	return func(yield func(Value) bool) {
		i := 0
		for range v.Len() {
			if !yield(Int(i)) {
				return
			}
			i += step
		}
	}
}

// iterator represents an iterator inside the runtime.
type iterator struct {
	next func() (Value, Value, bool)
	stop func()
}

func newIterator(it iter.Seq[Value]) *iterator {
	return newIterator2(func(yield func(Value, Value) bool) {
		i := 0
		for value := range it {
			if !yield(Int(i), value) {
				break
			}
			i++
		}
	})
}

func newIterator2(it iter.Seq2[Value, Value]) *iterator {
	next, stop := iter.Pull2(it)
	return &iterator{next: next, stop: stop}
}

func (v *iterator) Type() ValueType { return nil }
func (v *iterator) String() string  { return "<iterator>" }
func (v *iterator) IsFalsy() bool   { return true }
func (v *iterator) Clone() Value    { return v }

// valuePtr represents a free variable inside the runtime.
type valuePtr struct {
	p *Value
}

func (v *valuePtr) Type() ValueType { return nil }
func (v *valuePtr) String() string  { return "<free-var>" }
func (v *valuePtr) IsFalsy() bool   { return true }
func (v *valuePtr) Clone() Value    { return v }

// splatSequence represents a sequence in the runtime
// that supposed to be splat.
type splatSequence struct {
	s Sequence
}

func (v *splatSequence) Type() ValueType { return nil }
func (v *splatSequence) String() string  { return "<splat-sequence>" }
func (v *splatSequence) IsFalsy() bool   { return true }
func (v *splatSequence) Clone() Value    { return v }

// splatMapping represents a mapping in the runtime
// that supposed to be splat.
type splatMapping struct {
	m Mapping
}

func (v *splatMapping) Type() ValueType { return nil }
func (v *splatMapping) String() string  { return "<splat-mapping>" }
func (v *splatMapping) IsFalsy() bool   { return true }
func (v *splatMapping) Clone() Value    { return v }
