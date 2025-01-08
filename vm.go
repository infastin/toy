package toy

import (
	"fmt"
	"sync/atomic"

	"github.com/infastin/toy/parser"
	"github.com/infastin/toy/token"
)

// frame represents a function call frame.
type frame struct {
	fn          *CompiledFunction
	freeVars    []*objectPtr
	ip          int
	basePointer int
}

// VM is a virtual machine that executes the bytecode compiled by Compiler.
type VM struct {
	constants   []Object
	stack       []Object
	sp          int
	globals     []Object
	fileSet     *parser.SourceFileSet
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
		case parser.OpConstant:
			v.ip += 2
			cidx := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8
			v.stack[v.sp] = v.constants[cidx]
			v.sp++
		case parser.OpNull:
			v.stack[v.sp] = Undefined
			v.sp++
		case parser.OpBinaryOp:
			v.ip++
			right := v.stack[v.sp-1]
			left := v.stack[v.sp-2]
			tok := token.Token(v.curInsts[v.ip])

			res, err := BinaryOp(tok, left, right)
			if err != nil {
				v.sp -= 2
				v.err = fmt.Errorf("operation '%s %s %s' has failed: %w",
					left.TypeName(), tok.String(), right.TypeName(), err)
				return
			}

			v.stack[v.sp-2] = res
			v.sp--
		case parser.OpCompare:
			v.ip++
			right := v.stack[v.sp-1]
			left := v.stack[v.sp-2]
			tok := token.Token(v.curInsts[v.ip])

			res, err := Compare(tok, left, right)
			if err != nil {
				v.sp -= 2
				v.err = fmt.Errorf("operation '%s %s %s' has failed: %w",
					left.TypeName(), tok.String(), right.TypeName(), err)
				return
			}

			v.stack[v.sp-2] = Bool(res)
			v.sp--
		case parser.OpPop:
			v.sp--
		case parser.OpTrue:
			v.stack[v.sp] = True
			v.sp++
		case parser.OpFalse:
			v.stack[v.sp] = False
			v.sp++
		case parser.OpUnaryOp:
			v.ip++
			operand := v.stack[v.sp-1]
			tok := token.Token(v.curInsts[v.ip])
			v.sp--

			res, err := UnaryOp(tok, operand)
			if err != nil {
				v.err = fmt.Errorf("operation '%s%s' has failed: %w",
					tok.String(), operand.TypeName(), err)
				return
			}

			v.stack[v.sp] = res
			v.sp++
		case parser.OpJumpFalsy:
			v.ip += 4
			v.sp--
			if v.stack[v.sp].IsFalsy() {
				pos := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8 | int(v.curInsts[v.ip-2])<<16 | int(v.curInsts[v.ip-3])<<24
				v.ip = pos - 1
			}
		case parser.OpAndJump:
			v.ip += 4
			if v.stack[v.sp-1].IsFalsy() {
				pos := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8 | int(v.curInsts[v.ip-2])<<16 | int(v.curInsts[v.ip-3])<<24
				v.ip = pos - 1
			} else {
				v.sp--
			}
		case parser.OpOrJump:
			v.ip += 4
			if v.stack[v.sp-1].IsFalsy() {
				v.sp--
			} else {
				pos := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8 | int(v.curInsts[v.ip-2])<<16 | int(v.curInsts[v.ip-3])<<24
				v.ip = pos - 1
			}
		case parser.OpJump:
			pos := int(v.curInsts[v.ip+4]) | int(v.curInsts[v.ip+3])<<8 | int(v.curInsts[v.ip+2])<<16 | int(v.curInsts[v.ip+1])<<24
			v.ip = pos - 1
		case parser.OpSetGlobal:
			v.ip += 2
			v.sp--
			globalIndex := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8
			v.globals[globalIndex] = v.stack[v.sp]
		case parser.OpSetSelGlobal:
			v.ip += 3
			globalIndex := int(v.curInsts[v.ip-1]) | int(v.curInsts[v.ip-2])<<8
			numSelectors := int(v.curInsts[v.ip])
			// selectors and RHS value
			selectors := make([]Object, numSelectors)
			for i := 0; i < numSelectors; i++ {
				selectors[i] = v.stack[v.sp-numSelectors+i]
			}
			val := v.stack[v.sp-numSelectors-1]
			v.sp -= numSelectors + 1
			if err := indexAssign(v.globals[globalIndex], val, selectors); err != nil {
				v.err = err
				return
			}
		case parser.OpGetGlobal:
			v.ip += 2
			globalIndex := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8
			val := v.globals[globalIndex]
			v.stack[v.sp] = val
			v.sp++
		case parser.OpArray:
			v.ip += 3
			numElements := int(v.curInsts[v.ip-1]) | int(v.curInsts[v.ip-2])<<8
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
		case parser.OpMap:
			v.ip += 2
			numElements := 2 * (int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8)

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
		case parser.OpTuple:
			v.ip += 3
			numElements := int(v.curInsts[v.ip-1]) | int(v.curInsts[v.ip-2])<<8
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
		case parser.OpImmutable:
			value := v.stack[v.sp-1]
			v.stack[v.sp-1] = AsImmutable(value)
		case parser.OpIndex:
			index := v.stack[v.sp-1]
			left := v.stack[v.sp-2]
			v.sp -= 2
			val, err := IndexGet(left, index)
			if err != nil {
				v.err = fmt.Errorf("operation '%s[%s]' has failed: %w",
					left.TypeName(), index.TypeName(), err)
				return
			}
			if val == nil {
				val = Undefined
			}
			v.stack[v.sp] = val
			v.sp++
		case parser.OpField:
			name := v.stack[v.sp-1].(String)
			left := v.stack[v.sp-2]
			v.sp -= 2
			val, err := FieldGet(left, string(name))
			if err != nil {
				v.err = fmt.Errorf("operation '%s.%s' has failed: %w",
					left.TypeName(), string(name), err)
				return
			}
			if val == nil {
				val = Undefined
			}
			v.stack[v.sp] = val
			v.sp++
		case parser.OpSliceIndex:
			v.ip++
			left := v.stack[v.sp-1]
			op := v.curInsts[v.ip]
			v.sp--

			s, ok := left.(Sliceable)
			if !ok {
				v.err = fmt.Errorf("not sliceable: %s", left.TypeName())
			}
			n := s.Len()

			highIdx := n
			if op&0x2 != 0 {
				high := v.stack[v.sp-1]
				v.sp--

				if highInt, ok := high.(Int); ok {
					highIdx = int(highInt)
				} else {
					v.err = fmt.Errorf("invalid slice index type: %s", high.TypeName())
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
					v.err = fmt.Errorf("invalid slice index type: %s", low.TypeName())
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
		case parser.OpSplat:
			value := v.stack[v.sp-1]
			seq, ok := value.(Sequence)
			if !ok {
				v.err = fmt.Errorf("splat operator can only be used with sequence, got '%s' instead",
					seq.TypeName())
				return
			}
			v.stack[v.sp-1] = &splatSequence{s: seq}
		case parser.OpCall:
			numArgs := int(v.curInsts[v.ip+1])
			splat := int(v.curInsts[v.ip+2])
			v.ip += 2

			value := v.stack[v.sp-1-numArgs]

			callable, ok := value.(Callable)
			if !ok {
				v.err = fmt.Errorf("not callable: %s", value.TypeName())
				return
			}

			if splat == 1 {
				var newArgs []Object
				for i := v.sp - numArgs; i < v.sp; i++ {
					switch arg := v.stack[i].(type) {
					case *splatSequence:
						newArgs = append(newArgs, arg.s.Items()...)
					default:
						newArgs = append(newArgs, arg)
					}
				}
				copy(v.stack[v.sp-numArgs:], newArgs)
				v.sp += len(newArgs) - numArgs
				numArgs = len(newArgs)
			}

			if callee, ok := callable.(*CompiledFunction); ok {
				if callee.varArgs {
					// if the closure is variadic,
					// roll up all variadic parameters into an array
					realArgs := callee.numParameters - 1
					varArgs := numArgs - realArgs
					if varArgs >= 0 {
						numArgs = realArgs + 1
						args := make([]Object, varArgs)
						spStart := v.sp - varArgs
						for i := spStart; i < v.sp; i++ {
							args[i-spStart] = v.stack[i]
						}
						v.stack[spStart] = NewArray(args)
						v.sp = spStart + 1
					}
				}

				if numArgs != callee.numParameters {
					if callee.varArgs {
						v.err = &WrongNumArgumentsError{
							WantMin: callee.numParameters - 1,
							Got:     numArgs,
						}
					} else {
						v.err = &WrongNumArgumentsError{
							WantMin: callee.numParameters,
							WantMax: callee.numParameters,
							Got:     numArgs,
						}
					}
					return
				}

				// test if it's tail-call
				if callee == v.curFrame.fn { // recursion
					nextOp := v.curInsts[v.ip+1]
					if nextOp == parser.OpReturn ||
						(nextOp == parser.OpPop && parser.OpReturn == v.curInsts[v.ip+2]) {
						for p := 0; p < numArgs; p++ {
							v.stack[v.curFrame.basePointer+p] = v.stack[v.sp-numArgs+p]
						}
						v.sp -= numArgs + 1
						v.ip = -1 // reset IP to beginning of the frame
						continue
					}
				}
				if v.framesIndex >= MaxFrames {
					v.err = ErrStackOverflow
					return
				}

				// update call frame
				v.curFrame.ip = v.ip // store current ip before call
				v.curFrame = &(v.frames[v.framesIndex])
				v.curFrame.fn = callee
				v.curFrame.freeVars = callee.free
				v.curFrame.basePointer = v.sp - numArgs
				v.curInsts = callee.instructions
				v.ip = -1
				v.framesIndex++
				v.sp = v.sp - numArgs + callee.numLocals
			} else {
				var args []Object
				args = append(args, v.stack[v.sp-numArgs:v.sp]...)
				ret, err := callable.Call(args...)
				v.sp -= numArgs + 1

				// runtime error
				if err != nil {
					v.err = fmt.Errorf("error during call to '%s': %w",
						callable.TypeName(), err)
					return
				}

				// nil return -> undefined
				if ret == nil {
					ret = Undefined
				}
				v.stack[v.sp] = ret
				v.sp++
			}
		case parser.OpReturn:
			v.ip++
			var retVal Object
			if int(v.curInsts[v.ip]) == 1 {
				retVal = v.stack[v.sp-1]
			} else {
				retVal = Undefined
			}
			// v.sp--
			v.framesIndex--
			v.curFrame = &v.frames[v.framesIndex-1]
			v.curInsts = v.curFrame.fn.instructions
			v.ip = v.curFrame.ip
			// v.sp = lastFrame.basePointer - 1
			v.sp = v.frames[v.framesIndex].basePointer
			// skip stack overflow check because (newSP) <= (oldSP)
			v.stack[v.sp-1] = retVal
			// v.sp++
		case parser.OpDefineLocal:
			v.ip++
			localIndex := int(v.curInsts[v.ip])
			sp := v.curFrame.basePointer + localIndex

			// local variables can be mutated by other actions
			// so always store the copy of popped value
			val := v.stack[v.sp-1]
			v.sp--
			v.stack[sp] = val
		case parser.OpSetLocal:
			localIndex := int(v.curInsts[v.ip+1])
			v.ip++
			sp := v.curFrame.basePointer + localIndex

			// update pointee of v.stack[sp] instead of replacing the pointer
			// itself. this is needed because there can be free variables
			// referencing the same local variables.
			val := v.stack[v.sp-1]
			v.sp--
			if obj, ok := v.stack[sp].(*objectPtr); ok {
				*obj.p = val
				val = obj
			}
			v.stack[sp] = val // also use a copy of popped value
		case parser.OpSetSelLocal:
			localIndex := int(v.curInsts[v.ip+1])
			numSelectors := int(v.curInsts[v.ip+2])
			v.ip += 2

			// selectors and RHS value
			selectors := make([]Object, numSelectors)
			for i := 0; i < numSelectors; i++ {
				selectors[i] = v.stack[v.sp-numSelectors+i]
			}
			val := v.stack[v.sp-numSelectors-1]
			v.sp -= numSelectors + 1
			dst := v.stack[v.curFrame.basePointer+localIndex]
			if obj, ok := dst.(*objectPtr); ok {
				dst = *obj.p
			}
			if err := indexAssign(dst, val, selectors); err != nil {
				v.err = err
				return
			}
		case parser.OpGetLocal:
			v.ip++
			localIndex := int(v.curInsts[v.ip])
			val := v.stack[v.curFrame.basePointer+localIndex]
			if obj, ok := val.(*objectPtr); ok {
				val = *obj.p
			}
			v.stack[v.sp] = val
			v.sp++
		case parser.OpGetBuiltin:
			v.ip++
			builtinIndex := int(v.curInsts[v.ip])
			v.stack[v.sp] = BuiltinFuncs[builtinIndex]
			v.sp++
		case parser.OpIdxAssignAssert:
			v.ip += 2
			n := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8
			val := v.stack[v.sp-1]
			seq, ok := val.(Indexable)
			if !ok {
				v.err = fmt.Errorf("trying to assign non-indexable '%s' to %d variables", val.TypeName(), n)
				return
			}
			if n != seq.Len() {
				v.err = fmt.Errorf("trying to assign %d values to %d variables", seq.Len(), n)
				return
			}
		case parser.OpIdxElem:
			v.ip += 2
			eidx := int(v.curInsts[v.ip]) | int(v.curInsts[v.ip-1])<<8
			val := v.stack[v.sp-1]
			seq, ok := v.stack[v.sp-1].(Indexable)
			if !ok {
				v.err = fmt.Errorf("trying to get %d'th element from non-indexable '%s'", eidx, val.TypeName())
				return
			}
			v.stack[v.sp] = seq.At(eidx)
			v.sp++
		case parser.OpClosure:
			v.ip += 3
			constIndex := int(v.curInsts[v.ip-1]) | int(v.curInsts[v.ip-2])<<8
			numFree := int(v.curInsts[v.ip])
			fn, ok := v.constants[constIndex].(*CompiledFunction)
			if !ok {
				v.err = fmt.Errorf("not function: %s", fn.TypeName())
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
				free:          free,
			}
			v.stack[v.sp] = cl
			v.sp++
		case parser.OpGetFreePtr:
			v.ip++
			freeIndex := int(v.curInsts[v.ip])
			val := v.curFrame.freeVars[freeIndex]
			v.stack[v.sp] = val
			v.sp++
		case parser.OpGetFree:
			v.ip++
			freeIndex := int(v.curInsts[v.ip])
			val := *v.curFrame.freeVars[freeIndex].p
			v.stack[v.sp] = val
			v.sp++
		case parser.OpSetFree:
			v.ip++
			freeIndex := int(v.curInsts[v.ip])
			*v.curFrame.freeVars[freeIndex].p = v.stack[v.sp-1]
			v.sp--
		case parser.OpGetLocalPtr:
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
		case parser.OpSetSelFree:
			v.ip += 2
			freeIndex := int(v.curInsts[v.ip-1])
			numSelectors := int(v.curInsts[v.ip])

			// selectors and RHS value
			selectors := make([]Object, numSelectors)
			for i := 0; i < numSelectors; i++ {
				selectors[i] = v.stack[v.sp-numSelectors+i]
			}
			val := v.stack[v.sp-numSelectors-1]
			v.sp -= numSelectors + 1
			if err := indexAssign(*v.curFrame.freeVars[freeIndex].p, val, selectors); err != nil {
				v.err = err
				return
			}
		case parser.OpIteratorInit:
			dst := v.stack[v.sp-1]
			v.sp--
			iterable, ok := dst.(Iterable)
			if !ok {
				v.err = fmt.Errorf("not iterable: %s", dst.TypeName())
				return
			}
			iterator := iterable.Iterate()
			v.stack[v.sp] = iterator
			v.sp++
		case parser.OpIteratorNext:
			v.ip++
			iterator := v.stack[v.sp-1].(Iterator)
			op := v.curInsts[v.ip]
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
		case parser.OpIteratorClose:
			iterator, ok := v.stack[v.sp-1].(CloseableIterator)
			if ok {
				iterator.Close()
			}
			v.sp--
		case parser.OpSuspend:
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

func indexAssign(dst, src Object, selectors []Object) error {
	numSel := len(selectors)
	for sidx := numSel - 1; sidx > 0; sidx-- {
		next, err := IndexGet(dst, selectors[sidx])
		if err != nil {
			return fmt.Errorf("operation '%s[%s]' has failed: %w",
				dst.TypeName(), selectors[sidx].TypeName(), err)
		}
		dst = next
	}
	if err := IndexSet(dst, selectors[0], src); err != nil {
		return fmt.Errorf("assignment operation '%s[%s]' has failed: %w",
			dst.TypeName(), selectors[0].TypeName(), err)
	}
	return nil
}
