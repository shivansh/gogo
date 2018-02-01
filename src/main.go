package main

import (
	"fmt"
	"log"
	"os"

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
	fmt.Println(t.Stmts[0].Op)
}
