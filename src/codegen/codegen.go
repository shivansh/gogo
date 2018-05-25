// Package codegen implements routines for generating assembly code from IR.

package codegen

import (
	"container/heap"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/shivansh/gogo/src/tac"
)

type Addr struct {
	// The register value is represented as an integer
	// and an equivalent representation in MIPS will be -
	//	$r  ; r is the value of reg
	// For a variable which are not stored in any register,
	// the value of reg will be -1 for it.
	reg int
	// The memory address is currently an integer and
	// an equivalent representation in MIPS will be -
	//	($m)  ; m is the value of mem
	// TODO: Representing offsets from a memory location.
	mem int
}

// CodeGen updates the data structures for text and data segments with the
// generated assembly code.
func CodeGen(t tac.Tac) {
	ts := new(tac.TextSec)
	ds := new(tac.DataSec)
	ds.Lookup = make(map[string]bool)
	// arrLookup keeps track of all the arrays declared.
	arrLookup := make(map[string]bool)
	funcName := ""
	callerSaved := []string{}
	tab := "" // indentation for in-line comments.

	// Define the assembler directives for data and text.
	ds.Stmts = append(ds.Stmts, "\t.data")
	ts.Stmts = append(ts.Stmts, "\t.text")

	for _, blk := range t {
		blk.Rdesc = make(map[int]string)
		blk.Adesc = make(map[string]tac.Addr)
		blk.Pq = make(tac.PriorityQueue, tac.RegLimit)
		blk.NextUseTab = make([][]tac.UseInfo, len(blk.Stmts), len(blk.Stmts))
		// jumpStmt stores the intructions for jump statements which are
		// responsible for terminating a basic block. These statements
		// are added to the text segment only after all the block variables
		// have been stored back into memory.
		jumpStmt := []string{}
		// exitStmt stores the instructions which terminate a function.
		exitStmt := ""
		branchOp := "" // branch operator
		// branchStmt is the statement wrt a jump/branch instruction.
		branchStmt := ""

		if len(blk.Stmts) > 0 && blk.Stmts[0].Op == tac.FUNC {
			funcName = blk.Stmts[0].Dst
		}

		// Initialize the priority-queue with all the available free
		// registers with their next-use set to infinity.
		// NOTE: Register $1 is reserved by assembler for pseudo
		// instructions and hence is not assigned to variables.
		for i := 0; i < tac.RegLimit; i++ {
			switch i {
			case 0, 1, 2, 4, 29, 31:
				// The following registers are not allocated -
				//   * $0 is not a valid register.
				//   * $1 is reserved by the assembler for
				//   * $2 ($v0) stores function results.
				//     pseudo instructions.
				//   * $v0 and $a0 are special registers.
				//   * $29 ($sp) stores the stack pointer.
				//   * $31 ($ra) stores the return address.
				// The nextuse of these registers is set to -∞.
				blk.Pq[i] = &tac.UseInfo{
					Name:    strconv.Itoa(i),
					Nextuse: tac.MinInt,
				}
			default:
				blk.Pq[i] = &tac.UseInfo{
					Name: strconv.Itoa(i),
					// A higher priority is given to registers
					// with lower index, resulting in a
					// deterministic allocation. In case all
					// the registers have their Nextuse value
					// initialized to MaxInt, Pop() returns
					// one non-deterministically.
					Nextuse: tac.MaxInt - i,
				}
			}
		}
		heap.Init(&blk.Pq)
		// Update the next-use info for the given block.
		blk.EvalNextUseInfo()

		// Update data section data structures. For this, make a single
		// pass through the entire three-address code and for each
		// assignment statement, update the DS for data section.
		for _, stmt := range blk.Stmts {
			switch stmt.Op {
			case tac.LABEL,
				tac.FUNC,
				tac.RET,
				tac.CALL,
				tac.CMT,
				tac.BGT,
				tac.BGE,
				tac.BLT,
				tac.BLE,
				tac.BEQ,
				tac.BNE,
				tac.JMP:
				break
			default:
				if len(stmt.Dst) >= 8 {
					tab = "\t"
				} else {
					tab = "\t\t"
				}
				if stmt.Op == tac.DECL && !ds.Lookup[stmt.Dst] {
					ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:%s.space\t%d", stmt.Dst, tab, 4*stmt.Src[0].IntVal()))
					ds.Lookup[stmt.Dst] = true
					arrLookup[stmt.Dst] = true
				} else if stmt.Op == tac.DECLSTR {
					ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:%s.asciiz %s", stmt.Dst, tab, stmt.Src[0].StrVal()))
					ds.Lookup[stmt.Dst] = true
				} else if !ds.Lookup[stmt.Dst] {
					ds.Lookup[stmt.Dst] = true
					ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:%s.word\t0", stmt.Dst, tab))
				}
			}
		}

		for k, stmt := range blk.Stmts {
			switch stmt.Op {
			case tac.EQ, tac.DECLInt:
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[0].(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli\t$%d, %d\t\t%s",
						blk.Adesc[stmt.Dst].Reg, v, comment))
				case tac.Str:
					if blk.Adesc[stmt.Dst].Reg < 10 || blk.Adesc[v.StrVal()].Reg < 10 {
						tab = "\t\t"
					} else {
						tab = "\t"
					}
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove\t$%d, $%d%s%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[v.StrVal()].Reg, tab, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case tac.FROM:
				blk.GetReg(&stmt, ts, arrLookup)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					comment := fmt.Sprintf("# variable <- array")
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tlw\t$%d, %d($%d)\t%s",
						blk.Adesc[stmt.Dst].Reg, 4*stmt.Src[1].IntVal(), blk.Adesc[stmt.Src[0].StrVal()].Reg, comment))
				case tac.Str:
					comment := fmt.Sprintf("# iterator *= 4")
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll\t$s2, $%d, 2\t%s", blk.Adesc[v.StrVal()].Reg, comment))
					comment = fmt.Sprintf("# variable <- array")
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tlw\t$%d, %s($s2)\t%s",
						blk.Adesc[stmt.Dst].Reg, stmt.Src[0].StrVal(), comment))
				}

			case tac.INTO:
				blk.GetReg(&stmt, ts, arrLookup)
				switch u := stmt.Src[1].(type) {
				case tac.I32:
					switch v := stmt.Src[2].(type) {
					case tac.I32:
						comment := fmt.Sprintf("# const index -> $s1")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli\t$s1, %d \t%s", v.IntVal(), comment))
						comment = fmt.Sprintf("# variable -> array")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw\t$s1, %d($%d)\t%s",
							4*stmt.Src[1].IntVal(), blk.Adesc[stmt.Src[0].StrVal()].Reg, comment))
					case tac.Str:
						comment := fmt.Sprintf("# variable -> array")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw\t$%d, %d($%d)\t%s",
							blk.Adesc[v.StrVal()].Reg, 4*u.IntVal(), blk.Adesc[stmt.Src[0].StrVal()].Reg, comment))
					default:
						log.Fatal("Unknown type %T\n", v)
					}
				case tac.Str:
					switch v := stmt.Src[2].(type) {
					case tac.I32:
						comment := fmt.Sprintf("# const index -> $s1")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli\t$s1, %d \t%s", v.IntVal(), comment))
						comment = fmt.Sprintf("# iterator *= 4")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll $s2, $%d, 2\t%s", blk.Adesc[u.StrVal()].Reg, comment))
						comment = fmt.Sprintf("# variable -> array")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw\t$s1, %s($s2)\t%s",
							stmt.Src[0].StrVal(), comment))
					case tac.Str:
						comment := fmt.Sprintf("# iterator *= 4")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll $s2, $%d, 2\t%s", blk.Adesc[u.StrVal()].Reg, comment))
						comment = fmt.Sprintf("# variable -> array")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw\t$%d, %s($s2)\t%s",
							blk.Adesc[v.StrVal()].Reg, stmt.Src[0].StrVal(), comment))
					default:
						log.Fatal("Unknown type %T\n", v)
					}
				}

			case tac.ADD:
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\taddi\t$%d, $%d, %s\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tadd\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case tac.OR:
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tor\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tor\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case tac.AND:
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tand\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tand\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case tac.NOR:
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tnor\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tnor\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case tac.XOR:
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\txor\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\txor\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case tac.NOT:
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tnot\t$%d, $%d\t%s",
					blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, comment))

			case tac.MUL:
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmul\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmul\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case tac.DIV:
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tdiv\t$%d, $%d, %s\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tdiv\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case tac.SUB:
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsub\t$%d, $%d, %s\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsub\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case tac.REM:
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\trem\t$%d, $%d, %s\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\trem\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case tac.BGT:
				branchOp = "bgt"
				goto addBranchInstr

			case tac.BGE:
				branchOp = "bge"
				goto addBranchInstr

			case tac.BLT:
				branchOp = "blt"
				goto addBranchInstr

			case tac.BLE:
				branchOp = "ble"
				goto addBranchInstr

			case tac.BEQ:
				branchOp = "beq"
				goto addBranchInstr

			case tac.BNE:
				branchOp = "bne"
				goto addBranchInstr

			case tac.RST: // right shift
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsrl\t$%d, $%d, %s\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsrl\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case tac.LST: // left shift
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case tac.LABEL:
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("%s:", stmt.Dst))

			case tac.FUNC:
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\n\t.globl %s\n\t.ent %s", funcName, funcName))
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("%s:", stmt.Dst))
				if funcName != "main" {
					ts.Stmts = append(ts.Stmts, "\taddi\t$sp, $sp, -4\n\tsw\t$ra, 0($sp)")
				}

			case tac.JMP:
				// Defer adding the jump statement (basic block terminator)
				// until the modified variables have been stored in memory.
				jumpStmt = append(jumpStmt, fmt.Sprintf("\tj\t%s", stmt.Dst))

			case tac.CALL:
				// Since range loop over maps are not deterministic, maintain
				// a slice of sorted keys to preserve ordering across runs.
				keys := []int{}
				for k := range blk.Rdesc {
					keys = append(keys, k)
				}
				sort.Ints(keys)
				for _, r := range keys {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw\t$%d, %s", r, blk.Rdesc[r]))
					// It is the responsibility of the caller to save
					// all the registers before the callee starts.
					callerSaved = append(callerSaved, fmt.Sprintf("\tlw\t$%d, %s", r, blk.Rdesc[r]))
				}
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tjal\t%s", stmt.Dst))
				ts.Stmts = append(ts.Stmts, callerSaved...)

			case tac.STORE:
				blk.GetReg(&stmt, ts, arrLookup)
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove\t$%d, $2", blk.Adesc[stmt.Dst].Reg))

			case tac.CMT:
				if stmt.Line == 0 {
					ds.Stmts = append([]string{fmt.Sprintf("# %s\n", stmt.Dst)}, ds.Stmts...)
				} else {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\t# %s", stmt.Dst))
				}

			case tac.RET:
				if funcName == "main" {
					exitStmt = "\tli\t$2, 10\n\tsyscall\n\t.end main"
				} else {
					exitStmt = fmt.Sprintf("\n\tlw\t$ra, 0($sp)\n\taddi\t$sp, $sp, 4\n\tjr\t$ra\n\t.end %s", funcName)
				}
				// Check if the variable which is to hold the return value has a register -
				// 	* if it does then move register's content to $2 ($v0)
				//	* else load value of that variable to $2 ($v0) from memory
				if len(stmt.Dst) > 0 {
					if _, ok := blk.Adesc[stmt.Dst]; ok {
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove\t$2, $%d", blk.Adesc[stmt.Dst].Reg))
					} else {
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tlw\t$2, %s", stmt.Dst))
					}
				}

			case tac.SCANINT:
				ts.Stmts = append(ts.Stmts, "\tli\t$2, 5\n\tsyscall")
				blk.GetReg(&stmt, ts, arrLookup)
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove\t$%d, $2", blk.Adesc[stmt.Dst].Reg))

			case tac.PRINTINT:
				ts.Stmts = append(ts.Stmts, "\tli\t$2, 1")
				switch v := stmt.Src[0].(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli\t$4, %d", v.IntVal()))
				case tac.Str:
					blk.GetReg(&stmt, ts, arrLookup)
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove\t$4, $%d", blk.Adesc[v.StrVal()].Reg))
				}
				ts.Stmts = append(ts.Stmts, "\tsyscall")

			case tac.PRINTSTR:
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli\t$2, 4\n\tla\t$4, %s\n\tsyscall", stmt.Dst))
			}

			goto storeLiveRegs

		addBranchInstr:
			blk.GetReg(&stmt, ts, arrLookup)
			switch v := stmt.Src[1].(type) {
			case tac.I32:
				branchStmt = fmt.Sprintf("\t%s\t$%d, %s, %s", branchOp,
					blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal(), stmt.Dst)
			case tac.Str:
				branchStmt = fmt.Sprintf("\t%s\t$%d, $%d, %s", branchOp,
					blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst)
			default:
				log.Fatal("Unknown type %T\n", v)
			}
			// A jump/branch statement marks the end of a basic block. As a result
			// these statements are collected in the variable `jumpStmt` and added
			// only after all the live registers are store back in memory. When a
			// branch statement follows immediately after a label statement, a new
			// basic block is not created (see commit 61a9bde). If this is the case,
			// add the corresponding jump/branch statement right away as it does
			// not represent the end of basic block.
			if k-1 >= 0 && blk.Stmts[k-1].Op == tac.LABEL {
				ts.Stmts = append(ts.Stmts, branchStmt)
			} else {
				jumpStmt = append(jumpStmt, branchStmt)
			}

		storeLiveRegs:
			// In case on of the src variable's register was allocated to dst in GetReg(),
			// the src variable's lookup entry was temporarily marked. Find that variable
			// if it exists and delete its entry. It should be noted that the chosen
			// variable shouldn't have the same name as that of dst.
			if _, ok := blk.Adesc[stmt.Dst]; ok && stmt.Op == tac.PRINTINT {
				for _, v := range stmt.Src {
					switch v := v.(type) {
					case tac.Str:
						if blk.Adesc[v.StrVal()].Reg == blk.Adesc[stmt.Dst].Reg && v.StrVal() != stmt.Dst {
							delete(blk.Adesc, v.StrVal())
						}
					}
				}
			}
		}

		// Store non-empty registers back into memory at the end of basic block.
		if len(blk.Rdesc) > 0 {
			// Since range loop over maps are not deterministic, maintain
			// a slice of sorted keys to preserve ordering across runs.
			keys := []int{}
			for k := range blk.Rdesc {
				keys = append(keys, k)
			}
			sort.Ints(keys)
			ts.Stmts = append(ts.Stmts, fmt.Sprintf("\t# Store variables back into memory"))
			for _, k := range keys {
				if !arrLookup[blk.Rdesc[k]] {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw\t$%d, %s", k, blk.Rdesc[k]))
				}
			}
		}
		ts.Stmts = append(ts.Stmts, jumpStmt...)
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
