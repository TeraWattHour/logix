package ast

import (
	"fmt"
	"github.com/terawatthour/logix/tokenizer"
)

type Node interface {
	Literal() string
}

type Statement interface {
	Node
}

type Expression interface {
	Node
	Type() string
}

type Identifier struct {
	Token *tokenizer.Token
	Value string
}

func (s *Identifier) Type() string {
	return "identifier"
}

func (s *Identifier) Literal() string {
	return s.Value
}

type Boolean struct {
	Token *tokenizer.Token
	Value bool
}

func (s *Boolean) Type() string {
	return "boolean"
}

func (s *Boolean) Literal() string {
	if s.Value {
		return "1"
	} else {
		return "0"
	}
}

type SimplifyStatement struct {
	Token      *tokenizer.Token
	Expression Expression
}

func (s *SimplifyStatement) Literal() string {
	return fmt.Sprintf("simplify %s", s.Expression.Literal())
}

type TableStatement struct {
	Token      *tokenizer.Token
	Expression Expression
}

func (s *TableStatement) Literal() string {
	return fmt.Sprintf("table %s", s.Expression.Literal())
}

type PrefixExpression struct {
	Token  *tokenizer.Token
	Op     string
	Action string

	Right Expression
}

func (s *PrefixExpression) Type() string {
	return "prefix"
}

func (s *PrefixExpression) Literal() string {
	return fmt.Sprintf("%s%s", s.Op, s.Right.Literal())
}

type InfixExpression struct {
	Token  *tokenizer.Token
	Op     string
	Action string

	Left  Expression
	Right Expression
}

func (s *InfixExpression) Type() string {
	return "infix"
}

func (s *InfixExpression) Literal() string {
	return fmt.Sprintf("(%s %s %s)", s.Left.Literal(), s.Op, s.Right.Literal())
}
