package toy

import (
	"fmt"
	"reflect"

	"github.com/infastin/toy/parser"
)

// Bytecode is a compiled instructions and constants.
type Bytecode struct {
	FileSet      *parser.SourceFileSet
	MainFunction *CompiledFunction
	Constants    []Object
}

// FormatInstructions returns human readable string representations of
// compiled instructions.
func (b *Bytecode) FormatInstructions() []string {
	return FormatInstructions(b.MainFunction.instructions, 0)
}

// FormatConstants returns human readable string representations of
// compiled constants.
func (b *Bytecode) FormatConstants() (output []string) {
	for cidx, cn := range b.Constants {
		switch cn := cn.(type) {
		case *CompiledFunction:
			output = append(output, fmt.Sprintf(
				"[% 3d] (Compiled Function|%p)", cidx, &cn))
			for _, l := range FormatInstructions(cn.instructions, 0) {
				output = append(output, fmt.Sprintf("     %s", l))
			}
		default:
			output = append(output, fmt.Sprintf("[% 3d] %s (%s|%p)",
				cidx, cn, reflect.TypeOf(cn).Name(), &cn))
		}
	}
	return output
}

// RemoveDuplicates finds and removes the duplicate values in Constants.
// NOTE: this function mutates Bytecode.
func (b *Bytecode) RemoveDuplicates() {
	var deduped []Object

	indexMap := make(map[int]int) // mapping from old constant index to new index
	fns := make(map[*CompiledFunction]int)
	ints := make(map[Int]int)
	strings := make(map[String]int)
	floats := make(map[Float]int)
	chars := make(map[Char]int)
	modules := make(map[string]int)

	for curIdx, c := range b.Constants {
		switch c := c.(type) {
		case *CompiledFunction:
			if newIdx, ok := fns[c]; ok {
				indexMap[curIdx] = newIdx
			} else {
				newIdx = len(deduped)
				fns[c] = newIdx
				indexMap[curIdx] = newIdx
				deduped = append(deduped, c)
			}
		case *BuiltinModule:
			newIdx, ok := modules[c.Name]
			if c.Name != "" && ok {
				indexMap[curIdx] = newIdx
			} else {
				newIdx = len(deduped)
				modules[c.Name] = newIdx
				indexMap[curIdx] = newIdx
				deduped = append(deduped, c)
			}
		case Int:
			if newIdx, ok := ints[c]; ok {
				indexMap[curIdx] = newIdx
			} else {
				newIdx = len(deduped)
				ints[c] = newIdx
				indexMap[curIdx] = newIdx
				deduped = append(deduped, c)
			}
		case String:
			if newIdx, ok := strings[c]; ok {
				indexMap[curIdx] = newIdx
			} else {
				newIdx = len(deduped)
				strings[c] = newIdx
				indexMap[curIdx] = newIdx
				deduped = append(deduped, c)
			}
		case Float:
			if newIdx, ok := floats[c]; ok {
				indexMap[curIdx] = newIdx
			} else {
				newIdx = len(deduped)
				floats[c] = newIdx
				indexMap[curIdx] = newIdx
				deduped = append(deduped, c)
			}
		case Char:
			if newIdx, ok := chars[c]; ok {
				indexMap[curIdx] = newIdx
			} else {
				newIdx = len(deduped)
				chars[c] = newIdx
				indexMap[curIdx] = newIdx
				deduped = append(deduped, c)
			}
		default:
			panic(fmt.Errorf("unsupported top-level constant type: %s", TypeName(c)))
		}
	}

	// replace with de-duplicated constants
	b.Constants = deduped

	// update CONST instructions in main function with new indexes
	updateConstIndexes(b.MainFunction.instructions, indexMap)
	// update CONST instructions in other compiled functions
	for _, c := range b.Constants {
		switch c := c.(type) {
		case *CompiledFunction:
			updateConstIndexes(c.instructions, indexMap)
		}
	}
}

// RemoveUnused finds and removes the unused values in Constants.
// NOTE: this function mutates Bytecode.
func (b *Bytecode) RemoveUnused() {
	indexMap := make(map[int]int)
	// find constants in use by looking at main function
	stripped := b.removeUnused(b.MainFunction.instructions, nil, indexMap)
	// iterate over other visible compiled functions
	// and try to find more constants
	for i := 0; i < len(stripped); i++ {
		switch c := stripped[i]; c := c.(type) {
		case *CompiledFunction:
			stripped = b.removeUnused(c.instructions, stripped, indexMap)
		}
	}
	// at this point we are left with the list of constants in use
	b.Constants = stripped
}

func (b *Bytecode) removeUnused(insts []byte, stripped []Object, indexMap map[int]int) []Object {
	for i := 0; i < len(insts); {
		opcode := insts[i]
		operands, offset := parser.ReadOperands(parser.OpcodeOperands[opcode], insts[i+1:])
		switch opcode {
		case parser.OpConstant:
			curIdx := operands[0]
			newIdx, ok := indexMap[curIdx]
			if !ok {
				newIdx = len(stripped)
				stripped = append(stripped, b.Constants[curIdx])
				indexMap[curIdx] = newIdx
			}
			copy(insts[i:], MakeInstruction(opcode, newIdx))
		case parser.OpClosure:
			curIdx := operands[0]
			numFree := operands[1]
			newIdx, ok := indexMap[curIdx]
			if !ok {
				newIdx = len(stripped)
				stripped = append(stripped, b.Constants[curIdx])
				indexMap[curIdx] = newIdx
			}
			copy(insts[i:], MakeInstruction(opcode, newIdx, numFree))
		}
		i += 1 + offset
	}
	return stripped
}

func updateConstIndexes(insts []byte, indexMap map[int]int) {
	for i := 0; i < len(insts); {
		opcode := insts[i]
		operands, offset := parser.ReadOperands(parser.OpcodeOperands[opcode], insts[i+1:])
		switch opcode {
		case parser.OpConstant:
			curIdx := operands[0]
			newIdx, ok := indexMap[curIdx]
			if !ok {
				panic(fmt.Errorf("constant index not found: %d", curIdx))
			}
			copy(insts[i:], MakeInstruction(opcode, newIdx))
		case parser.OpClosure:
			curIdx := operands[0]
			numFree := operands[1]
			newIdx, ok := indexMap[curIdx]
			if !ok {
				panic(fmt.Errorf("constant index not found: %d", curIdx))
			}
			copy(insts[i:], MakeInstruction(opcode, newIdx, numFree))
		}
		i += 1 + offset
	}
}
