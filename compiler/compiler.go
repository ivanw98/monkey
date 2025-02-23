package compiler

import (
	"monkey/ast"
	"monkey/code"
	"monkey/object"
)

// Compiler evaluates AST nodes and turns them into Objects via code.OpConstant instructions
type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

// Bytecode represents the compiled output, containing instructions and a set of constants used during execution.
type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

// New  returns a new instance of Compiler with initialized instructions and constants.
func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
	}
}

// Compile traverses an AST node, generates bytecode instructions, and appends constants to the compiler's state.
func (c *Compiler) Compile(node ast.Node) error {
	return nil
}

// Bytecode returns the compiled output containing bytecode instructions and constants used during interpretation.
func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}
