package bytecode

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// MakeInstruction returns a bytecode for an opcode and the operands.
func MakeInstruction(opcode Opcode, operands ...int) []byte {
	numOperands := OpcodeOperands[opcode]

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
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 4:
			binary.BigEndian.PutUint32(instruction[offset:], uint32(o))
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
		operands, read := ReadOperands(OpcodeOperands[b[i]], b[i+1:])
		inst.WriteString(fmt.Sprintf("%04d %s", posOffset+i, OpcodeNames[b[i]]))
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
