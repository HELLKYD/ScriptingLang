package parser

import (
	"compiler/instructions"
	"compiler/tokenizer"
	"log"
	"strconv"
)

type Variable struct {
	VariableIndex int
	Type          string
}

type ExpNodeType int

const (
	ERROR ExpNodeType = iota
	NUMBER
	POSITIVE
	NEGATIVE
	ADD
	SUB
	MUL
	DIV
	POW
	IDENTIFIER
	FUNCTION_CALL
)

type precedence int

const (
	MIN precedence = iota
	TERM
	MULT
	DIVI
	POWER
	MAX
)

var precedenceLookupTable map[tokenizer.TokenType]precedence = map[tokenizer.TokenType]precedence{
	tokenizer.PLUS:  TERM,
	tokenizer.MINUS: TERM,
	tokenizer.MUL:   MULT,
	tokenizer.DIV:   DIVI,
	tokenizer.POW:   POWER,
}

type MathExpNode struct {
	Kind     ExpNodeType
	Number   tokenizer.Token
	FuncCall FunctionCall
	Unary    struct {
		Operand *MathExpNode
	}
	Binary struct {
		Left  *MathExpNode
		Right *MathExpNode
	}
}

func (mxp MathExpNode) GetExpressionType() string {
	return MATH_EXP
}

func (mxp MathExpNode) GenerateByteCode(context *GeneratorContext) []byte {
	byteCode := make([]byte, 0)
	if mxp.Kind == ADD {
		byteCode = append(byteCode, mxp.getOperationArgsByteCode(context)...)
		byteCode = append(byteCode, instructions.IADD)
	} else if mxp.Kind == SUB {
		byteCode = append(byteCode, mxp.getOperationArgsByteCode(context)...)
		byteCode = append(byteCode, instructions.ISUB)
	} else if mxp.Kind == MUL {
		byteCode = append(byteCode, mxp.getOperationArgsByteCode(context)...)
		byteCode = append(byteCode, instructions.IMUL)
	} else if mxp.Kind == DIV {
		byteCode = append(byteCode, mxp.getOperationArgsByteCode(context)...)
		byteCode = append(byteCode, instructions.IDIV)
	} else if mxp.Kind == NUMBER {
		number, err := strconv.ParseInt(mxp.Number.Value, 10, 32)
		if err != nil {
			log.Fatalf("error: expected number")
		}
		inst, ok := instructions.Iconsts[int32(number)]
		if !ok {
			log.Fatalf("error: number too large")
		}
		byteCode = append(byteCode, inst)
	} else if mxp.Kind == IDENTIFIER {
		variable, ok := context.Variables[mxp.Number.Value]
		if !ok {
			log.Fatalf("error: cannot use undeclared variable '%v'", mxp.Number.Value)
		}
		byteCode = append(byteCode, loadVariable(variable)...)
	} else if mxp.Kind == FUNCTION_CALL {
		fun, ok := discoveredFunctions[mxp.FuncCall.CalledFunctionName]
		if !ok {
			log.Fatalf("error: cannot call undefined function %v", mxp.FuncCall.CalledFunctionName)
		}
		if fun.ReturnType != "int" {
			log.Fatalf("error: cannot use function with return type %v in a mathmatical expression", fun.ReturnType)
		}
		byteCode = append(byteCode, mxp.FuncCall.GenerateByteCode(context)...)
	}
	return byteCode
}

func (mxp MathExpNode) getOperationArgsByteCode(context *GeneratorContext) (byteCode []byte) {
	byteCode = make([]byte, 0)
	number1, number2 := mxp.Binary.Left.GenerateByteCode(context), mxp.Binary.Right.GenerateByteCode(context)
	byteCode = append(byteCode, number1...)
	byteCode = append(byteCode, number2...)
	return
}

type MathmaticalParser struct {
	parser *Parser
}

func NewMathmaticalParser(parser *Parser) MathmaticalParser {
	return MathmaticalParser{parser: parser}
}

func (mp MathmaticalParser) Parse() MathExpNode {
	return *mp.parseExpression(MIN)
}

func (mp MathmaticalParser) parseExpression(curOpPrecedence precedence) *MathExpNode {
	left := mp.parsePrefixExpression()
	nextOp, err := mp.parser.reader.ReadToken()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	nextOpPrecedence := getPrecedenceOfOp(nextOp.Type)
	for nextOpPrecedence != MIN {
		if curOpPrecedence >= nextOpPrecedence {
			break
		} else {
			mp.parser.reader.NextToken()
			left = mp.parseInfixExpression(nextOp, left)
			nextOp, err = mp.parser.reader.ReadToken()
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			nextOpPrecedence = getPrecedenceOfOp(nextOp.Type)
		}
	}
	return left
}

func (mp MathmaticalParser) parsePrefixExpression() *MathExpNode {
	ret := MathExpNode{Kind: ERROR}
	curr, err := mp.parser.reader.ReadToken()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if curr.Type == tokenizer.NUMBER {
		ret = MathExpNode{Kind: NUMBER, Number: curr}
		mp.parser.reader.NextToken()
	} else if curr.Type == tokenizer.IDENTIFIER {
		next, err := mp.parser.reader.ReadTokenAtOffset(1)
		isUnexpectedEndOfInput(err)
		if next.Type == tokenizer.OPEN_PAR {
			mp.parser.reader.NextToken()
			fc := parseFunctionCallExp(mp.parser, curr, true)
			return &MathExpNode{Kind: FUNCTION_CALL, FuncCall: fc.(FunctionCall)}
		}
		ret = MathExpNode{Kind: IDENTIFIER, Number: curr}
		mp.parser.reader.NextToken()
	} else if curr.Type == tokenizer.OPEN_PAR {
		mp.parser.reader.NextToken()
		ret = *mp.parseExpression(MIN)
		temp, err := mp.parser.reader.ReadToken()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		if temp.Type == tokenizer.CLOSE_PAR {
			mp.parser.reader.NextToken()
		}
	} else if curr.Type == tokenizer.PLUS {
		mp.parser.reader.NextToken()
		ret = MathExpNode{Kind: POSITIVE, Unary: struct{ Operand *MathExpNode }{Operand: mp.parsePrefixExpression()}}
	} else if curr.Type == tokenizer.MINUS {
		mp.parser.reader.NextToken()
		ret = MathExpNode{Kind: NEGATIVE, Unary: struct{ Operand *MathExpNode }{Operand: mp.parsePrefixExpression()}}
	}
	return &ret
}

func getPrecedenceOfOp(t tokenizer.TokenType) precedence {
	value, ok := precedenceLookupTable[t]
	if !ok {
		return MIN
	}
	return value
}

func (mp MathmaticalParser) parseInfixExpression(op tokenizer.Token, left *MathExpNode) *MathExpNode {
	ret := MathExpNode{}
	switch op.Type {
	case tokenizer.PLUS:
		ret.Kind = ADD
	case tokenizer.MINUS:
		ret.Kind = SUB
	case tokenizer.MUL:
		ret.Kind = MUL
	case tokenizer.DIV:
		ret.Kind = DIV
	case tokenizer.POW:
		ret.Kind = POW
	}
	ret.Binary.Left = left
	ret.Binary.Right = mp.parseExpression(getPrecedenceOfOp(op.Type))
	return &ret
}
