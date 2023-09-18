package tokenizer

import (
	"io"
	"log"
	"strings"
	"unicode"
)

type TokenType int

const (
	RETURN TokenType = iota
	NUMBER
	SEMICOLON
	IDENTIFIER
	PLUS
	MINUS
	MUL
	DIV
	POW
	OPEN_PAR
	CLOSE_PAR
	ASSIGN
	VARDECL
	CURL_OPEN_PAR
	CURL_CLOSE_PAR
	FUN_DEF
	COLON
)

type Token struct {
	Value string
	Type  TokenType
}

type Tokenizer struct {
	Reader *strings.Reader
	src    []rune
}

func NewTokenizer(source string) *Tokenizer {
	tokenizer := Tokenizer{src: []rune(source), Reader: strings.NewReader(source)}
	return &tokenizer
}

func (t *Tokenizer) GetTokens() []Token {
	tokens := make([]Token, 0)
	tempWord := ""
	for {
		cur, _, err := t.Reader.ReadRune()
		if err == io.EOF {
			break
		}
		if unicode.IsLetter(cur) {
			tempWord += string(cur)
			for {
				temp, _, err := t.Reader.ReadRune()
				if err != nil {
					log.Fatalf("error: %v", err)
				}
				if !(unicode.IsLetter(temp) || unicode.IsDigit(temp)) {
					t.Reader.UnreadRune()
					break
				}
				tempWord += string(temp)
			}
			if tempWord == "return" {
				tokens = append(tokens, Token{Type: RETURN})
				tempWord = ""
				continue
			} else if tempWord == "let" {
				tokens = append(tokens, Token{Type: VARDECL})
				tempWord = ""
				continue
			} else if tempWord == "fun" {
				tokens = append(tokens, Token{Type: FUN_DEF})
				tempWord = ""
				continue
			}
			tokens = append(tokens, Token{Type: IDENTIFIER, Value: tempWord})
			tempWord = ""
			continue
		} else if unicode.IsDigit(cur) {
			tempWord += string(cur)
			for {
				temp, _, err := t.Reader.ReadRune()
				if err == io.EOF {
					break
				}
				if !unicode.IsDigit(temp) {
					t.Reader.UnreadRune()
					break
				}
				tempWord += string(temp)
			}
			tokens = append(tokens, Token{Type: NUMBER, Value: tempWord})
			tempWord = ""
			continue
		} else if cur == ';' {
			tokens = append(tokens, Token{Type: SEMICOLON})
			continue
		} else if unicode.IsSpace(cur) || cur == '\t' {
			continue
		} else if cur == '(' {
			tokens = append(tokens, Token{Type: OPEN_PAR})
		} else if cur == ')' {
			tokens = append(tokens, Token{Type: CLOSE_PAR})
		} else if cur == '+' {
			tokens = append(tokens, Token{Type: PLUS})
		} else if cur == '-' {
			tokens = append(tokens, Token{Type: MINUS})
		} else if cur == '*' {
			r, _, err := t.Reader.ReadRune()
			if err == io.EOF {
				break
			}
			if r == '*' {
				tokens = append(tokens, Token{Type: POW})
				continue
			} else {
				t.Reader.UnreadRune()
			}
			tokens = append(tokens, Token{Type: MUL})
		} else if cur == '/' {
			tokens = append(tokens, Token{Type: DIV})
		} else if cur == '=' {
			tokens = append(tokens, Token{Type: ASSIGN})
		} else if cur == '{' {
			tokens = append(tokens, Token{Type: CURL_OPEN_PAR})
		} else if cur == '}' {
			tokens = append(tokens, Token{Type: CURL_CLOSE_PAR})
		} else if cur == ',' {
			tokens = append(tokens, Token{Type: COLON})
		} else {
			log.Fatalf("error unrecognized token ('%v')", string(cur))
		}
	}
	return tokens
}

func IsOperator(t Token) bool {
	return (t.Type == PLUS || t.Type == MINUS || t.Type == MUL || t.Type == DIV)
}
