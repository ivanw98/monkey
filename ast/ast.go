// Package ast represents the data structure used for the internal representation of the source code.
package ast

import (
	"bytes"
	"monkey/token"
	"strings"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of every AST our parser produces.
type Program struct {
	Statements []Statement
}

// LetStatement represents a node for variable binding.
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Identifier
	Value Expression
}

// Identifier represents the identifier of the binding, it is a type of Expression.
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

// ReturnStatement contains a keyword for return and an Expression.
type ReturnStatement struct {
	Token       token.Token // the 'return' expression
	ReturnValue Expression
}

// ExpressionStatement is a wrapper and is statement consisting solely of one expression.
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

// PrefixExpression represents a prefix operation in an AST.
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

// InfixExpression represents a prefix operation in an AST.
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Right    Expression
	Operator string
}

// IntegerLiteral are expressions. The Value they produce is the integer itself.
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

// Boolean represents boolean literals.
type Boolean struct {
	Token token.Token
	Value bool
}

// IfExpression represents an if-else conditional expression in the abstract syntax tree.
type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

// BlockStatement represents a block of statements enclosed within braces.
type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

// StringLiteral represents a string literal in the code. Value is the string itself.
type StringLiteral struct {
	Token token.Token
	Value string
}

// FunctionLiteral represents a function definition with parameters and a body.
type FunctionLiteral struct {
	Token      token.Token // the 'fn' token
	Parameters []*Identifier
	Body       *BlockStatement
}

// CallExpression consists of an expression that results in a function when evaluated and a list of expressions that are the args to the function call.
type CallExpression struct {
	Token    token.Token // The '(' token
	Function Expression  // Identifier or FunctionLiteral
	Args     []Expression
}

// ArrayLiteral represents an array literal in the syntax tree. The list of expressions is for each element in the array.
type ArrayLiteral struct {
	Token    token.Token // token representing the '['
	Elements []Expression
}

// String creates a buffer and writes the return value of each statement's String() method to it.
func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// TokenLiteral returns the literal representation of the first token within the program's statements.
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// String allows for printing of AST nodes.
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

// TokenLiteral returns the Literal from the LetStatement being called on.
func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

func (ls *LetStatement) statementNode() {}

// String returns the Value of the Identifier.
func (i *Identifier) String() string {
	return i.Value
}

// TokenLiteral returns the Literal from the Identifier being called on.
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) expressionNode() {}

// String allows for printing of AST nodes.
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

// TokenLiteral returns the Literal from the ReturnStatement being called on.
func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *ReturnStatement) statementNode() {}

// String allows for printing of AST nodes.
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}

	return ""
}

// TokenLiteral returns the Literal from the ExpressionStatement being called on.
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

func (es *ExpressionStatement) statementNode() {}

// String allows for printing of AST nodes.
func (i *IntegerLiteral) String() string {
	return i.Token.Literal
}

// TokenLiteral returns the Literal from the IntegerLiteral being called on.
func (i *IntegerLiteral) TokenLiteral() string { return i.Token.Literal }

func (i *IntegerLiteral) expressionNode() {}

// String allows for printing of AST nodes.
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

// TokenLiteral returns the Literal from the PrefixExpression being called on.
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

func (pe *PrefixExpression) expressionNode() {}

// String allows for printing of AST nodes.
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

// TokenLiteral returns the Literal from the InfixExpression being called on.
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

func (ie *InfixExpression) expressionNode() {}

// String allows for printing of AST nodes.
func (b *Boolean) String() string {
	return b.Token.Literal
}

// TokenLiteral returns the Literal from the Boolean being called on.
func (b *Boolean) TokenLiteral() string {
	return b.Token.Literal
}

func (b *Boolean) expressionNode() {}

// String allows for printing of AST nodes.
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Condition.String())

	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

// TokenLiteral returns the Literal from the IfExpression being called on.
func (ie *IfExpression) TokenLiteral() string {
	return ie.Token.Literal
}

func (ie *IfExpression) expressionNode() {

}

// String allows for printing of AST nodes.
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// TokenLiteral returns the Literal from the BlockStatement being called on.
func (bs *BlockStatement) TokenLiteral() string {
	return bs.Token.Literal
}

func (bs *BlockStatement) statementNode() {

}

// String allows for printing of AST nodes.
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	var params []string
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	out.WriteString(fl.Body.String())

	return out.String()
}

// TokenLiteral returns the Literal from the FunctionLiteral being called on.
func (fl *FunctionLiteral) TokenLiteral() string {
	return fl.Token.Literal
}

func (fl *FunctionLiteral) expressionNode() {}

// String allows for printing of AST nodes.
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	var args []string
	for _, a := range ce.Args {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

// TokenLiteral returns the Literal from the CallExpression being called on.
func (ce *CallExpression) TokenLiteral() string {
	return ce.Token.Literal
}

func (ce *CallExpression) expressionNode() {

}

// String allows for printing of AST nodes.
func (sl *StringLiteral) String() string {
	return sl.Token.Literal
}

// TokenLiteral returns the Literal from the StringLiteral being called on.
func (sl *StringLiteral) TokenLiteral() string {
	return sl.Token.Literal
}

func (sl *StringLiteral) expressionNode() {

}

// String allows for printing of AST nodes.
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	var elements []string

	for _, element := range al.Elements {
		elements = append(elements, element.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// TokenLiteral returns the Literal from the StringLiteral being called on.
func (al *ArrayLiteral) TokenLiteral() string {
	return al.Token.Literal
}

func (al *ArrayLiteral) expressionNode() {

}
