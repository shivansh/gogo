package asm

import (
	"fmt"
	"gogo/src/tac"
	"strings"
)

type AddrDesc struct {
	// The register value is represented as an integer
	// and an equivalent representation in MIPS will be -
	//	$tr  ; r is the value of reg
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
	var ds tac.DataSec
	var ts tac.TextSec
	ds.Lookup = make(map[string]bool)

	// Define the assembler directives for data and text.
	ds.Stmts = append(ds.Stmts, "\t.data")
	ts.Stmts = append(ts.Stmts, "\t.text")

	for _, blk := range t {
		blk.Rdesc = make(map[int]string)
		blk.Adesc = make(map[string]tac.AddrDesc)
		blk.Table = make([]map[string]int, len(blk.Stmts), len(blk.Stmts))
		blk.Pq = make(tac.PriorityQueue, 0)
		// fmt.Println("don't print twice")
		blk.GetUseInfo()

		// for _, v := range blk.Table {
		// 	for v, lineno := range v {
		// 		fmt.Printf("name: %s ; nextuse: %d  ", v, lineno)
		// 	}
		// 	fmt.Println("\n")
		// }

		// Update data section data structures. For this, make a single
		// pass through the entire three-address code and for each
		// assignment statement, update the DS for data section.
		for _, stmt := range blk.Stmts {
			if !(stmt.Op == "label" || stmt.Op == "call") {
				if !ds.Lookup[stmt.Dst] {
					ds.Lookup[stmt.Dst] = true
					ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:\t.word\t0", stmt.Dst))
				}
				for _, src := range stmt.Src {
					if strings.Compare(src.Typ, "string") == 0 {
						if !ds.Lookup[src.Val] {
							ds.Lookup[src.Val] = true
							ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:\t.word\t0", src.Val))
						}
				    }
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
				blk.GetReg(stmt, &ts)
				comment := fmt.Sprintf("; %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				if strings.Compare(stmt.Src[0].Typ, "int") == 0 {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli $t%d, %s\t\t%s", blk.Adesc[stmt.Dst].Reg, stmt.Src[0].Val, comment))
				} else {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove $t%d, $t%d\t\t%s", blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].Val].Reg, comment))
				}
			case "+":
				blk.GetReg(stmt, &ts)
				comment := fmt.Sprintf("; %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				if strings.Compare(stmt.Src[1].Typ, "int") == 0 {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\taddi $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].Val].Reg, stmt.Src[1].Val, comment))
				} else {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tadd $t%d, $t%d, $t%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].Val].Reg, blk.Adesc[stmt.Src[1].Val].Reg, comment))
				}
			case "*":
				blk.GetReg(stmt, &ts)
				comment := fmt.Sprintf("; %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				if strings.Compare(stmt.Src[1].Typ, "int") == 0 {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmuli $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].Val].Reg, stmt.Src[1].Val, comment))
				} else {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmul $t%d, $t%d, $t%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].Val].Reg, blk.Adesc[stmt.Src[1].Val].Reg, comment))
				}
			case "label":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("%s:", stmt.Dst))
			case "call":
				fallthrough
			case "jump":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tj %s", stmt.Dst))
			}
			// fmt.Printf("Heap size codegen: %d\n", blk.Pq.Len())
		}

		// Store filled registers back into memory at the end of basic block.
		ts.Stmts = append(ts.Stmts, "\n\t; Store variables back into memory")
		for k, v := range blk.Adesc {
			if v.Reg > 0 {
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%d, %s", v.Reg, k))
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
