package tac

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"
)

// TODO Remove line numbers
type Tac []Blk

type AddrDesc struct {
	Reg int
	Mem int
}

type Blk struct {
	Stmts     []Stmt
	Adesc     map[string]AddrDesc
	Rdesc     map[int]string
	emptydesc map[int]bool
}

// NOTE: Placed here for the dummy register allocator GetReg
var Counter int

const regLimit = 4

// GetReg is (currently) a dummy register allocator which
// returns the index of the next free register.
func (blk Blk) GetReg(spillReg int) (retReg int) {
	if spillReg == 0 {
		retReg = 2
		blk.emptydesc[retReg] = false
		return
	}
	if len(blk.Rdesc) == regLimit {
		retReg = spillReg
		blk.emptydesc[spillReg] = true
	} else {
		for k, _ := range blk.emptydesc {
			if blk.emptydesc[k] {
				retReg = k
				blk.emptydesc[k] = false
				break
			}
		}
	}
	return
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

// GenTAC generates the three-address code (in-memory) data structure
// from the input file. The format of each statement in the input file
// is a tuple of the form -
// 	<line-number, operation, destination-variable, source-variable(s)>
func GenTAC(file *os.File) (tac Tac) {
	var blk *Blk = nil
	rgx, _ := regexp.Compile("(^-?[0-9]*$)") // integers
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
				blk.emptydesc = make(map[int]bool)
				tac = append(tac, *blk) // end the previous block
			}
			blk = new(Blk) // start a new block
			// label statement is the part of the newly created block
			blk.Stmts = append(blk.Stmts, Stmt{record[1], record[2], []SrcVar{}})
		case "jmp":
			blk.emptydesc = make(map[int]bool)
			tac = append(tac, *blk) // end the previous block
			blk = new(Blk)          // start a new block
			fallthrough             // move into next section to update blk.Src
		default:
			// Prepare a slice of source variables.
			var sv []SrcVar
			for i := 3; i < len(record); i++ {
				var typ string = "string"
				if rgx.MatchString(record[i]) {
					typ = "int"
				}
				sv = append(sv, SrcVar{typ, record[i]})
			}
			blk.Stmts = append(blk.Stmts, Stmt{record[1], record[2], sv})
		}
	}

	// Initialize emptydesc to have all elements as empty
	blk.emptydesc = make(map[int]bool)
	tac = append(tac, *blk) // push the last block

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}
