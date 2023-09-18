package main

import (
	"compiler/classfile"
	"compiler/generator"
	"compiler/parser"
	"compiler/tokenizer"
	"log"
	"os"
)

func main() {
	file, err := os.ReadFile("./testsource/main.e")
	if err != nil {
		log.Fatalf("error: could not open file (%v)", err)
	}
	class := classfile.NewClass("Main", "base/Object")
	log.Println(string(file))
	tokens := tokenizer.NewTokenizer(string(file)).GetTokens()
	log.Println(tokens)
	program := parser.NewParser(tokens).ParseProgram()
	log.Println(program)
	generator.NewGenerator(program).GenerateByteCode(class)
	classfile := class.ConvertToBytes()
	log.Println(classfile)
	os.WriteFile("./out.class", classfile, 0666)
}
