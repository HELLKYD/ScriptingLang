package parser

import (
	"compiler/classfile"
	"compiler/instructions"
	"compiler/tokenizer"
	"log"
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
	return "identifier"
}

type ReturnStatement struct {
	ReturnValue Expression
}

func (r ReturnStatement) GetStatementType() string {
	return "return"
}

func (r ReturnStatement) GenerateByteCode(context *GeneratorContext) []byte {
	byteCode := make([]byte, 0)
	switch r.ReturnValue.GetExpressionType() {
	case "mathExp":
		byteCode = append(byteCode, r.ReturnValue.(MathExpNode).GenerateByteCode(context.Variables)...)
		byteCode = append(byteCode, instructions.IRETURN)
	case "identifier":
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
	return "varDecl"
}

func (id VarDecl) GenerateByteCode(context *GeneratorContext) []byte {
	if id.Type.Value.Value == "int" {
		_, ok := context.Variables[id.Ident.Value.Value]
		if ok {
			log.Fatalf("error: cannot redeclare variable '%v'", id.Ident.Value.Value)
		}
		byteCode := make([]byte, 0)
		if id.Value.GetExpressionType() == "identifier" {
			variable, ok := context.Variables[id.Value.(Identifier).Value.Value]
			if !ok {
				log.Fatalf("error: cannot use undeclared variable '%v'", id.Value.(Identifier).Value.Value)
			}
			byteCode = append(byteCode, loadVariable(variable)...)
		} else if id.Value.GetExpressionType() == "mathExp" {
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
	return "varReAssignment"
}

func (vra VarReAssignment) GenerateByteCode(context *GeneratorContext) []byte {
	byteCode := make([]byte, 0)
	variable, ok := context.Variables[vra.Ident.Value.Value]
	if !ok {
		log.Fatalf("error: cannot reassign undeclared variable '%v'", vra.Ident.Value.Value)
	}
	if variable.Type == "int" {
		if vra.Value.GetExpressionType() == "mathExp" {
			byteCode = append(byteCode, vra.Value.(MathExpNode).GenerateByteCode(context.Variables)...)
		} else if vra.Value.GetExpressionType() == "identifier" {
			tempVar, ok := context.Variables[vra.Value.(Identifier).Value.Value]
			if !ok {
				log.Fatalf("error: cannot use undeclared variable %v", vra.Value.(Identifier).Value.Value)
			}
			if tempVar.Type != variable.Type {
				log.Fatalf("error: cannot add a variable of type %v to a variable of type %v", tempVar.Type, variable.Type)
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
	return "varAddToVariable"
}

func (vatv VarAddToValue) GenerateByteCode(context *GeneratorContext) []byte {
	byteCode := make([]byte, 0)
	variable, ok := context.Variables[vatv.Ident.Value.Value]
	if !ok {
		log.Fatalf("error: cannot use undeclared variable %v", vatv.Ident.Value.Value)
	}
	if variable.Type == "int" {
		if vatv.ValueToAdd.GetExpressionType() == "mathExp" {
			byteCode = append(byteCode, vatv.ValueToAdd.(MathExpNode).GenerateByteCode(context.Variables)...)
		} else if vatv.ValueToAdd.GetExpressionType() == "identifier" {
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
	Name Identifier
	Type string
}

func (fa FunctionArgument) GetExpressionType() string {
	return "functionArg"
}

type FunctionDefinition struct {
	Name       string
	ReturnType string
	Args       []FunctionArgument
	Scope      Scope
}

func (fd FunctionDefinition) GetStatementType() string {
	return "funcDef"
}

func (fd FunctionDefinition) GenerateByteCode(context *GeneratorContext) []byte {
	byteCode := make([]byte, 0)
	variables := make(map[string]string, 0)
	for k := range context.Variables {
		variables[k] = k
	}
	for _, arg := range fd.Args {
		byteCode = append(byteCode, declareVariable(arg.Name.Value.Value, arg.Type, context)...)
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

type Program struct {
	Statements []Statement
}
