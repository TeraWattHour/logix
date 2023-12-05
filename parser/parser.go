package parser

import (
	"github.com/terawatthour/logix/ast"
	"github.com/terawatthour/logix/tokenizer"
)

type Precedence int

const (
	_ Precedence = iota
	LOWEST
	INFIX
	PREFIX
)

var precedences = map[tokenizer.TokenKind]Precedence{
	tokenizer.TOK_OR:          INFIX,
	tokenizer.TOK_IMPLICATION: INFIX,
	tokenizer.TOK_AND:         INFIX,
	tokenizer.TOK_BICONDITION: INFIX,

	tokenizer.TOK_BANG: PREFIX,

	tokenizer.TOK_IDENT: LOWEST,
}

type ParsingError struct {
	Errors []string
}

func (e *ParsingError) Error() string {
	result := "\033[1m\033[31mParsing failed!\033[0m Encountered the following errors:\n"
	for _, err := range e.Errors {
		result += err + "\n"
	}
	return result
}

func NewParsingError(errors []string) *ParsingError {
	if len(errors) == 0 {
		return nil
	}
	return &ParsingError{Errors: errors}
}

type Parser struct {
	l *tokenizer.Tokenizer

	i            int
	currentToken *tokenizer.Token
	nextToken    *tokenizer.Token

	prefixParseFns map[tokenizer.TokenKind]func() ast.Expression
	infixParseFns  map[tokenizer.TokenKind]func(ast.Expression) ast.Expression

	errors []string
}

// NewParser creates a new parser, requires a tokenized tokenizer as an argument.
func NewParser(l *tokenizer.Tokenizer) *Parser {
	p := &Parser{
		l:              l,
		i:              -1,
		prefixParseFns: make(map[tokenizer.TokenKind]func() ast.Expression),
		infixParseFns:  make(map[tokenizer.TokenKind]func(ast.Expression) ast.Expression),
		errors:         make([]string, 0),
	}

	p.registerPrefix(tokenizer.TOK_IDENT, p.parseIdentifier)
	p.registerPrefix(tokenizer.TOK_FALSE, p.parseBoolean)
	p.registerPrefix(tokenizer.TOK_TRUE, p.parseBoolean)
	p.registerPrefix(tokenizer.TOK_BANG, p.parsePrefixExpression)
	p.registerPrefix(tokenizer.TOK_LPAREN, p.parseGroupExpression)

	p.registerInfix(tokenizer.TOK_OR, p.parseInfixExpression)
	p.registerInfix(tokenizer.TOK_IMPLICATION, p.parseInfixExpression)
	p.registerInfix(tokenizer.TOK_BICONDITION, p.parseInfixExpression)
	p.registerInfix(tokenizer.TOK_AND, p.parseInfixExpression)

	return p
}

func (p *Parser) Parse() (stmt ast.Statement, err error) {
	if !p.l.IsTokenized {
		p.errors = append(p.errors, "tokenizer must be tokenized before parsing")
		return nil, NewParsingError(p.errors)
	}
	p.advanceToken()

	switch p.currentToken.Kind {
	case tokenizer.TOK_TABLE:
		stmt = p.parseTableStatement()
	case tokenizer.TOK_SIMPLIFY:
		stmt = p.parseSimplifyStatement()
	default:
		p.errors = append(p.errors, "unexpected token "+p.currentToken.Literal)
	}

	if len(p.errors) > 0 {
		return nil, NewParsingError(p.errors)
	}

	return stmt, nil
}

func (p *Parser) parseSimplifyStatement() *ast.SimplifyStatement {
	stmt := &ast.SimplifyStatement{Token: p.currentToken}

	if p.nextIsEnd() {
		p.errors = append(p.errors, "unexpected EOF")
	} else {
		p.advanceToken()
		stmt.Expression = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseTableStatement() *ast.TableStatement {
	stmt := &ast.TableStatement{Token: p.currentToken}

	if p.nextIsEnd() {
		p.errors = append(p.errors, "unexpected EOF")
	} else {
		p.advanceToken()
		stmt.Expression = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	if p.currentToken == nil {
		p.errors = append(p.errors, "unexpected EOF")
		return nil
	}

	prefix := p.prefixParseFns[p.currentToken.Kind]
	if prefix == nil {
		p.errors = append(p.errors, "no prefix parse function for "+p.currentToken.Literal)
		return nil
	}

	leftExp := prefix()
	for p.currentToken != nil && precedence < p.nextPrecedence() {
		infix := p.infixParseFns[p.nextToken.Kind]
		if infix == nil {
			return leftExp
		}

		p.advanceToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseGroupExpression() ast.Expression {
	p.advanceToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectNext(tokenizer.TOK_RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Token: p.currentToken,
		Op:    p.currentToken.Literal,
	}
	p.advanceToken()
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	action, ok := tokenizer.ACTIONS[p.currentToken.Literal]
	if !ok {
		action = p.currentToken.Literal
	}
	expr := &ast.InfixExpression{
		Token:  p.currentToken,
		Op:     p.currentToken.Literal,
		Action: action,
		Left:   left,
	}
	precedence := p.currentPrecedence()
	p.advanceToken()
	expr.Right = p.parseExpression(precedence)
	return expr
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.currentToken, Value: p.currentToken.Kind == tokenizer.TOK_TRUE}
}

func (p *Parser) registerPrefix(forKind tokenizer.TokenKind, fn func() ast.Expression) {
	p.prefixParseFns[forKind] = fn
}

func (p *Parser) registerInfix(forKind tokenizer.TokenKind, fn func(left ast.Expression) ast.Expression) {
	p.infixParseFns[forKind] = fn
}

func (p *Parser) currentPrecedence() Precedence {
	if precedence, ok := precedences[p.currentToken.Kind]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) nextPrecedence() Precedence {
	if p.nextToken == nil {
		return LOWEST
	}

	if precedence, ok := precedences[p.nextToken.Kind]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) currentIs(t tokenizer.TokenKind) bool {
	return p.currentToken.Kind == t
}

func (p *Parser) nextIs(t tokenizer.TokenKind) bool {
	return p.nextToken.Kind == t
}

func (p *Parser) nextIsEnd() bool {
	return p.nextToken == nil
}

func (p *Parser) expectNext(kind tokenizer.TokenKind) bool {
	if p.nextToken.Kind == kind {
		p.advanceToken()
		return true
	}
	// handle error
	return false
}

func (p *Parser) advanceToken() {
	p.i++
	if p.i >= len(p.l.Tokens) {
		p.currentToken = nil
		p.nextToken = nil
		return
	}

	p.currentToken = &p.l.Tokens[p.i]

	if p.i+1 >= len(p.l.Tokens) {
		p.nextToken = nil
	} else {
		p.nextToken = &p.l.Tokens[p.i+1]
	}
}
