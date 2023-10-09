package command

import (
	"log"
	"os"
)

func GetSourceFile() string {
	if len(os.Args) <= 1 {
		log.Fatalf("error: expected input and output file")
	}
	allArgs := os.Args[1:]
	return allArgs[0]
}

func GetOutFile() string {
	allArgs := os.Args[1:]
	return allArgs[1]
}
