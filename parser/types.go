package parser

import (
	"compiler/classfile"
	"compiler/instructions"
	"compiler/tokenizer"
	"encoding/binary"
	"log"
)

const (
	IDENTIFIER_EXP   = "identifier"
	MATH_EXP         = "mathExp"
	RETURN           = "return"
	VARDECL          = "varDecl"
	VARREASSIGNMENT  = "varReAssignment"
	VARADDTOVARIABLE = "varAddToVariable"
	FUNCTIONARG      = "functionArg"
	FUNCDEF          = "funcDef"
	FUNCTIONCALL     = "functionCall"
)

type GeneratorContext struct {
	Class     *classfile.Class
	MaxLocals *int
	Variables map[string]Variable
}

type Expression interface {
	GetExpressionType() string
}

type Statement interface {
	GetStatementType() string
	GenerateByteCode(context *GeneratorContext) []byte
}

type Identifier struct {
	Value tokenizer.Token
}

func (i Identifier) GetExpressionType() string {
	return IDENTIFIER_EXP
}

type ReturnStatement struct {
	ReturnValue Expression
}

func (r ReturnStatement) GetStatementType() string {
	return RETURN
}

func (r ReturnStatement) GenerateByteCode(context *GeneratorContext) []byte {
	byteCode := make([]byte, 0)
	switch r.ReturnValue.GetExpressionType() {
	case MATH_EXP:
		byteCode = append(byteCode, r.ReturnValue.(MathExpNode).GenerateByteCode(context.Variables)...)
		byteCode = append(byteCode, instructions.IRETURN)
	case IDENTIFIER_EXP:
		variable, ok := context.Variables[r.ReturnValue.(Identifier).Value.Value]
		if !ok {
			log.Fatalf("error: cannot return undefined variable '%v'", r.ReturnValue.(Identifier).Value.Value)
		}
		byteCode = append(byteCode, loadVariable(variable)...)
		byteCode = append(byteCode, instructions.IRETURN)
	default:
		log.Fatalf("error: unsupported expression type (%v)", r.ReturnValue.GetExpressionType())
	}
	return byteCode
}

type VarDecl struct {
	Type  Identifier
	Value Expression
	Ident Identifier
}

func (id VarDecl) GetStatementType() string {
	return VARDECL
}

func (id VarDecl) GenerateByteCode(context *GeneratorContext) []byte {
	if id.Type.Value.Value == "int" {
		_, ok := context.Variables[id.Ident.Value.Value]
		if ok {
			log.Fatalf("error: cannot redeclare variable '%v'", id.Ident.Value.Value)
		}
		byteCode := make([]byte, 0)
		if id.Value.GetExpressionType() == IDENTIFIER_EXP {
			variable, ok := context.Variables[id.Value.(Identifier).Value.Value]
			if !ok {
				log.Fatalf("error: cannot use undeclared variable '%v'", id.Value.(Identifier).Value.Value)
			}
			byteCode = append(byteCode, loadVariable(variable)...)
		} else if id.Value.GetExpressionType() == MATH_EXP {
			byteCode = append(byteCode, id.Value.(MathExpNode).GenerateByteCode(context.Variables)...)
		}
		byteCode = append(byteCode, declareVariable(id.Ident.Value.Value, "int", context)...)
		return byteCode
	}
	log.Fatalf("error: unknown type")
	return nil
}

type VarReAssignment struct {
	Ident Identifier
	Value Expression
}

func (vra VarReAssignment) GetStatementType() string {
	return VARREASSIGNMENT
}

func (vra VarReAssignment) GenerateByteCode(context *GeneratorContext) []byte {
	byteCode := make([]byte, 0)
	variable, ok := context.Variables[vra.Ident.Value.Value]
	if !ok {
		log.Fatalf("error: cannot reassign undeclared variable '%v'", vra.Ident.Value.Value)
	}
	if variable.Type == "int" {
		if vra.Value.GetExpressionType() == MATH_EXP {
			byteCode = append(byteCode, vra.Value.(MathExpNode).GenerateByteCode(context.Variables)...)
		} else if vra.Value.GetExpressionType() == IDENTIFIER_EXP {
			tempVar, ok := context.Variables[vra.Value.(Identifier).Value.Value]
			if !ok {
				log.Fatalf("error: cannot use undeclared variable %v", vra.Value.(Identifier).Value.Value)
			}
			if tempVar.Type != variable.Type {
				log.Fatalf("error: cannot reassign a variable of type %v with a variable of type %v", tempVar.Type, variable.Type)
			}
			byteCode = append(byteCode, loadVariable(tempVar)...)
		}
	}
	byteCode = append(byteCode, storeVariable(variable)...)
	return byteCode
}

type VarAddToValue struct {
	Ident      Identifier
	ValueToAdd Expression
}

func (vatv VarAddToValue) GetStatementType() string {
	return VARADDTOVARIABLE
}

func (vatv VarAddToValue) GenerateByteCode(context *GeneratorContext) []byte {
	byteCode := make([]byte, 0)
	variable, ok := context.Variables[vatv.Ident.Value.Value]
	if !ok {
		log.Fatalf("error: cannot use undeclared variable %v", vatv.Ident.Value.Value)
	}
	if variable.Type == "int" {
		if vatv.ValueToAdd.GetExpressionType() == MATH_EXP {
			byteCode = append(byteCode, vatv.ValueToAdd.(MathExpNode).GenerateByteCode(context.Variables)...)
		} else if vatv.ValueToAdd.GetExpressionType() == IDENTIFIER_EXP {
			tempVar, ok := context.Variables[vatv.ValueToAdd.(Identifier).Value.Value]
			if !ok {
				log.Fatalf("error: cannot use undeclared variable %v", vatv.ValueToAdd.(Identifier).Value.Value)
			}
			if tempVar.Type != variable.Type {
				log.Fatalf("error: cannot add a variable of type %v to a variable of type %v", tempVar.Type, variable.Type)
			}
			byteCode = append(byteCode, loadVariable(tempVar)...)
		}
	}
	byteCode = append(byteCode, loadVariable(variable)...)
	byteCode = append(byteCode, instructions.IADD)
	byteCode = append(byteCode, storeVariable(variable)...)
	return byteCode
}

func storeVariable(variable Variable) []byte {
	byteCode := make([]byte, 0)
	inst, ok := instructions.Istores[int32(variable.VariableIndex)]
	if !ok {
		byteCode = append(byteCode, instructions.ISTORE)
		byteCode = append(byteCode, uint8(variable.VariableIndex))
	} else {
		byteCode = append(byteCode, inst)
	}
	return byteCode
}

func loadVariable(variable Variable) []byte {
	byteCode := make([]byte, 0)
	inst, ok := instructions.Iloads[int32(variable.VariableIndex)]
	if !ok {
		byteCode = append(byteCode, instructions.ILOAD)
		byteCode = append(byteCode, uint8(variable.VariableIndex))
	} else {
		byteCode = append(byteCode, inst)
	}
	return byteCode
}

// um imm scope deklarierte variablen zu löschen
// den unterschied der in var map befindlichen eintrge feststellen
// alles was davor noch nicht da war löschen
type Scope struct {
	Statements []Statement
}

type FunctionArgument struct {
	Name string
	Type string
}

func (fa FunctionArgument) GetExpressionType() string {
	return FUNCTIONARG
}

type FunctionDefinition struct {
	Name       string
	ReturnType string
	Args       []FunctionArgument
	Scope      Scope
}

type Function struct {
	ReturnType string
	Args       []FunctionArgument
}

func (fd FunctionDefinition) GetStatementType() string {
	return FUNCDEF
}

func (fd FunctionDefinition) GenerateByteCode(context *GeneratorContext) []byte {
	byteCode := make([]byte, 0)
	variables := make(map[string]string, 0)
	for k := range context.Variables {
		variables[k] = k
	}
	for _, arg := range fd.Args {
		//byteCode = append(byteCode, declareVariable(arg.Name.Value.Value, arg.Type, context)...)
		context.Variables[arg.Name] = Variable{VariableIndex: *context.MaxLocals, Type: arg.Type}
		*context.MaxLocals++
	}
	for _, stmt := range fd.Scope.Statements {
		byteCode = append(byteCode, stmt.GenerateByteCode(context)...)
	}
	for k := range context.Variables {
		if _, ok := variables[k]; !ok {
			delete(context.Variables, k)
		}
	}
	context.Class.AddMethod(fd.Name, fd.ReturnType, byteCode, uint16(*context.MaxLocals))
	*context.MaxLocals = 0
	return byteCode
}

func declareVariable(name, typ string, context *GeneratorContext) []byte {
	byteCode := make([]byte, 0)
	inst, ok := instructions.Istores[int32(*context.MaxLocals)]
	if !ok {
		byteCode = append(byteCode, instructions.ISTORE)
		byteCode = append(byteCode, uint8(*context.MaxLocals))
		context.Variables[name] = Variable{VariableIndex: *context.MaxLocals, Type: typ}
		*context.MaxLocals++
	} else {
		byteCode = append(byteCode, inst)
		context.Variables[name] = Variable{VariableIndex: *context.MaxLocals, Type: typ}
		*context.MaxLocals++
	}
	return byteCode
}

type FunctionCall struct {
	CalledFunctionName string
	Arguments          []Expression
}

func (fc FunctionCall) GetStatementType() string {
	return FUNCTIONCALL
}

func (fc FunctionCall) GenerateByteCode(context *GeneratorContext) []byte {
	byteCode := make([]byte, 0)
	fun, ok := discoveredFunctions[fc.CalledFunctionName]
	if !ok {
		log.Fatalf("error: cannot call undefined function %v", fc.CalledFunctionName)
	}
	if len(fun.Args) != len(fc.Arguments) {
		log.Fatalf("error: not enough/too many arguments to call function %v", fc.CalledFunctionName)
	}
	for index, arg := range fc.Arguments {
		if arg.GetExpressionType() == MATH_EXP {
			if fun.Args[index].Type != "int" {
				log.Fatalf("error: expected argument of type %v", fun.Args[index].Type)
			}
			byteCode = append(byteCode, arg.(MathExpNode).GenerateByteCode(context.Variables)...)
		} else if arg.GetExpressionType() == IDENTIFIER_EXP {
			v, ok := context.Variables[arg.(Identifier).Value.Value]
			if !ok {
				log.Fatalf("error: cannot use undefined variable %v", arg.(Identifier).Value.Value)
			}
			if fun.Args[index].Type != v.Type {
				log.Fatalf("error: expected argument of type %v", fun.Args[index].Type)
			}
			byteCode = append(byteCode, loadVariable(v)...)
		}
	}
	byteCode = append(byteCode, instructions.INVOKEVIRTUAL)
	methodRefIndex := context.Class.AddMethodRef(fc.CalledFunctionName, fun.ReturnType, "Default")
	byteCode = binary.BigEndian.AppendUint16(byteCode, methodRefIndex)
	return byteCode
}

type Program struct {
	Statements []Statement
}
