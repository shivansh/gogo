package asm

import (
	"container/heap"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"gogo/src/tac"
)

type Addr struct {
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
	ds := new(tac.DataSec)
	ts := new(tac.TextSec)
	ds.Lookup = make(map[string]bool)
	funcName := ""
	exitStmt := ""
	re := regexp.MustCompile("(^-?[0-9]*$)") // integers

	// Define the assembler directives for data and text.
	ds.Stmts = append(ds.Stmts, "\t.data")
	ts.Stmts = append(ts.Stmts, "\t.text\n\t.globl main\n\t.ent main")

	for _, blk := range t {
		blk.Rdesc = make(map[int]string)
		blk.Adesc = make(map[string]tac.Addr)
		blk.Pq = make(tac.PriorityQueue, tac.RegLimit)
		blk.NextUseTab = make([][]tac.UseInfo, len(blk.Stmts), len(blk.Stmts))

		if len(blk.Stmts) > 0 && strings.Compare(blk.Stmts[0].Op, "func") == 0 {
			funcName = blk.Stmts[0].Dst
		}

		for i := 0; i < tac.RegLimit; i++ {
			blk.Pq[i] = &tac.UseInfo{
				Name:    strconv.Itoa(i + 1),
				Nextuse: tac.MaxInt,
			}
		}
		heap.Init(&blk.Pq)
		blk.EvalNextUseInfo()
		// Update data section data structures. For this, make a single
		// pass through the entire three-address code and for each
		// assignment statement, update the DS for data section.
		for _, stmt := range blk.Stmts {
			switch stmt.Op {
			case "label", "func", "ret":
				continue
			}
			if !ds.Lookup[stmt.Dst] {
				ds.Lookup[stmt.Dst] = true
				ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:\t.word\t0", stmt.Dst))
			}
			// TODO It should be made possible to identify the contents of a variable.
			// For e.g. strings should be defined as following in MIPS -
			// 	str:	.byte	'a','b'
		}

		for _, stmt := range blk.Stmts {
			switch stmt.Op {
			case "=":
				blk.GetReg(&stmt, ts)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[0].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli $t%d, %d\t\t%s",
						blk.Adesc[stmt.Dst].Reg, v, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove $t%d, $t%d\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "+":
				blk.GetReg(&stmt, ts)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\taddi $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()], v, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tadd $t%d, $t%d, $t%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "label":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("%s:", stmt.Dst))
			case "func":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("%s:", stmt.Dst))
			case "jump", "call":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tj %s", stmt.Dst))
			case "#":
				if stmt.Line == 0 {
					ds.Stmts = append([]string{fmt.Sprintf("# %s\n", stmt.Dst)}, ds.Stmts...)
				} else {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\t# %s", stmt.Dst))
				}
			case "ret":
				if strings.Compare(funcName, "main") == 0 {
					exitStmt = "\tli $v0, 10\n\tsyscall\n\t.end main"
				} else {
					exitStmt = ""
				}
			case "printInt":
				ts.Stmts = append(ts.Stmts, "\tli $v0, 1")
				if re.MatchString(stmt.Dst) {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli $a0, %s", stmt.Dst))
				} else {
					blk.GetReg(&stmt, ts)
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove $a0, $t%d", blk.Adesc[stmt.Dst].Reg))
				}
				ts.Stmts = append(ts.Stmts, "\tsyscall")
			case "scanInt":
				ts.Stmts = append(ts.Stmts, "\tli $v0, 5\n\tsyscall")
				blk.GetReg(&stmt, ts)
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove $t%d, $v0", blk.Adesc[stmt.Dst].Reg))
			}
		}

		// Store non-empty registers back into memory at the end of basic block.
		if len(blk.Rdesc) > 0 {
			ts.Stmts = append(ts.Stmts, fmt.Sprintf("\n\t# Store variables back into memory"))
			for k, v := range blk.Rdesc {
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%d, %s", k, v))
			}
		}
		if len(exitStmt) > 0 {
			ts.Stmts = append(ts.Stmts, exitStmt)
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
