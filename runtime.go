package toy

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/infastin/toy/bytecode"
	"github.com/infastin/toy/token"
)

// deferredCall represents a deferred function call.
type deferredCall struct {
	fn   Callable
	args []Value
	pos  token.Pos
}

// frame represents a function call frame.
type frame struct {
	fn          *CompiledFunction
	freeVars    []*valuePtr
	ip          int
	basePointer int
	deferred    []*deferredCall
	curDefer    *deferredCall
}

// Runtime is a virtual machine that executes the bytecode.
type Runtime struct {
	constants   []Value
	stack       []Value
	sp          int
	globals     []Value
	fileSet     *token.FileSet
	frames      []frame
	framesIndex int
	curFrame    *frame
	curInsts    []byte
	ip          int
	aborting    *int64
}

// NewRuntime creates a Toy runtime.
func NewRuntime(bytecode *Bytecode, globals []Value) *Runtime {
	if globals == nil {
		globals = make([]Value, GlobalsSize)
	}
	r := &Runtime{
		constants:   bytecode.Constants,
		stack:       make([]Value, StackSize),
		sp:          0,
		globals:     globals,
		fileSet:     bytecode.FileSet,
		frames:      make([]frame, MaxFrames),
		framesIndex: 1,
		ip:          -1,
		aborting:    new(int64),
	}
	r.frames[0].fn = bytecode.MainFunction
	r.frames[0].ip = -1
	r.curFrame = &r.frames[0]
	r.curInsts = r.curFrame.fn.instructions
	return r
}

// Abort aborts the execution.
func (r *Runtime) Abort() {
	atomic.StoreInt64(r.aborting, 1)
}

// Run starts the execution.
func (r *Runtime) Run() error {
	// reset VM states
	r.sp = 0
	r.curFrame = &(r.frames[0])
	r.curInsts = r.curFrame.fn.instructions
	r.framesIndex = 1
	r.ip = -1
	_, err := r.run()
	atomic.StoreInt64(r.aborting, 0)
	if err != nil {
		rErr := r.unwindStack(err).(*runtimeError)
		var b strings.Builder
		fmt.Fprintf(&b, "runtime error: %s\n", rErr.Error())
		b.WriteString("\nstacktrace:")
		for i, pos := range rErr.Trace {
			if i != len(rErr.Trace)-1 {
				b.WriteString("\n├─ ")
			} else {
				b.WriteString("\n└─ ")
			}
			fmt.Fprintf(&b, "at %s", pos.String())
		}
		return errors.New(b.String())
	}
	return nil
}

func (r *Runtime) run() (_ Value, err error) {
	for atomic.LoadInt64(r.aborting) == 0 {
		r.ip++
		switch r.curInsts[r.ip] {
		case bytecode.OpConstant:
			r.ip += 2
			cidx := read2(r.curInsts, r.ip)
			r.stack[r.sp] = r.constants[cidx]
			r.sp++
		case bytecode.OpNull:
			r.stack[r.sp] = Nil
			r.sp++
		case bytecode.OpBinaryOp:
			r.ip++
			tok := token.Token(r.curInsts[r.ip])
			right := r.stack[r.sp-1]
			left := r.stack[r.sp-2]

			var res Value
			if tok == token.Nullish {
				if left != Nil {
					res = left
				} else {
					res = right
				}
			} else {
				res, err = BinaryOp(tok, left, right)
				if err != nil {
					r.sp -= 2
					return nil, err
				}
			}

			r.stack[r.sp-2] = res
			r.sp--
		case bytecode.OpCompare:
			r.ip++
			tok := token.Token(r.curInsts[r.ip])
			right := r.stack[r.sp-1]
			left := r.stack[r.sp-2]

			res, err := Compare(tok, left, right)
			if err != nil {
				r.sp -= 2
				return nil, err
			}

			r.stack[r.sp-2] = Bool(res)
			r.sp--
		case bytecode.OpPop:
			r.sp--
		case bytecode.OpTrue:
			r.stack[r.sp] = True
			r.sp++
		case bytecode.OpFalse:
			r.stack[r.sp] = False
			r.sp++
		case bytecode.OpUnaryOp:
			r.ip++
			tok := token.Token(r.curInsts[r.ip])
			operand := r.stack[r.sp-1]
			r.sp--

			res, err := UnaryOp(tok, operand)
			if err != nil {
				return nil, err
			}
			if res == nil {
				res = Nil
			}

			r.stack[r.sp] = res
			r.sp++
		case bytecode.OpJumpFalsy:
			r.ip += 4
			r.sp--
			if r.stack[r.sp].IsFalsy() {
				pos := read4(r.curInsts, r.ip)
				r.ip = pos - 1
			}
		case bytecode.OpAndJump:
			r.ip += 4
			if r.stack[r.sp-1].IsFalsy() {
				pos := read4(r.curInsts, r.ip)
				r.ip = pos - 1
			} else {
				r.sp--
			}
		case bytecode.OpOrJump:
			r.ip += 4
			if r.stack[r.sp-1].IsFalsy() {
				r.sp--
			} else {
				pos := read4(r.curInsts, r.ip)
				r.ip = pos - 1
			}
		case bytecode.OpJump:
			pos := read4(r.curInsts, r.ip+4)
			r.ip = pos - 1
		case bytecode.OpSetGlobal:
			r.ip += 2
			r.sp--
			globalIndex := read2(r.curInsts, r.ip)
			r.globals[globalIndex] = r.stack[r.sp]
		case bytecode.OpGetGlobal:
			r.ip += 2
			globalIndex := read2(r.curInsts, r.ip)
			val := r.globals[globalIndex]
			r.stack[r.sp] = val
			r.sp++
		case bytecode.OpString:
			r.ip += 3
			numParts := read2(r.curInsts, r.ip-1)
			unindent := int(r.curInsts[r.ip])

			var b strings.Builder
			for i := r.sp - numParts; i < r.sp; i++ {
				b.WriteString(AsString(r.stack[i]))
			}
			r.sp -= numParts

			str := b.String()
			if unindent == 1 {
				str = unindentString(str)
			}
			r.stack[r.sp] = String(str)
			r.sp++
		case bytecode.OpArray:
			r.ip += 3
			numElements := read2(r.curInsts, r.ip-1)
			splat := int(r.curInsts[r.ip])

			elements := r.takeListElements(numElements, splat)
			r.sp -= numElements

			r.stack[r.sp] = NewArray(elements)
			r.sp++
		case bytecode.OpTable:
			r.ip += 3
			numElements := read2(r.curInsts, r.ip-1)
			splat := int(r.curInsts[r.ip])

			t := NewTable(numElements)
			if splat == 1 {
				for i := r.sp - numElements; i < r.sp; {
					switch elem := r.stack[i].(type) {
					case *splatMapping:
						for key, value := range elem.m.Entries() {
							if err := t.ht.insert(key, value); err != nil {
								return nil, fmt.Errorf("table key '%s': %w", key.String(), err)
							}
						}
						i++
					default:
						value := r.stack[i+1]
						if err := t.ht.insert(elem, value); err != nil {
							return nil, fmt.Errorf("table key '%s': %w", elem.String(), err)
						}
						i += 2
					}
				}
			} else {
				for i := r.sp - numElements; i < r.sp; i += 2 {
					key, value := r.stack[i], r.stack[i+1]
					if err := t.ht.insert(key, value); err != nil {
						return nil, fmt.Errorf("table key '%s': %w", key.String(), err)
					}
				}
			}
			r.sp -= numElements

			r.stack[r.sp] = t
			r.sp++
		case bytecode.OpTuple:
			r.ip += 3
			numElements := read2(r.curInsts, r.ip-1)
			splat := int(r.curInsts[r.ip])

			elements := r.takeListElements(numElements, splat)
			r.sp -= numElements

			r.stack[r.sp] = Tuple(elements)
			r.sp++
		case bytecode.OpIndex:
			r.ip++
			withOk := int(r.curInsts[r.ip])
			key := r.stack[r.sp-1]
			left := r.stack[r.sp-2]

			val, found, err := Property(left, key)
			if err != nil {
				r.sp -= 2
				return nil, err
			}
			if val == nil {
				val = Nil
			}
			if withOk == 1 {
				val = Tuple{val, Bool(found)}
			}

			r.stack[r.sp-2] = val
			r.sp--
		case bytecode.OpSetIndex:
			key := r.stack[r.sp-1]
			left := r.stack[r.sp-2]
			right := r.stack[r.sp-3]
			r.sp -= 3
			if err := SetProperty(left, key, right); err != nil {
				return nil, err
			}
		case bytecode.OpSliceIndex:
			r.ip++
			op := r.curInsts[r.ip]
			left := r.stack[r.sp-1]
			r.sp--

			s, ok := left.(Sliceable)
			if !ok {
				return nil, fmt.Errorf("'%s' is not sliceable", TypeName(left))
			}
			n := s.Len()

			highIdx := n
			if op&0x2 != 0 {
				high := r.stack[r.sp-1]
				r.sp--

				if highInt, ok := high.(Int); ok {
					highIdx = int(highInt)
				} else {
					return nil, fmt.Errorf("invalid slice index type: %s", TypeName(high))
				}
			}

			lowIdx := 0
			if op&0x1 != 0 {
				low := r.stack[r.sp-1]
				r.sp--

				if lowInt, ok := low.(Int); ok {
					lowIdx = int(lowInt)
				} else {
					return nil, fmt.Errorf("invalid slice index type: %s", TypeName(low))
				}
			}

			res, err := Slice(s, lowIdx, highIdx)
			if err != nil {
				return nil, err
			}

			r.stack[r.sp] = res
			r.sp++
		case bytecode.OpSplat:
			r.ip++
			isMapping := int(r.curInsts[r.ip])
			value := r.stack[r.sp-1]
			if isMapping == 1 {
				m, ok := value.(Mapping)
				if !ok {
					return nil, fmt.Errorf("splat operator can only be used with mapping here, got '%s' instead",
						TypeName(value))
				}
				r.stack[r.sp-1] = &splatMapping{m: m}
			} else {
				seq, ok := value.(Sequence)
				if !ok {
					return nil, fmt.Errorf("splat operator can only be used with sequence here, got '%s' instead",
						TypeName(value))
				}
				r.stack[r.sp-1] = &splatSequence{s: seq}
			}
		case bytecode.OpCall:
			r.ip += 2
			numArgs := int(r.curInsts[r.ip-1])
			splat := int(r.curInsts[r.ip])

			value := r.stack[r.sp-1-numArgs]
			callable, ok := value.(Callable)
			if !ok {
				return nil, fmt.Errorf("not callable: %s", TypeName(value))
			}

			args := r.takeListElements(numArgs, splat)
			r.sp -= numArgs + 1

			if callee, ok := callable.(*CompiledFunction); ok {
				// do not need to pause the current runtime
				if _, err := callee.call(r, args, false); err != nil {
					return nil, fmt.Errorf("error during call to '%s': %w",
						TypeName(callable), err)
				}
			} else {
				ret, err := r.safeCall(callable, args)
				if err != nil {
					return nil, fmt.Errorf("error during call to '%s': %w",
						TypeName(callable), err)
				}
				// nil return -> nil
				if ret == nil {
					ret = Nil
				}
				r.stack[r.sp] = ret
				r.sp++
			}
		case bytecode.OpReturn:
			r.ip++
			numResults := int(r.curInsts[r.ip])

			if len(r.curFrame.deferred) > 0 {
				// must run deferred calls
				if err := r.runDefer(); err != nil {
					return nil, err
				}
			}

			var retVal Value
			if numResults == 1 {
				retVal = r.stack[r.sp-1]
			} else {
				// nil return -> nil
				retVal = Nil
			}

			if r.framesIndex == 1 {
				// we either leaving main function or hit stop point
				return retVal, nil
			}

			r.framesIndex--
			r.curFrame = &r.frames[r.framesIndex-1]
			r.curInsts = r.curFrame.fn.instructions
			r.ip = r.curFrame.ip
			r.sp = r.frames[r.framesIndex].basePointer
			r.stack[r.sp] = retVal
			r.sp++
		case bytecode.OpDefer:
			r.ip += 3
			numArgs := int(r.curInsts[r.ip-2])
			splat := int(r.curInsts[r.ip-1])
			deferIdx := int(r.curInsts[r.ip])

			value := r.stack[r.sp-1-numArgs]
			callable, ok := value.(Callable)
			if !ok {
				return nil, fmt.Errorf("not callable: %s", TypeName(value))
			}

			args := r.takeListElements(numArgs, splat)
			r.sp -= numArgs + 1

			r.curFrame.deferred = append(r.curFrame.deferred, &deferredCall{
				fn:   callable,
				args: args,
				pos:  r.curFrame.fn.deferMap[deferIdx],
			})
		case bytecode.OpTry:
			r.ip += 2
			numArgs := int(r.curInsts[r.ip-1])
			splat := int(r.curInsts[r.ip])

			value := r.stack[r.sp-1-numArgs]
			callable, ok := value.(Callable)
			if !ok {
				return nil, fmt.Errorf("not callable: %s", TypeName(value))
			}

			args := r.takeListElements(numArgs, splat)
			r.sp -= numArgs + 1

			var status Value
			ret, err := r.safeCall(callable, args)
			if err != nil {
				status = newExceptionTable(err)
			} else {
				status = Nil
			}
			if ret == nil {
				ret = Nil
			}

			r.stack[r.sp] = Tuple{ret, status}
			r.sp++
		case bytecode.OpThrow:
			r.ip++
			numErrors := int(r.curInsts[r.ip])

			var errVal Value
			if numErrors == 1 {
				errVal = r.stack[r.sp-1]
			} else {
				errVal = Nil
			}
			r.sp--

			return nil, &Exception{Value: errVal}
		case bytecode.OpDefineLocal:
			r.ip++
			localIndex := int(r.curInsts[r.ip])
			sp := r.curFrame.basePointer + localIndex
			val := r.stack[r.sp-1]
			r.sp--
			r.stack[sp] = val
		case bytecode.OpSetLocal:
			r.ip++
			localIndex := int(r.curInsts[r.ip])
			sp := r.curFrame.basePointer + localIndex
			val := r.stack[r.sp-1]
			r.sp--
			if obj, ok := r.stack[sp].(*valuePtr); ok {
				*obj.p = val
			} else {
				r.stack[sp] = val
			}
		case bytecode.OpGetLocal:
			r.ip++
			localIndex := int(r.curInsts[r.ip])
			val := r.stack[r.curFrame.basePointer+localIndex]
			if obj, ok := val.(*valuePtr); ok {
				val = *obj.p
			}
			r.stack[r.sp] = val
			r.sp++
		case bytecode.OpGetBuiltin:
			r.ip++
			builtinIndex := int(r.curInsts[r.ip])
			r.stack[r.sp] = Universe[builtinIndex].Value()
			r.sp++
		case bytecode.OpIdxAssignAssert:
			r.ip += 2
			n := read2(r.curInsts, r.ip)
			val := r.stack[r.sp-1]
			seq, ok := val.(IndexAccessible)
			if !ok {
				return nil, fmt.Errorf("trying to assign non-index-accessible '%s' to %d variable(s)",
					TypeName(val), n)
			}
			if n != seq.Len() {
				return nil, fmt.Errorf("trying to assign %d value(s) to %d variable(s)",
					seq.Len(), n)
			}
		case bytecode.OpIdxElem:
			r.ip += 2
			eidx := read2(r.curInsts, r.ip)
			val := r.stack[r.sp-1]
			seq, ok := r.stack[r.sp-1].(IndexAccessible)
			if !ok {
				return nil, fmt.Errorf("trying to get %d'th element from non-index-accessible '%s'",
					eidx, TypeName(val))
			}
			r.stack[r.sp] = seq.At(eidx)
			r.sp++
		case bytecode.OpClosure:
			r.ip += 3
			constIndex := read2(r.curInsts, r.ip-1)
			numFree := int(r.curInsts[r.ip])
			fn, ok := r.constants[constIndex].(*CompiledFunction)
			if !ok {
				return nil, fmt.Errorf("not function: %s", TypeName(fn))
			}
			free := make([]*valuePtr, numFree)
			for i := 0; i < numFree; i++ {
				// we always expect *objectPtr here
				// because the compiler only produces OpClosure
				// alongside OpGetLocalPtr and OpGetFreePtr,
				// which always push *objectPtr onto the stack
				free[i] = r.stack[r.sp-numFree+i].(*valuePtr)
			}
			r.sp -= numFree
			cl := &CompiledFunction{
				instructions:  fn.instructions,
				numLocals:     fn.numLocals,
				numParameters: fn.numParameters,
				numOptionals:  fn.numOptionals,
				varArgs:       fn.varArgs,
				sourceMap:     fn.sourceMap,
				free:          free,
			}
			r.stack[r.sp] = cl
			r.sp++
		case bytecode.OpGetFreePtr:
			r.ip++
			freeIndex := int(r.curInsts[r.ip])
			val := r.curFrame.freeVars[freeIndex]
			r.stack[r.sp] = val
			r.sp++
		case bytecode.OpGetFree:
			r.ip++
			freeIndex := int(r.curInsts[r.ip])
			val := *r.curFrame.freeVars[freeIndex].p
			r.stack[r.sp] = val
			r.sp++
		case bytecode.OpSetFree:
			r.ip++
			freeIndex := int(r.curInsts[r.ip])
			*r.curFrame.freeVars[freeIndex].p = r.stack[r.sp-1]
			r.sp--
		case bytecode.OpGetLocalPtr:
			r.ip++
			localIndex := int(r.curInsts[r.ip])
			sp := r.curFrame.basePointer + localIndex
			val := r.stack[sp]
			var freeVar *valuePtr
			if obj, ok := val.(*valuePtr); ok {
				freeVar = obj
			} else {
				freeVar = &valuePtr{p: &val}
				r.stack[sp] = freeVar
			}
			r.stack[r.sp] = freeVar
			r.sp++
		case bytecode.OpIteratorInit:
			dst := r.stack[r.sp-1]
			r.sp--

			var itValue *iterator
			switch iterable := dst.(type) {
			case KVIterable:
				itValue = newIterator2(iterable.Entries())
			case Iterable:
				itValue = newIterator(iterable.Elements())
			default:
				return nil, fmt.Errorf("not iterable: %s", TypeName(dst))
			}

			r.stack[r.sp] = itValue
			r.sp++
		case bytecode.OpIteratorNext:
			r.ip++
			op := r.curInsts[r.ip]
			it := r.stack[r.sp-1].(*iterator)
			r.sp--

			key, value, hasMore := it.next()
			if hasMore {
				if op&0x1 != 0 {
					r.stack[r.sp] = key
					r.sp++
				}
				if op&0x2 != 0 {
					r.stack[r.sp] = value
					r.sp++
				}
			}

			r.stack[r.sp] = Bool(hasMore)
			r.sp++
		case bytecode.OpIteratorClose:
			it, ok := r.stack[r.sp-1].(*iterator)
			if !ok {
				return nil, fmt.Errorf("not iterator: %s", TypeName(it))
			}
			it.stop()
			r.sp--
		default:
			return nil, fmt.Errorf("unknown opcode: %d", r.curInsts[r.ip])
		}
	}
	return Nil, nil
}

// safeCall calls callable with the provided arguments and properly recovers from panics.
func (r *Runtime) safeCall(callable Callable, args []Value) (_ Value, err error) {
	defer func() {
		if p := recover(); p != nil {
			switch e := p.(type) {
			case string:
				err = errors.New(e)
			case error:
				err = e
			case Value:
				err = &Exception{Value: e}
			default:
				err = fmt.Errorf("unknown panic: %v", e)
			}
		}
	}()
	ret, err := callable.Call(r, args...)
	return ret, err
}

// callCompiled calls a compiled function with the provided arguments
// and pauses the execution of the current runtime if pause = true.
func (r *Runtime) callCompiled(fn *CompiledFunction, args []Value, pause bool) (Value, error) {
	if r.framesIndex >= len(r.frames) {
		return nil, ErrStackOverflow
	}

	// update call frame
	r.curFrame.ip = r.ip // store current ip before call
	r.curFrame = &r.frames[r.framesIndex]
	r.curFrame.fn = fn
	r.curFrame.freeVars = fn.free
	r.curFrame.basePointer = r.sp
	r.curInsts = fn.instructions
	r.ip = -1
	r.framesIndex++
	copy(r.stack[r.sp:], args)
	r.sp = r.sp + fn.numLocals

	if pause {
		// save call stack
		frames := r.frames
		framesIndex := r.framesIndex

		// change call stack to stop execution and unwinding
		// after the current frame
		r.frames = frames[framesIndex-1:]
		r.framesIndex = 1

		res, err := r.run()
		if err != nil {
			err = r.unwindStack(err)
		}

		// restore call stack
		r.frames = frames
		r.framesIndex = framesIndex - 1
		r.curFrame = &r.frames[r.framesIndex-1]
		r.curInsts = r.curFrame.fn.instructions
		r.ip = r.curFrame.ip
		r.sp = r.frames[r.framesIndex].basePointer

		return res, err
	}

	return nil, nil
}

// takeListElements takes numElems elements from the stack
// and unpacks all splatSequence if splat = 1.
// NOTE: it doesn't pop them from the stack.
func (r *Runtime) takeListElements(numElems, splat int) []Value {
	elems := make([]Value, 0, numElems)
	if splat == 1 {
		for i := r.sp - numElems; i < r.sp; i++ {
			switch elem := r.stack[i].(type) {
			case *splatSequence:
				slices.AppendSeq(elems, elem.s.Elements())
			default:
				elems = append(elems, elem)
			}
		}
	} else {
		elems = append(elems, r.stack[r.sp-numElems:r.sp]...)
	}
	return elems
}

// unwindStack unwindes the call stack invoking all deferred calls along the way.
func (r *Runtime) unwindStack(reason error) error {
	var rErr *runtimeError
	if !errors.As(reason, &rErr) {
		rErr = &runtimeError{
			Errors: []error{reason},
			Trace:  make([]token.FilePos, 0),
		}
	}

	for r.framesIndex > 0 {
		var filePos token.FilePos
		if r.curFrame.curDefer != nil {
			filePos = r.fileSet.Position(r.curFrame.curDefer.pos)
		} else {
			filePos = r.fileSet.Position(r.curFrame.fn.sourcePos(r.ip - 1))
		}
		rErr.Trace = append(rErr.Trace, filePos)

		if len(r.curFrame.deferred) > 0 {
			if err := r.runDefer(); err != nil {
				var tmp *runtimeError
				if errors.As(err, &tmp) {
					rErr.Errors = append(rErr.Errors, tmp.Errors...)
					rErr.Trace = append(rErr.Trace, tmp.Trace...)
				} else {
					rErr.Errors = append(rErr.Errors, err)
				}
				continue // run remaining deferred calls
			}
		}

		r.framesIndex--
		if r.framesIndex > 0 {
			r.curFrame = &r.frames[r.framesIndex-1]
			r.curInsts = r.curFrame.fn.instructions
			r.ip = r.curFrame.ip
			r.sp = r.frames[r.framesIndex].basePointer
		}
	}

	return rErr
}

// runDefer invokes all deferred calls for the current frame.
func (r *Runtime) runDefer() (err error) {
	for len(r.curFrame.deferred) > 0 {
		// pop a deferred call
		call := r.curFrame.deferred[len(r.curFrame.deferred)-1]
		r.curFrame.deferred = r.curFrame.deferred[:len(r.curFrame.deferred)-1]
		r.curFrame.curDefer = call
		if _, err = r.safeCall(call.fn, call.args); err != nil {
			return err
		}
	}
	r.curFrame.curDefer = nil
	return nil
}

// runtimeError is a special error type returned by (*Runtime).unwind()
// that contains all errors encountered and a stacktrace.
type runtimeError struct {
	Errors []error
	Trace  []token.FilePos
}

func (e *runtimeError) Error() string {
	var b strings.Builder
	for i, err := range e.Errors {
		if i != 0 {
			if i != len(e.Errors)-1 {
				b.WriteString("\n├─ ")
			} else {
				b.WriteString("\n└─ ")
			}
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

func (e *runtimeError) Unwrap() error {
	// we only care about the initial error
	return e.Errors[0]
}

// newExceptionTable turns the error into an immutable table
// containing information about the exception/runtime error.
// This table is then returned by try keyword.
// The table consists of msg and val fields:
// - msg field contains the string representation of the error.
// - val field contains the value thrown by throw keyword.
func newExceptionTable(err error) Value {
	if rErr := (*runtimeError)(nil); errors.As(err, &rErr) {
		// we only care about the initial error
		err = rErr.Errors[0]
	}
	var val Value
	if exc := (*Exception)(nil); errors.As(err, &exc) {
		val = exc.Value
	} else {
		val = Nil
	}
	t := NewTable(2)
	t.SetProperty(String("msg"), String(err.Error()))
	t.SetProperty(String("val"), val)
	return t.Freeze()
}

// read2 reads a 2-byte operand assumming that
// ip is at the last byte of the operand.
func read2(ins []byte, ip int) int {
	return int(ins[ip]) | int(ins[ip-1])<<8
}

// read2 reads a 4-byte operand assumming that
// ip as the the last byte of the operand.
func read4(ins []byte, ip int) int {
	return int(ins[ip]) | int(ins[ip-1])<<8 | int(ins[ip-2])<<16 | int(ins[ip-3])<<24
}
