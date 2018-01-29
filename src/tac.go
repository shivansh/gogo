package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
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
	parentab *SymTab
}

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatalf("Usage: ./codegen ir-file")
	}

	tac := GenTAC(args[1])
	fmt.Println(tac[0].stmts[0].op)
}

// GenTAC generates the three-address code (in-memory) data structure
// from the input file. The format of each statement in the input file
// is a tuple of the form -
// 	<line-number, operation, destination-variable, source-variable(s)>
func GenTAC(irfile string) (tac TAC) {
	src, err := ioutil.ReadFile(irfile)
	if err != nil {
		log.Fatal(err)
	}

	var fxn *Fxn
	rgx, _ := regexp.Compile("(^[0-9]*$)") // regex for integers
	r := csv.NewReader(strings.NewReader(string(src)))

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// Sanitize the records
		for i := 0; i < len(record); i++ {
			record[i] = strings.TrimSpace(record[i])
		}

		switch record[1] {
		case "func":
			fxn = new(Fxn)
		case "ret":
			tac = append(tac, *fxn)
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
			stmt := Stmt{
				record[1],
				record[2],
				sv,
				nil, // not a block header
			}
			fxn.stmts = append(fxn.stmts, stmt)
		}
	}

	return
}
