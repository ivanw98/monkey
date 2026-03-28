package vm

import (
	"monkey/code"
	"monkey/object"
)

// Frame holds execution relevant information - a fn object and an instruction pointer.
type Frame struct {
	fn *object.CompiledFunction // The compiled function referenced by the frame.
	ip int                      // the instruction pointer in this frame for this function.
}

func NewFrame(fn *object.CompiledFunction) *Frame {
	return &Frame{
		fn: fn,
		ip: -1,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
