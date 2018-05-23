// Package tac implements heuristics and data structures to generate the three
// address code intermediate representation from a source file.

package tac

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Addr struct {
	Reg int
	Mem int
}

// Stmt defines the structure of a single statement in three-address code form.
type Stmt struct {
	Line int      // line number where the statement is available
	Op   string   // operator
	Dst  string   // destination variable
	Src  []SrcVar // source variables
}

// Data section
type DataSec struct {
	// Stmts is a slice of statements which will be flushed into the data
	// section of the generated assembly file.
	Stmts []string
	// Lookup keeps track of all the variables currently available in the
	// the data section.
	Lookup map[string]bool
}

type TextSec struct {
	// Stmts is a slice of statements which will be flushed into the text
	// section of the generated assembly file.
	Stmts []interface{}
}

// Tac represents the three-address code for the entire source program.
type Tac []Blk

// GenTAC generates the three-address code (in-memory) data structure
// from the input file. The format of each statement in the input file
// is a tuple of the form -
// 	<line-number, operation, destination-variable, source-variable(s)>
//
// The three-address code is a collection of basic block data structures,
// which are identified while reading the IR file as per following rules -
// 	A basic block starts:
//		* at label instruction
//		* after jump instruction
// 	and ends:
//		* before label instruction
//		* at jump instruction
func GenTAC(file string) (tac Tac) {
	blk := new(Blk)
	line := 0
	re := regexp.MustCompile("(^-?[0-9]+$)") // regex for integers
	startNewBlock := false

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		record := strings.Split(scanner.Text(), ",")
		// Sanitize the records.
		for i := 0; i < len(record); i++ {
			record[i] = strings.TrimSpace(record[i])
		}
		switch record[0] {
		case LABEL:
			// label statement is part of the newly created block.
			blk, line = NewBlock(blk, &tac)
			blk.Stmts = append(blk.Stmts, Stmt{line, record[0], record[1], []SrcVar{}})
			line++

		case FUNC:
			// func statement is part of the newly created block.
			blk, line = NewBlock(blk, &tac)
			blk.Stmts = append(blk.Stmts, Stmt{line, record[0], record[1], []SrcVar{}})
			line++

		case JMP, BGT, BGE, BLT, BLE, BEQ, BNE:
			// Start a new block after updating the current block
			// with the jump statement.
			startNewBlock = true
			fallthrough // move into next section to update blk.Src

		default:
			// Prepare a slice of source variables.
			var sv []SrcVar
			for i := 2; i < len(record); i++ {
				if re.MatchString(record[i]) {
					v, err := strconv.Atoi(record[i])
					if err != nil {
						fmt.Println(record[i])
						log.Fatal(err)
					}
					sv = append(sv, I32(v))
				} else {
					sv = append(sv, Str(record[i]))
				}
			}
			blk.Stmts = append(blk.Stmts, Stmt{line, record[0], record[1], sv})
			line++
			if startNewBlock {
				blk, line = NewBlock(blk, &tac)
				startNewBlock = false
			}
		}
	}

	// Push the last allocated basic block.
	tac = append(tac, *blk)
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}

// PrintTAC pretty-prints the three-address code IR.
func (tac Tac) PrintTAC() {
	for _, blk := range tac {
		for _, stmt := range blk.Stmts {
			fmt.Printf("%v, %v, ", stmt.Op, stmt.Dst)
			for _, v := range stmt.Src {
				fmt.Printf("%v, ", v)
			}
			fmt.Println()
		}
		fmt.Println()
	}

}
