package main

import (
	"log"
	"os"

	"gogo/src/asm"
	"gogo/src/gentoken"
	"gogo/src/parser"
	"gogo/src/tac"
)

// GenToken generates the tokens returned by lexer from the input program.
func GenToken(file string) {
	gentoken.PrintTokens(file)
}

// GenAsm generates the assembly code using the IR generated from the input program.
func GenAsm(file string) {
	asm.CodeGen(tac.GenTAC(file))
}

// GenHTML generates the rightmost derivations used in the bottom-up parsing
// and pretty-prints them in an HTML format.
func GenHTML(file string) {
	parser.RightmostDerivation(file)
}

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatalf("Usage: gogo <filename>")
	}
	parser.GenProductions(args[1])
}
