package main

import (
	"log"
	"os"

	"gogo/src/asm"
	"gogo/src/gentoken"
	"gogo/src/parser"
	"gogo/src/tac"
)

// GenToken generates the tokens returned by lexer for the input program.
func GenToken(file string) {
	gentoken.PrintTokens(file)
}

// GenAsm generates the assembly code using the IR generated from the input program.
func GenAsm(file string) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	asm.CodeGen(tac.GenTAC(f))
}

// GenRightmostDerivation generates the rightmost derivations used in the bottom-up
// parsing and pretty-prints them in an HTML format.
func GenRightmostDerivation(file string) {
	parser.GenHTML(file)
}

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatalf("Usage: gogo <filename>")
	}
	GenRightmostDerivation(args[1])
}
