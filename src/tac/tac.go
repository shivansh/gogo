package tac

import (
	"bufio"
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
//	* mode: The allocator operates in two modes -
//		- Mode 1: Handles expressions of the form "x = 1"
//		- Mode 2: Handles expressions of the form "x = y"
//	* dst: Destination variable name
//	* regType: There are two types of registers involved -
//		- Type 1: Value register, contain value of a variable
//		- Type 2: Address registers, contain memory address of a variable
//
// Return values:
//	* retReg: The register index which can now be used for allocation.
//	* isSpilled: Notifies if any register had to be spilled.
//
//	If a register was spilled when GetReg() was called, the following
//	values indicate back to the caller the steps to take to write back
//	the spilled values into memory.
//		* retMem: the (register) memory index of the spilled register
//		* retVar: the variable name of the spilled register
//
// GetReg handles all the side-effects induced due to register allocation.
func (blk Blk) GetReg(mode int, dst string, regType int) (retReg int, retMem int, retVar string, isSpilled bool) {
	if mode == 1 {
		// Check if there is an empty register.
		if len(blk.EmptyDesc) != 0 {
			isSpilled = false
			for i := 1; i <= RegLimit; i++ {
				if blk.EmptyDesc[i] {
					retReg = i
					delete(blk.EmptyDesc, i)
					break
				}
			}
			// Update the lookup tables with info of the newly acquired register.
			blk.Rdesc[retReg] = dst
			if regType == 1 {
				retMem = blk.Adesc[dst].Mem
				blk.Adesc[dst] = AddrDesc{retReg, blk.Adesc[dst].Mem}
			} else {
				retMem = retReg
				blk.Adesc[dst] = AddrDesc{blk.Adesc[dst].Reg, retReg}
			}
		} else {
			// No empty register left, spill a (non-empty) register with smallest index.
			isSpilled = true
			for i := 1; i <= RegLimit; i++ {
				if i == blk.Adesc[blk.Rdesc[i]].Mem {
					continue
				}
				if _, ok := blk.Rdesc[i]; ok {
					retReg = i
					retMem = blk.Adesc[blk.Rdesc[i]].Mem
					retVar = blk.Rdesc[i]
					// Update the lookup tables with info of the spilled register.
					m := blk.Adesc[blk.Rdesc[i]]
					delete(blk.Adesc, blk.Rdesc[i])
					delete(blk.Rdesc, m.Reg)
					delete(blk.Rdesc, m.Mem)
					blk.EmptyDesc[m.Reg] = true
					blk.EmptyDesc[m.Mem] = true
					break
				}
			}
			// Update the lookup tables with info of the newly acquired register.
			blk.Rdesc[retReg] = dst
			if regType == 1 {
				blk.Adesc[dst] = AddrDesc{retReg, blk.Adesc[dst].Mem}
			} else {
				blk.Adesc[dst] = AddrDesc{blk.Adesc[dst].Reg, retReg}
			}
		}
	} else {
		// TODO: Implement mode-2
	}

	return
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

	blk.EmptyDesc = make(map[int]bool)
	tac = append(tac, *blk) // push the last block

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}
