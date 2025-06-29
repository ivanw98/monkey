package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

// StackSize defines the default size allocated for a stack in bytes.
const StackSize = 2048

// True is an instance of true for the vm. Global variable that is immutable and unique.
var True = &object.Boolean{Value: true}

// False is an instance of false for the vm. Global variable that is immutable and unique.
var False = &object.Boolean{Value: false}

// VM represents a virtual machine for executing bytecode instructions, managing constants, and handling a stack.
type VM struct {
	constants    []object.Object
	instructions code.Instructions
	stack        []object.Object
	sp           int // Always points to the next value. Top of stack is stack[sp-1]
}

// New initializes a new instance of the VM.
func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

// Run executes the bytecode instructions stored in the VM and manages the stack using provided constants and opcodes.
func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])
		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}

		case code.OpPop:
			vm.pop()

		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// LastPoppedStackElem returns the last popped element from the stack, without modifying the stack pointer.
func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

// push adds the given object to the stack and increments the stack pointer.
func (vm *VM) push(o object.Object) error {
	// return an error if the stack exceeds the predefined StackSize.
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

// pop takes the first element from the top of the stack and decrements the stack pointer.
func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	var result int64
	switch op {
	case code.OpAdd:
		result = leftVal + rightVal

	case code.OpSub:
		result = leftVal - rightVal

	case code.OpMul:
		result = leftVal * rightVal

	case code.OpDiv:
		result = leftVal / rightVal

	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})

}
