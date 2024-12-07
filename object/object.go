package object

import (
	"bytes"
	"fmt"
	"monkey/ast"
	"strings"
)

// ObjectType defines the type for various objects in the system.
type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	STRING_OBJ       = "STRING"
	BUILTIN_OBJ      = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
)

// BuiltinFunction represents a function type that accepts a variable number of Object arguments and returns an Object.
type BuiltinFunction func(args ...Object) Object

type Object interface {
	Type() ObjectType
	Inspect() string
}

// Integer represents an integer object with a 64-bit Value field.
type Integer struct {
	Value int64
}

// Boolean represents a boolean value with true or false states.
type Boolean struct {
	Value bool
}

// String represents a string value for Monkey.
type String struct {
	Value string
}

// ReturnValue represents the return value object in the system, wrapping an Object that contains the actual returned value.
type ReturnValue struct {
	Value Object
}

// Function represents a function as an object.
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

// Builtin represents a structure that holds a BuiltinFunction which defines the behavior of the built-in functionality.
type Builtin struct {
	Fn BuiltinFunction
}

// Array represents a collection of elements implementing the Object interface.
type Array struct {
	Elements []Object
}

// Error represents an error object in the system with a message.
type Error struct {
	Message string
}

// Null represents the absence of a value.
type Null struct {
}

// Inspect returns a string representation of the Integer's value.
func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

// Type returns the object type.
func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ
}

// Inspect returns the string representation of the Boolean value.
func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%t", b.Value)
}

// Type returns the object type.
func (b *Boolean) Type() ObjectType {
	return BOOLEAN_OBJ
}

// Inspect returns the string representation of the String value.
func (s *String) Inspect() string {
	return s.Value
}

// Type returns the object type.
func (s *String) Type() ObjectType {
	return STRING_OBJ
}

// Inspect returns the string representation of the returned value by invoking the Inspect method on the wrapped Object.
func (rv *ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}

// Type returns the object type.
func (rv *ReturnValue) Type() ObjectType {
	return RETURN_VALUE_OBJ
}

// Inspect returns the error message formatted with a prefix "ERROR: ".
func (e *Error) Inspect() string {
	return "ERROR: " + e.Message
}

// Type returns the object type.
func (e *Error) Type() ObjectType {
	return ERROR_OBJ
}

// Inspect returns a string representation of a Null object.
func (n *Null) Inspect() string {
	return "null"
}

// Type returns the object type.
func (n *Null) Type() ObjectType {
	return NULL_OBJ
}

// Inspect returns a string representation of a Function object.
func (f *Function) Inspect() string {
	var out bytes.Buffer
	var params []string

	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

// Type returns the object type.
func (f *Function) Type() ObjectType {
	return FUNCTION_OBJ
}

// Inspect returns a string representation of a Builtin object.
func (b *Builtin) Inspect() string {
	return "builtin function"
}

// Type returns the BUILTIN_OBJ
func (b *Builtin) Type() ObjectType {
	return BUILTIN_OBJ
}

// Inspect returns a string representation of a Array object.
func (a *Array) Inspect() string {
	var out bytes.Buffer

	var elements []string

	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// Type returns the BUILTIN_OBJ
func (a *Array) Type() ObjectType {
	return ARRAY_OBJ
}
