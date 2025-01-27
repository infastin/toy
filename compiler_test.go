package toy

import (
	"fmt"
	"strings"
	"testing"

	"github.com/infastin/toy/bytecode"
	"github.com/infastin/toy/parser"
	"github.com/infastin/toy/token"
)

func TestCompiler_Compile(t *testing.T) {
	expectCompile(t, `1 + 2`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1; 2`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 - 2`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Sub)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 * 2`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Mul)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `2 / 1`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Quo)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(2),
			Int(1),
		),
	))

	expectCompile(t, `true`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpTrue),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `false`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpFalse),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `1 > 2`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpCompare, int(token.Greater)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 < 2`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpCompare, int(token.Less)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 >= 2`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpCompare, int(token.GreaterEq)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 <= 2`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpCompare, int(token.LessEq)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 == 2`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpCompare, int(token.Equal)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `1 != 2`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpCompare, int(token.NotEqual)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `true == false`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpTrue),
			bytecode.MakeInstruction(bytecode.OpFalse),
			bytecode.MakeInstruction(bytecode.OpCompare, int(token.Equal)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `true != false`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpTrue),
			bytecode.MakeInstruction(bytecode.OpFalse),
			bytecode.MakeInstruction(bytecode.OpCompare, int(token.NotEqual)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `-1`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpUnaryOp, int(token.Sub)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(Int(1)),
	))

	expectCompile(t, `!true`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpTrue),
			bytecode.MakeInstruction(bytecode.OpUnaryOp, int(token.Not)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `if true { 10 }; 3333`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpTrue),          // 0000
			bytecode.MakeInstruction(bytecode.OpJumpFalsy, 10), // 0001
			bytecode.MakeInstruction(bytecode.OpConstant, 0),   // 0004
			bytecode.MakeInstruction(bytecode.OpPop),           // 0007
			bytecode.MakeInstruction(bytecode.OpConstant, 1),   // 0008
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend), // 0011
		),
		objectsArray(
			Int(10),
			Int(3333),
		),
	))

	expectCompile(t, `if (true) { 10 } else { 20 }; 3333;`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpTrue),          // 0000
			bytecode.MakeInstruction(bytecode.OpJumpFalsy, 15), // 0001
			bytecode.MakeInstruction(bytecode.OpConstant, 0),   // 0004
			bytecode.MakeInstruction(bytecode.OpPop),           // 0007
			bytecode.MakeInstruction(bytecode.OpJump, 19),      // 0008
			bytecode.MakeInstruction(bytecode.OpConstant, 1),   // 0011
			bytecode.MakeInstruction(bytecode.OpPop),           // 0014
			bytecode.MakeInstruction(bytecode.OpConstant, 2),   // 0015
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend), // 0018
		),
		objectsArray(
			Int(10),
			Int(20),
			Int(3333),
		)),
	)

	expectCompile(t, `"kami"`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			String("kami"),
		),
	))

	expectCompile(t, `"ka" + "mi"`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			String("ka"),
			String("mi"),
		),
	))

	expectCompile(t, `a := 1; b := 2; a += b`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 1),
			bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpGetGlobal, 1),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `a := 1; b := 2; a /= b`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 1),
			bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpGetGlobal, 1),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Quo)),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `a, b := 1, 2`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 1),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(2),
			Int(1),
		),
	))

	expectCompile(t, `a, b, c := [1, 2, 3]`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpArray, 3, 0),
			bytecode.MakeInstruction(bytecode.OpIdxAssignAssert, 3),
			bytecode.MakeInstruction(bytecode.OpIdxElem, 0),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpIdxElem, 1),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 1),
			bytecode.MakeInstruction(bytecode.OpIdxElem, 2),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 2),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t, `[]`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpArray, 0, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `[1, 2, 3]`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpArray, 3, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t, `[1 + 2, 3 - 4, 5 * 6]`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpConstant, 3),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Sub)),
			bytecode.MakeInstruction(bytecode.OpConstant, 4),
			bytecode.MakeInstruction(bytecode.OpConstant, 5),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Mul)),
			bytecode.MakeInstruction(bytecode.OpArray, 3, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
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

	expectCompile(t, `[1, 2, ...[3], 4, ...[5, 6]]`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpArray, 1, 0),
			bytecode.MakeInstruction(bytecode.OpSplat),
			bytecode.MakeInstruction(bytecode.OpConstant, 3),
			bytecode.MakeInstruction(bytecode.OpConstant, 4),
			bytecode.MakeInstruction(bytecode.OpConstant, 5),
			bytecode.MakeInstruction(bytecode.OpArray, 2, 0),
			bytecode.MakeInstruction(bytecode.OpSplat),
			bytecode.MakeInstruction(bytecode.OpArray, 5, 1),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
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

	expectCompile(t, `{}`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpMap, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `{a: 2, b: 4, c: 6}`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpConstant, 3),
			bytecode.MakeInstruction(bytecode.OpConstant, 4),
			bytecode.MakeInstruction(bytecode.OpConstant, 5),
			bytecode.MakeInstruction(bytecode.OpMap, 3),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
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

	expectCompile(t, `{["a"]: 2, [2 * 3]: 4, ["b" + "c"]: 6}`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Mul)),
			bytecode.MakeInstruction(bytecode.OpConstant, 3),
			bytecode.MakeInstruction(bytecode.OpConstant, 4),
			bytecode.MakeInstruction(bytecode.OpConstant, 5),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
			bytecode.MakeInstruction(bytecode.OpConstant, 6),
			bytecode.MakeInstruction(bytecode.OpMap, 3),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
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

	expectCompile(t, `{a: 2 + 3, b: 5 * 6}`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
			bytecode.MakeInstruction(bytecode.OpConstant, 3),
			bytecode.MakeInstruction(bytecode.OpConstant, 4),
			bytecode.MakeInstruction(bytecode.OpConstant, 5),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Mul)),
			bytecode.MakeInstruction(bytecode.OpMap, 2),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
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

	expectCompile(t, `[1, 2, 3][1 + 1]`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpArray, 3, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
			bytecode.MakeInstruction(bytecode.OpIndex, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t, `{a: 2}[2 - 1]`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpMap, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Sub)),
			bytecode.MakeInstruction(bytecode.OpIndex, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			String("a"),
			Int(2),
			Int(1),
		),
	))

	expectCompile(t, `[1, 2, 3][:]`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpArray, 3, 0),
			bytecode.MakeInstruction(bytecode.OpSliceIndex, 0x0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t, `[1, 2, 3][0:2]`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 3),
			bytecode.MakeInstruction(bytecode.OpArray, 3, 0),
			bytecode.MakeInstruction(bytecode.OpSliceIndex, 0x3),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend)),
		objectsArray(
			Int(0),
			Int(2),
			Int(1),
			Int(3),
		),
	))

	expectCompile(t, `[1, 2, 3][:2]`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpArray, 3, 0),
			bytecode.MakeInstruction(bytecode.OpSliceIndex, 0x2),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(2),
			Int(1),
			Int(3),
		),
	))

	expectCompile(t, `[1, 2, 3][0:]`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpConstant, 3),
			bytecode.MakeInstruction(bytecode.OpArray, 3, 0),
			bytecode.MakeInstruction(bytecode.OpSliceIndex, 0x1),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(0),
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t, `freeze([1, 2, 3])`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpGetBuiltin, 2),
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpArray, 3, 0),
			bytecode.MakeInstruction(bytecode.OpCall, 1, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t, `a := freeze([1, 2, 3])`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpGetBuiltin, 2),
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpArray, 3, 0),
			bytecode.MakeInstruction(bytecode.OpCall, 1, 0),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t, `a := tuple(1, 2, 3)`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpGetBuiltin, 26),
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpCall, 3, 0),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))

	expectCompile(t, `tuple(1, 2, ...[3], 4, ...[5, 6])`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpGetBuiltin, 26),
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpArray, 1, 0),
			bytecode.MakeInstruction(bytecode.OpSplat),
			bytecode.MakeInstruction(bytecode.OpConstant, 3),
			bytecode.MakeInstruction(bytecode.OpConstant, 4),
			bytecode.MakeInstruction(bytecode.OpConstant, 5),
			bytecode.MakeInstruction(bytecode.OpArray, 2, 0),
			bytecode.MakeInstruction(bytecode.OpSplat),
			bytecode.MakeInstruction(bytecode.OpCall, 5, 1),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
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

	expectCompile(t,
		`f1 := fn(x, y, ...rest) { return x + y }; f1(...[1, 2], ...[3, 4, 5])`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpConstant, 2),
				bytecode.MakeInstruction(bytecode.OpArray, 2, 0),
				bytecode.MakeInstruction(bytecode.OpSplat),
				bytecode.MakeInstruction(bytecode.OpConstant, 3),
				bytecode.MakeInstruction(bytecode.OpConstant, 4),
				bytecode.MakeInstruction(bytecode.OpConstant, 5),
				bytecode.MakeInstruction(bytecode.OpArray, 3, 0),
				bytecode.MakeInstruction(bytecode.OpSplat),
				bytecode.MakeInstruction(bytecode.OpCall, 2, 1),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(3, 3, 0, true,
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 1),
					bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(1),
				Int(2),
				Int(3),
				Int(4),
				Int(5),
			),
		),
	)

	expectCompile(t, `fn() { return 5 + 10 }`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpConstant, 2),
				bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
				bytecode.MakeInstruction(bytecode.OpReturn, 1),
			),
			Int(5),
			Int(10),
		),
	))

	expectCompile(t, `fn() { 5 + 10 }`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpConstant, 2),
				bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpReturn, 0),
			),
			Int(5),
			Int(10),
		),
	))

	expectCompile(t, `fn() => 1 + 2`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpConstant, 2),
				bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
				bytecode.MakeInstruction(bytecode.OpReturn, 1),
			),
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `fn() => tuple(1, 2)`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpGetBuiltin, 26),
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpConstant, 2),
				bytecode.MakeInstruction(bytecode.OpCall, 2, 0),
				bytecode.MakeInstruction(bytecode.OpReturn, 1),
			),
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `fn() { 1; 2 }`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpConstant, 2),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpReturn, 0),
			),
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `fn() { 1; return 2 }`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpConstant, 2),
				bytecode.MakeInstruction(bytecode.OpReturn, 1),
			),
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `fn() { return 1, 2 }`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpConstant, 2),
				bytecode.MakeInstruction(bytecode.OpTuple, 2, 0),
				bytecode.MakeInstruction(bytecode.OpReturn, 1),
			),
			Int(1),
			Int(2),
		),
	))

	expectCompile(t, `a, b := (fn() { return 1, 2 })()`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpCall, 0, 0),
			bytecode.MakeInstruction(bytecode.OpIdxAssignAssert, 2),
			bytecode.MakeInstruction(bytecode.OpIdxElem, 0),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpIdxElem, 1),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 1),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpConstant, 2),
				bytecode.MakeInstruction(bytecode.OpTuple, 2, 0),
				bytecode.MakeInstruction(bytecode.OpReturn, 1),
			),
			Int(1),
			Int(2),
		),
	))

	expectCompile(t,
		`fn() { if (true) { return 1 } else { return 2 } }`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(0, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpTrue),          // 0000
					bytecode.MakeInstruction(bytecode.OpJumpFalsy, 11), // 0001
					bytecode.MakeInstruction(bytecode.OpConstant, 1),   // 0004
					bytecode.MakeInstruction(bytecode.OpReturn, 1),     // 0007
					bytecode.MakeInstruction(bytecode.OpConstant, 2),   // 0009
					bytecode.MakeInstruction(bytecode.OpReturn, 1),     // 0012
				),
				Int(1),
				Int(2),
			),
		),
	)

	expectCompile(t,
		`fn() { 1; if (true) { 2 } else { 3 }; 4 }`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(0, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpConstant, 1),   // 0000
					bytecode.MakeInstruction(bytecode.OpPop),           // 0003
					bytecode.MakeInstruction(bytecode.OpTrue),          // 0004
					bytecode.MakeInstruction(bytecode.OpJumpFalsy, 19), // 0005
					bytecode.MakeInstruction(bytecode.OpConstant, 2),   // 0008
					bytecode.MakeInstruction(bytecode.OpPop),           // 0011
					bytecode.MakeInstruction(bytecode.OpJump, 23),      // 0012
					bytecode.MakeInstruction(bytecode.OpConstant, 3),   // 0015
					bytecode.MakeInstruction(bytecode.OpPop),           // 0018
					bytecode.MakeInstruction(bytecode.OpConstant, 4),   // 0019
					bytecode.MakeInstruction(bytecode.OpPop),           // 0022
					bytecode.MakeInstruction(bytecode.OpReturn, 0),     // 0023
				),
				Int(1),
				Int(2),
				Int(3),
				Int(4),
			),
		),
	)

	expectCompile(t, `fn() {}`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpReturn, 0),
			),
		),
	))

	expectCompile(t, `fn() { 24 }()`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpCall, 0, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpReturn, 0),
			),
			Int(24),
		),
	))

	expectCompile(t, `fn() { return 24 }()`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpCall, 0, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpReturn, 1),
			),
			Int(24),
		),
	))

	expectCompile(t, `noArg := fn() { 24 }; noArg();`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpCall, 0, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpReturn, 0),
			),
			Int(24),
		),
	))

	expectCompile(t,
		`noArg := fn() { return 24 }; noArg();`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpCall, 0, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(0, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpConstant, 1),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(24),
			),
		),
	)

	expectCompile(t, `n := 55; fn() { n };`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(55),
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpReturn, 0),
			),
		),
	))

	expectCompile(t, `fn() { n := 55; return n }`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(1, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpDefineLocal, 0),
				bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
				bytecode.MakeInstruction(bytecode.OpReturn, 1),
			),
			Int(55),
		),
	))

	expectCompile(t,
		`fn() { a := 55; b := 77; return a + b }`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(2, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpConstant, 1),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 0),
					bytecode.MakeInstruction(bytecode.OpConstant, 2),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 1),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 1),
					bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(55),
				Int(77),
			),
		),
	)

	expectCompile(t,
		`f1 := fn(a) { return a }; f1(24);`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpCall, 1, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(1, 1, 0, false,
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(24),
			),
		),
	)

	expectCompile(t,
		`varTest := fn(...a) { return a }; varTest(1,2,3);`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpConstant, 2),
				bytecode.MakeInstruction(bytecode.OpConstant, 3),
				bytecode.MakeInstruction(bytecode.OpCall, 3, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(1, 1, 0, true,
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpReturn, 1)),
				Int(1),
				Int(2),
				Int(3),
			),
		),
	)

	expectCompile(t,
		`f1 := fn(a, b, c) { a; b; return c; }; f1(24, 25, 26);`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpConstant, 2),
				bytecode.MakeInstruction(bytecode.OpConstant, 3),
				bytecode.MakeInstruction(bytecode.OpCall, 3, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(3, 3, 0, false,
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpPop),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 1),
					bytecode.MakeInstruction(bytecode.OpPop),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 2),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(24),
				Int(25),
				Int(26),
			),
		),
	)

	expectCompile(t,
		`fn() { n := 55; n = 23; return n }`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(1, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpConstant, 1),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 0),
					bytecode.MakeInstruction(bytecode.OpConstant, 2),
					bytecode.MakeInstruction(bytecode.OpSetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(55),
				Int(23),
			),
		),
	)

	expectCompile(t, `len([]);`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpGetBuiltin, 5),
			bytecode.MakeInstruction(bytecode.OpArray, 0, 0),
			bytecode.MakeInstruction(bytecode.OpCall, 1, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t, `fn() { return len([]) }`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpPop),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			compiledFunction(0, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpGetBuiltin, 5),
				bytecode.MakeInstruction(bytecode.OpArray, 0, 0),
				bytecode.MakeInstruction(bytecode.OpCall, 1, 0),
				bytecode.MakeInstruction(bytecode.OpReturn, 1),
			),
		),
	))

	expectCompile(t,
		`fn(a) { fn(b) { return a + b } }`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(1, 1, 0, false,
					bytecode.MakeInstruction(bytecode.OpGetLocalPtr, 0),
					bytecode.MakeInstruction(bytecode.OpClosure, 1, 1),
					bytecode.MakeInstruction(bytecode.OpPop),
					bytecode.MakeInstruction(bytecode.OpReturn, 0),
				),
				compiledFunction(1, 1, 0, false,
					bytecode.MakeInstruction(bytecode.OpGetFree, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
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
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(1, 1, 0, false,
					bytecode.MakeInstruction(bytecode.OpGetLocalPtr, 0),
					bytecode.MakeInstruction(bytecode.OpClosure, 1, 1),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				compiledFunction(1, 1, 0, false,
					bytecode.MakeInstruction(bytecode.OpGetFreePtr, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocalPtr, 0),
					bytecode.MakeInstruction(bytecode.OpClosure, 2, 2),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				compiledFunction(1, 1, 0, false,
					bytecode.MakeInstruction(bytecode.OpGetFree, 0),
					bytecode.MakeInstruction(bytecode.OpGetFree, 1),
					bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
			),
		),
	)

	expectCompile(t, `
g := 55
fn() {
	a := 66
	return fn() {
		b := 77
		return fn() {
			c := 88
			return g + a + b + c
		}
	}
}`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				Int(55),
				compiledFunction(1, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpConstant, 2),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocalPtr, 0),
					bytecode.MakeInstruction(bytecode.OpClosure, 3, 1),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(66),
				compiledFunction(1, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpConstant, 4),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 0),
					bytecode.MakeInstruction(bytecode.OpGetFreePtr, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocalPtr, 0),
					bytecode.MakeInstruction(bytecode.OpClosure, 5, 2),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(77),
				compiledFunction(1, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpConstant, 6),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 0),
					bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
					bytecode.MakeInstruction(bytecode.OpGetFree, 0),
					bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
					bytecode.MakeInstruction(bytecode.OpGetFree, 1),
					bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(88),
			),
		),
	)

	expectCompile(t, `for i := 0; i < 10; i++ {}`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpCompare, int(token.Less)),
			bytecode.MakeInstruction(bytecode.OpJumpFalsy, 35),
			bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpJump, 6),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(
			Int(0),
			Int(10),
			Int(1),
		),
	))

	expectCompile(t, `m := {}; for k, v in m {}`, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpMap, 0),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
			bytecode.MakeInstruction(bytecode.OpIteratorInit),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 1),
			bytecode.MakeInstruction(bytecode.OpGetGlobal, 1),
			bytecode.MakeInstruction(bytecode.OpIteratorNext, 0x3),
			bytecode.MakeInstruction(bytecode.OpJumpFalsy, 34),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 3),
			bytecode.MakeInstruction(bytecode.OpSetGlobal, 2),
			bytecode.MakeInstruction(bytecode.OpJump, 13),
			bytecode.MakeInstruction(bytecode.OpGetGlobal, 1),
			bytecode.MakeInstruction(bytecode.OpIteratorClose),
			bytecode.MakeInstruction(bytecode.OpSuspend),
		),
		objectsArray(),
	))

	expectCompile(t,
		`a := 0; a == 0 && a != 1 || a < 1`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpCompare, int(token.Equal)),
				bytecode.MakeInstruction(bytecode.OpAndJump, 27),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpCompare, int(token.NotEqual)),
				bytecode.MakeInstruction(bytecode.OpOrJump, 40),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpCompare, int(token.Less)),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				Int(0),
				Int(1),
			),
		),
	)

	expectCompile(t, `
fn() {
	x := 10
	for i := 0; i < 10; i++ {
		defer fn() { x = i }()
	}
	return x
}`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(2, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpConstant, 1),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 0),
					bytecode.MakeInstruction(bytecode.OpConstant, 2),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 1),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 1),
					bytecode.MakeInstruction(bytecode.OpConstant, 1),
					bytecode.MakeInstruction(bytecode.OpCompare, int(token.Less)),
					bytecode.MakeInstruction(bytecode.OpJumpFalsy, 48),
					bytecode.MakeInstruction(bytecode.OpGetLocalPtr, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocalPtr, 1),
					bytecode.MakeInstruction(bytecode.OpClosure, 3, 2),
					bytecode.MakeInstruction(bytecode.OpDefer, 0, 0, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 1),
					bytecode.MakeInstruction(bytecode.OpConstant, 4),
					bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
					bytecode.MakeInstruction(bytecode.OpSetLocal, 1),
					bytecode.MakeInstruction(bytecode.OpJump, 10),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpRunDefer),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(10),
				Int(0),
				compiledFunction(0, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpGetFree, 1),
					bytecode.MakeInstruction(bytecode.OpSetFree, 0),
					bytecode.MakeInstruction(bytecode.OpReturn, 0),
				),
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
`, "Parse Error: illegal character U+0040 '@'\n\tat test:3:3 (and 10 more errors)")

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
		"Compile Error: 'a' redeclared in this block\n\tat test:1:9")
	expectCompileError(t, `a,b := 1, 2; a,b := 2, 4`,
		"Compile Error: no new variables on the left side of :=\n\tat test:1:14")
	expectCompileError(t, `a, b := 1, 2, 3`,
		"Compile Error: trying to assign 3 value(s) to 2 variable(s)\n\tat test:1:1")

	expectCompileError(t, `return 5`,
		"Compile Error: return not allowed outside function\n\tat test:1:1")
	expectCompileError(t, `fn() { break }`,
		"Compile Error: break not allowed outside loop\n\tat test:1:8")
	expectCompileError(t, `fn() { continue }`,
		"Compile Error: continue not allowed outside loop\n\tat test:1:8")
	expectCompileError(t, `fn() { export 5 }`,
		"Compile Error: export not allowed inside function\n\tat test:1:8")
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
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(1, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpConstant, 1),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(4),
			),
		),
	)

	expectCompile(t, `
fn(x) {
	if x > 10 {
		return 10
		a := 4 // dead code from here
	}
	a := 5 // not dead code for now
	return a
}`,
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(2, 1, 0, false,
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpConstant, 1),
					bytecode.MakeInstruction(bytecode.OpCompare, int(token.Greater)),
					bytecode.MakeInstruction(bytecode.OpJumpFalsy, 17),
					bytecode.MakeInstruction(bytecode.OpConstant, 1),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
					bytecode.MakeInstruction(bytecode.OpConstant, 2),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 1),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 1),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(10),
				Int(5),
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
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(0, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpTrue),
					bytecode.MakeInstruction(bytecode.OpJumpFalsy, 11),
					bytecode.MakeInstruction(bytecode.OpConstant, 1),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
					bytecode.MakeInstruction(bytecode.OpConstant, 2),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(5),
				Int(4),
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
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend)),
			objectsArray(
				compiledFunction(1, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpConstant, 1),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpConstant, 2),
					bytecode.MakeInstruction(bytecode.OpCompare, int(token.Equal)),
					bytecode.MakeInstruction(bytecode.OpJumpFalsy, 22),
					bytecode.MakeInstruction(bytecode.OpConstant, 3),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
					bytecode.MakeInstruction(bytecode.OpConstant, 2),
					bytecode.MakeInstruction(bytecode.OpConstant, 2),
					bytecode.MakeInstruction(bytecode.OpBinaryOp, int(token.Add)),
					bytecode.MakeInstruction(bytecode.OpPop),
					bytecode.MakeInstruction(bytecode.OpConstant, 4),
					bytecode.MakeInstruction(bytecode.OpReturn, 1),
				),
				Int(1),
				Int(5),
				Int(10),
				Int(20),
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
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(0, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpTrue),
					bytecode.MakeInstruction(bytecode.OpJumpFalsy, 8),
					bytecode.MakeInstruction(bytecode.OpReturn, 0),
					bytecode.MakeInstruction(bytecode.OpReturn, 0),
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
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpJumpFalsy, 31),
				bytecode.MakeInstruction(bytecode.OpConstant, 1),
				bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpSetGlobal, 1),
				bytecode.MakeInstruction(bytecode.OpJump, 43),
				bytecode.MakeInstruction(bytecode.OpConstant, 2),
				bytecode.MakeInstruction(bytecode.OpSetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpSetGlobal, 2),
				bytecode.MakeInstruction(bytecode.OpSuspend),
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
		makeBytecode(
			concatInsts(
				bytecode.MakeInstruction(bytecode.OpConstant, 0),
				bytecode.MakeInstruction(bytecode.OpPop),
				bytecode.MakeInstruction(bytecode.OpSuspend),
			),
			objectsArray(
				compiledFunction(2, 0, 0, false,
					bytecode.MakeInstruction(bytecode.OpConstant, 1),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpJumpFalsy, 26),
					bytecode.MakeInstruction(bytecode.OpConstant, 2),
					bytecode.MakeInstruction(bytecode.OpSetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 1),
					bytecode.MakeInstruction(bytecode.OpJump, 35),
					bytecode.MakeInstruction(bytecode.OpConstant, 3),
					bytecode.MakeInstruction(bytecode.OpSetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpGetLocal, 0),
					bytecode.MakeInstruction(bytecode.OpDefineLocal, 1),
					bytecode.MakeInstruction(bytecode.OpReturn, 0),
				),
				Int(1),
				Int(2),
				Int(3),
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
// 	expectNoError(t, err)
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
// 	expectNoError(t, err)
//
// 	c := NewCompiler(srcFile, nil, nil, modules, nil)
// 	c.EnableFileImport(true)
// 	c.SetImportDir(filepath.Dir(pathFileSource))
//
// 	// Search for "*.toy" and ".mshk"(custom extension)
// 	c.SetImportFileExt(".toy", ".mshk")
//
// 	err = c.Compile(file)
// 	expectNoError(t, err)
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
// 	expectEqual(t, []string{".toy"}, c.GetImportFileExt(),
// 		"newly created compiler object must contain the default extension")
// }

// func TestCompilerSetImportExt_extension_name_validation(t *testing.T) {
// 	c := new(Compiler) // Instantiate a new compiler object with no initialization
//
// 	// Test of empty arg
// 	err := c.SetImportFileExt()
// 	expectError(t, err, "empty arg should return an error")
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
// 			expectError(t, err, test.msgFail)
// 		}
// 		expect := test.expect
// 		actual := c.GetImportFileExt()
// 		expectEqual(t, expect, actual, test.msgFail)
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
	expectEqual(t, expected, actual)
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
	fileSet := token.NewFileSet()
	file := fileSet.AddFile("test", -1, len(input))

	p := parser.NewParser(file, []byte(input), nil)

	symTable := NewSymbolTable()
	for name := range symbols {
		symTable.Define(name)
	}
	for i, v := range Universe {
		symTable.DefineBuiltin(i, v.name)
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
	res.RemoveUnused()

	trace = append(trace, fmt.Sprintf("Compiler Trace:\n%s",
		strings.Join(tr.Out, "")))
	trace = append(trace, fmt.Sprintf("Compiled Constants:\n%s",
		strings.Join(res.FormatConstants(), "\n")))
	trace = append(trace, fmt.Sprintf("Compiled Instructions:\n%s\n",
		strings.Join(res.FormatInstructions(), "\n")))

	return res, trace, err
}
