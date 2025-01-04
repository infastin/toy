package toy

import (
	"fmt"
	"strings"
	"testing"

	"github.com/infastin/toy/parser"
	"github.com/infastin/toy/token"
)

func TestCompiler_Compile(t *testing.T) {
	expectCompile(t, `1 + 2`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.Add)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1; 2`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 - 2`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.Sub)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 * 2`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.Mul)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `2 / 1`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.Quo)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(2),
			Int(1),
		),
	))

	expectCompile(t, `true`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpTrue),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `false`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpFalse),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `1 > 2`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.Greater)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 < 2`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.Less)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 >= 2`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.GreaterEq)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 <= 2`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.LessEq)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 == 2`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.Equal)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 != 2`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.NotEqual)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `true == false`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpTrue),
			MakeInstruction(parser.OpFalse),
			MakeInstruction(parser.OpBinaryOp, int(token.Equal)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `true != false`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpTrue),
			MakeInstruction(parser.OpFalse),
			MakeInstruction(parser.OpBinaryOp, int(token.NotEqual)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `-1`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpUnaryOp, int(token.Sub)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(Int(1)),
	))

	expectCompile(t, `!true`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpTrue),
			MakeInstruction(parser.OpUnaryOp, int(token.Not)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `if true { 10 }; 3333`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpTrue),          // 0000
			MakeInstruction(parser.OpJumpFalsy, 10), // 0001
			MakeInstruction(parser.OpConstant, 0),   // 0004
			MakeInstruction(parser.OpPop),           // 0007
			MakeInstruction(parser.OpConstant, 1),   // 0008
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend), // 0011
		),
		objectsArray(
			Int(10),
			Int(3333),
		),
	))

	expectCompile(t, `if (true) { 10 } else { 20 }; 3333;`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpTrue),          // 0000
			MakeInstruction(parser.OpJumpFalsy, 15), // 0001
			MakeInstruction(parser.OpConstant, 0),   // 0004
			MakeInstruction(parser.OpPop),           // 0007
			MakeInstruction(parser.OpJump, 19),      // 0008
			MakeInstruction(parser.OpConstant, 1),   // 0011
			MakeInstruction(parser.OpPop),           // 0014
			MakeInstruction(parser.OpConstant, 2),   // 0015
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend), // 0018
		),
		objectsArray(
			Int(10),
			Int(20),
			Int(3333),
		)),
	)

	expectCompile(t, `"kami"`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			String("kami"),
		),
	))

	expectCompile(t, `"ka" + "mi"`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.Add)),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			String("ka"),
			String("mi"),
		),
	))

	expectCompile(t, `a := 1; b := 2; a += b`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpSetGlobal, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpSetGlobal, 1),
			MakeInstruction(parser.OpGetGlobal, 0),
			MakeInstruction(parser.OpGetGlobal, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.Add)),
			MakeInstruction(parser.OpSetGlobal, 0),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `a := 1; b := 2; a /= b`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpSetGlobal, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpSetGlobal, 1),
			MakeInstruction(parser.OpGetGlobal, 0),
			MakeInstruction(parser.OpGetGlobal, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.Quo)),
			MakeInstruction(parser.OpSetGlobal, 0),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `a, b := 1, 2`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpSetGlobal, 0),
			MakeInstruction(parser.OpSetGlobal, 1),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `a, b, c := unpack([1, 2, 3])`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpConstant, 3),
			MakeInstruction(parser.OpArray, 3, 0),
			MakeInstruction(parser.OpSeqToTuple),
			MakeInstruction(parser.OpTupleAssignAssert, 3),
			MakeInstruction(parser.OpTupleElem, 0),
			MakeInstruction(parser.OpSetGlobal, 0),
			MakeInstruction(parser.OpTupleElem, 1),
			MakeInstruction(parser.OpSetGlobal, 1),
			MakeInstruction(parser.OpTupleElem, 2),
			MakeInstruction(parser.OpSetGlobal, 2),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t, `[]`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpArray, 0, 0),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `[1, 2, 3]`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpArray, 3, 0),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t, `[1 + 2, 3 - 4, 5 * 6]`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.Add)),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpConstant, 3),
			MakeInstruction(parser.OpBinaryOp, int(token.Sub)),
			MakeInstruction(parser.OpConstant, 4),
			MakeInstruction(parser.OpConstant, 5),
			MakeInstruction(parser.OpBinaryOp, int(token.Mul)),
			MakeInstruction(parser.OpArray, 3, 0),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
			Int(4),
			Int(5),
			Int(6),
		),
	))

	expectCompile(t, `[1, 2, ...[3], 4, ...[5, 6]]`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpArray, 1, 0),
			MakeInstruction(parser.OpSplat),
			MakeInstruction(parser.OpConstant, 3),
			MakeInstruction(parser.OpConstant, 4),
			MakeInstruction(parser.OpConstant, 5),
			MakeInstruction(parser.OpArray, 2, 0),
			MakeInstruction(parser.OpSplat),
			MakeInstruction(parser.OpArray, 5, 1),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
			Int(4),
			Int(5),
			Int(6),
		),
	))

	expectCompile(t, `{}`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpMap, 0),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `{a: 2, b: 4, c: 6}`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpConstant, 3),
			MakeInstruction(parser.OpConstant, 4),
			MakeInstruction(parser.OpConstant, 5),
			MakeInstruction(parser.OpMap, 3),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			String("a"),
			Int(2),
			String("b"),
			Int(4),
			String("c"),
			Int(6),
		),
	))

	expectCompile(t, `{["a"]: 2, [2 * 3]: 4, ["b" + "c"]: 6}`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpBinaryOp, int(token.Mul)),
			MakeInstruction(parser.OpConstant, 3),
			MakeInstruction(parser.OpConstant, 4),
			MakeInstruction(parser.OpConstant, 5),
			MakeInstruction(parser.OpBinaryOp, int(token.Add)),
			MakeInstruction(parser.OpConstant, 6),
			MakeInstruction(parser.OpMap, 3),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			String("a"),
			Int(2),
			Int(3),
			Int(4),
			String("b"),
			String("c"),
			Int(6),
		),
	))

	expectCompile(t, `{a: 2 + 3, b: 5 * 6}`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpBinaryOp, int(token.Add)),
			MakeInstruction(parser.OpConstant, 3),
			MakeInstruction(parser.OpConstant, 4),
			MakeInstruction(parser.OpConstant, 5),
			MakeInstruction(parser.OpBinaryOp, int(token.Mul)),
			MakeInstruction(parser.OpMap, 2),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			String("a"),
			Int(2),
			Int(3),
			String("b"),
			Int(5),
			Int(6),
		),
	))

	expectCompile(t, `[1, 2, 3][1 + 1]`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpArray, 3, 0),
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpBinaryOp, int(token.Add)),
			MakeInstruction(parser.OpIndex),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t, `{a: 2}[2 - 1]`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpMap, 2),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpBinaryOp, int(token.Sub)),
			MakeInstruction(parser.OpIndex),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			String("a"),
			Int(2),
			Int(1),
		),
	))

	expectCompile(t, `[1, 2, 3][:]`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpArray, 3, 0),
			MakeInstruction(parser.OpSliceIndex, 0x0),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t, `[1, 2, 3][0:2]`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpArray, 3, 0),
			MakeInstruction(parser.OpConstant, 3),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpSliceIndex, 0x3),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend)),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
			Int(0),
		),
	))

	expectCompile(t, `[1, 2, 3][:2]`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpArray, 3, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpSliceIndex, 0x2),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t, `[1, 2, 3][0:]`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpArray, 3, 0),
			MakeInstruction(parser.OpConstant, 3),
			MakeInstruction(parser.OpSliceIndex, 0x1),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
			Int(0),
		),
	))

	expectCompile(t, `a := immutable([1, 2, 3])`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpArray, 3, 0),
			MakeInstruction(parser.OpImmutable),
			MakeInstruction(parser.OpSetGlobal, 0),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t,
		`f1 := fn(x, y, ...rest) { return x + y }; f1(...[1, 2], ...[3, 4, 5])`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpSetGlobal, 0),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpConstant, 2),
				MakeInstruction(parser.OpArray, 2, 0),
				MakeInstruction(parser.OpSplat),
				MakeInstruction(parser.OpConstant, 3),
				MakeInstruction(parser.OpConstant, 4),
				MakeInstruction(parser.OpConstant, 5),
				MakeInstruction(parser.OpArray, 3, 0),
				MakeInstruction(parser.OpSplat),
				MakeInstruction(parser.OpCall, 2, 1),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				compiledFunction(1, 1,
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpGetLocal, 1),
					MakeInstruction(parser.OpBinaryOp, int(token.Add)),
					MakeInstruction(parser.OpReturn, 1),
				),
				Int(1),
				Int(2),
				Int(3),
				Int(4),
				Int(5),
			),
		),
	)

	expectCompile(t, `fn() { return 5 + 10 }`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(5),
			Int(10),
			compiledFunction(0, 0,
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpBinaryOp, int(token.Add)),
				MakeInstruction(parser.OpReturn, 1),
			),
		),
	))

	expectCompile(t, `fn() { 5 + 10 }`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(5),
			Int(10),
			compiledFunction(0, 0,
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpBinaryOp, int(token.Add)),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpReturn, 0),
			),
		),
	))

	expectCompile(t, `fn() { 1; 2 }`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			compiledFunction(0, 0,
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpReturn, 0),
			),
		),
	))

	expectCompile(t, `fn() { 1; return 2 }`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			compiledFunction(0, 0,
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpReturn, 1),
			),
		),
	))

	expectCompile(t, `fn() { return 1, 2 }`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			compiledFunction(0, 0,
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpTuple, 2),
				MakeInstruction(parser.OpReturn, 1),
			),
		),
	))

	expectCompile(t, `a, b := (fn() { return 1, 2 })()`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpCall, 0, 0),
			MakeInstruction(parser.OpTupleAssignAssert, 2),
			MakeInstruction(parser.OpTupleElem, 0),
			MakeInstruction(parser.OpSetGlobal, 0),
			MakeInstruction(parser.OpTupleElem, 1),
			MakeInstruction(parser.OpSetGlobal, 1),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			compiledFunction(0, 0,
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpTuple, 2),
				MakeInstruction(parser.OpReturn, 1),
			),
		),
	))

	expectCompile(t,
		`fn() { if(true) { return 1 } else { return 2 } }`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 2),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				Int(1),
				Int(2),
				compiledFunction(0, 0,
					MakeInstruction(parser.OpTrue),          // 0000
					MakeInstruction(parser.OpJumpFalsy, 11), // 0001
					MakeInstruction(parser.OpConstant, 0),   // 0004
					MakeInstruction(parser.OpReturn, 1),     // 0007
					MakeInstruction(parser.OpConstant, 1),   // 0009
					MakeInstruction(parser.OpReturn, 1),     // 0012
				),
			),
		),
	)

	expectCompile(t,
		`fn() { 1; if(true) { 2 } else { 3 }; 4 }`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 4),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				Int(1),
				Int(2),
				Int(3),
				Int(4),
				compiledFunction(0, 0,
					MakeInstruction(parser.OpConstant, 0),   // 0000
					MakeInstruction(parser.OpPop),           // 0003
					MakeInstruction(parser.OpTrue),          // 0004
					MakeInstruction(parser.OpJumpFalsy, 19), // 0005
					MakeInstruction(parser.OpConstant, 1),   // 0008
					MakeInstruction(parser.OpPop),           // 0011
					MakeInstruction(parser.OpJump, 23),      // 0012
					MakeInstruction(parser.OpConstant, 2),   // 0015
					MakeInstruction(parser.OpPop),           // 0018
					MakeInstruction(parser.OpConstant, 3),   // 0019
					MakeInstruction(parser.OpPop),           // 0022
					MakeInstruction(parser.OpReturn, 0),     // 0023
				),
			),
		),
	)

	expectCompile(t, `fn() { }`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0,
				MakeInstruction(parser.OpReturn, 0),
			),
		),
	))

	expectCompile(t, `fn() { 24 }()`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpCall, 0, 0),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(24),
			compiledFunction(0, 0,
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpReturn, 0),
			),
		),
	))

	expectCompile(t, `fn() { return 24 }()`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpCall, 0, 0),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(24),
			compiledFunction(0, 0,
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpReturn, 1),
			),
		),
	))

	expectCompile(t, `noArg := fn() { 24 }; noArg();`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpSetGlobal, 0),
			MakeInstruction(parser.OpGetGlobal, 0),
			MakeInstruction(parser.OpCall, 0, 0),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(24),
			compiledFunction(0, 0,
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpReturn, 0),
			),
		),
	))

	expectCompile(t,
		`noArg := fn() { return 24 }; noArg();`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpSetGlobal, 0),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpCall, 0, 0),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				Int(24),
				compiledFunction(0, 0,
					MakeInstruction(parser.OpConstant, 0),
					MakeInstruction(parser.OpReturn, 1),
				),
			),
		),
	)

	expectCompile(t, `n := 55; fn() { n };`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpSetGlobal, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(55),
			compiledFunction(0, 0,
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpReturn, 0),
			),
		),
	))

	expectCompile(t, `fn() { n := 55; return n }`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(55),
			compiledFunction(1, 0,
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpDefineLocal, 0),
				MakeInstruction(parser.OpGetLocal, 0),
				MakeInstruction(parser.OpReturn, 1),
			),
		),
	))

	expectCompile(t,
		`fn() { a := 55; b := 77; return a + b }`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 2),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				Int(55),
				Int(77),
				compiledFunction(2, 0,
					MakeInstruction(parser.OpConstant, 0),
					MakeInstruction(parser.OpDefineLocal, 0),
					MakeInstruction(parser.OpConstant, 1),
					MakeInstruction(parser.OpDefineLocal, 1),
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpGetLocal, 1),
					MakeInstruction(parser.OpBinaryOp, int(token.Add)),
					MakeInstruction(parser.OpReturn, 1),
				),
			),
		),
	)

	expectCompile(t,
		`f1 := fn(a) { return a }; f1(24);`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpSetGlobal, 0),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpCall, 1, 0),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				compiledFunction(1, 1,
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpReturn, 1),
				),
				Int(24),
			),
		),
	)

	expectCompile(t,
		`varTest := fn(...a) { return a }; varTest(1,2,3);`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpSetGlobal, 0),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpConstant, 2),
				MakeInstruction(parser.OpConstant, 3),
				MakeInstruction(parser.OpCall, 3, 0),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				compiledFunction(1, 1,
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpReturn, 1)),
				Int(1),
				Int(2),
				Int(3),
			),
		),
	)

	expectCompile(t,
		`f1 := fn(a, b, c) { a; b; return c; }; f1(24, 25, 26);`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpSetGlobal, 0),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpConstant, 2),
				MakeInstruction(parser.OpConstant, 3),
				MakeInstruction(parser.OpCall, 3, 0),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				compiledFunction(3, 3,
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpPop),
					MakeInstruction(parser.OpGetLocal, 1),
					MakeInstruction(parser.OpPop),
					MakeInstruction(parser.OpGetLocal, 2),
					MakeInstruction(parser.OpReturn, 1),
				),
				Int(24),
				Int(25),
				Int(26),
			),
		),
	)

	expectCompile(t,
		`fn() { n := 55; n = 23; return n }`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 2),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				Int(55),
				Int(23),
				compiledFunction(1, 0,
					MakeInstruction(parser.OpConstant, 0),
					MakeInstruction(parser.OpDefineLocal, 0),
					MakeInstruction(parser.OpConstant, 1),
					MakeInstruction(parser.OpSetLocal, 0),
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpReturn, 1),
				),
			),
		),
	)

	expectCompile(t, `len([]);`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpGetBuiltin, 0),
			MakeInstruction(parser.OpArray, 0, 0),
			MakeInstruction(parser.OpCall, 1, 0),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `fn() { return len([]) }`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpPop),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0,
				MakeInstruction(parser.OpGetBuiltin, 0),
				MakeInstruction(parser.OpArray, 0, 0),
				MakeInstruction(parser.OpCall, 1, 0),
				MakeInstruction(parser.OpReturn, 1),
			),
		),
	))

	expectCompile(t,
		`fn(a) { fn(b) { return a + b } }`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				compiledFunction(1, 1,
					MakeInstruction(parser.OpGetFree, 0),
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpBinaryOp, int(token.Add)),
					MakeInstruction(parser.OpReturn, 1),
				),
				compiledFunction(1, 1,
					MakeInstruction(parser.OpGetLocalPtr, 0),
					MakeInstruction(parser.OpClosure, 0, 1),
					MakeInstruction(parser.OpPop),
					MakeInstruction(parser.OpReturn, 0),
				),
			),
		),
	)

	expectCompile(t, `
fn(a) {
	return fn(b) {
		return fn(c) {
			return a + b + c
		}
	}
}`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 2),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				compiledFunction(1, 1,
					MakeInstruction(parser.OpGetFree, 0),
					MakeInstruction(parser.OpGetFree, 1),
					MakeInstruction(parser.OpBinaryOp, int(token.Add)),
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpBinaryOp, int(token.Add)),
					MakeInstruction(parser.OpReturn, 1),
				),
				compiledFunction(1, 1,
					MakeInstruction(parser.OpGetFreePtr, 0),
					MakeInstruction(parser.OpGetLocalPtr, 0),
					MakeInstruction(parser.OpClosure, 0, 2),
					MakeInstruction(parser.OpReturn, 1),
				),
				compiledFunction(1, 1,
					MakeInstruction(parser.OpGetLocalPtr, 0),
					MakeInstruction(parser.OpClosure, 1, 1),
					MakeInstruction(parser.OpReturn, 1),
				),
			),
		),
	)

	expectCompile(t, `
g := 55;

fn() {
	a := 66;

	return fn() {
		b := 77;

		return fn() {
			c := 88;

			return g + a + b + c;
		}
	}
}`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpSetGlobal, 0),
				MakeInstruction(parser.OpConstant, 6),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				Int(55),
				Int(66),
				Int(77),
				Int(88),
				compiledFunction(1, 0,
					MakeInstruction(parser.OpConstant, 3),
					MakeInstruction(parser.OpDefineLocal, 0),
					MakeInstruction(parser.OpGetGlobal, 0),
					MakeInstruction(parser.OpGetFree, 0),
					MakeInstruction(parser.OpBinaryOp, int(token.Add)),
					MakeInstruction(parser.OpGetFree, 1),
					MakeInstruction(parser.OpBinaryOp, int(token.Add)),
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpBinaryOp, int(token.Add)),
					MakeInstruction(parser.OpReturn, 1),
				),
				compiledFunction(1, 0,
					MakeInstruction(parser.OpConstant, 2),
					MakeInstruction(parser.OpDefineLocal, 0),
					MakeInstruction(parser.OpGetFreePtr, 0),
					MakeInstruction(parser.OpGetLocalPtr, 0),
					MakeInstruction(parser.OpClosure, 4, 2),
					MakeInstruction(parser.OpReturn, 1),
				),
				compiledFunction(1, 0,
					MakeInstruction(parser.OpConstant, 1),
					MakeInstruction(parser.OpDefineLocal, 0),
					MakeInstruction(parser.OpGetLocalPtr, 0),
					MakeInstruction(parser.OpClosure, 5, 1),
					MakeInstruction(parser.OpReturn, 1),
				),
			),
		),
	)

	expectCompile(t, `for i := 0; i < 10; i++ {}`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpSetGlobal, 0),
			MakeInstruction(parser.OpGetGlobal, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpBinaryOp, int(token.Less)),
			MakeInstruction(parser.OpJumpFalsy, 35),
			MakeInstruction(parser.OpGetGlobal, 0),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpBinaryOp, int(token.Add)),
			MakeInstruction(parser.OpSetGlobal, 0),
			MakeInstruction(parser.OpJump, 6),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(
			Int(0),
			Int(10),
			Int(1),
		),
	))

	expectCompile(t, `m := {}; for k, v in m {}`, bytecode(
		concatInsts(
			MakeInstruction(parser.OpMap, 0),
			MakeInstruction(parser.OpSetGlobal, 0),
			MakeInstruction(parser.OpGetGlobal, 0),
			MakeInstruction(parser.OpIteratorInit),
			MakeInstruction(parser.OpSetGlobal, 1),
			MakeInstruction(parser.OpGetGlobal, 1),
			MakeInstruction(parser.OpIteratorNext),
			MakeInstruction(parser.OpJumpFalsy, 34),
			MakeInstruction(parser.OpSetGlobal, 3),
			MakeInstruction(parser.OpSetGlobal, 2),
			MakeInstruction(parser.OpJump, 13),
			MakeInstruction(parser.OpGetGlobal, 1),
			MakeInstruction(parser.OpIteratorClose),
			MakeInstruction(parser.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t,
		`a := 0; a == 0 && a != 1 || a < 1`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpSetGlobal, 0),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpBinaryOp, int(token.Equal)),
				MakeInstruction(parser.OpAndJump, 25),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpBinaryOp, int(token.NotEqual)),
				MakeInstruction(parser.OpOrJump, 38),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpBinaryOp, int(token.Less)),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				Int(0),
				Int(1),
			),
		),
	)

	// unknown module name
	expectCompileError(t, `import("user1")`, "module 'user1' not found")

	// too many errors
	expectCompileError(t, `
r["x"] = {
    @a:1,
    @b:1,
    @c:1,
    @d:1,
    @e:1,
    @f:1,
    @g:1,
    @h:1,
    @i:1,
    @j:1,
    @k:1
}
`, "Parse Error: illegal character U+0040 '@'\n\tat test:3:5 (and 10 more errors)")

	expectCompileError(t, `import("")`, "empty module name")

	// https://github.com/d5/tengo/issues/314
	expectCompileError(t, `(fn() { f := f() })()`, "unresolved reference 'f")
}

func TestCompilerErrorReport(t *testing.T) {
	expectCompileError(t, `import("user1")`,
		"Compile Error: module 'user1' not found\n\tat test:1:1")

	expectCompileError(t, `a = 1`,
		"Compile Error: unresolved reference 'a'\n\tat test:1:1")
	expectCompileError(t, `a := a`,
		"Compile Error: unresolved reference 'a'\n\tat test:1:6")
	expectCompileError(t, `a.b := 1`,
		"not allowed with selector")
	expectCompileError(t, `a := 1; a := 3`,
		"Compile Error: 'a' redeclared in this block\n\tat test:1:7")
	expectCompileError(t, `a,b := 1, 2; a,b := 2, 4`,
		"Compile Error: no new variables on the left side of :=\n\tat test:1:8")
	expectCompileError(t, `a, b := 1, 2, 3`,
		"Compile Error: trying to assign 3 values to 2 variables\n\tat test:1:9")

	expectCompileError(t, `return 5`,
		"Compile Error: return not allowed outside function\n\tat test:1:1")
	expectCompileError(t, `fn() { break }`,
		"Compile Error: break not allowed outside loop\n\tat test:1:10")
	expectCompileError(t, `fn() { continue }`,
		"Compile Error: continue not allowed outside loop\n\tat test:1:10")
	expectCompileError(t, `fn() { export 5 }`,
		"Compile Error: export not allowed inside function\n\tat test:1:10")
}

func TestCompilerDeadCode(t *testing.T) {
	expectCompile(t, `
fn() {
	a := 4
	return a

	b := 5 // dead code from here
	c := a
	return b
}`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 2),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				Int(4),
				Int(5),
				compiledFunction(0, 0,
					MakeInstruction(parser.OpConstant, 0),
					MakeInstruction(parser.OpDefineLocal, 0),
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpReturn, 1),
				),
			),
		),
	)

	expectCompile(t, `
fn() {
	if true {
		return 5
		a := 4  // dead code from here
		b := a
		return b
	} else {
		return 4
		c := 5  // dead code from here
		d := c
		return d
	}
}`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 2),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				Int(5),
				Int(4),
				compiledFunction(0, 0,
					MakeInstruction(parser.OpTrue),
					MakeInstruction(parser.OpJumpFalsy, 11),
					MakeInstruction(parser.OpConstant, 0),
					MakeInstruction(parser.OpReturn, 1),
					MakeInstruction(parser.OpConstant, 1),
					MakeInstruction(parser.OpReturn, 1),
				),
			),
		),
	)

	expectCompile(t, `
fn() {
	a := 1
	for {
		if a == 5 {
			return 10
		}
		5 + 5
		return 20
		b := a
		return b
	}
}`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 4),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend)),
			objectsArray(
				Int(1),
				Int(5),
				Int(10),
				Int(20),
				compiledFunction(0, 0,
					MakeInstruction(parser.OpConstant, 0),
					MakeInstruction(parser.OpDefineLocal, 0),
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpConstant, 1),
					MakeInstruction(parser.OpBinaryOp, int(token.Equal)),
					MakeInstruction(parser.OpJumpFalsy, 21),
					MakeInstruction(parser.OpConstant, 2),
					MakeInstruction(parser.OpReturn, 1),
					MakeInstruction(parser.OpConstant, 1),
					MakeInstruction(parser.OpConstant, 1),
					MakeInstruction(parser.OpBinaryOp, int(token.Add)),
					MakeInstruction(parser.OpPop),
					MakeInstruction(parser.OpConstant, 3),
					MakeInstruction(parser.OpReturn, 1),
				),
			),
		),
	)

	expectCompile(t, `
fn() {
	if true {
		return 5
		a := 4  // dead code from here
		b := a
		return b
	} else {
		return 4
		c := 5  // dead code from here
		d := c
		return d
	}
}`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 2),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				Int(5),
				Int(4),
				compiledFunction(0, 0,
					MakeInstruction(parser.OpTrue),
					MakeInstruction(parser.OpJumpFalsy, 11),
					MakeInstruction(parser.OpConstant, 0),
					MakeInstruction(parser.OpReturn, 1),
					MakeInstruction(parser.OpConstant, 1),
					MakeInstruction(parser.OpReturn, 1),
				),
			),
		),
	)

	expectCompile(t, `
fn() {
	if true {
		return
	}

    return

    return 123
}`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				Int(123),
				compiledFunction(0, 0,
					MakeInstruction(parser.OpTrue),
					MakeInstruction(parser.OpJumpFalsy, 8),
					MakeInstruction(parser.OpReturn, 0),
					MakeInstruction(parser.OpReturn, 0),
					MakeInstruction(parser.OpConstant, 0),
					MakeInstruction(parser.OpReturn, 1),
				),
			),
		),
	)
}

func TestCompilerScopes(t *testing.T) {
	expectCompile(t, `
if a := 1; a {
    a = 2
	b := a
} else {
    a = 3
	b := a
}`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 0),
				MakeInstruction(parser.OpSetGlobal, 0),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpJumpFalsy, 31),
				MakeInstruction(parser.OpConstant, 1),
				MakeInstruction(parser.OpSetGlobal, 0),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpSetGlobal, 1),
				MakeInstruction(parser.OpJump, 43),
				MakeInstruction(parser.OpConstant, 2),
				MakeInstruction(parser.OpSetGlobal, 0),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpSetGlobal, 2),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				Int(1),
				Int(2),
				Int(3),
			),
		),
	)

	expectCompile(t, `
fn() {
	if a := 1; a {
    	a = 2
		b := a
	} else {
    	a = 3
		b := a
	}
}`,
		bytecode(
			concatInsts(
				MakeInstruction(parser.OpConstant, 3),
				MakeInstruction(parser.OpPop),
				MakeInstruction(parser.OpSuspend),
			),
			objectsArray(
				Int(1),
				Int(2),
				Int(3),
				compiledFunction(0, 0,
					MakeInstruction(parser.OpConstant, 0),
					MakeInstruction(parser.OpDefineLocal, 0),
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpJumpFalsy, 26),
					MakeInstruction(parser.OpConstant, 1),
					MakeInstruction(parser.OpSetLocal, 0),
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpDefineLocal, 1),
					MakeInstruction(parser.OpJump, 35),
					MakeInstruction(parser.OpConstant, 2),
					MakeInstruction(parser.OpSetLocal, 0),
					MakeInstruction(parser.OpGetLocal, 0),
					MakeInstruction(parser.OpDefineLocal, 1),
					MakeInstruction(parser.OpReturn, 0),
				),
			),
		),
	)
}

// func TestCompiler_custom_extension(t *testing.T) {
// 	pathFileSource := "./testdata/issue286/test.mshk"
//
// 	modules := stdlib.GetModuleMap(stdlib.AllModuleNames()...)
//
// 	src, err := os.ReadFile(pathFileSource)
// 	require.NoError(t, err)
//
// 	// Escape shegang
// 	if len(src) > 1 && string(src[:2]) == "#!" {
// 		copy(src, "//")
// 	}
//
// 	fileSet := parser.NewFileSet()
// 	srcFile := fileSet.AddFile(filepath.Base(pathFileSource), -1, len(src))
//
// 	p := parser.NewParser(srcFile, src, nil)
// 	file, err := p.ParseFile()
// 	require.NoError(t, err)
//
// 	c := NewCompiler(srcFile, nil, nil, modules, nil)
// 	c.EnableFileImport(true)
// 	c.SetImportDir(filepath.Dir(pathFileSource))
//
// 	// Search for "*.toy" and ".mshk"(custom extension)
// 	c.SetImportFileExt(".toy", ".mshk")
//
// 	err = c.Compile(file)
// 	require.NoError(t, err)
// }

// func TestCompilerNewCompiler_default_file_extension(t *testing.T) {
// 	modules := stdlib.GetModuleMap(stdlib.AllModuleNames()...)
// 	input := "{}"
// 	fileSet := parser.NewFileSet()
// 	file := fileSet.AddFile("test", -1, len(input))
//
// 	c := NewCompiler(file, nil, nil, modules, nil)
// 	c.EnableFileImport(true)
//
// 	require.Equal(t, []string{".toy"}, c.GetImportFileExt(),
// 		"newly created compiler object must contain the default extension")
// }

// func TestCompilerSetImportExt_extension_name_validation(t *testing.T) {
// 	c := new(Compiler) // Instantiate a new compiler object with no initialization
//
// 	// Test of empty arg
// 	err := c.SetImportFileExt()
// 	require.Error(t, err, "empty arg should return an error")
//
// 	// Test of various arg types
// 	for _, test := range []struct {
// 		extensions []string
// 		expect     []string
// 		requireErr bool
// 		msgFail    string
// 	}{
// 		{[]string{".toy"}, []string{".toy"}, false,
// 			"well-formed extension should not return an error"},
// 		{[]string{""}, []string{".toy"}, true,
// 			"empty extension name should return an error"},
// 		{[]string{"foo"}, []string{".toy"}, true,
// 			"name without dot prefix should return an error"},
// 		{[]string{"foo.bar"}, []string{".toy"}, true,
// 			"malformed extension should return an error"},
// 		{[]string{"foo."}, []string{".toy"}, true,
// 			"malformed extension should return an error"},
// 		{[]string{".mshk"}, []string{".mshk"}, false,
// 			"name with dot prefix should be added"},
// 		{[]string{".foo", ".bar"}, []string{".foo", ".bar"}, false,
// 			"it should replace instead of appending"},
// 	} {
// 		err := c.SetImportFileExt(test.extensions...)
// 		if test.requireErr {
// 			require.Error(t, err, test.msgFail)
// 		}
// 		expect := test.expect
// 		actual := c.GetImportFileExt()
// 		require.Equal(t, expect, actual, test.msgFail)
// 	}
// }

func expectCompile(t *testing.T, input string, expected *Bytecode) {
	t.Helper()
	actual, trace, err := traceCompile(input, nil)

	var ok bool
	defer func() {
		if !ok {
			for _, tr := range trace {
				t.Log(tr)
			}
		}
	}()

	expectNoError(t, err)
	expectEqualBytecode(t, expected, actual)
	ok = true
}

func expectCompileError(t *testing.T, input, expected string) {
	_, trace, err := traceCompile(input, nil)
	var ok bool
	defer func() {
		if !ok {
			for _, tr := range trace {
				t.Log(tr)
			}
		}
	}()
	expectError(t, err)
	expectContains(t, err.Error(), expected, "invalid error string")
	ok = true
}

type compileTracer struct {
	Out []string
}

func (o *compileTracer) Write(p []byte) (n int, err error) {
	o.Out = append(o.Out, string(p))
	return len(p), nil
}

func traceCompile(input string, symbols map[string]Object,
) (res *Bytecode, trace []string, err error) {
	fileSet := parser.NewFileSet()
	file := fileSet.AddFile("test", -1, len(input))

	p := parser.NewParser(file, []byte(input), nil)

	symTable := NewSymbolTable()
	for name := range symbols {
		symTable.Define(name)
	}
	for idx, fn := range BuiltinFuncs {
		symTable.DefineBuiltin(idx, fn.Name)
	}

	tr := &compileTracer{}
	c := NewCompiler(file, symTable, nil, nil, tr)

	parsed, err := p.ParseFile()
	if err != nil {
		return nil, nil, err
	}

	err = c.Compile(parsed)

	res = c.Bytecode()
	res.RemoveDuplicates()

	trace = append(trace, fmt.Sprintf("Compiler Trace:\n%s",
		strings.Join(tr.Out, "")))
	trace = append(trace, fmt.Sprintf("Compiled Constants:\n%s",
		strings.Join(res.FormatConstants(), "\n")))
	trace = append(trace, fmt.Sprintf("Compiled Instructions:\n%s\n",
		strings.Join(res.FormatInstructions(), "\n")))

	return res, trace, err
}
