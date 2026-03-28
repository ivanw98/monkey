package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const (
	// StackSize defines the default size allocated for a stack in bytes.
	StackSize = 2048

	// GlobalSize is the number of global bindings for `let` in the virtual machine.
	GlobalSize = 65536

	// MaxFrames is the maximum number of call frames the VM can hold on the frames stack, limiting the depth of nested function calls.
	MaxFrames = 1024
)

var (
	// True is an instance of true for the vm. Global variable that is immutable and unique.
	True = &object.Boolean{Value: true}

	// False is an instance of false for the vm. Global variable that is immutable and unique.
	False = &object.Boolean{Value: false}

	// Null is an instance of null for the vm. Global variable that is immutable and unique.
	Null = &object.Null{}
)

// VM represents a virtual machine for executing bytecode instructions, managing constants, and handling a stack.
type VM struct {
	constants   []object.Object
	stack       []object.Object // holds temporary values during expression evaluation
	sp          int             // Always points to the next value. Top of stack is stack[sp-1]
	globals     []object.Object // the VM's storage for all `let` bindings
	frames      []*Frame
	framesIndex int
}

// New initializes a new instance of the VM.
func New(bytecode *compiler.Bytecode) *VM {
	// pre-allocate frames slice
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame
	return &VM{
		constants:   bytecode.Constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     make([]object.Object, GlobalSize),
		frames:      frames,
		framesIndex: 1, // if we allocate a frame, we have to increase our index for the stack implementation
	}
}

// NewWithGlobalStore a new instance of the VM that tracks global state for the REPL.
func NewWithGlobalStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

// Run executes the bytecode instructions stored in the VM and manages the stack using provided constants and opcodes.
func (vm *VM) Run() error {
	var ip int                // current instruction pointer position within the active frame
	var ins code.Instructions // the instruction bytes of the active frame
	var op code.Opcode        // the opcode decoded from the current instruction

	// Continue executing as long as the instruction pointer hasn't reached the end of the current frame's instructions.
	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		// Advance the instruction pointer to the next instruction before decoding.
		vm.currentFrame().ip++

		// Cache the instruction pointer and instructions slice for this cycle to avoid repeated method calls.
		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()

		// Decode the opcode at the current instruction pointer position.
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
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

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}

		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}

		case code.OpJump:
			// decode the operand located right after the opcode.
			pos := int(code.ReadUint16(ins[ip+1:]))
			// set the instruction pointer, `ip`, to the target of our jump
			// we need to set `ip` to the offset right before the one we want.
			// the loop will increment it to the value we want on the next cycle.
			// write back update to ip
			vm.currentFrame().ip = pos - 1

		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			// skip over the two bytes of the operand in the next cycle
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				// if not truthy, we jump - similar to case above
				vm.currentFrame().ip = pos - 1
			}

		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}

		case code.OpArray:
			numOfElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			array := vm.buildArray(vm.sp-numOfElements, vm.sp)
			vm.sp = vm.sp - numOfElements

			err := vm.push(array)
			if err != nil {
				return err
			}

		case code.OpHash:
			numOfElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			hash, err := vm.buildHash(vm.sp-numOfElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numOfElements

			err = vm.push(hash)
			if err != nil {
				return err
			}

		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}

		case code.OpCall:
			// get the compiled function off the stack and check type
			fn, ok := vm.peekStack().(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("calling non-function")
			}

			// Create a new frame for the compiledFn.
			frame := NewFrame(fn)
			// add the frame to the vm frame stack.
			vm.pushFrame(frame)

		case code.OpReturnValue:
			// Get fn return val from stack
			returnVal := vm.pop()
			// remove frame from stack
			vm.popFrame()

			// discard CompiledFunction object
			vm.pop()
			if err := vm.push(returnVal); err != nil {
				return err
			}

		case code.OpReturn:
			// no return val so popFrame immediately
			vm.popFrame()
			vm.pop() // discard fn

			if err := vm.push(Null); err != nil {
				return err
			}
		}

	}

	return nil
}

func (vm *VM) peekStack() object.Object {
	return vm.stack[vm.sp-1]
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

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeIntegerBinaryOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeStringBinaryOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operator: %s %s", leftType, rightType)
	}
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())

	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightVal == leftVal))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightVal != leftVal))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftVal > rightVal))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()
	switch operand {
	case False:
		return vm.push(True)
	case True:
		return vm.push(False)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()
	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	val := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -val})
}

func (vm *VM) executeIntegerBinaryOperation(op code.Opcode, left object.Object, right object.Object) error {
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

func (vm *VM) executeStringBinaryOperation(op code.Opcode, left object.Object, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	return vm.push(&object.String{Value: leftValue + rightValue})
}

// buildArray iterates through the elements in the specified section of the stack, adding each to an *object.Array.
// This array is then pushed on to the stack after the elements have been taken off.
func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &object.Array{Elements: elements}
}

// buildHash iterates through elements between startIndex and endIndex in pairs creating a object.HashPair out of them.
// It generates the HashKey and adds to hashedPairs, then builds the *object.Hash with them.
func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]
		pair := object.HashPair{
			Key:   key,
			Value: value,
		}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{
		Pairs: hashedPairs,
	}, nil
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)

	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)

	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	maxBounds := int64(len(arrayObject.Elements) - 1)

	// bounds check
	if i < 0 || i > maxBounds {
		return vm.push(Null)
	}

	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)

	// Check whether the index provided can be used as a hash key
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

// currentFrame returns the last value of the stack (e.g. peek)
func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {

	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}
