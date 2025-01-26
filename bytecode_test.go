package toy

import (
	"testing"

	"github.com/infastin/toy/bytecode"
)

func TestBytecode_RemoveDuplicates(t *testing.T) {
	testBytecodeRemoveDuplicates(t, makeBytecode(
		concatInsts(),
		objectsArray(
			Char('y'),
			Float(93.11),
			compiledFunction(1, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 3),
				bytecode.MakeInstruction(bytecode.OpSetLocal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetFree, 0),
			),
			Float(39.2),
			Int(192),
			String("bar"),
		),
	), makeBytecode(
		concatInsts(),
		objectsArray(
			Char('y'),
			Float(93.11),
			compiledFunction(1, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 3),
				bytecode.MakeInstruction(bytecode.OpSetLocal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetFree, 0)),
			Float(39.2),
			Int(192),
			String("bar"),
		),
	))

	testBytecodeRemoveDuplicates(t, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpConstant, 3),
			bytecode.MakeInstruction(bytecode.OpConstant, 4),
			bytecode.MakeInstruction(bytecode.OpConstant, 5),
			bytecode.MakeInstruction(bytecode.OpConstant, 6),
			bytecode.MakeInstruction(bytecode.OpConstant, 7),
			bytecode.MakeInstruction(bytecode.OpConstant, 8),
			bytecode.MakeInstruction(bytecode.OpClosure, 4, 1),
		),
		objectsArray(
			Int(1),
			Float(2.0),
			Char('3'),
			String("four"),
			compiledFunction(1, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 3),
				bytecode.MakeInstruction(bytecode.OpConstant, 7),
				bytecode.MakeInstruction(bytecode.OpSetLocal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetFree, 0),
			),
			Int(1),
			Float(2.0),
			Char('3'),
			String("four"),
		),
	), makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpConstant, 3),
			bytecode.MakeInstruction(bytecode.OpConstant, 4),
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpConstant, 3),
			bytecode.MakeInstruction(bytecode.OpClosure, 4, 1),
		),
		objectsArray(
			Int(1),
			Float(2.0),
			Char('3'),
			String("four"),
			compiledFunction(1, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 3),
				bytecode.MakeInstruction(bytecode.OpConstant, 2),
				bytecode.MakeInstruction(bytecode.OpSetLocal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetFree, 0),
			),
		),
	))

	testBytecodeRemoveDuplicates(t, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpConstant, 3),
			bytecode.MakeInstruction(bytecode.OpConstant, 4),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
			Int(1),
			Int(3),
		),
	), makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	))
}

func testBytecodeRemoveDuplicates(t *testing.T, input, expected *Bytecode) {
	input.RemoveDuplicates()
	expectEqual(t, expected.FileSet, input.FileSet)
	expectEqual(t, expected.MainFunction, input.MainFunction)
	expectEqual(t, expected.Constants, input.Constants)
}

func TestBytecode_RemoveUnused(t *testing.T) {
	testBytecodeRemoveUnused(t, makeBytecode(
		concatInsts(),
		objectsArray(
			Char('y'),
			Float(93.11),
			compiledFunction(1, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 3),
				bytecode.MakeInstruction(bytecode.OpSetLocal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetFree, 0),
			),
			Float(39.2),
			Int(192),
			String("bar"),
		),
	), makeBytecode(
		concatInsts(),
		objectsArray(),
	))

	testBytecodeRemoveUnused(t, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
		),
		objectsArray(
			Int(1),
			Float(2.0),
			Char('3'),
			String("four"),
			compiledFunction(1, 0, 0, false,
				bytecode.MakeInstruction(bytecode.OpConstant, 3),
				bytecode.MakeInstruction(bytecode.OpConstant, 7),
				bytecode.MakeInstruction(bytecode.OpSetLocal, 0),
				bytecode.MakeInstruction(bytecode.OpGetGlobal, 0),
				bytecode.MakeInstruction(bytecode.OpGetFree, 0),
			),
		),
	), makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
		),
		objectsArray(
			Int(1),
		),
	))

	testBytecodeRemoveUnused(t, makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 2),
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
		),
	), makeBytecode(
		concatInsts(
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
			bytecode.MakeInstruction(bytecode.OpConstant, 1),
			bytecode.MakeInstruction(bytecode.OpConstant, 0),
		),
		objectsArray(
			Int(1),
			Int(3),
		),
	))
}

func testBytecodeRemoveUnused(t *testing.T, input, expected *Bytecode) {
	input.RemoveUnused()
	expectEqual(t, expected.FileSet, input.FileSet)
	expectEqual(t, expected.MainFunction, input.MainFunction)
	expectEqual(t, expected.Constants, input.Constants)
}
