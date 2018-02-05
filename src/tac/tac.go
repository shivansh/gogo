package tac

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type Tac []Blk

type AddrDesc struct {
	Reg int
	Mem int
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

// Data section
type DataSec struct {
	// Stmts is a slice of statements which will be flushed
	// into the data section of the generated assembly file.
	Stmts []string
	// Lookup keeps track of all the variables currently
	// available in the data section.
	Lookup map[string]bool
}

type TextSec struct {
	// Stmts is a slice of statements which will be flushed
	// into the text section of the generated assembly file.
	Stmts []string
}

type Blk struct {
	Stmts []Stmt
	// Address descriptor:
	//	* Keeps track of location where current value of the
	//	  name can be found at runtime.
	//	* The location can be either one or a set of -
	//		- register
	//		- memory address
	//		- stack (TODO)
	Adesc map[string]AddrDesc
	// Register descriptor:
	//	* Keeps track of what is currently in each register.
	//	* Initially all registers are empty.
	Rdesc     map[int]string
	EmptyDesc map[int]bool
}

// RegLimit determines the upper bound on the number of free registers at any
// given instant supported by the concerned architecture (MIPS in this case).
// Currently, for testing purposes the value is set "too" low.
const RegLimit = 4

// Register allocator
// ~~~~~~~~~~~~~~~~~~
// Arguments:
//	* stmt: The allocator ensures that all the variables available in Stmt
//		object has been allocated a register.
//	* ts: If a register had to be spilled when GetReg() was called, the text
//	      segment should be updated with an equivalent statement (store-word).
//
// GetReg handles all the side-effects induced due to register allocation.
func (blk Blk) GetReg(stmt Stmt, ts *TextSec) {
	localVar := []string{stmt.Dst}
	for _, v := range stmt.Src {
		if strings.Compare(v.Typ, "string") == 0 {
			localVar = append(localVar, v.Val)
		}
	}

	regMap := make(map[int]bool)
	for _, v := range localVar {
		if _, ok := blk.Adesc[v]; ok {
			regMap[blk.Adesc[v].Reg] = true
		}
	}

	var i int
	for _, v := range localVar {
		if _, ok := blk.Adesc[v]; !ok {
			if len(blk.EmptyDesc) > 0 {
				for i = 1; i <= RegLimit; i++ {
					if blk.EmptyDesc[i] {
						delete(blk.EmptyDesc, i)
						break
					}
				}
			} else {
				// Spill a register.
				// TODO: Replacing RegLimit with len(blk.Rdesc) should work too.
				for i = 1; i <= RegLimit; i++ {
					if !regMap[i] {
						comment := fmt.Sprintf("; spilled %s, freed $t%d", blk.Rdesc[i], i)
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%d, %s\t\t%s", i, blk.Rdesc[i], comment))
						delete(blk.Adesc, blk.Rdesc[i])
						delete(blk.Rdesc, i)
						break
					}
				}
			}
			blk.Rdesc[i] = v
			blk.Adesc[v] = AddrDesc{i, blk.Adesc[v].Mem}
			regMap[i] = true
		}
	}
}

// GenTAC generates the three-address code (in-memory) data structure
// from the input file. The format of each statement in the input file
// is a tuple of the form -
// 	<line-number, operation, destination-variable, source-variable(s)>
func GenTAC(file *os.File) (tac Tac) {
	var blk *Blk = nil
	re := regexp.MustCompile("(^-?[0-9]*$)") // integers
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
				if re.MatchString(record[i]) {
					typ = "int"
				}
				sv = append(sv, SrcVar{typ, record[i]})
			}
			blk.Stmts = append(blk.Stmts, Stmt{record[1], record[2], sv})
		}
	}
	// Push the last allocated basic block
	blk.EmptyDesc = make(map[int]bool)
	tac = append(tac, *blk)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}
