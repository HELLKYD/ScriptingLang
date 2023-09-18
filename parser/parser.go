package parser

import (
	"compiler/tokenizer"
	"log"
)

type Parser struct {
	Source    []tokenizer.Token
	reader    tokenizer.TokenReader
	variables map[string]string
}

func NewParser(src []tokenizer.Token) Parser {
	return Parser{Source: src, reader: tokenizer.NewTokenReader(src), variables: make(map[string]string)}
}

func (p *Parser) parseExpression() Expression {
	cur, err := p.reader.ReadToken()
	isUnexpectedEndOfInput(err)
	p.reader.NextToken()
	next, err := p.reader.ReadToken()
	isUnexpectedEndOfInput(err)
	if isStartOfMathExp(cur, next) {
		p.reader.UnreadToken()
		return NewMathmeticalParser(&p.reader).Parse()
	} else if cur.Type == tokenizer.IDENTIFIER {
		return Identifier{Value: cur}
	}
	return nil
}

func isStartOfMathExp(cur, next tokenizer.Token) bool {
	return cur.Type == tokenizer.NUMBER || cur.Type == tokenizer.PLUS ||
		cur.Type == tokenizer.MINUS || cur.Type == tokenizer.OPEN_PAR ||
		(cur.Type == tokenizer.IDENTIFIER && tokenizer.IsOperator(next))
}

func isSemicolon(t tokenizer.Token) {
	if t.Type != tokenizer.SEMICOLON {
		log.Fatalf("error: expected ';'")
	}
}

func (p *Parser) parseStatement() Statement {
	cur, err := p.reader.ReadToken()
	isUnexpectedEndOfInput(err)
	p.reader.NextToken()
	if cur.Type == tokenizer.RETURN {
		expr := p.parseExpression()
		if expr == nil {
			log.Fatalf("error: could not parse expression")
		}
		tok, _ := p.reader.ReadToken()
		isSemicolon(tok)
		p.reader.NextToken()
		stmt := ReturnStatement{ReturnValue: expr}
		return stmt
	} else if cur.Type == tokenizer.VARDECL {
		ident := p.parseExpression()
		if ident == nil || ident.GetExpressionType() != "identifier" {
			log.Fatalf("error: expected identifier")
		}
		typeOfVar := p.parseExpression()
		if typeOfVar == nil || ident.GetExpressionType() != "identifier" {
			log.Fatalf("error: expected type")
		}
		next, err := p.reader.ReadToken()
		isUnexpectedEndOfInput(err)
		if next.Type != tokenizer.ASSIGN {
			log.Fatalf("error: expected '='")
		}
		p.reader.NextToken()
		varValue := p.parseExpression()
		next, err = p.reader.ReadToken()
		isUnexpectedEndOfInput(err)
		isSemicolon(next)
		p.reader.NextToken()
		varDecl := VarDecl{Ident: ident.(Identifier), Value: varValue, Type: typeOfVar.(Identifier)}
		p.variables[ident.(Identifier).Value.Value] = typeOfVar.(Identifier).Value.Value
		return varDecl
	} else if cur.Type == tokenizer.FUN_DEF {
		ident := p.parseExpression()
		if ident.GetExpressionType() != "identifier" {
			log.Fatalf("error: expected identifier")
		}
		next, err := p.reader.ReadToken()
		isUnexpectedEndOfInput(err)
		if next.Type != tokenizer.OPEN_PAR {
			log.Fatalf("error: expected open parentheses")
		}
		p.reader.NextToken()
		args := make([]FunctionArgument, 0)
		p.parseFuncArgs(&args)
		retType := ""
		next, err = p.reader.ReadToken()
		isUnexpectedEndOfInput(err)
		getFuncReturnType(&retType, next, p)
		next, err = p.reader.ReadToken()
		isUnexpectedEndOfInput(err)
		p.reader.NextToken()
		if next.Type != tokenizer.CURL_OPEN_PAR {
			log.Fatalf("error: expected '{'")
		}
		stmts := make([]Statement, 0)
		p.parseScope(&stmts)
		funcDef := FunctionDefinition{Name: ident.(Identifier).Value.Value, Args: args,
			Scope: Scope{Statements: stmts}, ReturnType: retType}
		log.Println(funcDef.Name, funcDef.ReturnType, funcDef.Args)
		return funcDef
	} else if cur.Type == tokenizer.IDENTIFIER {
		next, err := p.reader.ReadToken()
		if err != nil {
			log.Fatalf("error: unexpected end of input")
		}
		p.reader.NextToken()
		if next.Type == tokenizer.ASSIGN {
			varIdent := cur
			newValue := p.parseExpression()
			next, err = p.reader.ReadToken()
			isUnexpectedEndOfInput(err)
			isSemicolon(next)
			p.reader.NextToken()
			varReassign := VarReAssignment{Ident: Identifier{Value: varIdent}, Value: newValue}
			return varReassign
		} else if next.Type == tokenizer.PLUS {
			next, err = p.reader.ReadToken()
			isUnexpectedEndOfInput(err)
			if next.Type != tokenizer.ASSIGN {
				log.Fatalf("error: expected '='")
			}
			p.reader.NextToken()
			varIdent := cur
			newValue := p.parseExpression()
			next, err := p.reader.ReadToken()
			isUnexpectedEndOfInput(err)
			isSemicolon(next)
			p.reader.NextToken()
			return VarAddToValue{Ident: Identifier{Value: varIdent}, ValueToAdd: newValue}
		}
	}
	return nil
}

func isUnexpectedEndOfInput(err error) {
	if err != nil {
		log.Fatalf("error: unexpected end of input")
	}
}

func getFuncReturnType(retType *string, t tokenizer.Token, p *Parser) {
	if t.Type == tokenizer.IDENTIFIER {
		*retType = t.Value
		p.reader.NextToken()
	} else if t.Type == tokenizer.CURL_OPEN_PAR {
		*retType = "void"
	} else {
		log.Fatalf("error: expected return type")
	}
}

func (p *Parser) parseFuncArgs(args *[]FunctionArgument) {
	for {
		next, err := p.reader.ReadToken()
		isUnexpectedEndOfInput(err)
		p.reader.NextToken()
		if next.Type == tokenizer.CLOSE_PAR {
			break
		} else if next.Type == tokenizer.IDENTIFIER {
			temp, err := p.reader.ReadToken()
			isUnexpectedEndOfInput(err)
			p.reader.NextToken()
			if temp.Type != tokenizer.IDENTIFIER {
				log.Fatalf("error: expected identifier")
			}
			*args = append(*args, FunctionArgument{Name: Identifier{Value: next}, Type: temp.Value})
			continue
		} else if next.Type == tokenizer.COLON {
			continue
		}
	}
}

func (p *Parser) parseScope(stmts *[]Statement) {
	for {
		next, err := p.reader.ReadToken()
		isUnexpectedEndOfInput(err)
		if next.Type == tokenizer.CURL_CLOSE_PAR {
			p.reader.NextToken()
			break
		}
		stmt := p.parseStatement()
		if stmt.GetStatementType() == "funcDef" {
			log.Fatalf("error: cannot define function inside another function")
		}
		*stmts = append(*stmts, stmt)
	}
}

func (p Parser) ParseProgram() Program {
	program := Program{Statements: make([]Statement, 0)}
	for {
		_, err := p.reader.ReadToken()
		if err != nil {
			break
		}
		stmt := p.parseStatement()
		if stmt == nil {
			log.Fatalf("error: could not identify statement")
		}
		program.Statements = append(program.Statements, stmt)
	}
	return program
}
