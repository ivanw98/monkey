// Package ast represents the data structure used for the internal representation of the source code.
package ast

import (
	"bytes"
	"monkey/token"
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

// Identifier represents the identifier of the binding.
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

// ReturnStatement contains a keyword for return and an expression.
type ReturnStatement struct {
	Token       token.Token // the 'return' expression
	ReturnValue Expression
}

// ExpressionStatement is a wrapper and is statement consisting solely of one expression.
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

// IntegerLiteral are expressions. The Value they produce is the integer itself.
type IntegerLiteral struct {
	Token token.Token
	Value int64
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

// TokenLiteral returns the Literal from the ReturnStatement being called on.
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

func (es *ExpressionStatement) statementNode() {}

// String allows for printing of AST nodes.
func (i *IntegerLiteral) String() string {
	return i.Token.Literal
}

// TokenLiteral returns the Literal from the ReturnStatement being called on.
func (i *IntegerLiteral) TokenLiteral() string { return i.Token.Literal }

func (i *IntegerLiteral) expressionNode() {}
