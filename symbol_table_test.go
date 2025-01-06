package toy_test

import (
	"testing"

	"github.com/infastin/toy"
)

func TestSymbolTable(t *testing.T) {
	/*
		GLOBAL
		[0] a
		[1] b

			LOCAL 1
			[0] d

				LOCAL 2
				[0] e
				[1] f

				LOCAL 2 BLOCK 1
				[2] g
				[3] h

				LOCAL 2 BLOCK 2
				[2] i
				[3] j
				[4] k

			LOCAL 1 BLOCK 1
			[1] l
			[2] m
			[3] n
			[4] o
			[5] p

				LOCAL 3
				[0] q
				[1] r
	*/

	global := symbolTable()
	expectEqual(t, globalSymbol("a", 0), global.Define("a"))
	expectEqual(t, globalSymbol("b", 1), global.Define("b"))

	local1 := global.Fork(false)
	expectEqual(t, localSymbol("d", 0), local1.Define("d"))

	local1Block1 := local1.Fork(true)
	expectEqual(t, localSymbol("l", 1), local1Block1.Define("l"))
	expectEqual(t, localSymbol("m", 2), local1Block1.Define("m"))
	expectEqual(t, localSymbol("n", 3), local1Block1.Define("n"))
	expectEqual(t, localSymbol("o", 4), local1Block1.Define("o"))
	expectEqual(t, localSymbol("p", 5), local1Block1.Define("p"))

	local2 := local1.Fork(false)
	expectEqual(t, localSymbol("e", 0), local2.Define("e"))
	expectEqual(t, localSymbol("f", 1), local2.Define("f"))

	local2Block1 := local2.Fork(true)
	expectEqual(t, localSymbol("g", 2), local2Block1.Define("g"))
	expectEqual(t, localSymbol("h", 3), local2Block1.Define("h"))

	local2Block2 := local2.Fork(true)
	expectEqual(t, localSymbol("i", 2), local2Block2.Define("i"))
	expectEqual(t, localSymbol("j", 3), local2Block2.Define("j"))
	expectEqual(t, localSymbol("k", 4), local2Block2.Define("k"))

	local3 := local1Block1.Fork(false)
	expectEqual(t, localSymbol("q", 0), local3.Define("q"))
	expectEqual(t, localSymbol("r", 1), local3.Define("r"))

	expectEqual(t, 2, global.MaxSymbols())
	expectEqual(t, 6, local1.MaxSymbols())
	expectEqual(t, 6, local1Block1.MaxSymbols())
	expectEqual(t, 5, local2.MaxSymbols())
	expectEqual(t, 4, local2Block1.MaxSymbols())
	expectEqual(t, 5, local2Block2.MaxSymbols())
	expectEqual(t, 2, local3.MaxSymbols())

	resolveExpect(t, global, "a", globalSymbol("a", 0), 0)
	resolveExpect(t, local1, "d", localSymbol("d", 0), 0)
	resolveExpect(t, local1, "a", globalSymbol("a", 0), 1)
	resolveExpect(t, local3, "a", globalSymbol("a", 0), 3)
	resolveExpect(t, local3, "d", freeSymbol("d", 0), 2)
	resolveExpect(t, local3, "r", localSymbol("r", 1), 0)
	resolveExpect(t, local2Block2, "k", localSymbol("k", 4), 0)
	resolveExpect(t, local2Block2, "e", localSymbol("e", 0), 1)
	resolveExpect(t, local2Block2, "b", globalSymbol("b", 1), 3)
}

func symbol(name string, scope toy.SymbolScope, index int) *toy.Symbol {
	return &toy.Symbol{
		Name:  name,
		Scope: scope,
		Index: index,
	}
}

func globalSymbol(name string, index int) *toy.Symbol {
	return symbol(name, toy.ScopeGlobal, index)
}

func localSymbol(name string, index int) *toy.Symbol {
	return symbol(name, toy.ScopeLocal, index)
}

func freeSymbol(name string, index int) *toy.Symbol {
	return symbol(name, toy.ScopeFree, index)
}

func symbolTable() *toy.SymbolTable {
	return toy.NewSymbolTable()
}

func resolveExpect(
	t *testing.T,
	symbolTable *toy.SymbolTable,
	name string,
	expectedSymbol *toy.Symbol,
	expectedDepth int,
) {
	actualSymbol, actualDepth, ok := symbolTable.Resolve(name, true)
	expectTrue(t, ok)
	expectEqual(t, expectedSymbol, actualSymbol)
	expectEqual(t, expectedDepth, actualDepth)
}
