package vm

import (
	"monkey/code"
	"monkey/object"
)

// Frame holds execution relevant information - a fn object and an instruction pointer.
// It is temporary storage that lives as long as a function call.
type Frame struct {
	fn          *object.CompiledFunction // The compiled function referenced by the frame.
	ip          int                      // the instruction pointer in this frame for this function.
	basePointer int                      // points to the bottom of the stack of the current call frame
}

// NewFrame returns a new Frame for the given compiled function, with the instruction pointer initialised to -1
// so that the first increment in the execution loop moves it to the first instruction.
func NewFrame(fn *object.CompiledFunction, basePointer int) *Frame {
	return &Frame{
		fn:          fn,
		ip:          -1,
		basePointer: basePointer,
	}
}

// Instructions returns the bytecode instructions of the function associated with this frame.
func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
