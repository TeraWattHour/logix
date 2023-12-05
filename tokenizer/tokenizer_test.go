package tokenizer

import "testing"

func TestTokenize(t *testing.T) {
	tokenizer := NewTokenizer("!(a -> b) || a")
	if err := tokenizer.Tokenize(); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	tokens := tokenizer.Tokens

	expected := []string{TOK_BANG, TOK_LPAREN, TOK_IDENT, TOK_IMPLICATION, TOK_IDENT, TOK_RPAREN, TOK_OR, TOK_IDENT}

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, expectedToken := range expected {
		if tokens[i].Kind != expectedToken {
			t.Errorf("Expected %d token to be %s, got %s", i, expectedToken, tokens[0].Kind)
		}
	}
}
