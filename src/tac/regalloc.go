package tac

import (
	"container/heap"
	"fmt"
	"log"
	"strconv"

	"github.com/shivansh/gogo/src/types"
)

// RegLimit determines the upper bound on the number of free registers at any
// given instant supported by the concerned architecture (MIPS in this case).
const RegLimit = 32

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
func (blk Blk) GetReg(stmt *Stmt, ts *TextSec, typeInfo map[string]types.RegType) {
	// allocReg is a slice of all the register DS which are popped from the
	// heap and have been assigned a variable's data. These DS are updated
	// with the newly assigned variable's next-use info and after all the
	// variables (x,y,z) are assigned a register, all entities in allocReg
	// are pushed back into the heap. This ensures that the registers
	// belonging to source variables don't spill each other.
	var allocReg []*UseInfo
	var srcVars []string
	var lenSource int // number of source variables

	// Collect all "variables" available in stmt. Register allocation is
	// first done for the source variables and then for the destination
	// variable.
	for _, v := range stmt.Src {
		switch v := v.(type) {
		case Str:
			srcVars = append(srcVars, v.StrVal())
		}
	}
	switch stmt.Op {
	case BGT, BGE, BLT, BLE, BEQ, BNE, JMP:
		lenSource = len(srcVars) + 1
	default:
		srcVars = append(srcVars, stmt.Dst)
		lenSource = len(srcVars)
	}

	for k, v := range srcVars {
		if _, hasReg := blk.Adesc[v]; !hasReg {
			// Element with next-use farthest in future is popped.
			item := heap.Pop(&blk.Pq).(*UseInfo)
			reg, err := strconv.Atoi(item.Name)
			if err != nil {
				log.Fatal(err)
			}
			if entry, ok := blk.Rdesc[reg]; ok && typeInfo[entry.Name] != types.ARR && entry.Dirty {
				comment := fmt.Sprintf("# spilled %s, freed $%s", blk.Rdesc[reg].Name, item.Name)
				tab := "\t\t" // indentation for in-line comments
				if len(blk.Rdesc[reg].Name) > 3 {
					tab = "\t"
				}
				fmt.Fprintf(&ts.Stmts, "\tsw\t$%s, %s\n", item.Name, blk.Rdesc[reg].Name+tab+comment)
			}
			allocReg = append(allocReg, &UseInfo{strconv.Itoa(reg), blk.FindNextUse(stmt.Line, v)})
			delete(blk.Adesc, blk.Rdesc[reg].Name)
			delete(blk.Rdesc, reg)
			isDirty, isLoaded := false, false
			blk.Rdesc[reg] = RegDesc{v, isDirty, isLoaded}
			blk.Adesc[v] = Addr{reg, blk.Adesc[v].Mem}
			// Load the variable from memory.
			if k < lenSource-1 {
				if typeInfo[v] == types.ARR {
					fmt.Fprintf(&ts.Stmts, "\tla\t$%d, %s\n", reg, v)
				} else {
					tab := "\t\t" // indentation for in-line comments
					if len(v) > 3 {
						tab = "\t"
					}
					comment := fmt.Sprintf("# %s -> $%d", v, reg)
					fmt.Fprintf(&ts.Stmts, "\tlw\t$%d, %s%s\n", reg, v, tab+comment)
					blk.MarkLoaded(reg)
				}
			}
		}
	}

	// Push the popped items with updated priorities back into heap.
	for _, v := range allocReg {
		heap.Push(&blk.Pq, v)
	}

	// Check if any src variable is without a register. If there is, then
	// temporarily mark the lookup table corresponding to it to ensure that
	// the relevant statement is correctly inserted into the text segment
	// data structure. Once that is done, this entry will be deleted by the
	// caller (CodeGen) of GetReg().
	for i := 0; i < len(srcVars)-1; i++ {
		if _, ok := blk.Adesc[srcVars[i]]; !ok {
			blk.Adesc[srcVars[i]] = Addr{blk.Adesc[stmt.Dst].Reg, 0}
		}
	}
}
