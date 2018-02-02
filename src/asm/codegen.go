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

func CodeGen(t tac.Tac) {
	var ds DataSec
	ds.lookup = make(map[string]bool)
	for _, blk := range t {
		// Update data section data structures
		for _, stmt := range blk.Stmts {
			switch stmt.Op {
			case "=":
				if !ds.lookup[stmt.Dst] {
					ds.lookup[stmt.Dst] = true
					ds.Stmts = append(ds.Stmts,
						fmt.Sprintf("%s: .word", stmt.Dst))
				}
				for _, s := range stmt.Src {
					if !ds.lookup[s.Val] && s.Typ == "string" {
						ds.lookup[s.Val] = true
						ds.Stmts = append(ds.Stmts,
							fmt.Sprintf("%s: .word", s.Val))
					}
				}
			}
		}
	}

	for _, s := range ds.Stmts {
		fmt.Println(s)
	}
}
