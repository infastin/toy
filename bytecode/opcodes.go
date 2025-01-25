package bytecode

import "encoding/binary"

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
	OpString                        // String literal
	OpArray                         // Array object
	OpMap                           // Map object
	OpTuple                         // Tuple object
	OpFreeze                        // Make object immutable
	OpIndex                         // Index access operation
	OpSetIndex                      // Index assignment operation
	OpField                         // Field access operation
	OpSetField                      // Field assignment operation
	OpSliceIndex                    // Slice operation
	OpSplat                         // Splat operation
	OpCall                          // Call function
	OpReturn                        // Return
	OpDefer                         // Defer function call
	OpRunDefer                      // Run defered function calls
	OpGetGlobal                     // Get global variable
	OpSetGlobal                     // Set global variable
	OpGetLocal                      // Get local variable
	OpSetLocal                      // Set local variable
	OpDefineLocal                   // Define local variable
	OpGetFreePtr                    // Get free variable pointer object
	OpGetFree                       // Get free variables
	OpSetFree                       // Set free variables
	OpGetLocalPtr                   // Get local variable as a pointer
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
	OpString:          "STR",
	OpArray:           "ARR",
	OpMap:             "MAP",
	OpTuple:           "TUPLE",
	OpFreeze:          "FREEZE",
	OpIndex:           "INDEX",
	OpSetIndex:        "SETINDEX",
	OpField:           "FIELD",
	OpSetField:        "SETFIELD",
	OpSliceIndex:      "SLICE",
	OpSplat:           "SPLAT",
	OpCall:            "CALL",
	OpReturn:          "RET",
	OpDefer:           "DEFER",
	OpRunDefer:        "RUNDEFER",
	OpGetLocal:        "GETL",
	OpSetLocal:        "SETL",
	OpDefineLocal:     "DEFL",
	OpGetFreePtr:      "GETFP",
	OpGetFree:         "GETF",
	OpGetLocalPtr:     "GETLP",
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
	OpString:          {2},
	OpArray:           {2, 1},
	OpMap:             {2},
	OpTuple:           {2, 1},
	OpFreeze:          {},
	OpIndex:           {1},
	OpSetIndex:        {},
	OpField:           {},
	OpSetField:        {},
	OpSliceIndex:      {1},
	OpSplat:           {},
	OpCall:            {1, 1},
	OpReturn:          {1},
	OpDefer:           {1, 1, 1},
	OpRunDefer:        {},
	OpGetLocal:        {1},
	OpSetLocal:        {1},
	OpDefineLocal:     {1},
	OpGetFreePtr:      {1},
	OpGetFree:         {1},
	OpSetFree:         {1},
	OpGetLocalPtr:     {1},
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

// Read2 reads a 2-byte operand.
func Read2(ins []byte) int {
	return int(binary.BigEndian.Uint16(ins))
}

// Read2 reads a 4-byte operand.
func Read4(ins []byte) int {
	return int(binary.BigEndian.Uint32(ins))
}

// ReadOperands reads operands from the bytecode.
func ReadOperands(numOperands []int, ins []byte) (operands []int, offset int) {
	for _, width := range numOperands {
		switch width {
		case 1:
			operands = append(operands, int(ins[offset]))
		case 2:
			operands = append(operands, Read2(ins[offset:]))
		case 4:
			operands = append(operands, Read4(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}
