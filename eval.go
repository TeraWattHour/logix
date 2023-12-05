package main

import (
	"fmt"
	"github.com/terawatthour/logix/ast"
	"math"
	"slices"
	"strings"
)

type Evaluator struct {
	simplificationRules []func(ast.Expression) (bool, ast.Expression)
}

func NewEvaluator() *Evaluator {
	e := &Evaluator{
		simplificationRules: make([]func(ast.Expression) (bool, ast.Expression), 0),
	}

	e.registerSimplificationRule(IdentityRule)
	e.registerSimplificationRule(IdempotenceRule)
	e.registerSimplificationRule(NegatedAlternativeRule)
	e.registerSimplificationRule(NegatedConjunctionRule)
	e.registerSimplificationRule(ImplicationRule)
	e.registerSimplificationRule(DeMorganRule)
	e.registerSimplificationRule(DoubleNegationRule)
	e.registerSimplificationRule(DuplicateAlternativeRule)

	return e
}

func (e *Evaluator) registerSimplificationRule(fn func(ast.Expression) (bool, ast.Expression)) {
	e.simplificationRules = append(e.simplificationRules, fn)
}

func (e *Evaluator) evaluate(statement ast.Statement) string {
	switch stmt := statement.(type) {
	case *ast.TableStatement:
		return generateTruthTable(stmt.Expression)
	case *ast.SimplifyStatement:
		return e.Simplify(stmt.Expression).Literal()
	}

	panic("implement me")
}

func (e *Evaluator) Simplify(expression ast.Expression) ast.Expression {
	for {
		previous := expression.Literal()
		expression = e.simplify(expression)
		if previous == expression.Literal() {
			break
		}
	}
	return expression
}

func (e *Evaluator) simplify(expression ast.Expression) ast.Expression {
	switch expr := expression.(type) {
	case *ast.InfixExpression:
		expr.Left = e.simplify(expr.Left)
		expr.Right = e.simplify(expr.Right)
		return e.applyRules(expr)
	case *ast.PrefixExpression:
		expr.Right = e.simplify(expr.Right)
		return e.applyRules(expr)
	case *ast.Identifier:
		return expr
	case *ast.Boolean:
		return expr
	}

	panic("unreachable")
}

func (e *Evaluator) applyRules(expr ast.Expression) ast.Expression {
	for _, rule := range e.simplificationRules {
		_, expr = rule(expr)
	}
	return expr
}

func (e *Evaluator) simplifyInfixExpression(left ast.Expression, operator string, right ast.Expression) ast.Expression {
	return nil
}

func (e *Evaluator) simplifyPrefixExpression(operator string, right ast.Expression) ast.Expression {
	return nil
}

func evaluateExpression(input map[string]bool, expression ast.Expression) bool {
	switch expr := expression.(type) {
	case *ast.InfixExpression:
		left := evaluateExpression(input, expr.Left)
		right := evaluateExpression(input, expr.Right)
		return evaluateInfixExpression(left, expr.Op, right)
	case *ast.PrefixExpression:
		return evaluatePrefixExpression(expr.Op, evaluateExpression(input, expr.Right))
	case *ast.Identifier:
		return input[expr.Value]
	case *ast.Boolean:
		return expr.Value
	}

	panic("unreachable")
}

func evaluateInfixExpression(left bool, operator string, right bool) bool {
	switch operator {
	case "+", "|":
		return left || right
	case "*", "&":
		return left && right
	case "->":
		return !left || right
	case "<->":
		return (!left || right) && (!right || left)
	}
	return false
}

func evaluatePrefixExpression(operator string, right bool) bool {
	switch operator {
	case "!":
		return !right
	}
	return false
}

func generateTruthTable(expression ast.Expression) string {
	idents := getAllIdentifiers(expression, []string{})
	permutations := generateBinaryPermutations(len(idents))

	table := ""

	header := ""
	border := "┌"
	for _, ident := range idents {
		header += fmt.Sprintf("│ %s ", ident)
	}
	for _, ident := range idents {
		border += fmt.Sprintf("%s┬", generatePadding("─", len(ident)+2))
	}
	border += fmt.Sprintf("%s┐\n", generatePadding("─", len(expression.Literal())+2))
	header += fmt.Sprintf("│ %s │\n", bold(expression.Literal()))
	table += border
	table += header
	table += strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(border, "┬", "┼"), "┌", "├"), "┐", "┤")

	literalLength := len(expression.Literal())

	for _, permutation := range permutations {
		values := make(map[string]bool)
		for i, v := range permutation {
			values[idents[i]] = v
			table += fmt.Sprintf("│ %s ", fillSpace(colourBool(v), 1, len(idents[i])))
		}
		result := evaluateExpression(values, expression)
		table += fmt.Sprintf("│ %s │\n", fillSpace(bold(colourBool(result)), 1, literalLength))
	}

	table += strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(border, "┬", "┴"), "┌", "└"), "┐", "┘")
	return table
}

func fillSpace(s string, l int, space int) string {
	leftSpace := space - l
	leading := leftSpace / 2
	trailing := leftSpace - leading
	return fmt.Sprintf("%s%s%s", generatePadding(" ", leading), s, generatePadding(" ", trailing))
}

func generatePadding(character string, n int) string {
	res := ""
	for i := 0; i < n; i++ {
		res += character
	}
	return res
}

func colourBool(b bool) string {
	if b {
		return "\033[32m1\033[0m"
	} else {
		return "\033[31m0\033[0m"
	}
}

func bold(s string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}

func integerToBits(x int, digits int) []bool {
	var result []bool
	for x > 0 {
		result = append(result, x%2 == 1)
		x /= 2
	}

	slices.Reverse(result)

	for len(result) < digits {
		result = append([]bool{false}, result...)
	}

	return result
}

func generateBinaryPermutations(x int) [][]bool {
	result := make([][]bool, 0)

	for i := 0; i < int(math.Pow(2, float64(x))); i++ {
		result = append(result, integerToBits(i, x))
	}

	slices.Reverse(result)

	return result
}

func merge(a []string, b []string) []string {
	for _, v := range b {
		if !contains(a, v) {
			a = append(a, v)
		}
	}
	return a
}

func contains[T comparable](a []T, b T) bool {
	for _, v := range a {
		if v == b {
			return true
		}
	}
	return false
}

func getAllIdentifiers(expression ast.Expression, current []string) []string {
	if current == nil {
		current = make([]string, 0)
	}

	switch expr := expression.(type) {
	case *ast.InfixExpression:
		current = merge(current, getAllIdentifiers(expr.Left, current))
		current = merge(current, getAllIdentifiers(expr.Right, current))
	case *ast.PrefixExpression:
		current = merge(current, getAllIdentifiers(expr.Right, current))
	case *ast.Identifier:
		current = merge(current, []string{expr.Value})
	}
	return current
}
