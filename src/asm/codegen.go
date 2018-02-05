package asm

import (
	"fmt"

	"gogo/src/tac"
)

// Data section
type DataSec struct {
	// Stmts is a slice of statements which will be flushed
	// into the data section of the generated assembly file.
	Stmts []string
	// lookup keeps track of all the variables currently
	// available in the data section.
	lookup map[string]bool
}

type TextSec struct {
	// Stmts is a slice of statements which will be flushed
	// into the text section of the generated assembly file.
	Stmts []string
}

type AddrDesc struct {
	// The register value is represented as an integer
	// and an equivalent representation in MIPS will be -
	//	$tr  ; where r is the value of reg
	// For a variable which are not stored in any register,
	// the value of reg will be -1 for it.
	reg int
	// The memory address is currently an integer and
	// an equivalent representation in MIPS will be -
	//	($tm)  ; m is the value of mem
	// TODO: Representing offsets from a memory location.
	mem int
}

func CodeGen(t tac.Tac) {
	var ds DataSec
	var ts TextSec
	ds.lookup = make(map[string]bool)

	// Define the assembler directives for data and text.
	ds.Stmts = append(ds.Stmts, "\t.data")
	ts.Stmts = append(ts.Stmts, "\t.text")

	for _, blk := range t {
		blk.Rdesc = make(map[int]string)
		blk.Adesc = make(map[string]tac.AddrDesc)

		// Update data section data structures. For this, make a single
		// pass through the entire three-address code and for each
		// assignment statement, update the DS for data section.
		for _, stmt := range blk.Stmts {
			if stmt.Op == "=" {
				if !ds.lookup[stmt.Dst] {
					ds.lookup[stmt.Dst] = true
					ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:\t.word\t0", stmt.Dst))
				}
				// TODO It should be made possible to identify the contents of a variable.
				// For e.g. strings should be defined as following in MIPS -
				// 	str:	.byte	'a','b'
			}
		}

		// Initialize all registers to be empty at the starting of basic block.
		for i := 1; i <= tac.RegLimit; i++ {
			blk.EmptyDesc[i] = true
		}

		for _, stmt := range blk.Stmts {
			switch stmt.Op {
			case "=":
				if stmt.Src[0].Typ == "int" {
					if blk.Adesc[stmt.Dst].Reg == 0 {
						// When loading a variable from memory, two registers are required -
						//	* The first register loads the memory address of the variable. This
						//	  address serves as a reference where the variable value (which will
						// 	  be loaded in a separate register in the next step) is stored back
						//	  when spilling or at the end of the basic block.
						// 	* The second register loads the value of the register from the
						//	  memory address loaded in the previous step.
						//
						// Load variable's memory address into a register.
						retReg, retMem, retVar, isSpilled := blk.GetReg(1, stmt.Dst, 2)
						if isSpilled {
							// The register was spilled.
							comment := fmt.Sprintf("; spilled %s and freed {$t%d, $t%d}", retVar, retReg, retMem)
							ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%d, ($t%d)\t\t%s", retReg, retMem, comment))
						}
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tla $t%d, %s", retReg, stmt.Dst))

						// Load variable value from the address loaded in the previous step.
						retReg, retMem, retVar, isSpilled = blk.GetReg(1, stmt.Dst, 1)
						if isSpilled {
							// The register was spilled.
							comment := fmt.Sprintf("; spilled %s and freed {$t%d, $t%d}", retVar, retReg, retMem)
							ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%d, ($t%d)\t\t%s", retReg, retMem, comment))
						}
						comment := fmt.Sprintf("; %s -> {reg: $t%d, mem: $t%d}", stmt.Dst, retReg, blk.Adesc[blk.Rdesc[retReg]].Mem)
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli, $t%d, %s\t\t%s", retReg, stmt.Src[0].Val, comment))
					} else {
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli, $t%d, %s", blk.Adesc[stmt.Dst].Reg, stmt.Src[0].Val))
					}
				} else {
					// TODO Handle mode 2
				}
			case "<":
				// TODO Handle the case when the argument variables are not in registers
				if stmt.Src[0].Typ == "int" {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tjlt, $t%d, %s, %s", blk.Adesc[stmt.Dst].Reg, stmt.Src[0].Val, stmt.Src[1].Val))
				} else {
					ts.Stmts = append(ts.Stmts,
						fmt.Sprintf("\tjlt, $t%d, $t%d, %s", blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].Val].Reg, stmt.Src[1].Val))
				}
			case "label":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("%s:", stmt.Dst))
			case "call":
				fallthrough
			case "jump":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tj %s", stmt.Dst))
			}
		}

		// Store variables back into memory for the previous basic block
		ts.Stmts = append(ts.Stmts, "\n\t; Store variables back into memory")

		// Only store the variables which were loaded in the first place
		for _, v := range blk.Adesc {
			if v.Mem > 0 {
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%d, ($t%d)", v.Reg, v.Mem))
			}
		}
	}

	ds.Stmts = append(ds.Stmts, "") // data section terminator

	for _, s := range ds.Stmts {
		fmt.Println(s)
	}
	for _, s := range ts.Stmts {
		fmt.Println(s)
	}
}
