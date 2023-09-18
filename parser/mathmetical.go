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
	Kind   ExpNodeType
	Number tokenizer.Token
	Unary  struct {
		Operand *MathExpNode
	}
	Binary struct {
		Left  *MathExpNode
		Right *MathExpNode
	}
}

func (mxp MathExpNode) GetExpressionType() string {
	return "mathExp"
}

func (mxp MathExpNode) GenerateByteCode(variables map[string]Variable) []byte {
	byteCode := make([]byte, 0)
	if mxp.Kind == ADD {
		byteCode = append(byteCode, mxp.getOperationArgsByteCode(variables)...)
		byteCode = append(byteCode, instructions.IADD)
	} else if mxp.Kind == SUB {
		byteCode = append(byteCode, mxp.getOperationArgsByteCode(variables)...)
		byteCode = append(byteCode, instructions.ISUB)
	} else if mxp.Kind == MUL {
		byteCode = append(byteCode, mxp.getOperationArgsByteCode(variables)...)
		byteCode = append(byteCode, instructions.IMUL)
	} else if mxp.Kind == DIV {
		byteCode = append(byteCode, mxp.getOperationArgsByteCode(variables)...)
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
		variable, ok := variables[mxp.Number.Value]
		if !ok {
			log.Fatalf("error: cannot use undeclared variable '%v'", mxp.Number.Value)
		}
		inst, ok := instructions.Iloads[int32(variable.VariableIndex)]
		if !ok {
			byteCode = append(byteCode, instructions.ILOAD)
			byteCode = append(byteCode, uint8(variable.VariableIndex))
		} else {
			byteCode = append(byteCode, inst)
		}
	}
	return byteCode
}

func (mxp MathExpNode) getOperationArgsByteCode(variables map[string]Variable) (byteCode []byte) {
	byteCode = make([]byte, 0)
	number1, number2 := mxp.Binary.Left.GenerateByteCode(variables), mxp.Binary.Right.GenerateByteCode(variables)
	byteCode = append(byteCode, number1...)
	byteCode = append(byteCode, number2...)
	return
}

type MathmeticalParser struct {
	reader *tokenizer.TokenReader
}

func NewMathmeticalParser(reader *tokenizer.TokenReader) MathmeticalParser {
	return MathmeticalParser{reader: reader}
}

func (mp MathmeticalParser) Parse() MathExpNode {
	return *mp.parseExpression(MIN)
}

func (mp MathmeticalParser) parseExpression(curOpPrecedence precedence) *MathExpNode {
	left := mp.parsePrefixExpression()
	nextOp, err := mp.reader.ReadToken()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	nextOpPrecedence := getPrecedenceOfOp(nextOp.Type)
	for nextOpPrecedence != MIN {
		if curOpPrecedence >= nextOpPrecedence {
			break
		} else {
			mp.reader.NextToken()
			left = mp.parseInfixExpression(nextOp, left)
			nextOp, err = mp.reader.ReadToken()
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			nextOpPrecedence = getPrecedenceOfOp(nextOp.Type)
		}
	}
	return left
}

func (mp MathmeticalParser) parsePrefixExpression() *MathExpNode {
	ret := MathExpNode{Kind: ERROR}
	curr, err := mp.reader.ReadToken()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if curr.Type == tokenizer.NUMBER {
		ret = MathExpNode{Kind: NUMBER, Number: curr}
		mp.reader.NextToken()
	} else if curr.Type == tokenizer.IDENTIFIER {
		ret = MathExpNode{Kind: IDENTIFIER, Number: curr}
		mp.reader.NextToken()
	} else if curr.Type == tokenizer.OPEN_PAR {
		mp.reader.NextToken()
		ret = *mp.parseExpression(MIN)
		temp, err := mp.reader.ReadToken()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		if temp.Type == tokenizer.CLOSE_PAR {
			mp.reader.NextToken()
		}
	} else if curr.Type == tokenizer.PLUS {
		mp.reader.NextToken()
		ret = MathExpNode{Kind: POSITIVE, Unary: struct{ Operand *MathExpNode }{Operand: mp.parsePrefixExpression()}}
	} else if curr.Type == tokenizer.MINUS {
		mp.reader.NextToken()
		ret = MathExpNode{Kind: NEGATIVE, Unary: struct{ Operand *MathExpNode }{Operand: mp.parsePrefixExpression()}}
	}

	//todo: function calls implementieren
	return &ret
}

func getPrecedenceOfOp(t tokenizer.TokenType) precedence {
	value, ok := precedenceLookupTable[t]
	if !ok {
		return MIN
	}
	return value
}

func (mp MathmeticalParser) parseInfixExpression(op tokenizer.Token, left *MathExpNode) *MathExpNode {
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
