package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type TAC []Fxn                 // Three-address code
type SymTab map[string]*SrcVar // Symbol table

type Fxn struct {
	stmts    []Stmt
	symtab   SymTab
	labelmap map[string]int
}

type Stmt struct {
	op  string
	dst string
	src []SrcVar
	blk *Block
}

type SrcVar struct {
	typ string
	val string
}

type Block struct {
	stmts    []Stmt
	symtab   SymTab
	parentab *SymTab // parent's symbol table
}

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatalf("Usage: ./codegen ir-file")
	}

	tac := GenTAC(args[1])
	fmt.Println(tac[0].stmts[0].op) // testcase for function statement
	// TODO: Add testcase for block statement
}

// GenTAC generates the three-address code (in-memory) data structure
// from the input file. The format of each statement in the input file
// is a tuple of the form -
// 	<line-number, operation, destination-variable, source-variable(s)>
func GenTAC(irfile string) (tac TAC) {
	file, err := os.Open(irfile)
	if err != nil {
		log.Fatal(err)
	}

	// "context" represents the context the program is in, governed
	// by the following values -
	// 	* 0: function context
	// 	* 1: block context
	// This is required because a statement in block and function is
	// handled in a similar manner when generating three-address code.
	// Thus, while updating the data structures, it should be known
	// whether it is a function that is being updated or is it a block.
	// TODO: Ideally, a union-like DS should be used.
	var context int
	var fxn *Fxn
	var blk *Block

	rgx, _ := regexp.Compile("(^[0-9]*$)") // regex for integers
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		record := strings.Split(scanner.Text(), ",")
		// Sanitize the records
		for i := 0; i < len(record); i++ {
			record[i] = strings.TrimSpace(record[i])
		}

		switch record[1] {
		case "func":
			fxn = new(Fxn)
			context = 0
		case "ret":
			tac = append(tac, *fxn)
		case "else":
			fallthrough
		case "elif":
			fallthrough
		case "if":
			context = 1
			blk = new(Block)
			// Prepare a slice of source variables.
			var sv []SrcVar
			for i := 4; i < len(record); i++ {
				var typ string = "string"
				if rgx.MatchString(record[i]) {
					typ = "Int"
				}
				sv = append(sv, SrcVar{typ, record[i]})
			}
			// First statement of block describes the conditional.
			blk.stmts = append(blk.stmts,
				Stmt{record[2], record[3], sv, nil})
			blk.parentab = &(fxn.symtab)
		case "end":
			// Add the block header to functions DS.
			fxn.stmts = append(fxn.stmts, Stmt{"", "", nil, blk})
			context = 0
		default:
			// Prepare a slice of source variables.
			var sv []SrcVar
			for i := 3; i < len(record); i++ {
				var typ string = "string"
				if rgx.MatchString(record[i]) {
					typ = "Int"
				}
				sv = append(sv, SrcVar{typ, record[i]})
			}

			if context == 0 {
				fxn.stmts = append(fxn.stmts,
					Stmt{record[1], record[2], sv, nil})
			} else {
				blk.stmts = append(blk.stmts,
					Stmt{record[1], record[2], sv, blk})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}
