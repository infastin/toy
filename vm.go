package toy

import (
	"fmt"
	"sync/atomic"

	"github.com/infastin/toy/bytecode"
	"github.com/infastin/toy/token"
)

// deferredCall represents a deferred function call.
type deferredCall struct {
	fn   Callable
	args []Object
	pos  token.Pos
}

// frame represents a function call frame.
type frame struct {
	fn          *CompiledFunction
	freeVars    []*objectPtr
	ip          int
	basePointer int
	deferred    []*deferredCall
	// this flag tells compiled functions
	// to run in the subVM
	subvm bool
}

// VM is a virtual machine that executes the bytecode compiled by Compiler.
type VM struct {
	constants   []Object
	stack       []Object
	sp          int
	globals     []Object
	fileSet     *token.FileSet
	frames      []frame
	framesIndex int
	curFrame    *frame
	curInsts    []byte
	ip          int
	aborting    int64
	err         error
}

// NewVM creates a VM.
func NewVM(bytecode *Bytecode, globals []Object) *VM {
	if globals == nil {
		globals = make([]Object, GlobalsSize)
	}
	v := &VM{
		constants:   bytecode.Constants,
		stack:       make([]Object, StackSize),
		sp:          0,
		globals:     globals,
		fileSet:     bytecode.FileSet,
		frames:      make([]frame, MaxFrames),
		framesIndex: 1,
		ip:          -1,
	}
	v.frames[0].fn = bytecode.MainFunction
	v.frames[0].ip = -1
	v.curFrame = &v.frames[0]
	v.curInsts = v.curFrame.fn.instructions
	return v
}

// Abort aborts the execution.
func (v *VM) Abort() {
	atomic.StoreInt64(&v.aborting, 1)
}

// Run starts the execution.
func (v *VM) Run() error {
	// reset VM states
	v.sp = 0
	v.curFrame = &(v.frames[0])
	v.curInsts = v.curFrame.fn.instructions
	v.framesIndex = 1
	v.ip = -1
	v.run()
	atomic.StoreInt64(&v.aborting, 0)
	if err := v.err; err != nil {
		filePos := v.fileSet.Position(v.curFrame.fn.sourcePos(v.ip - 1))
		err := fmt.Errorf("Runtime Error: %w\n\tat %s", err, filePos)
		for v.framesIndex > 1 {
			v.framesIndex--
			v.curFrame = &v.frames[v.framesIndex-1]
			filePos = v.fileSet.Position(v.curFrame.fn.sourcePos(v.curFrame.ip - 1))
			err = fmt.Errorf("%w\n\tat %s", err, filePos)
		}
		return err
	}
	return nil
}

func (v *VM) run() {
	for atomic.LoadInt64(&v.aborting) == 0 {
		v.ip++
		switch v.curInsts[v.ip] {
		case bytecode.OpConstant:
			v.ip += 2
			cidx := read2(v.curInsts, v.ip)
			v.stack[v.sp] = v.constants[cidx]
			v.sp++
		case bytecode.OpNull:
			v.stack[v.sp] = Nil
			v.sp++
		case bytecode.OpBinaryOp:
			v.ip++
			tok := token.Token(v.curInsts[v.ip])
			right := v.stack[v.sp-1]
			left := v.stack[v.sp-2]

			res, err := BinaryOp(tok, left, right)
			if err != nil {
				v.sp -= 2
				v.err = err
				return
			}

			v.stack[v.sp-2] = res
			v.sp--
		case bytecode.OpCompare:
			v.ip++
			tok := token.Token(v.curInsts[v.ip])
			right := v.stack[v.sp-1]
			left := v.stack[v.sp-2]

			res, err := Compare(tok, left, right)
			if err != nil {
				v.sp -= 2
				v.err = err
				return
			}

			v.stack[v.sp-2] = Bool(res)
			v.sp--
		case bytecode.OpPop:
			v.sp--
		case bytecode.OpTrue:
			v.stack[v.sp] = True
			v.sp++
		case bytecode.OpFalse:
			v.stack[v.sp] = False
			v.sp++
		case bytecode.OpUnaryOp:
			v.ip++
			tok := token.Token(v.curInsts[v.ip])
			operand := v.stack[v.sp-1]
			v.sp--

			res, err := UnaryOp(tok, operand)
			if err != nil {
				v.err = err
				return
			}

			v.stack[v.sp] = res
			v.sp++
		case bytecode.OpJumpFalsy:
			v.ip += 4
			v.sp--
			if v.stack[v.sp].IsFalsy() {
				pos := read4(v.curInsts, v.ip)
				v.ip = pos - 1
			}
		case bytecode.OpAndJump:
			v.ip += 4
			if v.stack[v.sp-1].IsFalsy() {
				pos := read4(v.curInsts, v.ip)
				v.ip = pos - 1
			} else {
				v.sp--
			}
		case bytecode.OpOrJump:
			v.ip += 4
			if v.stack[v.sp-1].IsFalsy() {
				v.sp--
			} else {
				pos := read4(v.curInsts, v.ip)
				v.ip = pos - 1
			}
		case bytecode.OpJump:
			pos := read4(v.curInsts, v.ip+4)
			v.ip = pos - 1
		case bytecode.OpSetGlobal:
			v.ip += 2
			v.sp--
			globalIndex := read2(v.curInsts, v.ip)
			v.globals[globalIndex] = v.stack[v.sp]
		case bytecode.OpGetGlobal:
			v.ip += 2
			globalIndex := read2(v.curInsts, v.ip)
			val := v.globals[globalIndex]
			v.stack[v.sp] = val
			v.sp++
		case bytecode.OpArray:
			v.ip += 3
			numElements := read2(v.curInsts, v.ip-1)
			splat := int(v.curInsts[v.ip])

			var elements []Object
			if splat == 1 {
				for i := v.sp - numElements; i < v.sp; i++ {
					switch elem := v.stack[i].(type) {
					case *splatSequence:
						elements = append(elements, elem.s.Items()...)
					default:
						elements = append(elements, elem)
					}
				}
			} else {
				for i := v.sp - numElements; i < v.sp; i++ {
					elements = append(elements, v.stack[i])
				}
			}
			v.sp -= numElements

			v.stack[v.sp] = NewArray(elements)
			v.sp++
		case bytecode.OpMap:
			v.ip += 2
			numElements := 2 * read2(v.curInsts, v.ip)

			m := NewMap(numElements)
			for i := v.sp - numElements; i < v.sp; i += 2 {
				key := v.stack[i]
				value := v.stack[i+1]
				if err := m.ht.insert(key, value); err != nil {
					v.err = fmt.Errorf("map key '%s': %w", key.String(), err)
					return
				}
			}
			v.sp -= numElements

			v.stack[v.sp] = m
			v.sp++
		case bytecode.OpTuple:
			v.ip += 3
			numElements := read2(v.curInsts, v.ip-1)
			splat := int(v.curInsts[v.ip])

			var tup Tuple
			if splat == 1 {
				for i := v.sp - numElements; i < v.sp; i++ {
					switch elem := v.stack[i].(type) {
					case *splatSequence:
						tup = append(tup, elem.s.Items()...)
					default:
						tup = append(tup, elem)
					}
				}
			} else {
				for i := v.sp - numElements; i < v.sp; i++ {
					tup = append(tup, v.stack[i])
				}
			}
			v.sp -= numElements

			v.stack[v.sp] = tup
			v.sp++
		case bytecode.OpImmutable:
			value := v.stack[v.sp-1]
			v.stack[v.sp-1] = AsImmutable(value)
		case bytecode.OpIndex:
			v.ip++
			withOk := int(v.curInsts[v.ip])
			index := v.stack[v.sp-1]
			left := v.stack[v.sp-2]

			val, found, err := IndexGet(left, index)
			if err != nil {
				v.sp -= 2
				v.err = err
				return
			}
			if val == nil {
				val = Nil
			}
			if withOk == 1 {
				val = Tuple{val, Bool(found)}
			}

			v.stack[v.sp-2] = val
			v.sp--
		case bytecode.OpSetIndex:
			index := v.stack[v.sp-1]
			left := v.stack[v.sp-2]
			right := v.stack[v.sp-3]
			v.sp -= 3
			if err := IndexSet(left, index, right); err != nil {
				v.err = err
				return
			}
		case bytecode.OpField:
			name := v.stack[v.sp-1].(String)
			left := v.stack[v.sp-2]
			v.sp -= 2
			val, err := FieldGet(left, string(name))
			if err != nil {
				v.err = err
				return
			}
			if val == nil {
				val = Nil
			}
			v.stack[v.sp] = val
			v.sp++
		case bytecode.OpSetField:
			sel := v.stack[v.sp-1].(String)
			left := v.stack[v.sp-2]
			right := v.stack[v.sp-3]
			v.sp -= 3
			if err := FieldSet(left, string(sel), right); err != nil {
				v.err = err
				return
			}
		case bytecode.OpSliceIndex:
			v.ip++
			op := v.curInsts[v.ip]
			left := v.stack[v.sp-1]
			v.sp--

			s, ok := left.(Sliceable)
			if !ok {
				v.err = fmt.Errorf("not sliceable: %s", TypeName(left))
			}
			n := s.Len()

			highIdx := n
			if op&0x2 != 0 {
				high := v.stack[v.sp-1]
				v.sp--

				if highInt, ok := high.(Int); ok {
					highIdx = int(highInt)
				} else {
					v.err = fmt.Errorf("invalid slice index type: %s", TypeName(high))
					return
				}
			}

			lowIdx := 0
			if op&0x1 != 0 {
				low := v.stack[v.sp-1]
				v.sp--

				if lowInt, ok := low.(Int); ok {
					lowIdx = int(lowInt)
				} else {
					v.err = fmt.Errorf("invalid slice index type: %s", TypeName(low))
					return
				}
			}

			if lowIdx > highIdx {
				v.err = fmt.Errorf("invalid slice indices: %d > %d", lowIdx, highIdx)
				return
			}
			if lowIdx < 0 || lowIdx > n {
				v.err = fmt.Errorf("slice bounds out of range [%d:%d]", lowIdx, n)
				return
			}
			if highIdx < 0 || highIdx > n {
				v.err = fmt.Errorf("slice bounds out of range [%d:%d] with len %d", lowIdx, highIdx, n)
				return
			}

			v.stack[v.sp] = s.Slice(lowIdx, highIdx)
			v.sp++
		case bytecode.OpSplat:
			value := v.stack[v.sp-1]
			seq, ok := value.(Sequence)
			if !ok {
				v.err = fmt.Errorf("splat operator can only be used with sequence, got '%s' instead",
					TypeName(seq))
				return
			}
			v.stack[v.sp-1] = &splatSequence{s: seq}
		case bytecode.OpCall:
			v.ip += 2
			numArgs := int(v.curInsts[v.ip-1])
			splat := int(v.curInsts[v.ip])

			value := v.stack[v.sp-1-numArgs]
			callable, ok := value.(Callable)
			if !ok {
				v.err = fmt.Errorf("not callable: %s", TypeName(value))
				return
			}

			args := make([]Object, 0, numArgs)
			if splat == 1 {
				for i := v.sp - numArgs; i < v.sp; i++ {
					switch arg := v.stack[i].(type) {
					case *splatSequence:
						args = append(args, arg.s.Items()...)
					default:
						args = append(args, arg)
					}
				}
			} else {
				args = append(args, v.stack[v.sp-numArgs:v.sp]...)
			}
			v.sp -= numArgs + 1

			if callee, ok := callable.(*CompiledFunction); ok {
				if _, err := callee.Call(v, args...); err != nil {
					v.err = fmt.Errorf("error during call to '%s': %w",
						TypeName(callee), err)
					return
				}
			} else {
				// user functions can potentially call compiled functions,
				// so we tell compiled functions that they should run in the subVM
				v.curFrame.subvm = true
				ret, err := callable.Call(v, args...)
				if err != nil {
					v.err = fmt.Errorf("error during call to '%s': %w",
						TypeName(callable), err)
					return
				}
				if v.err != nil {
					// subVM closed with an error
					return
				}
				v.curFrame.subvm = false
				// nil return -> nil
				if ret == nil {
					ret = Nil
				}
				v.stack[v.sp] = ret
				v.sp++
			}
		case bytecode.OpReturn:
			v.ip++
			numResults := int(v.curInsts[v.ip])
			var retVal Object
			if numResults == 1 {
				retVal = v.stack[v.sp-1]
			} else {
				retVal = Nil
			}
			v.framesIndex--
			v.curFrame = &v.frames[v.framesIndex-1]
			v.curInsts = v.curFrame.fn.instructions
			v.ip = v.curFrame.ip
			v.sp = v.frames[v.framesIndex].basePointer
			v.stack[v.sp] = retVal
			v.sp++
			if v.curFrame.subvm {
				return
			}
		case bytecode.OpDefer:
			v.ip += 3
			numArgs := int(v.curInsts[v.ip-2])
			splat := int(v.curInsts[v.ip-1])
			deferIdx := int(v.curInsts[v.ip])

			value := v.stack[v.sp-1-numArgs]
			callable, ok := value.(Callable)
			if !ok {
				v.err = fmt.Errorf("not callable: %s", TypeName(value))
				return
			}

			args := make([]Object, 0, numArgs)
			if splat == 1 {
				for i := v.sp - numArgs; i < v.sp; i++ {
					switch arg := v.stack[i].(type) {
					case *splatSequence:
						args = append(args, arg.s.Items()...)
					default:
						args = append(args, arg)
					}
				}
			} else {
				args = append(args, v.stack[v.sp-numArgs:v.sp]...)
			}
			v.sp -= numArgs + 1

			v.curFrame.deferred = append(v.curFrame.deferred, &deferredCall{
				fn:   callable,
				args: args,
				pos:  v.curFrame.fn.deferMap[deferIdx],
			})
		case bytecode.OpRunDefer:
			for i := len(v.curFrame.deferred) - 1; i >= 0; i-- {
				call := v.curFrame.deferred[i]
				// tell compiled functions that they should run in the subVM
				v.curFrame.subvm = true
				// map RUNDEFER instruction to the position
				// of the defered call, so that we get
				// a correct error message in case of an error
				v.curFrame.fn.sourceMap[v.ip-1] = call.pos
				if _, err := call.fn.Call(v, call.args...); err != nil {
					v.err = fmt.Errorf("error during call to '%s': %w",
						TypeName(call.fn), err)
					return
				}
				if v.err != nil {
					// subVM closed with an error
					return
				}
				v.curFrame.subvm = false
			}
		case bytecode.OpDefineLocal:
			v.ip++
			localIndex := int(v.curInsts[v.ip])
			sp := v.curFrame.basePointer + localIndex
			// local variables can be mutated by other actions
			// so always store the copy of popped value
			val := v.stack[v.sp-1]
			v.sp--
			v.stack[sp] = val
		case bytecode.OpSetLocal:
			v.ip++
			localIndex := int(v.curInsts[v.ip])
			sp := v.curFrame.basePointer + localIndex
			// update pointee of v.stack[sp] instead of replacing the pointer
			// itself. this is needed because there can be free variables
			// referencing the same local variables
			val := v.stack[v.sp-1]
			v.sp--
			if obj, ok := v.stack[sp].(*objectPtr); ok {
				*obj.p = val
				val = obj
			}
			v.stack[sp] = val // also use a copy of popped value
		case bytecode.OpGetLocal:
			v.ip++
			localIndex := int(v.curInsts[v.ip])
			val := v.stack[v.curFrame.basePointer+localIndex]
			if obj, ok := val.(*objectPtr); ok {
				val = *obj.p
			}
			v.stack[v.sp] = val
			v.sp++
		case bytecode.OpGetBuiltin:
			v.ip++
			builtinIndex := int(v.curInsts[v.ip])
			v.stack[v.sp] = Universe[builtinIndex].Value()
			v.sp++
		case bytecode.OpIdxAssignAssert:
			v.ip += 2
			n := read2(v.curInsts, v.ip)
			val := v.stack[v.sp-1]
			seq, ok := val.(Indexable)
			if !ok {
				v.err = fmt.Errorf("trying to assign non-indexable '%s' to %d variable(s)", TypeName(val), n)
				return
			}
			if n != seq.Len() {
				v.err = fmt.Errorf("trying to assign %d value(s) to %d variable(s)", seq.Len(), n)
				return
			}
		case bytecode.OpIdxElem:
			v.ip += 2
			eidx := read2(v.curInsts, v.ip)
			val := v.stack[v.sp-1]
			seq, ok := v.stack[v.sp-1].(Indexable)
			if !ok {
				v.err = fmt.Errorf("trying to get %d'th element from non-indexable '%s'", eidx, TypeName(val))
				return
			}
			v.stack[v.sp] = seq.At(eidx)
			v.sp++
		case bytecode.OpClosure:
			v.ip += 3
			constIndex := read2(v.curInsts, v.ip-1)
			numFree := int(v.curInsts[v.ip])
			fn, ok := v.constants[constIndex].(*CompiledFunction)
			if !ok {
				v.err = fmt.Errorf("not function: %s", TypeName(fn))
				return
			}
			free := make([]*objectPtr, numFree)
			for i := 0; i < numFree; i++ {
				switch freeVar := (v.stack[v.sp-numFree+i]).(type) {
				case *objectPtr:
					free[i] = freeVar
				default:
					free[i] = &objectPtr{p: &v.stack[v.sp-numFree+i]}
				}
			}
			v.sp -= numFree
			cl := &CompiledFunction{
				instructions:  fn.instructions,
				numLocals:     fn.numLocals,
				numParameters: fn.numParameters,
				varArgs:       fn.varArgs,
				sourceMap:     fn.sourceMap,
				deferMap:      fn.deferMap,
				free:          free,
			}
			v.stack[v.sp] = cl
			v.sp++
		case bytecode.OpGetFreePtr:
			v.ip++
			freeIndex := int(v.curInsts[v.ip])
			val := v.curFrame.freeVars[freeIndex]
			v.stack[v.sp] = val
			v.sp++
		case bytecode.OpGetFree:
			v.ip++
			freeIndex := int(v.curInsts[v.ip])
			val := *v.curFrame.freeVars[freeIndex].p
			v.stack[v.sp] = val
			v.sp++
		case bytecode.OpSetFree:
			v.ip++
			freeIndex := int(v.curInsts[v.ip])
			*v.curFrame.freeVars[freeIndex].p = v.stack[v.sp-1]
			v.sp--
		case bytecode.OpGetLocalPtr:
			v.ip++
			localIndex := int(v.curInsts[v.ip])
			sp := v.curFrame.basePointer + localIndex
			val := v.stack[sp]
			var freeVar *objectPtr
			if obj, ok := val.(*objectPtr); ok {
				freeVar = obj
			} else {
				freeVar = &objectPtr{p: &val}
				v.stack[sp] = freeVar
			}
			v.stack[v.sp] = freeVar
			v.sp++
		case bytecode.OpIteratorInit:
			dst := v.stack[v.sp-1]
			v.sp--
			iterable, ok := dst.(Iterable)
			if !ok {
				v.err = fmt.Errorf("not iterable: %s", TypeName(dst))
				return
			}
			iterator := iterable.Iterate()
			v.stack[v.sp] = iterator
			v.sp++
		case bytecode.OpIteratorNext:
			v.ip++
			op := v.curInsts[v.ip]
			iterator := v.stack[v.sp-1].(Iterator)
			v.sp--

			var (
				key, value       Object
				keyPtr, valuePtr *Object
			)
			if op&0x1 != 0 {
				keyPtr = &key
			}
			if op&0x2 != 0 {
				valuePtr = &value
			}

			hasMore := iterator.Next(keyPtr, valuePtr)
			if hasMore {
				if keyPtr != nil {
					v.stack[v.sp] = key
					v.sp++
				}
				if valuePtr != nil {
					v.stack[v.sp] = value
					v.sp++
				}
			}

			v.stack[v.sp] = Bool(hasMore)
			v.sp++
		case bytecode.OpIteratorClose:
			iterator, ok := v.stack[v.sp-1].(CloseableIterator)
			if ok {
				iterator.Close()
			}
			v.sp--
		case bytecode.OpSuspend:
			return
		default:
			v.err = fmt.Errorf("unknown opcode: %d", v.curInsts[v.ip])
			return
		}
	}
}

// IsStackEmpty tests if the stack is empty or not.
func (v *VM) IsStackEmpty() bool {
	return v.sp == 0
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
