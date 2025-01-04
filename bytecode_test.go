package toy

import (
	"testing"

	"github.com/infastin/toy/parser"
)

func TestBytecode_RemoveDuplicates(t *testing.T) {
	testBytecodeRemoveDuplicates(t, bytecode(
		concatInsts(),
		objectsArray(
			Char('y'),
			Float(93.11),
			compiledFunction(1, 0,
				MakeInstruction(parser.OpConstant, 3),
				MakeInstruction(parser.OpSetLocal, 0),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpGetFree, 0),
			),
			Float(39.2),
			Int(192),
			String("bar"),
		),
	), bytecode(
		concatInsts(),
		objectsArray(
			Char('y'),
			Float(93.11),
			compiledFunction(1, 0,
				MakeInstruction(parser.OpConstant, 3),
				MakeInstruction(parser.OpSetLocal, 0),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpGetFree, 0)),
			Float(39.2),
			Int(192),
			String("bar"),
		),
	))

	testBytecodeRemoveDuplicates(t, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpConstant, 3),
			MakeInstruction(parser.OpConstant, 4),
			MakeInstruction(parser.OpConstant, 5),
			MakeInstruction(parser.OpConstant, 6),
			MakeInstruction(parser.OpConstant, 7),
			MakeInstruction(parser.OpConstant, 8),
			MakeInstruction(parser.OpClosure, 4, 1),
		),
		objectsArray(
			Int(1),
			Float(2.0),
			Char('3'),
			String("four"),
			compiledFunction(1, 0,
				MakeInstruction(parser.OpConstant, 3),
				MakeInstruction(parser.OpConstant, 7),
				MakeInstruction(parser.OpSetLocal, 0),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpGetFree, 0),
			),
			Int(1),
			Float(2.0),
			Char('3'),
			String("four"),
		),
	), bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpConstant, 3),
			MakeInstruction(parser.OpConstant, 4),
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpConstant, 3),
			MakeInstruction(parser.OpClosure, 4, 1),
		),
		objectsArray(
			Int(1),
			Float(2.0),
			Char('3'),
			String("four"),
			compiledFunction(1, 0,
				MakeInstruction(parser.OpConstant, 3),
				MakeInstruction(parser.OpConstant, 2),
				MakeInstruction(parser.OpSetLocal, 0),
				MakeInstruction(parser.OpGetGlobal, 0),
				MakeInstruction(parser.OpGetFree, 0),
			),
		),
	))

	testBytecodeRemoveDuplicates(t, bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpConstant, 3),
			MakeInstruction(parser.OpConstant, 4),
		),
		objectsArray(
			Int(1),
			Int(2),
			Int(3),
			Int(1),
			Int(3),
		),
	), bytecode(
		concatInsts(
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 1),
			MakeInstruction(parser.OpConstant, 2),
			MakeInstruction(parser.OpConstant, 0),
			MakeInstruction(parser.OpConstant, 2),
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
