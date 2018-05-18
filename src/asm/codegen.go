package asm

import (
	"container/heap"
	"fmt"
	"log"
	"strconv"
	"strings"

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

func CodeGen(t tac.Tac) {
	ts := new(tac.TextSec)
	ds := new(tac.DataSec)
	ds.Lookup = make(map[string]bool)
	// arrLookup keeps track of all the arrays declared.
	arrLookup := make(map[string]bool)
	funcName := ""
	callerSaved := []string{}
	// tab specifies the indentation to be given for in-line comments.
	tab := ""

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

		// Initialize the priority-queue with all the available free
		// registers with their next-use set to infinity.
		// NOTE: Register $1 is reserved by assembler for pseudo
		// instructions and hence is not assigned to variables.
		blk.Pq[0] = &tac.UseInfo{}
		for i := 1; i < tac.RegLimit; i++ {
			switch i {
			case 1, 2, 4:
				// The following registers are not allocated -
				//   * Register $1 is reserved by the assembler
				//     for pseudo instructions.
				//   * $v0 and $a0 are special registers.
				// The nextuse of these registers is set to -∞.
				blk.Pq[i] = &tac.UseInfo{
					Name:    strconv.Itoa(i),
					Nextuse: tac.MinInt,
				}
			default:
				blk.Pq[i] = &tac.UseInfo{
					Name:    strconv.Itoa(i),
					Nextuse: tac.MaxInt,
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
			case "label", "func", "ret", "call", "#", "bgt", "bge", "blt", "ble", "beq", "bne", "j":
				break
			default:
				if strings.Compare(stmt.Op, "decl") == 0 && !ds.Lookup[stmt.Dst] {
					ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:\t.space\t%d", stmt.Dst, 4*stmt.Src[0].U.IntVal()))
					ds.Lookup[stmt.Dst] = true
					arrLookup[stmt.Dst] = true
					break
				}
				if strings.Compare(stmt.Op, "declStr") == 0 {
					ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:\t.asciiz %s", stmt.Dst, stmt.Src[0].U.StrVal()))
					ds.Lookup[stmt.Dst] = true
					break
				}
				if !ds.Lookup[stmt.Dst] {
					ds.Lookup[stmt.Dst] = true
					if len(stmt.Dst) >= 8 {
						tab = "\t"
					} else {
						tab = "\t\t"
					}
					ds.Stmts = append(ds.Stmts, fmt.Sprintf("%s:%s.word\t0", stmt.Dst, tab))
				}
			}
		}

		for _, stmt := range blk.Stmts {
			switch stmt.Op {
			case "=":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[0].U.(type) {
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

			case "from":
				blk.GetReg(&stmt, ts, arrLookup)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					comment := fmt.Sprintf("# variable <- array")
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tlw\t$%d, %d($%d)\t\t%s",
						blk.Adesc[stmt.Dst].Reg, 4*stmt.Src[1].U.IntVal(), blk.Adesc[stmt.Src[0].U.StrVal()].Reg, comment))
				case tac.Str:
					comment := fmt.Sprintf("# iterator *= 4")
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll\t$s2, $%d, 2\t%s", blk.Adesc[v.StrVal()].Reg, comment))
					comment = fmt.Sprintf("# variable <- array")
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tlw\t$%d, %s($s2)\t\t%s",
						blk.Adesc[stmt.Dst].Reg, stmt.Src[0].U.StrVal(), comment))
				}

			case "into":
				blk.GetReg(&stmt, ts, arrLookup)
				switch u := stmt.Src[1].U.(type) {
				case tac.I32:
					switch v := stmt.Src[2].U.(type) {
					case tac.I32:
						comment := fmt.Sprintf("# const index -> $s1")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli\t$s1, %d \t%s", v.IntVal(), comment))
						comment = fmt.Sprintf("# variable -> array")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw\t$s1, %d($%d)\t\t%s",
							4*stmt.Src[1].U.IntVal(), blk.Adesc[stmt.Src[0].U.StrVal()].Reg, comment))
					case tac.Str:
						comment := fmt.Sprintf("# variable -> array")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw\t$%d, %d($%d)\t\t%s",
							blk.Adesc[v.StrVal()].Reg, 4*u.IntVal(), blk.Adesc[stmt.Src[0].U.StrVal()].Reg, comment))
					default:
						log.Fatal("Unknown type %T\n", v)
					}
				case tac.Str:
					switch v := stmt.Src[2].U.(type) {
					case tac.I32:
						comment := fmt.Sprintf("# const index -> $s1")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli\t$s1, %d \t%s", v.IntVal(), comment))
						comment = fmt.Sprintf("# iterator *= 4")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll $s2, $%d, 2\t%s", blk.Adesc[u.StrVal()].Reg, comment))
						comment = fmt.Sprintf("# variable -> array")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw\t$s1, %s($s2)\t\t%s",
							stmt.Src[0].U.StrVal(), comment))
					case tac.Str:
						comment := fmt.Sprintf("# iterator *= 4")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll $s2, $%d, 2\t%s", blk.Adesc[u.StrVal()].Reg, comment))
						comment = fmt.Sprintf("# variable -> array")
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw\t$%d, %s($s2)\t\t%s",
							blk.Adesc[v.StrVal()].Reg, stmt.Src[0].U.StrVal(), comment))
					default:
						log.Fatal("Unknown type %T\n", v)
					}
				}

			case "+":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\taddi\t$%d, $%d, %s\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tadd\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "or":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tor\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tor\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "and":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tand\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tand\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "nor":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tnor\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tnor\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "xor":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\txor\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\txor\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "not":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tnot\t$%d, $%d\t%s",
					blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, comment))

			case "*":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmul\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmul\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "/":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tdiv\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tdiv\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "-":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsub\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsub\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "rem":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\trem\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\trem\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "bgt":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbgt\t$%d, %s, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), stmt.Dst, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbgt\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "bge":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbge\t$%d, %s, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), stmt.Dst, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbge\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "blt":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tblt\t$%d, %s, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), stmt.Dst, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tblt\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "ble":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tble\t$%d, %s, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), stmt.Dst, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tble\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "beq":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbeq\t$%d, %s, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), stmt.Dst, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbeq\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "bne":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbne\t$%d, %s, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), stmt.Dst, comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tbne\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, stmt.Dst, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case ">>":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsrl\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsrl\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "<<":
				blk.GetReg(&stmt, ts, arrLookup)
				comment := fmt.Sprintf("# %s -> $%d", stmt.Dst, blk.Adesc[stmt.Dst].Reg)
				switch v := stmt.Src[1].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll\t$%d, $%d, %s\t\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, v.StrVal(), comment))
				case tac.Str:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsll\t$%d, $%d, $%d\t%s",
						blk.Adesc[stmt.Dst].Reg, blk.Adesc[stmt.Src[0].U.StrVal()].Reg, blk.Adesc[v.StrVal()].Reg, comment))
				default:
					log.Fatal("Unknown type %T\n", v)
				}

			case "label":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("%s:", stmt.Dst))

			case "func":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\t.globl %s\n\t.ent %s", funcName, funcName))
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("%s:", stmt.Dst))
				if funcName != "main" {
					ts.Stmts = append(ts.Stmts, "\taddi\t$sp, $sp, -4\n\tsw\t$ra, 0($sp)")
				}

			case "j":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tj\t%s", stmt.Dst))

			case "call":
				for r, _ := range blk.Rdesc {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw\t$%d, %s", r, blk.Rdesc[r]))
					// It is the responsibility of the caller to save all the registers
					// before the callee starts.
					callerSaved = append(callerSaved, fmt.Sprintf("\tlw\t$%d, %s", r, blk.Rdesc[r]))
				}
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tjal\t%s", stmt.Dst))
				ts.Stmts = append(ts.Stmts, callerSaved...)

			case "store":
				blk.GetReg(&stmt, ts, arrLookup)
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove\t$%d, $v0", blk.Adesc[stmt.Dst].Reg))

			case "#":
				if stmt.Line == 0 {
					ds.Stmts = append([]string{fmt.Sprintf("# %s\n", stmt.Dst)}, ds.Stmts...)
				} else {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\t# %s", stmt.Dst))
				}

			case "ret":
				if strings.Compare(funcName, "main") == 0 {
					exitStmt = "\tli\t$v0, 10\n\tsyscall\n\t.end main"
				} else {
					exitStmt = fmt.Sprintf("\n\tlw\t$ra, 0($sp)\n\taddi\t$sp, $sp, 4\n\tjr\t$ra\n\t.end %s", funcName)
				}
				// Check if the variable which is to hold the return value has a register -
				// 	* if it does then move register's content to $v0
				//	* else load value of that variable to $v0 from memory
				if len(stmt.Dst) > 0 {
					if _, ok := blk.Adesc[stmt.Dst]; ok {
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove\t$v0, $%d", blk.Adesc[stmt.Dst].Reg))
					} else {
						ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tlw\t$v0, %s", stmt.Dst))
					}
				}

			case "scanInt":
				ts.Stmts = append(ts.Stmts, "\tli\t$v0, 5\n\tsyscall")
				blk.GetReg(&stmt, ts, arrLookup)
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove\t$%d, $v0", blk.Adesc[stmt.Dst].Reg))

			case "printInt":
				ts.Stmts = append(ts.Stmts, "\tli\t$v0, 1")
				switch v := stmt.Src[0].U.(type) {
				case tac.I32:
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli\t$a0, %s", v.IntVal()))
				case tac.Str:
					blk.GetReg(&stmt, ts, arrLookup)
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tmove\t$a0, $%d", blk.Adesc[v.StrVal()].Reg))
				}
				ts.Stmts = append(ts.Stmts, "\tsyscall")

			case "printStr":
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tli\t$v0, 4\n\tla $a0, %s\n\tsyscall", stmt.Dst))
			}

			// In case on of the src variable's register was allocated to dst in GetReg(),
			// the src variable's lookup entry was temporarily marked. Find that variable
			// if it exists and delete its entry. It should be noted that the chosen
			// variable shouldn't have the same name as that of dst.
			if _, ok := blk.Adesc[stmt.Dst]; ok && strings.Compare(stmt.Op, "printInt") != 0 {
				for _, v := range stmt.Src {
					switch v := v.U.(type) {
					case tac.Str:
						if blk.Adesc[v.StrVal()].Reg == blk.Adesc[stmt.Dst].Reg && strings.Compare(v.StrVal(), stmt.Dst) != 0 {
							delete(blk.Adesc, v.StrVal())
						}
					}
				}
			}
		}

		// Store non-empty registers back into memory at the end of basic block.
		if len(blk.Rdesc) > 0 {
			ts.Stmts = append(ts.Stmts, fmt.Sprintf("\t# Store variables back into memory"))
			for k, v := range blk.Rdesc {
				if !arrLookup[v] {
					ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw\t$%d, %s", k, v))
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
