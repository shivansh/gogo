// Package codegen implements routines for generating assembly code from IR.

package codegen

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/shivansh/gogo/src/ast"
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
func CodeGen(t tac.Tac, runtime bool) {
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
	// entryPoint determines whether the source program contains a main
	// function (entry point).
	entryPoint := false
	// globals stores the code for the global declarations. This code is
	// inserted at the beginning of the entry point, i.e. the main routine.
	globals := new(tac.TextSec)
	// globalLineNo stores the line number of global declarations.
	globalLineNo := make(map[int]bool)

	if !runtime {
		// Define the assembler directives for data and text.
		fmt.Fprintln(&ds.Stmts, "\t.data")
		fmt.Fprintln(&ts.Stmts, "\t.text")
		// Place the runtime code in the generated assembly.
		// Avoid generating runtime code when compiling runtime itself.
		if content, err := ioutil.ReadFile("src/runtime/runtime.asm"); err == nil {
			fmt.Fprintf(&ts.Stmts, "%s", content)
		} else {
			log.Fatal(err)
		}
	}

	// Make a single pass across the entire three-address code and collect
	// code for the global declarations. This code will be inserted in the
	// main routine, i.e. the entry point.
	funcScope := false
	for _, blk := range t {
		if len(blk.Stmts) == 0 {
			continue
		}
		switch blk.Stmts[0].Op {
		case tac.FUNC, tac.LABEL:
			// Since labels can only occur inside a function, we are
			// still in a function's scope. This handles multiple
			// occurrences of return statements within the same function
			// in the body of conditionals.
			// For reference see 'test/codegen/labels.go'.
			funcScope = true
		}
		for _, stmt := range blk.Stmts {
			switch stmt.Op {
			case tac.RET:
				funcScope = false

			case tac.EQ, tac.DECLInt:
				if funcScope {
					break
				}
				blk.InitRegDS()
				blk.InitHeap()
				// Update the next-use info for the given block.
				blk.EvalNextUseInfo()
				blk.GetReg(&stmt, globals, typeInfo)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[0].(type) {
				case tac.I32:
					fmt.Fprintf(&globals.Stmts, "\tli\t$%d, %d\t\t%s\n", blk.Adesc[stmt.Dst].Reg, v, comment)
					comment = "# global decl -> memory"
					fmt.Fprintf(&globals.Stmts, "\tsw\tt%d, %s\t\t%s\n", blk.Adesc[stmt.Dst].Reg, stmt.Dst, comment)
					globalLineNo[stmt.Line] = true
				case tac.Str:
					// It is not required to handle strings separately here
					// as the only side-effect of string declaration are
					// limited to data section and not text section.
				default:
					log.Fatal("CodeGen: unknown type %T\n", v)
				}
			}
		}
	}

	for _, blk := range t {
		blk.InitRegDS()
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

		blk.InitHeap()
		// Update the next-use info for the given block.
		blk.EvalNextUseInfo()

		// Update data section data structures. For this, make a single
		// pass through the entire three-address code and for each
		// assignment statement, update the DS for data section.
		for _, stmt := range blk.Stmts {
			if ds.Lookup[stmt.Dst] {
				continue
			}
			tab := "\t\t" // indentation for in-line comments.
			if len(stmt.Dst) > 6 {
				tab = "\t"
			}
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
			case tac.DECL:
				typeInfo[stmt.Dst] = types.ARR
				fmt.Fprintf(&ds.Stmts, "%s:%s.space\t%d\n", stmt.Dst, tab, 4*stmt.Src[0].IntVal())
			case tac.DECLSTR:
				typeInfo[stmt.Dst] = types.STR
				fmt.Fprintf(&ds.Stmts, "%s:%s.asciiz %s\n", stmt.Dst, tab, stmt.Src[0].StrVal())
			default:
				typeInfo[stmt.Dst] = types.INT
				fmt.Fprintf(&ds.Stmts, "%s:%s.word\t0\n", stmt.Dst, tab)
			}
			ds.Lookup[stmt.Dst] = true
		}

		for k, stmt := range blk.Stmts {
			switch stmt.Op {
			case tac.EQ, tac.DECLInt:
				// Avoid re-generating code for global declarations.
				if globalLineNo[stmt.Line] {
					continue
				}
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
						fmt.Fprintf(&ts.Stmts, "\tsw\t$s1, %d($%d)\t%s\n", 4*stmt.Src[1].IntVal(),
							blk.Adesc[stmt.Src[0].StrVal()].Reg, comment)
					case tac.Str:
						comment := "# variable -> array"
						fmt.Fprintf(&ts.Stmts, "\tsw\t$%d, %d($%d)\t%s\n", blk.Adesc[v.StrVal()].Reg,
							4*u.IntVal(), blk.Adesc[stmt.Src[0].StrVal()].Reg, comment)
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
						fmt.Fprintf(&ts.Stmts, "\tsw\t$%d, %s($s2)\t%s\n", blk.Adesc[v.StrVal()].Reg,
							stmt.Src[0].StrVal(), comment)
					default:
						log.Fatal("Codegen: unknown type %T\n", v)
					}
				}

			case tac.ADD:
				blk.GetReg(&stmt, ts, typeInfo)
				switch v := stmt.Src[1].(type) {
				case tac.I32:
					fmt.Fprintf(&ts.Stmts, "\taddi\t$%d, $%d, %s\n", blk.Adesc[stmt.Dst].Reg,
						blk.Adesc[stmt.Src[0].StrVal()].Reg, v.StrVal())
				case tac.Str:
					fmt.Fprintf(&ts.Stmts, "\tadd\t$%d, $%d, $%d\n", blk.Adesc[stmt.Dst].Reg,
						blk.Adesc[stmt.Src[0].StrVal()].Reg, blk.Adesc[v.StrVal()].Reg)
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
				if funcName == "main" {
					entryPoint = true
					fmt.Fprintf(&ts.Stmts, "\n\t.globl %s\n\t.ent %s\n", funcName, funcName)
				}
				fmt.Fprintf(&ts.Stmts, "%s:\n", stmt.Dst)
				if funcName != "main" {
					fmt.Fprintln(&ts.Stmts, "\taddi\t$sp, $sp, -4\n\tsw\t$ra, 0($sp)")
				} else {
					// Add code for global declarations.
					fmt.Fprintf(&ts.Stmts, "%s", globals.Stmts.String())
					globals.Stmts.Reset()
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
					// Topmost comment in IR goes in the topmost position in
					// the generated assembly.
					s := ds.Stmts.String()
					ds.Stmts.Reset()
					fmt.Fprintf(&ds.Stmts, "# %s\n\n%s", stmt.Dst, s)
				} else {
					fmt.Fprintf(&ts.Stmts, "\t# %s\n", stmt.Dst)
				}

			case tac.DECL, tac.DECLSTR:
				// Handled above in the first pass while updating data segment as
				// data segment is required to be updated in case of declarations.

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

	// Check if the entry point (main) has been encountered.
	if !entryPoint && ast.PkgName != "runtime" {
		log.Fatal("function main not defined\n")
	}

	fmt.Fprintln(&ds.Stmts, "")

	if _, err := ds.Stmts.WriteTo(os.Stdout); err != nil {
		log.Fatal(err)
	}
	if _, err := ts.Stmts.WriteTo(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
