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
	"sort"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

// Compiler evaluates AST nodes and turns them into Objects via code.OpConstant instructions
type Compiler struct {
	constants   []object.Object // represents the constants pool
	symbolTable *SymbolTable
	scopes      []CompilationScope
	scopeIndex  int
}

// Bytecode represents the compiled output, containing instructions and a set of constants used during execution.
type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

// CompilationScope represents an isolated compilation context for a single scope (e.g. the top-level program or a function body).
type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

// New  returns a new instance of Compiler with initialized instructions and constants.
func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	return &Compiler{
		constants:   []object.Object{},
		symbolTable: NewSymbolTable(),
		scopeIndex:  0,
		scopes:      []CompilationScope{mainScope},
	}
}

// NewWithState  returns a new instance of Compiler with symbol table and constants to keep global state for the REPL.
func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
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
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}

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
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)

		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))

	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}

	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit an `OpJumpNotTruthy` with a bogus value
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		// Emit an `OpJump` with a bogus value
		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err = c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.LetStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}

		symbol := c.symbolTable.Define(node.Name.Value)
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}

	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			// compile-time error
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		if symbol.Scope == GlobalScope {
			c.emit(code.OpGetGlobal, symbol.Index)
		} else {
			c.emit(code.OpGetLocal, symbol.Index)
		}

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))

	case *ast.ArrayLiteral:
		for _, element := range node.Elements {
			err := c.Compile(element)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))

	case *ast.HashLiteral:
		var keys []ast.Expression
		for k := range node.Pairs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}
			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}
		c.emit(code.OpHash, len(node.Pairs)*2)

	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)

	case *ast.FunctionLiteral:
		c.enterScope()
		for _, parameter := range node.Parameters {
			c.symbolTable.Define(parameter.Value)
		}
		if err := c.Compile(node.Body); err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}

		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		numLocals := c.symbolTable.numDefinitions
		// change where emitted instructions are stored when compiling a function.
		ins := c.leaveScope()
		compiledFn := &object.CompiledFunction{
			Instructions:  ins,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}
		c.emit(code.OpConstant, c.addConstant(compiledFn))

	case *ast.ReturnStatement:
		if err := c.Compile(node.ReturnValue); err != nil {
			return err
		}

		c.emit(code.OpReturnValue)

	case *ast.CallExpression:
		if err := c.Compile(node.Function); err != nil {
			return err
		}
		for _, arg := range node.Args {
			if err := c.Compile(arg); err != nil {
				return err
			}
		}
		c.emit(code.OpCall, len(node.Args))
	}

	return nil
}

// Bytecode returns the compiled output containing bytecode instructions and constants used during interpretation.
func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

// addConstant appends an object to the constants pool and returns its index.
// The index is used as the operand for OpConstant instructions, allowing the VM to load the value onto the stack at runtime.
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

	c.setLastInstruction(op, pos)

	// return the starting point of the just-emitted instruction
	return pos
}

// currentScope returns a pointer to the CompilationScope currently being compiled.
func (c *Compiler) currentScope() *CompilationScope {
	return &c.scopes[c.scopeIndex]
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)
	c.currentScope().instructions = updatedInstructions
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	prev := c.currentScope().lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.currentScope().previousInstruction = prev
	c.currentScope().lastInstruction = last
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.currentScope().lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.currentScope().lastInstruction
	prev := c.currentScope().previousInstruction
	old := c.currentInstructions()
	newInstructions := old[0:last.Position]

	c.currentScope().instructions = newInstructions
	// As you have removed a position, you need to set last instruction to what it used to be before that.
	c.currentScope().lastInstruction = prev
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.currentScope().instructions
}

// enterScope pushes a new scope onto the end of the scopes stack with append
func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	c.scopes = append(c.scopes, scope)
	c.scopeIndex++

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

// leaveScope pops the last scope off the end of the scopes stack
func (c *Compiler) leaveScope() code.Instructions {
	ins := c.currentInstructions()
	c.scopes = c.scopes[0 : len(c.scopes)-1]
	c.scopeIndex--

	c.symbolTable = c.symbolTable.Outer

	return ins
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.currentScope().lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))

	c.currentScope().lastInstruction.Opcode = code.OpReturnValue
}
