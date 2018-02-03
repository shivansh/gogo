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
		// Register descriptor:
		//	* Keeps track of what is currently in each register.
		//	* Initially all registers are empty.
		regDesc := make(map[int]string)
		// Address descriptor:
		//	* Keeps track of location where current value of the
		//	  name can be found at runtime.
		//	* The location can be either one or a set of -
		//		- register
		//		- memory address
		//		- stack (TODO)
		addrDesc := make(map[string]AddrDesc)

		// At the end of each basic block, all the registers are flushed
		// back to memory which means that they can be reused inside a
		// different basic block (it can also be the same basic block,
		// depending on the control flow). Hence at the beginning of
		// each basic block, reset the counter used to keep track of
		// "free" registers by the "dummy" register allocator.
		tac.Counter = 0

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

		for _, stmt := range blk.Stmts {
			switch stmt.Op {
			case "=":
				if stmt.Src[0].Typ == "int" {
					if addrDesc[stmt.Dst].reg == 0 {
						// Load variables from memory into registers
						addrIndex := tac.GetReg()
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tla $t%d, %s", addrIndex, stmt.Dst))
						regIndex := tac.GetReg()
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tlw $t%d, ($t%d)", regIndex, addrIndex))
						// Update lookup tables
						regDesc[regIndex] = stmt.Dst
						addrDesc[stmt.Dst] = AddrDesc{regIndex, addrIndex}
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli, $t%d, %s", regIndex, stmt.Src[0].Val))
					} else {
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli, $t%d, %s", addrDesc[stmt.Dst].reg, stmt.Src[0].Val))
					}
				} else {
					if addrDesc[stmt.Dst].reg == 0 {
						// Load variables from memory into registers
						addrIndex := tac.GetReg()
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tla $t%d, %s", addrIndex, stmt.Dst))
						regIndex := tac.GetReg()
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tlw $t%d, ($t%d)", regIndex, addrIndex))
						// Update lookup tables
						regDesc[regIndex] = stmt.Dst
						addrDesc[stmt.Dst] = AddrDesc{regIndex, addrIndex}
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove, $t%d, $t%d", regIndex, addrDesc[stmt.Src[0].Val].reg))
					} else {
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove, $t%d, $t%d", addrDesc[stmt.Dst].reg, addrDesc[stmt.Src[0].Val].reg))
					}
				}
			case "<":
				if stmt.Src[0].Typ == "int" {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tjlt, $t%d, %s, %s", addrDesc[stmt.Dst].reg, stmt.Src[0].Val, stmt.Src[1].Val))
				} else {
					ts.Stmts = append(ts.Stmts,
						fmt.Sprintf("\tjlt, $t%d, $t%d, %s", addrDesc[stmt.Dst].reg, addrDesc[stmt.Src[0].Val].reg, stmt.Src[1].Val))
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
		for v, _ := range ds.lookup {
			regIndex := addrDesc[v].reg
			addrIndex := addrDesc[v].mem
			// Only store the variables which were loaded in the first place
			if addrIndex > 0 {
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%d, ($t%d)", regIndex, addrIndex))
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
