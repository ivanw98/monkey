// Package ast represents the data structure used for the internal representation of the source code.
package ast

import (
	"monkey/token"
)

type Node interface {
	TokenLiteral() string
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

// TokenLiteral returns the literal representation of the first token within the program's statements.
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (ls *LetStatement) statementNode() {}

// TokenLiteral returns the Literal from the LetStatement being called on.
func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

func (i *Identifier) expressionNode() {}

// TokenLiteral returns the Literal from the Identifier being called on.
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}
