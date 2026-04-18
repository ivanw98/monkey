package code_test

import (
	"monkey/code"
	"testing"
)

func TestMake(t *testing.T) {
	tests := []struct {
		name     string
		op       code.Opcode
		operands []int
		expected []byte
	}{
		{
			name:     "OpConstant",
			op:       code.OpConstant,
			operands: []int{65534},
			expected: []byte{byte(code.OpConstant), 255, 254},
		},
		{
			"OpAdd",
			code.OpAdd,
			[]int{},
			[]byte{byte(code.OpAdd)},
		},
		{
			name:     "OpGetLocal",
			op:       code.OpGetLocal,
			operands: []int{255},
			expected: []byte{byte(code.OpGetLocal), 255},
		},
		{
			name:     "OpConstant",
			op:       code.OpClosure,
			operands: []int{65534, 255},
			// OpClosure takes two operands with widths {2, 1} — a 2-byte operand followed by a 1-byte operand, encoded big-endian.
			// - Operand 1 = 65534 → 2 bytes, big-endian:
			// - 65534 in hex is 0xFFFE
			// - High byte first: 0xFF = 255, then 0xFE = 254
			// - Operand 2 = 255 → 1 byte:
			// - 0xFF = 255
			expected: []byte{byte(code.OpClosure), 255, 254, 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instruction := code.Make(tt.op, tt.operands...)

			if len(instruction) != len(tt.expected) {
				t.Errorf("instruction wrong length.got=%d.want=%d", len(instruction), len(tt.expected))
			}

			for i, b := range tt.expected {
				if instruction[i] != tt.expected[i] {
					t.Errorf("wrong byte at position %d.got=%d.want=%d", i, instruction[i], b)
				}
			}
		})

	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		name      string
		op        code.Opcode
		operands  []int
		bytesRead int
	}{
		{
			name:      "Read OpConstant",
			op:        code.OpConstant,
			operands:  []int{65535},
			bytesRead: 2,
		},
		{
			name:      "Read OpGetLocal",
			op:        code.OpGetLocal,
			operands:  []int{255},
			bytesRead: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instruction := code.Make(tt.op, tt.operands...)
			def, err := code.Lookup(byte(tt.op))
			if err != nil {
				t.Fatalf("definition not found: %q\n", err)
			}

			operandsRead, n := code.ReadOperands(def, instruction[1:])
			if n != tt.bytesRead {
				t.Fatalf("n wrong. want=%d, got=%d", tt.bytesRead, n)
			}

			for i, want := range tt.operands {
				if operandsRead[i] != want {
					t.Errorf("operand wrong. want=%d, got=%d", want, operandsRead[i])
				}
			}
		})
	}

}

func TestInstructionsString(t *testing.T) {
	instructions := []code.Instructions{
		code.Make(code.OpAdd),
		code.Make(code.OpGetLocal, 1),
		code.Make(code.OpConstant, 2),
		code.Make(code.OpConstant, 65535),
	}

	// represents offset, operand, bytes used
	/**
	So the sequence is:
	OpAdd
	starts at offset 0
	size 1 (OpAdd is one byte)
	next instruction starts at 1
	OpGetLocal 1
	starts at offset 1
	size 2 (OpGetLocal is one byte + has an operand of size 1 byte)
	next instruction starts at 3
	OpConstant 2
	starts at offset 3
	size 3 (OpConstant is one byte + has an operand of size 2 bytes)
	next instruction starts at 6
	OpConstant 65535
	starts at offset 6
	size 3
	*/

	expected := `0000 OpAdd
0001 OpGetLocal 1
0003 OpConstant 2
0006 OpConstant 65535
`
	concatted := code.Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}

	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted.\nwant=%q\ngot=%q", expected, concatted.String())
	}
}
