package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type SymTab map[string]*SrcVar // Symbol table

type Tac struct {
	stmts    []Stmt
	symtab   SymTab
	labelmap map[string]int
}

type Stmt struct {
	op  string
	dst string
	src []SrcVar
}

type SrcVar struct {
	typ string
	val string
}

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatalf("Usage: ./tac ir-file")
	}
	tac := GenTAC(args[1])
	fmt.Println(tac.stmts[9].op) // testcase for function statement
}

// GenTAC generates the three-address code (in-memory) data structure
// from the input file. The format of each statement in the input file
// is a tuple of the form -
// 	<line-number, operation, destination-variable, source-variable(s)>
func GenTAC(irfile string) (tac Tac) {
	file, err := os.Open(irfile)
	if err != nil {
		log.Fatal(err)
	}

	rgx, _ := regexp.Compile("(^[0-9]*$)") // natural numbers
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		record := strings.Split(scanner.Text(), ",")
		// Sanitize the records
		for i := 0; i < len(record); i++ {
			record[i] = strings.TrimSpace(record[i])
		}

		switch record[1] {
		case "label":
			tac.labelmap[record[2]], err = strconv.Atoi(record[0])
			if err != nil {
				log.Fatalf("Atoi")
			}
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
			tac.stmts = append(tac.stmts,
				Stmt{record[1], record[2], sv})
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}
