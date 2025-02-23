package code_test

import (
	"monkey/code"
	"testing"
)

func TestMake(t *testing.T) {
	tests := []struct {
		op       code.Opcode
		operands []int
		expected []byte
	}{
		{
			op:       code.OpConstant,
			operands: []int{65534},
			expected: []byte{byte(code.OpConstant), 255, 254},
		},
	}

	for _, tt := range tests {
		instruction := code.Make(tt.op, tt.operands...)

		if len(instruction) != len(tt.expected) {
			t.Errorf("instruction wrong length.got=%d.want=%d", len(instruction), len(tt.expected))
		}

		for i, b := range tt.expected {
			if instruction[i] != tt.expected[i] {
				t.Errorf("wrong byte at position %d.got=%d.want=%d", i, instruction[i], b)
			}
		}
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        code.Opcode
		operands  []int
		bytesRead int
	}{
		{
			code.OpConstant, []int{65535}, 2,
		},
	}

	for _, tt := range tests {
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
	}

}
