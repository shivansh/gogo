// Package tac implements heuristics and data structures to generate the three
// address code intermediate representation from a source file.

package tac

import (
	"bufio"
	"bytes"
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

type RegDesc struct {
	// Name determines the variable name whose value the register stores.
	Name string
	// Dirty determines whether the value of the variable stored by the
	// register has changed after loading from memory.
	Dirty bool
	// Loaded determines whether the variable whose value the register is
	// supposed to store has been loaded from memory.
	Loaded bool
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
	Stmts bytes.Buffer
	// Lookup keeps track of all the variables currently available in the
	// the data section.
	Lookup map[string]bool
}

type TextSec struct {
	// Stmts is a slice of statements which will be flushed into the text
	// section of the generated assembly file.
	Stmts bytes.Buffer
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
	blk, line := InitBlock(), 0
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
		case LABEL, FUNC:
			// label statement is part of the newly created block.
			blk, line = blk.NewBlock(&tac)
			blk.Stmts = append(blk.Stmts, Stmt{line, record[0], record[1], []SrcVar{}})
			line++

		case JMP, BGT, BGE, BLT, BLE, BEQ, BNE:
			// Start a new block after appending the jump statement
			// to the current block.
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
				blk, line = blk.NewBlock(&tac)
				startNewBlock = false
			}
		}
	}

	// Push the last allocated basic block.
	tac = append(tac, *blk)
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	labelMap := tac.LabelInfo()
	tac.EvalFlowGraph(labelMap)
	tac.EvalDataFlowSets()

	// Perform peephole optimization on the generated three-address code IR.
	return tac.PeepHole()
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
		fmt.Printf("Left child index: %d\n", blk.child.left)
		fmt.Printf("Right child index: %d\n\n", blk.child.right)
	}
	fmt.Println("--- [ End of TAC ] -----------------")
}

// FixLineNumbers fixes the line numbers in each basic block in case they get
// disturbed after the optimization passes finish. Correctness of line numbers
// is essential to ensure proper calculation of next-use information.
func (tac Tac) FixLineNumbers() {
	for i, blk := range tac {
		lineno := 0
		for j, stmt := range blk.Stmts {
			if stmt.Line != lineno {
				stmt.Line = lineno
			}
			tac[i].Stmts[j] = stmt
			lineno++
		}
	}
}
