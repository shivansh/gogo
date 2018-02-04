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
	// The register allocator requires a register index as an argument
	// which "might" be spilled in case there is no empty register left.
	// The argument (spillReg) should have a non-empty register.
	var spillReg int
	ds.lookup = make(map[string]bool)

	// Define the assembler directives for data and text.
	ds.Stmts = append(ds.Stmts, "\t.data")
	ts.Stmts = append(ts.Stmts, "\t.text")

	for _, blk := range t {
		// Register descriptor:
		//	* Keeps track of what is currently in each register.
		//	* Initially all registers are empty.
		blk.Rdesc = make(map[int]string)
		// Address descriptor:
		//	* Keeps track of location where current value of the
		//	  name can be found at runtime.
		//	* The location can be either one or a set of -
		//		- register
		//		- memory address
		//		- stack (TODO)
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
		for i := 1; i <= 4; i++ {
			blk.EmptyDesc[i] = true
		}

		for _, stmt := range blk.Stmts {
			switch stmt.Op {
			case "=":
				if stmt.Src[0].Typ == "int" {
					if blk.Adesc[stmt.Dst].Reg == 0 {
						// The following is a non-deterministic O(1) algorithm
						// for k, _ := range blk.Rdesc {
						// 	spillReg = k
						// 	break
						// }

						// The following is a deterministic O(RegLimit) algorithm.
						for i := 1; i <= tac.RegLimit; i++ {
							_, ok := blk.Rdesc[i]
							if ok {
								spillReg = i
								break
							}
						}
						addrIndex := blk.GetReg(spillReg)
						if addrIndex == spillReg {
							// The register needs to be spilled.
							comment := fmt.Sprintf("; spilled %s and freed {$t%d, $t%d}",
								blk.Rdesc[spillReg], spillReg, blk.Adesc[blk.Rdesc[spillReg]].Mem)
							ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%d, ($t%d)\t\t%s",
								spillReg, blk.Adesc[blk.Rdesc[spillReg]].Mem, comment))
							blk.EmptyDesc[blk.Adesc[blk.Rdesc[spillReg]].Mem] = true
							delete(blk.Adesc, blk.Rdesc[spillReg])
							delete(blk.Rdesc, spillReg)
						}

						// Load variables from memory into registers
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tla $t%d, %s", addrIndex, stmt.Dst))

						// The following is a non-deterministic O(1) algorithm
						// for k, _ := range blk.Rdesc {
						// 	spillReg = k
						// 	break
						// }

						// The following is a deterministic O(RegLimit) algorithm.
						for i := 1; i <= tac.RegLimit; i++ {
							_, ok := blk.Rdesc[i]
							if ok {
								spillReg = i
								break
							}
						}
						regIndex := blk.GetReg(spillReg)
						if regIndex == spillReg {
							// The register needs to be spilled.
							ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%d, ($t%d)", spillReg, blk.Adesc[blk.Rdesc[spillReg]].Mem))
							blk.EmptyDesc[blk.Adesc[blk.Rdesc[spillReg]].Mem] = true
							delete(blk.Rdesc, spillReg)
							delete(blk.Adesc, blk.Rdesc[spillReg])
						}
						// Update lookup tables
						blk.Rdesc[regIndex] = stmt.Dst
						blk.Adesc[stmt.Dst] = tac.AddrDesc{regIndex, addrIndex}
						comment := fmt.Sprintf("; %s -> {reg: $t%d, mem: $t%d}", blk.Rdesc[regIndex], regIndex, addrIndex)
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli, $t%d, %s\t\t%s", regIndex, stmt.Src[0].Val, comment))
					} else {
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli, $t%d, %s", blk.Adesc[stmt.Dst].Reg, stmt.Src[0].Val))
					}
				} else {
					if blk.Adesc[stmt.Dst].Reg == 0 {
						// By the current heuristic of register allocation, the register
						// which "might" be spilled is always first source variable.
						addrIndex := blk.GetReg(blk.Adesc[stmt.Src[0].Val].Reg)
						// Load variables from memory into registers
						addrIndex = blk.GetReg(blk.Adesc[stmt.Src[0].Val].Reg)
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tla $t%d, %s", addrIndex, stmt.Dst))
						regIndex := blk.GetReg(blk.Adesc[stmt.Src[0].Val].Reg)
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tlw $t%d, ($t%d)", regIndex, addrIndex))
						// Update lookup tables
						blk.Rdesc[regIndex] = stmt.Dst
						blk.Adesc[stmt.Dst] = tac.AddrDesc{regIndex, addrIndex}
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove, $t%d, $t%d", regIndex, blk.Adesc[stmt.Src[0].Val].Reg))
					} else {
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove, $t%d, $t%d", blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].Val].Reg))
					}
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
