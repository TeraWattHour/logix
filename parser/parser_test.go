package parser

import (
	"fmt"
	"github.com/terawatthour/logix/ast"
	"github.com/terawatthour/logix/tokenizer"
	"testing"
)

func TestTruthTableExpressionParsing(t *testing.T) {
	content := "table (a + b)"
	to := tokenizer.NewTokenizer(content)
	if err := to.Tokenize(); err != nil {
		t.Fatalf("failed to tokenize: %v", err)
		return
	}
	p := NewParser(to)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
		return
	}
	fmt.Println(program.Literal())
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  string
		operator   string
		rightValue string
	}{
		{"table a | b", "a", "|", "b"},
		{"table a & b", "a", "&", "b"},
		{"table a + b", "a", "+", "b"},
		{"table a * b", "a", "*", "b"},
	}

	for _, tt := range infixTests {
		tok := tokenizer.NewTokenizer(tt.input)
		tok.Tokenize()
		p := NewParser(tok)
		stmt, _ := p.Parse()
		tableStmt, ok := stmt.(*ast.TableStatement)
		if !ok {
			t.Fatalf("statement is not *ast.TableStatement. got=%T", stmt)
		}
		expr := tableStmt.Expression.(*ast.InfixExpression)
		if expr.Left.Literal() != tt.leftValue {
			t.Fatalf("exp.Left is not %s. got=%s", tt.leftValue, expr.Left.Literal())
		}

		if expr.Op != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, expr.Op)
		}

		if expr.Right.Literal() != tt.rightValue {
			t.Fatalf("exp.Right is not %s. got=%s", tt.rightValue, expr.Right.Literal())
		}

	}
}

func TestGroupParsing(t *testing.T) {
	content := "table ((!b | c) | a) * a"
	to := tokenizer.NewTokenizer(content)
	if err := to.Tokenize(); err != nil {
		t.Fatalf("failed to tokenize: %v", err)
		return
	}
	p := NewParser(to)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
		return
	}
	expected := `table (((!b | c) | a) * a)`
	if expected != program.Literal() {
		t.Fatalf("expected %s, got %s", expected, program.Literal())
	}
}
