package main

import (
	"fmt"
	"github.com/terawatthour/logix/ast"
)

var DEBUG = false

func debugPrint(messages ...any) {
	if DEBUG {
		fmt.Println(messages...)
	}
}

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

	debugPrint("turning", expression.Literal(), "into", expr.Left.Literal(), "by idempotence")
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

	if matched, expr := negatedAlternativeRule(expr.Left, expr.Right); matched {
		debugPrint("turning", expression.Literal(), "into", expr.Literal())
		return true, expr
	}

	return false, expression
}

func negatedAlternativeRule(left ast.Expression, right ast.Expression) (bool, ast.Expression) {
	if left.Type() == "identifier" && right.Type() == "prefix" {
		if left.Literal() == right.Literal()[1:] {
			return true, &ast.Boolean{Value: true}
		}
	} else if left.Type() == "prefix" && right.Type() == "identifier" {
		if right.Literal() == left.Literal()[1:] {
			return true, &ast.Boolean{Value: true}
		}
	}

	if left.Type() == "infix" && (right.Type() == "identifier" || right.Type() == "prefix") {
		expr := left.(*ast.InfixExpression)
		if expr.Action == "or" {
			debugPrint("checking", expr.Left.Literal(), right.Literal())
			if matched, _ := negatedAlternativeRule(expr.Left, right); matched {
				return true, &ast.Boolean{Value: true}
			}
			if matched, _ := negatedAlternativeRule(expr.Right, right); matched {
				return true, &ast.Boolean{Value: true}
			}
		}
	} else if (left.Type() == "identifier" || left.Type() == "prefix") && right.Type() == "infix" {
		expr := right.(*ast.InfixExpression)
		if expr.Action == "or" {
			if matched, _ := negatedAlternativeRule(left, expr.Left); matched {
				return true, &ast.Boolean{Value: true}

			}
			if matched, _ := negatedAlternativeRule(expr.Right, right); matched {
				return true, &ast.Boolean{Value: true}
			}
		}
	} else if left.Type() == "infix" && right.Type() == "infix" {
		leftExpr := left.(*ast.InfixExpression)
		rightExpr := right.(*ast.InfixExpression)
		if leftExpr.Action == "or" && rightExpr.Action == "or" {
			if matched, _ := negatedAlternativeRule(leftExpr.Left, rightExpr.Left); matched {
				return true, &ast.Boolean{Value: true}
			}
			if matched, _ := negatedAlternativeRule(leftExpr.Left, rightExpr.Right); matched {
				return true, &ast.Boolean{Value: true}
			}
			if matched, _ := negatedAlternativeRule(leftExpr.Right, rightExpr.Left); matched {
				return true, &ast.Boolean{Value: true}
			}
			if matched, _ := negatedAlternativeRule(leftExpr.Right, rightExpr.Right); matched {
				return true, &ast.Boolean{Value: true}
			}
		}
	}

	return false, nil
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

	if matched, expr := duplicateAlternativeRule(expr.Left, expr.Right); matched {
		debugPrint("turning", expression.Literal(), "into", expr.Literal())
		return true, expr
	}
	return false, expression
}

func duplicateAlternativeRule(left ast.Expression, right ast.Expression) (bool, ast.Expression) {
	if left.Type() == "identifier" && right.Type() == "identifier" {
		if left.Literal() == right.Literal() {
			return true, left
		}
	}
	if left.Type() == "prefix" && right.Type() == "prefix" {
		leftExpr := left.(*ast.PrefixExpression)
		rightExpr := right.(*ast.PrefixExpression)
		if leftExpr.Op == "!" && rightExpr.Op == "!" {
			if leftExpr.Right.Type() == "identifier" && rightExpr.Right.Type() == "identifier" {
				if leftExpr.Right.Literal() == rightExpr.Right.Literal() {
					return true, left
				}
			}
		}
	}

	if left.Type() == "infix" && (right.Type() == "identifier" || right.Type() == "prefix") {
		expr := left.(*ast.InfixExpression)
		if expr.Action == "or" {
			if matched, _ := duplicateAlternativeRule(expr.Left, right); matched {
				return true, &ast.InfixExpression{
					Op:     "+",
					Action: "or",
					Left:   left,
					Right:  &ast.Boolean{Value: false},
				}
			}
			if matched, _ := duplicateAlternativeRule(expr.Right, right); matched {
				return true, &ast.InfixExpression{
					Op:     "+",
					Action: "or",
					Left:   left,
					Right:  &ast.Boolean{Value: false},
				}
			}
		}
	} else if (left.Type() == "identifier" || left.Type() == "prefix") && right.Type() == "infix" {
		expr := right.(*ast.InfixExpression)
		if expr.Action == "or" {
			if matched, expr := duplicateAlternativeRule(left, expr.Left); matched {
				return true, &ast.InfixExpression{
					Op:     "+",
					Action: "or",
					Left:   expr,
					Right:  &ast.Boolean{Value: false},
				}
			}
			if matched, expr := duplicateAlternativeRule(expr.Right, right); matched {
				return true, &ast.InfixExpression{
					Op:     "+",
					Action: "or",
					Left:   &ast.Boolean{Value: false},
					Right:  expr,
				}
			}
		}
	} else if left.Type() == "infix" && right.Type() == "infix" {
		leftExpr := left.(*ast.InfixExpression)
		rightExpr := right.(*ast.InfixExpression)
		if leftExpr.Action == "or" && rightExpr.Action == "or" {
			if matched, expr := duplicateAlternativeRule(leftExpr.Left, rightExpr.Left); matched {
				return true, &ast.InfixExpression{
					Op:     "+",
					Action: "or",
					Left:   expr,
					Right:  &ast.Boolean{Value: false},
				}
			}
			if matched, expr := duplicateAlternativeRule(leftExpr.Left, rightExpr.Right); matched {
				return true, &ast.InfixExpression{
					Op:     "+",
					Action: "or",
					Left:   expr,
					Right:  &ast.Boolean{Value: false},
				}
			}
			if matched, expr := duplicateAlternativeRule(leftExpr.Right, rightExpr.Left); matched {
				return true, &ast.InfixExpression{
					Op:     "+",
					Action: "or",
					Left:   expr,
					Right:  &ast.Boolean{Value: false},
				}
			}
			if matched, expr := duplicateAlternativeRule(leftExpr.Right, rightExpr.Right); matched {
				return true, &ast.InfixExpression{
					Op:     "+",
					Action: "or",
					Left:   expr,
					Right:  &ast.Boolean{Value: false},
				}
			}
		}
	}

	return false, nil
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
