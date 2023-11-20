package parser

import (
	"compiler/classfile"
	"compiler/tokenizer"
	"log"
)

type Parser struct {
	Source []tokenizer.Token
	reader tokenizer.TokenReader
	class  *classfile.Class
}

var discoveredFunctions map[string]Function = make(map[string]Function)

func NewParser(src []tokenizer.Token, class *classfile.Class) Parser {
	discoveredFunctions["println"] = Function{
		ReturnType: "void",
		Args: []FunctionArgument{
			{Name: "value", Type: "int"},
		},
	}
	return Parser{Source: src, reader: tokenizer.NewTokenReader(src), class: class}
}

func (p *Parser) parseExpression() Expression {
	prev, err := p.reader.ReadTokenAtOffset(-1)
	isUnexpectedEndOfInput(err)
	cur, err := p.reader.ReadToken()
	isUnexpectedEndOfInput(err)
	p.reader.NextToken()
	next, err := p.reader.ReadToken()
	isUnexpectedEndOfInput(err)
	if isStartOfMathExp(cur, next) {
		p.reader.UnreadToken()
		return NewMathmaticalParser(p).Parse()
	} else if isFunctionCallStart(cur, next, prev) {
		return parseFunctionCallExp(p, cur, false)
	} else if cur.Type == tokenizer.IDENTIFIER {
		return Identifier{Value: cur}
	}
	return nil
}

func parseFunctionCallExp(p *Parser, cur tokenizer.Token, isInMathmeticalExp bool) Expression {
	positionOfFuncName := p.reader.GetCurrentPosition() - 1
	p.reader.NextToken()
	functionName := cur.Value
	args := make([]Expression, 0)
	parseFuncCallArgs(p, &args)
	p.reader.NextToken()
	next, err := p.reader.ReadToken()
	isUnexpectedEndOfInput(err)
	if tokenizer.IsOperator(next) && !isInMathmeticalExp {
		positionOfOp := p.reader.GetCurrentPosition()
		p.reader.UnreadTokens(positionOfOp - positionOfFuncName)
		mathExp := NewMathmaticalParser(p).Parse()
		return mathExp
	}
	return FunctionCall{CalledFunctionName: functionName, Arguments: args}
}

func isStartOfMathExp(cur, next tokenizer.Token) bool {
	return cur.Type == tokenizer.NUMBER || cur.Type == tokenizer.PLUS ||
		cur.Type == tokenizer.MINUS || cur.Type == tokenizer.OPEN_PAR ||
		(cur.Type == tokenizer.IDENTIFIER && tokenizer.IsOperator(next))
}

func isFunctionCallStart(cur, next, prev tokenizer.Token) bool {
	return cur.Type == tokenizer.IDENTIFIER && next.Type == tokenizer.OPEN_PAR && prev.Type != tokenizer.FUN_DEF
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
		return parseReturnStatement(p)
	} else if cur.Type == tokenizer.VARDECL {
		return parseVarDecl(p)
	} else if cur.Type == tokenizer.FUN_DEF {
		return parseFunDef(p)
	} else if cur.Type == tokenizer.IDENTIFIER {
		next, err := p.reader.ReadToken()
		if err != nil {
			log.Fatalf("error: unexpected end of input")
		}
		p.reader.NextToken()
		if next.Type == tokenizer.ASSIGN {
			return parseVarReassignment(p, cur)
		} else if next.Type == tokenizer.PLUS {
			return parseVarAddToValue(p, cur)
		} else if next.Type == tokenizer.OPEN_PAR {
			return parseFunctionCall(p, cur)
		}
	}
	return nil
}

func parseReturnStatement(p *Parser) ReturnStatement {
	expr := p.parseExpression()
	if expr == nil {
		log.Fatalf("error: could not parse expression")
	}
	tok, _ := p.reader.ReadToken()
	isSemicolon(tok)
	p.reader.NextToken()
	stmt := ReturnStatement{ReturnValue: expr}
	return stmt
}

func parseVarDecl(p *Parser) VarDecl {
	ident := p.parseExpression()
	if ident == nil || ident.GetExpressionType() != IDENTIFIER_EXP {
		log.Fatalf("error: expected identifier")
	}
	typeOfVar := p.parseExpression()
	if typeOfVar == nil || ident.GetExpressionType() != IDENTIFIER_EXP {
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
	return varDecl
}

func parseFunDef(p *Parser) FunctionDefinition {
	ident := p.parseExpression()
	if ident.GetExpressionType() != IDENTIFIER_EXP {
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
	addDiscoveredFunction(ident.(Identifier).Value.Value, retType, args)
	return funcDef
}

func parseVarReassignment(p *Parser, cur tokenizer.Token) VarReAssignment {
	varIdent := cur
	newValue := p.parseExpression()
	next, err := p.reader.ReadToken()
	isUnexpectedEndOfInput(err)
	isSemicolon(next)
	p.reader.NextToken()
	varReassign := VarReAssignment{Ident: Identifier{Value: varIdent}, Value: newValue}
	return varReassign
}

func parseVarAddToValue(p *Parser, cur tokenizer.Token) VarAddToValue {
	next, err := p.reader.ReadToken()
	isUnexpectedEndOfInput(err)
	if next.Type != tokenizer.ASSIGN {
		log.Fatalf("error: expected '='")
	}
	p.reader.NextToken()
	varIdent := cur
	newValue := p.parseExpression()
	next, err = p.reader.ReadToken()
	isUnexpectedEndOfInput(err)
	isSemicolon(next)
	p.reader.NextToken()
	return VarAddToValue{Ident: Identifier{Value: varIdent}, ValueToAdd: newValue}
}

func parseFunctionCall(p *Parser, cur tokenizer.Token) FunctionCall {
	args := make([]Expression, 0)
	next, err := p.reader.ReadToken()
	isUnexpectedEndOfInput(err)
	if next.Type != tokenizer.CLOSE_PAR {
		parseFuncCallArgs(p, &args)
	}
	p.reader.NextToken()
	next, err = p.reader.ReadToken()
	isUnexpectedEndOfInput(err)
	isSemicolon(next)
	p.reader.NextToken()
	funcCall := FunctionCall{CalledFunctionName: cur.Value, Arguments: args}
	return funcCall
}

func parseFuncCallArgs(p *Parser, args *[]Expression) {
	for {
		exp := p.parseExpression()
		*args = append(*args, exp)
		cur, err := p.reader.ReadToken()
		isUnexpectedEndOfInput(err)
		if cur.Type == tokenizer.CLOSE_PAR {
			break
		} else if cur.Type == tokenizer.COLON {
			p.reader.NextToken()
			continue
		}
	}
}

func isUnexpectedEndOfInput(err error) {
	if err != nil {
		log.Fatalf("error: unexpected end of input")
	}
}

func addDiscoveredFunction(name, retType string, args []FunctionArgument) {
	if _, ok := discoveredFunctions[name]; ok {
		log.Fatalf("error: cannot define a function with the name %v (function with that name already exists)", name)
	}
	discoveredFunctions[name] = Function{ReturnType: retType, Args: args}
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
			*args = append(*args, FunctionArgument{Name: next.Value, Type: temp.Value})
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
		if stmt.GetStatementType() == FUNCDEF {
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
