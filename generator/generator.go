package generator

import (
	"compiler/classfile"
	"compiler/parser"
)

type Generator struct {
	programAsAST parser.Program
}

func NewGenerator(program parser.Program) Generator {
	return Generator{programAsAST: program}
}

func (g Generator) GenerateByteCode(class *classfile.Class) {
	maxLocals := 0
	var variables map[string]parser.Variable = make(map[string]parser.Variable)
	genContext := parser.GeneratorContext{Class: class, MaxLocals: &maxLocals, Variables: variables}
	for _, stmt := range g.programAsAST.Statements {
		stmt.GenerateByteCode(&genContext)
	}
}
