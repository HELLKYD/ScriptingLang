package tokenizer

import (
	"errors"
	"fmt"
)

type TokenReader struct {
	tokens []Token
	index  int
}

func NewTokenReader(tokens []Token) TokenReader {
	return TokenReader{tokens: tokens, index: 0}
}

func (tr TokenReader) ReadToken() (Token, error) {
	if tr.index >= len(tr.tokens) || tr.index < 0 {
		return Token{}, errors.New("error: index out of bounds")
	}
	return tr.tokens[tr.index], nil
}

func (tr TokenReader) ReadTokenAtOffset(offset int) (Token, error) {
	if tr.index+offset >= len(tr.tokens) || tr.index+offset < 0 {
		return Token{}, fmt.Errorf("error: cannot read tokens at position %v (index out of bounds)", tr.index+offset)
	}
	return tr.tokens[tr.index+offset], nil
}

func (tr *TokenReader) NextToken() {
	tr.index++
}

func (tr *TokenReader) SkipTokens(tokens int) error {
	if tr.index+tokens >= len(tr.tokens) || tr.index+tokens < 0 {
		return fmt.Errorf("error: cannot skip to index %v", tr.index+tokens)
	}
	tr.index += tokens
	return nil
}

func (tr *TokenReader) UnreadToken() error {
	if tr.index-1 < 0 {
		return errors.New("error: no previous token")
	}
	tr.index--
	return nil
}

func (tr *TokenReader) UnreadTokens(tokens int) error {
	if tr.index-tokens < 0 {
		return fmt.Errorf("error: cannot unread %v tokens", tokens)
	}
	for i := 0; i < tokens; i++ {
		tr.UnreadToken()
	}
	return nil
}

func (tr TokenReader) GetCurrentPosition() int {
	return tr.index
}
