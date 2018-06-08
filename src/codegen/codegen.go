// Package codegen implements routines for generating assembly code from IR.

package codegen

import (
	"container/heap"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"

	"github.com/shivansh/gogo/src/tac"
	"github.com/shivansh/gogo/src/types"
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
	// typeInfo keeps track of all the types of declared variables. This
	// information is useful during register allocation, as for example a
	// register storing an integer will have different load/store operations
	// than a register storing an array type.
	typeInfo := make(map[string]types.RegType)
	funcName := ""
	callerSaved := []string{}

	// Define the assembler directives for data and text.
	ds.Stmts = append(ds.Stmts, "\t.data")
	fmt.Fprintln(&ts.Stmts, "\t.text")

	for _, blk := range t {
		blk.Rdesc = make(map[int]tac.RegDesc)
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
		// dirtyRegCount determines the number of registers which have been
		// modified after loading their values from memory.
		dirtyRegCount := 0

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
				tab := "\t\t" // indentation for in-line comments.
				if len(stmt.Dst) >= 7 {
					tab = "\t"
				}
				if stmt.Op == tac.DECL && !ds.Lookup[stmt.Dst] {
					ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:%s.space\t%d", stmt.Dst, tab, 4*stmt.Src[0].IntVal()))
					ds.Lookup[stmt.Dst] = true
					typeInfo[stmt.Dst] = types.ARR
				} else if stmt.Op == tac.DECLSTR {
					ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:%s.asciiz %s", stmt.Dst, tab, stmt.Src[0].StrVal()))
					ds.Lookup[stmt.Dst] = true
					typeInfo[stmt.Dst] = types.STR
				} else if !ds.Lookup[stmt.Dst] {
					ds.Lookup[stmt.Dst] = true
					typeInfo[stmt.Dst] = types.INT
					ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:%s.word\t0", stmt.Dst, tab))
				}
			}
		}

		for k, stmt := range blk.Stmts {
			switch stmt.Op {
			case tac.EQ, tac.DECLInt:
				blk.GetReg(&stmt, ts, typeInfo)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[0].(type) {
				case tac.I32:
					fmt.Fprintf(&ts.Stmts, "\tli\t$%d, %d\t\t%s\n", blk.Adesc[stmt.Dst].Reg, v, comment)
				case tac.Str:
					tab := "\t" // indentation for in-line comments.
					if blk.Adesc[stmt.Dst].Reg < 10 || blk.Adesc[v.StrVal()].Reg < 10 {
						tab = "\t\t"
					}
					fmt.Fprintf(&ts.Stmts, "\tmove\t$%d, $%d%s%s\n", blk.Adesc[stmt.Dst].Reg,
						blk.Adesc[v.StrVal()].Reg, tab, comment)
				default:
					log.Fatal("Codegen: unknown type %T\n", v)
				}
				blk.MarkDirty(blk.Adesc[stmt.Dst].Reg)
				dirtyRegCount++

			case tac.FROM:
				blk.GetReg(&stmt, ts, typeInfo)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					comment := fmt.Sprintf("# variable <- array")
					fmt.Fprintf(&ts.Stmts, "\tlw\t$%d, %d($%d)\t%s\n",
						blk.Adesc[stmt.Dst].Reg, 4*stmt.Src[1].IntVal(), blk.Adesc[stmt.Src[0].StrVal()].Reg, comment)
				case tac.Str:
					comment := fmt.Sprintf("# iterator *= 4")
					fmt.Fprintf(&ts.Stmts, "\tsll\t$s2, $%d, 2\t%s\n", blk.Adesc[v.StrVal()].Reg, comment)
					comment = fmt.Sprintf("# variable <- array")
					fmt.Fprintf(&ts.Stmts, "\tlw\t$%d, %s($s2)\t%s\n",
						blk.Adesc[stmt.Dst].Reg, stmt.Src[0].StrVal(), comment)
				}
				blk.MarkDirty(blk.Adesc[stmt.Dst].Reg)
				dirtyRegCount++

			case tac.INTO:
				blk.GetReg(&stmt, ts, typeInfo)
				switch u := stmt.Src[1].(type) {
				case tac.I32:
					switch v := stmt.Src[2].(type) {
					case tac.I32:
						comment := "# const index -> $s1"
						fmt.Fprintf(&ts.Stmts, "\tli\t$s1, %d \t%s\n", v.IntVal(), comment)
						comment = "# variable -> array"
						fmt.Fprintf(&ts.Stmts, "\tsw\t$s1, %d($%d)\t%s\n",
							4*stmt.Src[1].IntVal(), blk.Adesc[stmt.Src[0].StrVal()].Reg, comment)
					case tac.Str:
						comment := "# variable -> array"
						fmt.Fprintf(&ts.Stmts, "\tsw\t$%d, %d($%d)\t%s\n",
							blk.Adesc[v.StrVal()].Reg, 4*u.IntVal(), blk.Adesc[stmt.Src[0].StrVal()].Reg, comment)
					default:
						log.Fatal("Codegen: unknown type %T\n", v)
					}
				case tac.Str:
					switch v := stmt.Src[2].(type) {
					case tac.I32:
						comment := "# const index -> $s1"
						fmt.Fprintf(&ts.Stmts, "\tli\t$s1, %d \t%s\n", v.IntVal(), comment)
						comment = "# iterator *= 4"
						fmt.Fprintf(&ts.Stmts, "\tsll $s2, $%d, 2\t%s\n", blk.Adesc[u.StrVal()].Reg, comment)
						comment = "# variable -> array"
						fmt.Fprintf(&ts.Stmts, "\tsw\t$s1, %s($s2)\t%s\n", stmt.Src[0].StrVal(), comment)
					case tac.Str:
						comment := "# iterator *= 4"
						fmt.Fprintf(&ts.Stmts, "\tsll $s2, $%d, 2\t%s\n", blk.Adesc[u.StrVal()].Reg, comment)
						comment = "# variable -> array"
						fmt.Fprintf(&ts.Stmts, "\tsw\t$%d, %s($s2)\t%s\n",
							blk.Adesc[v.StrVal()].Reg, stmt.Src[0].StrVal(), comment)
					default:
						log.Fatal("Codegen: unknown type %T\n", v)
					}
				}

			case tac.ADD:
				blk.GetReg(&stmt, ts, typeInfo)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					fmt.Fprintf(&ts.Stmts, "\taddi\t$%d, $%d, %s\n",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal())
				case tac.Str:
					fmt.Fprintf(&ts.Stmts, "\tadd\t$%d, $%d, $%d\n",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg)
				default:
					log.Fatal("Codegen: unknown type %T\n", v)
				}
				blk.MarkDirty(blk.Adesc[stmt.Dst].Reg)
				dirtyRegCount++

			case tac.SUB,
				tac.MUL,
				tac.DIV,
				tac.REM,
				tac.RST,
				tac.LST,
				tac.AND,
				tac.OR,
				tac.NOR,
				tac.XOR:
				blk.GetReg(&stmt, ts, typeInfo)
				op := ConvertOp(stmt.Op)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					fmt.Fprintf(&ts.Stmts, "\t%s\t$%d, $%d, %s\n", op, blk.Adesc[stmt.Dst].Reg,
						blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal())
				case tac.Str:
					fmt.Fprintf(&ts.Stmts, "\t%s\t$%d, $%d, $%d\n", op,
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg)
				default:
					log.Fatal("Codegen: unknown type %T\n", v)
				}
				blk.MarkDirty(blk.Adesc[stmt.Dst].Reg)
				dirtyRegCount++

			case tac.NOT:
				blk.GetReg(&stmt, ts, typeInfo)
				fmt.Fprintf(&ts.Stmts, "\tnot\t$%d, $%d\n",
					blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].StrVal()].Reg)
				blk.MarkDirty(blk.Adesc[stmt.Dst].Reg)
				dirtyRegCount++

			case tac.BEQ,
				tac.BNE,
				tac.BGT,
				tac.BGE,
				tac.BLT,
				tac.BLE:
				blk.GetReg(&stmt, ts, typeInfo)
				branchOp := ConvertOp(stmt.Op)
				branchStmt := ""
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					branchStmt = fmt.Sprintf("\t%s\t$%d, %s, %s", branchOp,
						blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal(), stmt.Dst)
				case tac.Str:
					branchStmt = fmt.Sprintf("\t%s\t$%d, $%d, %s", branchOp,
						blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst)
				default:
					log.Fatal("Codegen: unknown type %T\n", v)
				}
				// A jump/branch statement marks the end of a basic block. As a result
				// these statements are collected in the variable `jumpStmt` and added
				// only after all the live registers are store back in memory. When a
				// branch statement follows immediately after a label statement, a new
				// basic block is not created (see commit 61a9bde). If this is the case,
				// add the corresponding jump/branch statement right away as it does
				// not represent the end of basic block.
				if k >= 1 && blk.Stmts[k-1].Op == tac.LABEL {
					fmt.Fprintln(&ts.Stmts, branchStmt)
				} else {
					jumpStmt = append(jumpStmt, branchStmt)
				}

			case tac.LABEL:
				fmt.Fprintf(&ts.Stmts, "%s:\n", stmt.Dst)

			case tac.FUNC:
				fmt.Fprintf(&ts.Stmts, "\n\t.globl %s\n\t.ent %s\n", funcName, funcName)
				fmt.Fprintf(&ts.Stmts, "%s:\n", stmt.Dst)
				if funcName != "main" {
					fmt.Fprintln(&ts.Stmts, "\taddi\t$sp, $sp, -4\n\tsw\t$ra, 0($sp)")
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
				for _, reg := range keys {
					varName := blk.Rdesc[reg].Name
					fmt.Fprintf(&ts.Stmts, "\tsw\t$%d, %s\n", reg, varName)
					blk.UnmarkDirty(reg)
					dirtyRegCount--
					// It is the responsibility of the caller to save
					// all the registers before the callee starts.
					if !blk.IsLoaded(reg, varName) {
						callerSaved = append(callerSaved, fmt.Sprintf("\tlw\t$%d, %s\n", reg, varName))
						blk.MarkLoaded(reg)
					}
				}
				fmt.Fprintf(&ts.Stmts, "\tjal\t%s\n", stmt.Dst)
				// Load the caller-saved registers after returning from
				// the function call.
				for _, v := range callerSaved {
					fmt.Fprint(&ts.Stmts, v)
				}

			case tac.STORE:
				blk.GetReg(&stmt, ts, typeInfo)
				fmt.Fprintf(&ts.Stmts, "\tmove\t$%d, $2\n", blk.Adesc[stmt.Dst].Reg)
				blk.MarkDirty(blk.Adesc[stmt.Dst].Reg)
				dirtyRegCount++

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
						fmt.Fprintf(&ts.Stmts, "\tmove\t$2, $%d\n", blk.Adesc[stmt.Dst].Reg)
					} else {
						fmt.Fprintf(&ts.Stmts, "\tlw\t$2, %s\n", stmt.Dst)
					}
				}

			case tac.SCANINT:
				fmt.Fprintln(&ts.Stmts, "\tli\t$2, 5\n\tsyscall")
				blk.GetReg(&stmt, ts, typeInfo)
				fmt.Fprintf(&ts.Stmts, "\tmove\t$%d, $2\n", blk.Adesc[stmt.Dst].Reg)
				blk.MarkDirty(blk.Adesc[stmt.Dst].Reg)
				dirtyRegCount++

			case tac.PRINTINT:
				fmt.Fprintln(&ts.Stmts, "\tli\t$2, 1")
				switch v := stmt.Src[0].(type) {
				case tac.I32:
					fmt.Fprintf(&ts.Stmts, "\tli\t$4, %d\n", v.IntVal())
				case tac.Str:
					blk.GetReg(&stmt, ts, typeInfo)
					fmt.Fprintf(&ts.Stmts, "\tmove\t$4, $%d\n", blk.Adesc[v.StrVal()].Reg)
				}
				fmt.Fprintln(&ts.Stmts, "\tsyscall")

			case tac.PRINTSTR:
				fmt.Fprintf(&ts.Stmts, "\tli\t$2, 4\n\tla\t$4, %s\n\tsyscall\n", stmt.Dst)

			case tac.CMT:
				if stmt.Line == 0 {
					ds.Stmts = append([]string{fmt.Sprintf("# %s\n", stmt.Dst)}, ds.Stmts...)
				} else {
					fmt.Fprintf(&ts.Stmts, "\t# %s\n", stmt.Dst)
				}

			case tac.DECL, tac.DECLSTR:
				// Handled above in the first pass while updating data segment.

			default:
				log.Fatalf("Codegen: invalid operator %s\n", stmt.Op)
			}

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
			if dirtyRegCount > 0 {
				fmt.Fprintln(&ts.Stmts, "\t# Store dirty variables back into memory")
				for _, k := range keys {
					if typeInfo[blk.Rdesc[k].Name] != types.ARR && blk.Rdesc[k].Dirty {
						fmt.Fprintf(&ts.Stmts, "\tsw\t$%d, %s\n", k, blk.Rdesc[k].Name)
						blk.UnmarkDirty(k)
					}
				}
			}
		}
		for _, v := range jumpStmt {
			fmt.Fprintln(&ts.Stmts, v)
		}
		fmt.Fprintln(&ts.Stmts, exitStmt)
	}

	ds.Stmts = append(ds.Stmts, "") // data section terminator

	for _, s := range ds.Stmts {
		fmt.Println(s)
	}
	ts.Stmts.WriteTo(os.Stdout)
}
