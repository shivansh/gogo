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
	EmptyDesc map[int]bool
}

const RegLimit = 4

func (blk Blk) GetReg(mode int, dst string, regType int) (retReg int, retMem int, retVar string, isSpilled bool) {
	// regType: 1 -> value register and 2 -> memory register
	if mode == 1 {
		// x = 1
		// Check if there is an empty register
		if len(blk.EmptyDesc) != 0 {
			isSpilled = false
			for i := 1; i <= RegLimit; i++ {
				if blk.EmptyDesc[i] {
					retReg = i
					delete(blk.EmptyDesc, i)
					break
				}
			}
			// Update lookup tables
			blk.Rdesc[retReg] = dst
			if regType == 1 {
				retMem = blk.Adesc[dst].Mem
				blk.Adesc[dst] = AddrDesc{retReg, blk.Adesc[dst].Mem}
			} else {
				retMem = retReg
				blk.Adesc[dst] = AddrDesc{blk.Adesc[dst].Reg, retReg}
			}
		} else {
			// No empty register hence spill
			isSpilled = true
			for i := 1; i <= RegLimit; i++ {
				if i == blk.Adesc[blk.Rdesc[i]].Mem {
					continue
				}
				if _, ok := blk.Rdesc[i]; ok {
					retReg = i
					retMem = blk.Adesc[blk.Rdesc[i]].Mem
					retVar = blk.Rdesc[i]

					// Update Rdesc and Adesc
					m := blk.Adesc[blk.Rdesc[i]]
					delete(blk.Adesc, blk.Rdesc[i])
					delete(blk.Rdesc, m.Reg)
					delete(blk.Rdesc, m.Mem)
					blk.EmptyDesc[m.Reg] = true
					blk.EmptyDesc[m.Mem] = true

					break
				}
			}
			// Update the newly acquired register in the lookup tables
			blk.Rdesc[retReg] = dst
			if regType == 1 {
				blk.Adesc[dst] = AddrDesc{retReg, blk.Adesc[dst].Mem}
			} else {
				blk.Adesc[dst] = AddrDesc{blk.Adesc[dst].Reg, retReg}
			}
		}
	} else {
		// x = y
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
				blk.EmptyDesc = make(map[int]bool)
				tac = append(tac, *blk) // end the previous block
			}
			blk = new(Blk) // start a new block
			// label statement is the part of the newly created block
			blk.Stmts = append(blk.Stmts, Stmt{record[1], record[2], []SrcVar{}})
		case "jmp":
			blk.EmptyDesc = make(map[int]bool)
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

	blk.EmptyDesc = make(map[int]bool)
	tac = append(tac, *blk) // push the last block

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}
