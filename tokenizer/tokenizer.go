package tokenizer

import (
	"unicode"
)

type TokenKind string

const (
	TOK_SIMPLIFY    TokenKind = "simplify"
	TOK_FALSE       TokenKind = "false"
	TOK_TRUE        TokenKind = "true"
	TOK_ILLEGAL     TokenKind = "illegal"
	TOK_IDENT       TokenKind = "ident"
	TOK_LPAREN      TokenKind = "lparen"
	TOK_RPAREN      TokenKind = "rparen"
	TOK_BANG        TokenKind = "bang"
	TOK_AND         TokenKind = "and"
	TOK_OR          TokenKind = "or"
	TOK_IMPLICATION TokenKind = "implication"
	TOK_BICONDITION TokenKind = "bicondition"
	TOK_EQ          TokenKind = "eq"
	TOK_NEQ         TokenKind = "neq"
	TOK_INTRODUCE   TokenKind = "introduce"
	TOK_TABLE       TokenKind = "table"
)

var KEYWORDS = []TokenKind{
	TOK_INTRODUCE,
	TOK_SIMPLIFY,
	TOK_TABLE,
	TOK_FALSE,
	TOK_TRUE,
}

var ACTIONS = map[string]string{
	"*": string(TOK_AND),
	"&": string(TOK_AND),
	"+": string(TOK_OR),
	"|": string(TOK_OR),
}

type Tokenizer struct {
	Content         string
	Runes           []rune
	Tokens          []Token
	cursor          int
	char            rune
	nextChar        rune
	isInsideComment bool
	IsTokenized     bool
}

type Token struct {
	Kind    TokenKind
	Literal string
	Start   int
	Length  int
}

func Contains[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func NewTokenizer(content string) *Tokenizer {
	t := &Tokenizer{
		Content:     content,
		Runes:       []rune(content),
		Tokens:      make([]Token, 0),
		cursor:      -1,
		IsTokenized: false,
	}

	t.Next()

	return t
}

func (t *Tokenizer) Tokenize() error {
	for t.char != 0 {
		pushNext := true

		t.skipWhitespace()

		token := Token{Start: t.cursor, Length: 1, Literal: string(t.char)}

		switch t.char {
		case '(':
			token.Kind = TOK_LPAREN
		case ')':
			token.Kind = TOK_RPAREN
		case '<':
			if t.nextChar == '-' {
				t.Next()
				if t.nextChar == '>' {
					t.Next()
					token.Kind = TOK_BICONDITION
					token.Literal = "<->"
					token.Length = 3
				} else {
					token.Kind = TOK_ILLEGAL
				}
			} else {
				token.Kind = TOK_ILLEGAL
			}
		case '-':
			if t.nextChar == '>' {
				t.Next()
				token.Kind = TOK_IMPLICATION
				token.Literal = "->"
				token.Length = 2
			} else {
				token.Kind = TOK_ILLEGAL
			}
		case '&', '*':
			token.Kind = TOK_AND
		case '|', '+':
			token.Kind = TOK_OR
		case '=':
			if t.nextChar == '=' {
				t.Next()
				token.Kind = TOK_EQ
				token.Literal = "=="
				token.Length = 2
			}
		case '0':
			token.Kind = TOK_FALSE
		case '1':
			token.Kind = TOK_TRUE
		case '!':
			if t.nextChar == '=' {
				t.Next()
				token.Kind = TOK_NEQ
				token.Literal = "!="
				token.Length = 2
			} else {
				token.Kind = TOK_BANG
			}
		default:
			if t.isValidVariableName() {
				start := t.cursor
				for t.isValidVariableName() || t.isValidNumber() {
					t.Next()
				}
				literal := TokenKind(t.Runes[start:t.cursor])
				if Contains(KEYWORDS, literal) {
					token = Token{
						Kind:    literal,
						Literal: string(literal),
					}
				} else {
					token = Token{
						Kind:    TOK_IDENT,
						Literal: string(t.Runes[start:t.cursor]),
					}
				}
				token.Start = start
				token.Length = t.cursor - start
				pushNext = false
			} else {
				token.Kind = TOK_ILLEGAL
				token.Literal = string(t.char)
			}
		}

		t.Tokens = append(t.Tokens, token)
		if pushNext {
			t.Next()
		}
	}

	t.IsTokenized = true

	return nil
}

func (t *Tokenizer) Next() {
	t.cursor += 1
	if t.cursor >= len(t.Runes) {
		t.char = 0
	} else {
		t.char = t.Runes[t.cursor]
	}

	if t.cursor+1 >= len(t.Runes) {
		t.nextChar = 0
	} else {
		t.nextChar = t.Runes[t.cursor+1]
	}
}

func (t *Tokenizer) skipWhitespace() {
	for t.char == ' ' || t.char == '\t' || t.char == '\n' || t.char == '\r' {
		t.Next()
	}
}

func (t *Tokenizer) isValidVariableName() bool {
	return unicode.IsLetter(t.char) || t.char == '_'
}

func (t *Tokenizer) isValidNumber() bool {
	return t.char >= '0' && t.char <= '9'
}
