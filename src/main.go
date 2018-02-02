package main

import (
	"log"
	"os"

	"gogo/src/asm"
	"gogo/src/tac"
)

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatalf("Usage: gogo ir-file")
	}
	file, err := os.Open(args[1])
	if err != nil {
		log.Fatal(err)
	}
	t := tac.GenTAC(file)
	asm.CodeGen(t)
}
