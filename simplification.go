package main

import (
	"github.com/terawatthour/logix/ast"
)

func checkForRightOrLeft(fn func(expression ast.Expression) (bool, ast.Expression), base *ast.InfixExpression) (bool, ast.Expression) {
	if matched, expr := fn(base.Left); matched {
		return true, expr
	}
	if matched, expr := fn(base.Right); matched {
		return true, expr
	}
	return false, base

}

// IdempotenceRule
// a + a = a
// a * a = a
func IdempotenceRule(expression ast.Expression) (bool, ast.Expression) {
	if expression.Type() != "infix" {
		return false, expression
	}
	expr := expression.(*ast.InfixExpression)
	if (expr.Action != "or" && expr.Action != "and") || expr.Left.Literal() != expr.Right.Literal() {
		return false, expression
	}

	return true, &ast.Identifier{Value: expr.Left.Literal()}
}

// NegatedAlternativeRule
// a + !a = 1
func NegatedAlternativeRule(expression ast.Expression) (bool, ast.Expression) {
	if expression.Type() != "infix" {
		return false, expression
	}
	expr := expression.(*ast.InfixExpression)
	if expr.Action != "or" {
		return false, expression
	}

	if expr.Left.Type() == "identifier" && expr.Right.Type() == "prefix" {
		prefix := expr.Right.(*ast.PrefixExpression)
		if prefix.Op == "!" && prefix.Right.Type() == "identifier" {
			if expr.Left.Literal() == prefix.Right.Literal() {
				return true, &ast.Boolean{Value: true}
			}
		}
	} else if expr.Right.Type() == "identifier" && expr.Left.Type() == "prefix" {
		prefix := expr.Left.(*ast.PrefixExpression)
		if prefix.Op == "!" && prefix.Right.Type() == "identifier" {
			if expr.Right.Literal() == prefix.Right.Literal() {
				return true, &ast.Boolean{Value: true}
			}
		}
	}

	return false, expression
}

// DuplicateAlternativeRule
// a + a = a
// a + b + a = a + b
func DuplicateAlternativeRule(expression ast.Expression) (bool, ast.Expression) {
	if expression.Type() != "infix" {
		return false, expression
	}
	expr := expression.(*ast.InfixExpression)
	if expr.Action != "or" {
		return false, expression
	}

	// implement me

	return false, expression
}

// NegatedConjunctionRule
// a * !a = 0
func NegatedConjunctionRule(expression ast.Expression) (bool, ast.Expression) {
	if expression.Type() != "infix" {
		return false, expression
	}
	expr := expression.(*ast.InfixExpression)
	if expr.Action != "and" {
		return false, expression
	}
	if expr.Left.Type() == "identifier" && expr.Right.Type() == "prefix" {
		prefix := expr.Right.(*ast.PrefixExpression)
		if prefix.Op == "!" && prefix.Right.Type() == "identifier" {
			if expr.Left.Literal() == prefix.Right.Literal() {
				return true, &ast.Boolean{Value: false}
			}
		}
	} else if expr.Right.Type() == "identifier" && expr.Left.Type() == "prefix" {
		prefix := expr.Left.(*ast.PrefixExpression)
		if prefix.Op == "!" && prefix.Right.Type() == "identifier" {
			if expr.Right.Literal() == prefix.Right.Literal() {
				return true, &ast.Boolean{Value: false}
			}
		}
	}

	return false, expression
}

// IdentityRule
// a + 0 = a
// a * 1 = a
// a * 0 = 0
func IdentityRule(expression ast.Expression) (bool, ast.Expression) {
	if expression.Type() != "infix" {
		return false, expression
	}
	expr := expression.(*ast.InfixExpression)
	if expr.Action == "and" {
		if matched, expr := checkForRightOrLeft(func(expression ast.Expression) (bool, ast.Expression) {
			if expression.Type() == "boolean" && expression.Literal() == "0" {
				return true, &ast.Boolean{Value: false}
			}
			return false, expression
		}, expr); matched {
			return true, expr
		}

		if expr.Left.Literal() == "1" {
			return true, expr.Right
		}

		if expr.Right.Literal() == "1" {
			return true, expr.Left
		}

	} else if expr.Action == "or" {
		if matched, expr := checkForRightOrLeft(func(expression ast.Expression) (bool, ast.Expression) {
			if expression.Type() == "boolean" && expression.Literal() == "1" {
				return true, &ast.Boolean{Value: true}
			}
			return false, expression
		}, expr); matched {
			return true, expr
		}

		if expr.Left.Literal() == "0" {
			return true, expr.Right
		}

		if expr.Right.Literal() == "0" {
			return true, expr.Left
		}
	}

	return false, expression
}

func ImplicationRule(expression ast.Expression) (bool, ast.Expression) {
	if expression.Type() != "infix" {
		return false, expression
	}
	expr := expression.(*ast.InfixExpression)
	if expr.Action != "->" {
		return false, expression
	}

	return true, &ast.InfixExpression{
		Op:     "+",
		Action: "or",
		Left:   &ast.PrefixExpression{Op: "!", Right: expr.Left},
		Right:  expr.Right,
	}
}

func DeMorganRule(expression ast.Expression) (bool, ast.Expression) {
	if expression.Type() != "prefix" {
		return false, expression
	}
	expr := expression.(*ast.PrefixExpression)
	if expr.Op != "!" {
		return false, expression
	}
	if expr.Right.Type() != "infix" {
		return false, expression
	}
	infix := expr.Right.(*ast.InfixExpression)
	if infix.Action != "and" && infix.Action != "or" {
		return false, expression
	}

	var op string
	if infix.Action == "and" {
		op = "+"
	} else {
		op = "*"
	}

	return true, &ast.InfixExpression{
		Op:     op,
		Action: infix.Action,
		Left:   &ast.PrefixExpression{Op: "!", Right: infix.Left},
		Right:  &ast.PrefixExpression{Op: "!", Right: infix.Right},
	}
}

func DoubleNegationRule(expression ast.Expression) (bool, ast.Expression) {
	if expression.Type() != "prefix" {
		return false, expression
	}
	expr := expression.(*ast.PrefixExpression)
	if expr.Op != "!" {
		return false, expression
	}
	if expr.Right.Type() != "prefix" {
		return false, expression
	}
	prefix := expr.Right.(*ast.PrefixExpression)
	if prefix.Op != "!" {
		return false, expression
	}

	return true, prefix.Right
}
