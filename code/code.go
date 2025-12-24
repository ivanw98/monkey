package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Instructions represent a sequence of bytes intended for processing or interpretation.
//
// A single Instruction consists of an Opcode and an optional number of operands.
type Instructions []byte

// Opcode represents a single instruction code in a virtual machine or processing unit.
type Opcode byte

// Definition helps make Opcode readable and stores the number of bytes each operand takes up.
type Definition struct {
	Name          string
	OperandWidths []int
}

const (
	// OpConstant retrieves the constant using the operand as an index and pushes it on to the stack.
	OpConstant Opcode = iota

	// OpAdd instructs the VM to pop the two topmost elements off the stack, add them together and push the result back.
	// It doesn't have any operands, it is one byte (a single Opcode).
	OpAdd

	//OpPop instructs the VM to remove the topmost element off the stack. It will be emitted after every expression stmt.
	OpPop

	// OpSub instructs the VM to pop the two topmost elements off the stack, subtract them and push the result back.
	OpSub

	// OpMul instructs the VM to pop the two topmost elements off the stack, multiply them together and push the result back.
	OpMul

	// OpDiv instructs the VM to pop the two topmost elements off the stack, divide them and push the result back.
	OpDiv

	// OpTrue instructs the VM to load boolean `true` onto the stack.
	OpTrue

	// OpFalse instructs the VM to load boolean `false` onto the stack.
	OpFalse

	// OpEqual instructs the VM to use an equal to comparison operator.
	OpEqual

	// OpNotEqual instructs the VM to use a not equal to comparison operator.
	OpNotEqual

	// OpGreaterThan instructs the VM to use a greater than comparison
	// Note that we do not need a less than operator as 3 < 5 can be re-ordered to 5 > 3.ß
	OpGreaterThan

	// OpMinus instructs the VM to negate an integer.
	OpMinus

	// OpBang instructs the VM to negate a boolean.
	OpBang

	// OpJumpNotTruthy instructs the VM to skip a set of instructions if a condition fails to eval as truthy.
	OpJumpNotTruthy

	// OpJump instructs the VM to skip a set of instructions regardless.
	OpJump

	// OpNull instructs the VM to insert an instance of vm.Null on the stack.
	OpNull

	// OpGetGlobal instructs the VM to retrieve a global variable (using operand as index) and push its value onto the stack.
	OpGetGlobal

	// OpSetGlobal instructs the VM to pop a value from the stack and store it in a global variable (using operand as index).
	OpSetGlobal

	// OpArray defines an opcode to instruct the VM to leave N values on the stack.
	OpArray

	// OpHash is an opcode that specifies to the VM the number of keys and values sitting on the stack
	OpHash

	// OpIndex tells the VM to take the two values sitting on top of the stack to perform an index operation.
	OpIndex
)

var definitions = map[Opcode]*Definition{
	OpConstant:      {"OpConstant", []int{2}},
	OpAdd:           {"OpAdd", []int{}},
	OpPop:           {"OpPop", []int{}},
	OpSub:           {"OpSub", []int{}},
	OpMul:           {"OpMul", []int{}},
	OpDiv:           {"OpDiv", []int{}},
	OpTrue:          {"OpTrue", []int{}},
	OpFalse:         {"OpFalse", []int{}},
	OpEqual:         {"OpEqual", []int{}},
	OpNotEqual:      {"OpNotEqual", []int{}},
	OpGreaterThan:   {"OpGreaterThan", []int{}},
	OpMinus:         {"OpMinus", []int{}},
	OpBang:          {"OpBang", []int{}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},
	OpJump:          {"OpJump", []int{2}},
	OpNull:          {"OpNull", []int{}},
	OpGetGlobal:     {"OpGetGlobal", []int{2}},
	OpSetGlobal:     {"OpSetGlobal", []int{2}},
	OpArray:         {"OpArray", []int{2}},
	OpHash:          {"OpHash", []int{2}},
	OpIndex:         {"OpIndex", []int{}},
}

// String outputs a readable format of Instructions.
func (ins Instructions) String() string {
	var out bytes.Buffer

	for i := 0; i < len(ins); {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			i++
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

// Lookup retrieves the Definition associated with the given opcode.
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

// Make encodes the operands of a single bytecode instruction.
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	// find out how long the resulting instruction needs to be.
	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	// iterate over operand widths, taking the matching element from operands and place in instruction.
	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			// Endianness is the order in which bytes within a word are addressed in computer memory,
			// counting only byte significance compared to earliness.
			// A big-endian system stores the most significant byte of a word at the smallest memory address
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		offset += width
	}

	return instruction
}

// ReadOperands decodes the operands of an instruction based on its definition and also tells us how many bytes were read.
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}

		offset += width
	}
	return operands, offset
}

// ReadUint16 reads two bytes from the provided Instructions and returns them as an uint16 in big-endian order.
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
