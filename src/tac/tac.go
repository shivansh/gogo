package tac

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type SymTab map[string]*SrcVar // Symbol table

type Tac struct {
	Stmts    []Stmt
	symtab   SymTab
	labelmap map[string]int
}

type Stmt struct {
	Op  string
	dst string
	src []SrcVar
}

type SrcVar struct {
	typ string
	val string
}

// GenTAC generates the three-address code (in-memory) data structure
// from the input file. The format of each statement in the input file
// is a tuple of the form -
// 	<line-number, operation, destination-variable, source-variable(s)>
func GenTAC(file *os.File) (tac Tac) {
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
			tac.labelmap[record[2]], _ = strconv.Atoi(record[0])
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
			tac.Stmts = append(tac.Stmts,
				Stmt{record[1], record[2], sv})
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}
