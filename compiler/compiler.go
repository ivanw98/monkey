// Package compiler defines the code that allows our compiler to:
// walk an AST recursively to find Literals
// evaluate the literals and turn them into objects
// Add those to the constants field
// add OpConstant instructions to its internal instructions slice.
package compiler

import (
	"fmt"
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

// Compile recursively traverses an AST node, generates bytecode instructions, and appends constants to the compiler's state.
func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)

	case *ast.InfixExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)

		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	}

	return nil
}

// Bytecode returns the compiled output containing bytecode instructions and constants used during interpretation.
func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)

	// return the index of the object.Object appended to the constants.
	// This identifier will be used as the operand for the OpConstant instruction
	// that should cause the VM to load this constant from the constants pool on to the stack.
	return len(c.constants) - 1
}

// emit generates an instruction and adds it to the results.
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	// return the starting point of the just-emitted instruction
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}
