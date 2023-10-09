package main

import (
	"compiler/classfile"
	"compiler/command"
	"compiler/generator"
	"compiler/parser"
	"compiler/tokenizer"
	"log"
	"os"
)

func main() {
	file, err := os.ReadFile(command.GetSourceFile())
	if err != nil {
		log.Fatalf("error: could not open file (%v)", err)
	}
	class := classfile.NewClass(command.GetSourceFile(), "base/Object")
	log.Println(string(file))
	tokens := tokenizer.NewTokenizer(string(file)).GetTokens()
	log.Println(tokens)
	program := parser.NewParser(tokens, class).ParseProgram()
	log.Println(program)
	generator.NewGenerator(program).GenerateByteCode(class)
	classfile := class.ConvertToBytes()
	log.Println(classfile)
	os.WriteFile(command.GetOutFile(), classfile, 0666)
}
