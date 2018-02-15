package asm

import (
	"container/heap"
	"fmt"
	"log"
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
	as := make(map[string]bool)
	ds := new(tac.DataSec)
	ts := new(tac.TextSec)
	ds.Lookup = make(map[string]bool)
	funcName := ""
	callerSaved := []string{}

	// Define the assembler directives for data and text.
	ds.Stmts = append(ds.Stmts, "\t.data")
	ts.Stmts = append(ts.Stmts, "\t.text")

	for _, blk := range t {
		exitStmt := ""
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
			case "label", "func", "ret", "call", "#", "bgt", "bge", "blt", "ble", "beq", "bne", "j":
				break
			default:
				if strings.Compare(stmt.Op, "decl") == 0 && !ds.Lookup[stmt.Dst] {
					ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:\t.space\t%d", stmt.Dst, 4*stmt.Src[0].U.IntVal()))
					ds.Lookup[stmt.Dst] = true
					as[stmt.Dst] = true
					break
				}
				if strings.Compare(stmt.Op, "declStr") == 0 {
					ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:\t.asciiz %s", stmt.Dst, stmt.Src[0].U.StrVal()))
					ds.Lookup[stmt.Dst] = true
					break
				}
				if !ds.Lookup[stmt.Dst] {
					ds.Lookup[stmt.Dst] = true
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
				blk.GetReg(&stmt, ts, as)
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
			case "from":
				blk.GetReg(&stmt, ts, as)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					comment := fmt.Sprintf("#Assigning value to the variable from array")
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tlw $t%d, %d($t%d)\t\t%s",
						blk.Adesc[stmt.Dst].Reg, 4*stmt.Src[1].U.IntVal(), blk.Adesc[stmt.Src[0].U.StrVal()].Reg, comment))
				case tac.Str:
					comment := fmt.Sprintf("# Multiplying iterator by 4")
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll $s2, $t%d, 2\t%s", blk.Adesc[v.StrVal()].Reg, comment))
					comment = fmt.Sprintf("#Assigning value to the variable from array")
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tlw $t%d, %s($s2)\t\t%s",
						blk.Adesc[stmt.Dst].Reg, stmt.Src[0].U.StrVal(), comment))
				}
			case "into":
				blk.GetReg(&stmt, ts, as)
				switch stmt.Src[1].U.(type) {
				case tac.I32:
					switch stmt.Src[2].U.(type) {
					case tac.I32:
						comment := fmt.Sprintf("# Storing const index into $s1")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli $s1, %d \t%s", stmt.Src[2].U.IntVal(), comment))
						comment = fmt.Sprintf("#Assigning variable to the array")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $s1, %d($t%d)\t\t%s",
							4*stmt.Src[1].U.IntVal(), blk.Adesc[stmt.Src[0].U.StrVal()].Reg, comment))
					case tac.Str:
						comment := fmt.Sprintf("#Assigning variable to the array")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%d, %d($t%d)\t\t%s",
							blk.Adesc[stmt.Src[2].U.StrVal()].Reg, 4*stmt.Src[1].U.IntVal(), blk.Adesc[stmt.Src[0].U.StrVal()].Reg, comment))
					}
				case tac.Str:
					switch stmt.Src[2].U.(type) {
					case tac.I32:
						comment := fmt.Sprintf("# Storing const index into $s1")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli $s1, %d \t%s", stmt.Src[2].U.IntVal(), comment))
						comment = fmt.Sprintf("# Multiplying iterator by 4")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll $s2, $t%d, 2\t%s", blk.Adesc[stmt.Src[1].U.StrVal()].Reg, comment))
						comment = fmt.Sprintf("#Assigning variable to the array")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $s1, %s($s2)\t\t%s",
							stmt.Src[0].U.StrVal(), comment))

					case tac.Str:
						comment := fmt.Sprintf("# Multiplying iterator by 4")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll $s2, $t%d, 2\t%s", blk.Adesc[stmt.Src[1].U.StrVal()].Reg, comment))
						comment = fmt.Sprintf("#Assigning variable to the array")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%d, %s($s2)\t\t%s",
							blk.Adesc[stmt.Src[2].U.StrVal()].Reg, stmt.Src[0].U.StrVal(), comment))
					}
				}
			case "+":
				blk.GetReg(&stmt, ts, as)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\taddi $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tadd $t%d, $t%d, $t%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "*":
				blk.GetReg(&stmt, ts, as)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmul $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmul $t%d, $t%d, $t%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "/":
				blk.GetReg(&stmt, ts, as)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tdiv $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tdiv $t%d, $t%d, $t%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "-":
				blk.GetReg(&stmt, ts, as)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsub $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsub $t%d, $t%d, $t%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "rem":
				blk.GetReg(&stmt, ts, as)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\trem $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\trem $t%d, $t%d, $t%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "bgt":
				blk.GetReg(&stmt, ts, as)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbgt $t%d, %s, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), stmt.Dst, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbgt $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "bge":
				blk.GetReg(&stmt, ts, as)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbge $t%d, %s, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), stmt.Dst, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbge $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "blt":
				blk.GetReg(&stmt, ts, as)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tblt $t%d, %s, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), stmt.Dst, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tblt $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "ble":
				blk.GetReg(&stmt, ts, as)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tble $t%d, %s, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), stmt.Dst, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tble $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "beq":
				blk.GetReg(&stmt, ts, as)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbeq $t%d, %s, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), stmt.Dst, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbeq $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "bne":
				blk.GetReg(&stmt, ts, as)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbne $t%d, %s, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), stmt.Dst, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbne $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case ">>":
				blk.GetReg(&stmt, ts, as)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsrl $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsrl $t%d, $t%d, $t%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "<<":
				blk.GetReg(&stmt, ts, as)
				comment := fmt.Sprintf("# %s -> $t%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll $t%d, $t%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll $t%d, $t%d, $t%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}
			case "label":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("%s:", stmt.Dst))
			case "func":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\n\t.globl %s\n\t.ent %s", funcName, funcName))
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("%s:", stmt.Dst))
			case "j":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tj %s", stmt.Dst))
			case "call":
				for r, _ := range blk.Rdesc {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%d, %s", r, blk.Rdesc[r]))
					callerSaved = append(callerSaved, fmt.Sprintf("\tlw $t%d, %s", r, blk.Rdesc[r]))
				}
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tjal %s", stmt.Dst))
				ts.Stmts = append(ts.Stmts, callerSaved...)
			case "store":
				blk.GetReg(&stmt, ts, as)
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove $t%d, $v0", blk.Adesc[stmt.Dst].Reg))
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
					exitStmt = fmt.Sprintf("\n\tjr $ra\n\t.end %s", funcName)
				}
				// Check if the variable which is to hold the return value has a register. If it does
				// then move register's content to v0 else load value of that variable to v0 from memory.
				if len(stmt.Dst) > 0 {
					if _, ok := blk.Adesc[stmt.Dst]; ok {
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove $v0, $t%d", blk.Adesc[stmt.Dst].Reg))
					} else {
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tlw $v0, %s", stmt.Dst))
					}
				}

			case "printInt":
				ts.Stmts = append(ts.Stmts, "\tli $v0, 1")
				switch v := stmt.Src[0].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli $a0, %s", v.IntVal()))
				case tac.Str:
					blk.GetReg(&stmt, ts, as)
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove $a0, $t%d", blk.Adesc[v.StrVal()].Reg))
				}
				ts.Stmts = append(ts.Stmts, "\tsyscall")
			case "scanInt":
				ts.Stmts = append(ts.Stmts, "\tli $v0, 5\n\tsyscall")
				blk.GetReg(&stmt, ts, as)
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove $t%d, $v0", blk.Adesc[stmt.Dst].Reg))
			case "printStr":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli $v0, 4\n\tla $a0, %s\n\tsyscall", stmt.Dst))
			}

			// In case on of the src variable's register was allocated to dst in GetReg(),
			// the src variable's lookup entry was temporarily marked. Find that variable
			// if it exists and delete its entry.
			if _, ok := blk.Adesc[stmt.Dst]; ok && strings.Compare(stmt.Op, "printInt") != 0 {
				for _, v := range stmt.Src {
					switch v := v.U.(type) {
					case tac.Str:
						if blk.Adesc[v.StrVal()].Reg == blk.Adesc[stmt.Dst].Reg && strings.Compare(v.StrVal(), stmt.Dst) != 0 {
							// delete lookup entry of v
							delete(blk.Adesc, v.StrVal())
							break
						}
					}
				}
			}
		}

		// Store non-empty registers back into memory at the end of basic block.
		if len(blk.Rdesc) > 0 {
			ts.Stmts = append(ts.Stmts, fmt.Sprintf("\t# Store variables back into memory"))
			for k, v := range blk.Rdesc {
				if !as[v] {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%d, %s", k, v))
				}
			}
		}
		ts.Stmts = append(ts.Stmts, exitStmt)

	}
	ds.Stmts = append(ds.Stmts, "") // data section terminator

	for _, s := range ds.Stmts {
		fmt.Println(s)
	}
	for _, s := range ts.Stmts {
		fmt.Println(s)
	}
}
