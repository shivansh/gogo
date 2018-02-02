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
type LabelMap map[string]int

type Tac []Blk

type Blk struct {
	Stmts    []Stmt
	symtab   SymTab
	labelmap LabelMap
}

type Stmt struct {
	Op  string
	Dst string
	Src []SrcVar
}

type SrcVar struct {
	Typ string
	Val string
}

// NOTE: Placed here for the dummy register allocator GetReg
var counter int = -1

// GenTAC generates the three-address code (in-memory) data structure
// from the input file. The format of each statement in the input file
// is a tuple of the form -
// 	<line-number, operation, destination-variable, source-variable(s)>
func GenTAC(file *os.File) (tac Tac) {
	var blk *Blk = nil
	rgx, _ := regexp.Compile("(^[0-9]*$)") // natural numbers
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		record := strings.Split(scanner.Text(), ",")
		// Sanitize the records
		for i := 0; i < len(record); i++ {
			record[i] = strings.TrimSpace(record[i])
		}

		// A basic block starts:
		//	* at label instruction
		//	* after jump instruction
		// and ends:
		//	* before label instruction
		//	* at jump instruction
		switch record[1] {
		case "label":
			if blk != nil {
				tac = append(tac, *blk) // end the previous block
			}
			blk = new(Blk) // start a new block
			blk.symtab = make(SymTab)
			blk.labelmap = make(LabelMap)
			blk.labelmap[record[2]], _ = strconv.Atoi(record[0])
		case "jmp":
			// It is possible that the target of jump instruction
			// has not yet been encountered. Hence instead of trying
			// resolving it via a lookup, insert the target name itself
			// which can be resolved once the entire TAC is loaded.
			tac = append(tac, *blk) // end the previous block
			blk = new(Blk)          // start a new block
			blk.symtab = make(SymTab)
			blk.labelmap = make(LabelMap)
			fallthrough // move into next section to update blk.Src
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
			blk.Stmts = append(blk.Stmts,
				Stmt{record[1], record[2], sv})
		}
	}

	tac = append(tac, *blk) // push the last block

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}

// GetReg is (currently) a dummy register allocator which
// returns the index of the next free register.
func GetReg() int {
	counter++
	return counter
}
