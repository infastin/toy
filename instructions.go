package toy

import (
	"fmt"
	"strings"

	"github.com/infastin/toy/parser"
)

// MakeInstruction returns a bytecode for an opcode and the operands.
func MakeInstruction(opcode parser.Opcode, operands ...int) []byte {
	numOperands := parser.OpcodeOperands[opcode]

	totalLen := 1
	for _, w := range numOperands {
		totalLen += w
	}

	instruction := make([]byte, totalLen)
	instruction[0] = opcode

	offset := 1
	for i, o := range operands {
		width := numOperands[i]
		switch width {
		case 1:
			instruction[offset] = byte(o)
		case 2:
			n := uint16(o)
			instruction[offset] = byte(n >> 8)
			instruction[offset+1] = byte(n)
		case 4:
			n := uint32(o)
			instruction[offset] = byte(n >> 24)
			instruction[offset+1] = byte(n >> 16)
			instruction[offset+2] = byte(n >> 8)
			instruction[offset+3] = byte(n)
		}
		offset += width
	}

	return instruction
}

// FormatInstructions returns string representation of bytecode instructions.
func FormatInstructions(b []byte, posOffset int) []string {
	var inst strings.Builder
	var out []string
	i := 0
	for i < len(b) {
		operands, read := parser.ReadOperands(parser.OpcodeOperands[b[i]], b[i+1:])
		inst.WriteString(fmt.Sprintf("%04d %s", posOffset+i, parser.OpcodeNames[b[i]]))
		if len(operands) != 0 {
			inst.WriteString(" [")
			for i, operand := range operands {
				if i != 0 {
					inst.WriteString(", ")
				}
				inst.WriteString(fmt.Sprintf("%d", operand))
			}
			inst.WriteByte(']')
		}
		out = append(out, inst.String())
		inst.Reset()
		i += 1 + read
	}
	return out
}
