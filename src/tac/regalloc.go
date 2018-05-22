package tac

import (
	"container/heap"
	"fmt"
	"strconv"
)

// RegLimit determines the upper bound on the number of free registers at any
// given instant supported by the concerned architecture (MIPS in this case).
// NOTE: Register $1 is reserved by assembler for pseudo instructions and hence
// is not assigned to variables.
const RegLimit = 31

// Register allocation
// ~~~~~~~~~~~~~~~~~~~
// Arguments:
//	* stmt: The allocator ensures that all the variables available in Stmt
//		object have been allocated a register.
//	* ts: If a register had to be spilled when GetReg() was called, the text
//	      segment should be updated with an equivalent "sw" instruction.
//
// GetReg handles all the side-effects induced due to register allocation -
//	* Updating lookup tables.
//	* Generating additional instructions resulting due to register spilling.
//
// A note on register spilling
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~
// A variable which doesn't have a next-use in the current basic block is
// spilled right away even if there are free registers available, resulting in
// one "sw" instruction. In case spilling was avoided and one of the free
// registers was used instead, that too would have resulted in one "sw"
// instruction at the end of the basic block.
func (blk Blk) GetReg(stmt *Stmt, ts *TextSec, arrLookup map[string]bool) {
	// allocReg is a slice of all the register DS which are popped from the
	// heap and have been assigned a variable's data. These DS are updated
	// with the newly assigned variable's next-use info and after all the
	// variables (x,y,z) are assigned a register, all entities in allocReg
	// are pushed back into the heap. This ensures that the registers
	// belonging to source variables don't spill each other.
	var allocReg []*UseInfo
	var srcVars []string
	var lenSource int // number of source variables
	var tab string    // indentation for in-line comments

	// Collect all "variables" available in stmt. Register allocation is
	// first done for the source variables and then for the destination
	// variable.
	for _, v := range stmt.Src {
		switch v := v.(type) {
		case Str:
			srcVars = append(srcVars, v.StrVal())
		case I32:
			srcVars = append(srcVars, v.StrVal())
		default:
			panic("GetReg: invalid type")
		}
	}
	switch stmt.Op {
	case BGT, BGE, BLT, BLE, BEQ, BNE, JMP:
		lenSource = len(srcVars) + 1
		break
	default:
		srcVars = append(srcVars, stmt.Dst)
		lenSource = len(srcVars)
	}

	for k, v := range srcVars {
		if _, hasReg := blk.Adesc[v]; !hasReg {
			// element with highest next-use is popped
			item := heap.Pop(&blk.Pq).(*UseInfo)
			reg, _ := strconv.Atoi(item.Name)
			if _, ok := blk.Rdesc[reg]; ok && !arrLookup[blk.Rdesc[reg]] {
				comment := fmt.Sprintf("# spilled %s, freed $%s", blk.Rdesc[reg], item.Name)
				if len(blk.Rdesc[reg]) >= 6 {
					tab = "\t"
				} else {
					tab = "\t\t"
				}
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw\t$%s, %s", item.Name, blk.Rdesc[reg]+tab+comment))
			}
			allocReg = append(allocReg, &UseInfo{strconv.Itoa(reg), blk.FindNextUse(stmt.Line, v)})
			delete(blk.Adesc, blk.Rdesc[reg])
			delete(blk.Rdesc, reg)
			blk.Rdesc[reg] = v
			blk.Adesc[v] = Addr{reg, blk.Adesc[v].Mem}
			if k < lenSource-1 {
				if !arrLookup[v] {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tlw\t$%d, %s", blk.Adesc[v].Reg, v))
				} else {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tla\t$%d, %s", blk.Adesc[v].Reg, v))
				}
			}
		}
	}

	// Push the popped items with updated priorities back into heap.
	for _, v := range allocReg {
		heap.Push(&blk.Pq, v)
	}

	// Check if any src variable is without a register. If there is,
	// then temporarily mark the lookup table corresponding to it to
	// ensure that the relevant statement is correctly inserted into
	// the text segment data structure. Once that is done, this entry
	// will be deleted by the caller.
	for i := 0; i < len(srcVars)-1; i++ {
		if _, ok := blk.Adesc[srcVars[i]]; !ok {
			blk.Adesc[srcVars[i]] = Addr{blk.Adesc[stmt.Dst].Reg, 0}
		}
	}
}
