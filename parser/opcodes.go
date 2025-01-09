package parser

// Opcode represents a single byte operation code.
type Opcode = byte

// List of opcodes
const (
	OpConstant        Opcode = iota // Load constant
	OpPop                           // Pop
	OpTrue                          // Push true
	OpFalse                         // Push false
	OpJumpFalsy                     // Jump if falsy
	OpAndJump                       // Logical AND jump
	OpOrJump                        // Logical OR jump
	OpJump                          // Jump
	OpNull                          // Push null
	OpArray                         // Array object
	OpMap                           // Map object
	OpTuple                         // Tuple object
	OpImmutable                     // Immutable object
	OpIndex                         // Index operation
	OpField                         // Field operation
	OpSliceIndex                    // Slice operation
	OpSplat                         // Splat operation
	OpCall                          // Call function
	OpReturn                        // Return
	OpSaveDefer                     // Save deferred call
	OpPushDefer                     // Push deferred call to the stack
	OpGetGlobal                     // Get global variable
	OpSetGlobal                     // Set global variable
	OpSetSelGlobal                  // Set global variable using selectors
	OpGetLocal                      // Get local variable
	OpSetLocal                      // Set local variable
	OpDefineLocal                   // Define local variable
	OpSetSelLocal                   // Set local variable using selectors
	OpGetFreePtr                    // Get free variable pointer object
	OpGetFree                       // Get free variables
	OpSetFree                       // Set free variables
	OpGetLocalPtr                   // Get local variable as a pointer
	OpSetSelFree                    // Set free variables using selectors
	OpGetBuiltin                    // Get builtin function
	OpIdxAssignAssert               // Assert indexable size during tuple-assignment
	OpIdxElem                       // Push element from indexable
	OpClosure                       // Push closure
	OpIteratorInit                  // Iterator init
	OpIteratorNext                  // Iterator next
	OpIteratorClose                 // Iterator close
	OpBinaryOp                      // Binary operation
	OpUnaryOp                       // Unary operation
	OpCompare                       // Comparison operation
	OpSuspend                       // Suspend VM
)

// OpcodeNames are string representation of opcodes.
var OpcodeNames = [...]string{
	OpConstant:        "CONST",
	OpPop:             "POP",
	OpTrue:            "TRUE",
	OpFalse:           "FALSE",
	OpJumpFalsy:       "JMPF",
	OpAndJump:         "ANDJMP",
	OpOrJump:          "ORJMP",
	OpJump:            "JMP",
	OpNull:            "NULL",
	OpGetGlobal:       "GETG",
	OpSetGlobal:       "SETG",
	OpSetSelGlobal:    "SETSG",
	OpArray:           "ARR",
	OpMap:             "MAP",
	OpTuple:           "TUPLE",
	OpImmutable:       "IMMUT",
	OpIndex:           "INDEX",
	OpField:           "FIELD",
	OpSliceIndex:      "SLICE",
	OpSplat:           "SPLAT",
	OpCall:            "CALL",
	OpReturn:          "RET",
	OpSaveDefer:       "SAVEDEFER",
	OpPushDefer:       "PUSHDEFER",
	OpGetLocal:        "GETL",
	OpSetLocal:        "SETL",
	OpDefineLocal:     "DEFL",
	OpSetSelLocal:     "SETSL",
	OpGetFreePtr:      "GETFP",
	OpGetFree:         "GETF",
	OpGetLocalPtr:     "GETLP",
	OpSetSelFree:      "SETSF",
	OpSetFree:         "SETF",
	OpGetBuiltin:      "BUILTIN",
	OpIdxAssignAssert: "IDXASSERT",
	OpIdxElem:         "IDXELEM",
	OpClosure:         "CLOSURE",
	OpIteratorInit:    "ITER",
	OpIteratorNext:    "ITNEXT",
	OpIteratorClose:   "ITCLOSE",
	OpBinaryOp:        "BINARYOP",
	OpUnaryOp:         "UNARYOP",
	OpCompare:         "CMP",
	OpSuspend:         "SUSPEND",
}

// OpcodeOperands is the number of operands.
var OpcodeOperands = [...][]int{
	OpConstant:        {2},
	OpPop:             {},
	OpTrue:            {},
	OpFalse:           {},
	OpJumpFalsy:       {4},
	OpAndJump:         {4},
	OpOrJump:          {4},
	OpJump:            {4},
	OpNull:            {},
	OpGetGlobal:       {2},
	OpSetGlobal:       {2},
	OpSetSelGlobal:    {2, 1},
	OpArray:           {2, 1},
	OpMap:             {2},
	OpTuple:           {2, 1},
	OpImmutable:       {},
	OpIndex:           {},
	OpField:           {},
	OpSliceIndex:      {1},
	OpSplat:           {},
	OpCall:            {1, 1, 1},
	OpReturn:          {1},
	OpSaveDefer:       {1, 1},
	OpPushDefer:       {},
	OpGetLocal:        {1},
	OpSetLocal:        {1},
	OpDefineLocal:     {1},
	OpSetSelLocal:     {1, 1},
	OpGetFreePtr:      {1},
	OpGetFree:         {1},
	OpSetFree:         {1},
	OpGetLocalPtr:     {1},
	OpSetSelFree:      {1, 1},
	OpGetBuiltin:      {1},
	OpIdxAssignAssert: {2},
	OpIdxElem:         {2},
	OpClosure:         {2, 1},
	OpIteratorInit:    {},
	OpIteratorNext:    {1},
	OpIteratorClose:   {},
	OpBinaryOp:        {1},
	OpUnaryOp:         {1},
	OpCompare:         {1},
	OpSuspend:         {},
}

// ReadOperands reads operands from the bytecode.
func ReadOperands(numOperands []int, ins []byte) (operands []int, offset int) {
	for _, width := range numOperands {
		switch width {
		case 1:
			operands = append(operands, int(ins[offset]))
		case 2:
			operands = append(operands, int(ins[offset+1])|int(ins[offset])<<8)
		case 4:
			operands = append(operands, int(ins[offset+3])|int(ins[offset+2])<<8|int(ins[offset+1])<<16|int(ins[offset])<<24)
		}
		offset += width
	}
	return operands, offset
}
