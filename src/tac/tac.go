package tac

import (
	"bufio"
	"container/heap"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Tac []Blk

type AddrDesc struct {
	Reg int
	Mem int
}

type Stmt struct {
	line int
	Op   string
	Dst  string
	Src  []SrcVar
}

type SrcVar struct {
	Typ string
	Val string
}

// Data section
type DataSec struct {
	// Stmts is a slice of statements which will be flushed
	// into the data section of the generated assembly file.
	Stmts []string
	// Lookup keeps track of all the variables currently
	// available in the data section.
	Lookup map[string]bool
}

type TextSec struct {
	// Stmts is a slice of statements which will be flushed
	// into the text section of the generated assembly file.
	Stmts []string
}

type UseInfo struct {
	Name    string // name of the register in case of priority queue and variable in case of table
	Nextuse int    // 1024 if dead
	Index   int
}

type PriorityQueue []*UseInfo

type Blk struct {
	Stmts []Stmt
	// Address descriptor:
	//	* Keeps track of location where current value of the
	//	  name can be found at runtime.
	//	* The location can be either one or a set of -
	//		- register
	//		- memory address
	//		- stack (TODO)
	Adesc map[string]AddrDesc
	// Register descriptor:
	//	* Keeps track of what is currently in each register.
	//	* Initially all registers are empty.
	Rdesc     map[int]string
	EmptyDesc map[int]bool
	Table     [][]UseInfo
	Pq        PriorityQueue
}

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, nextuse so we use greater than here.
	return pq[i].Nextuse > pq[j].Nextuse
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*UseInfo)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.Index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the nextuse and value of an Item in the queue.
func (pq *PriorityQueue) update(item *UseInfo, name string, nextuse int) {
	item.Name = name
	item.Nextuse = nextuse
	heap.Fix(pq, item.Index)
}

// RegLimit determines the upper bound on the number of free registers at any
// given instant supported by the concerned architecture (MIPS in this case).
// Currently, for testing purposes the value is set "too" low.
const RegLimit = 4

// Register allocator
// ~~~~~~~~~~~~~~~~~~
// Arguments:
//	* stmt: The allocator ensures that all the variables available in Stmt
//		object has been allocated a register.
//	* ts: If a register had to be spilled when GetReg() was called, the text
//	      segment should be updated with an equivalent statement (store-word).
//
// GetReg handles all the side-effects induced due to register allocation.
func (blk Blk) GetReg(stmt Stmt, ts *TextSec) {
	localVar := []string{stmt.Dst}
	for _, v := range stmt.Src {
		if strings.Compare(v.Typ, "string") == 0 {
			localVar = append(localVar, v.Val)
		}
	}

	regMap := make(map[int]bool)
	for _, v := range localVar {
		if _, ok := blk.Adesc[v]; ok {
			regMap[blk.Adesc[v].Reg] = true
		}
	}

	var i int
	for _, v := range localVar {
		if _, ok := blk.Adesc[v]; !ok {
			if len(blk.EmptyDesc) > 0 {
				for i = 1; i <= RegLimit; i++ {
					if blk.EmptyDesc[i] {
						// evaluate nextuse from table
						var nu int
						for _, l := range blk.Table[stmt.line] {
							if strings.Compare(v, l.Name) == 0 {
								nu = l.Nextuse
								break
							}
						}
						fmt.Printf("[Allocate]\treg: %d ; use: %d\n", i, nu)
						ui := &UseInfo{
							Name:    strconv.Itoa(i),
							Nextuse: nu,
						}
						blk.Pq.update(ui, ui.Name, nu)
						// Update lookup tables
						delete(blk.EmptyDesc, i)
						blk.Rdesc[i] = v
						blk.Adesc[v] = AddrDesc{i, blk.Adesc[v].Mem}
						break
					}
				}
			} else {
				// Spill a register.
				// TODO: Replacing RegLimit with len(blk.Rdesc) should work too.
				// get max element from heap
				item := heap.Pop(&blk.Pq).(*UseInfo)
				// Don't spill the source variable registers until dst is assigned
				// In case a src variable register was popped, store it in a slice and insert it at the end.
				var itemslice []*UseInfo
				regname, _ := strconv.Atoi(item.Name)
				validItemFound := false

				for !validItemFound {
					for _, l := range blk.Table[stmt.line] {
						if strings.Compare(l.Name, blk.Rdesc[regname]) == 0 {
							// collect the source variables and push them afterwards.
							itemslice = append(itemslice, item)
							fmt.Printf("[Avoid spill]\t%d\n", regname)
							item = heap.Pop(&blk.Pq).(*UseInfo)
							regname, _ = strconv.Atoi(item.Name)
							validItemFound = false
						} else {
							validItemFound = true
						}
					}
				}

				iname, _ := strconv.Atoi(item.Name)
				comment := fmt.Sprintf("; spilled %s, freed $t%s", blk.Rdesc[iname], item.Name)
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%s, %s\t\t%s", item.Name, blk.Rdesc[iname], comment))
				delete(blk.Adesc, blk.Rdesc[iname])
				delete(blk.Rdesc, iname)
				blk.Rdesc[iname] = v
				blk.Adesc[v] = AddrDesc{iname, blk.Adesc[v].Mem}

				var nu int
				for _, l := range blk.Table[stmt.line] {
					if strings.Compare(v, l.Name) == 0 {
						nu = l.Nextuse
						break
					}
				}
				// TODO push back into heap with updated priority
				ui := &UseInfo{
					Name:    strconv.Itoa(iname),
					Nextuse: nu,
				}
				heap.Push(&blk.Pq, ui)
				fmt.Printf("[NextUse]\treg: %d ; use: %d\n", iname, nu)
				for _, sliceitems := range itemslice {
					heap.Push(&blk.Pq, sliceitems)
				}
			}
		}
	}

	for _, v := range localVar {
		var nu int
		for _, y := range blk.Table[stmt.line] {
			if strings.Compare(y.Name, v) == 0 {
				nu = y.Nextuse
				break
			}
		}
		ui := &UseInfo{
			Name:    strconv.Itoa(blk.Adesc[v].Reg),
			Nextuse: nu,
		}
		blk.Pq.update(ui, ui.Name, nu)
		fmt.Printf("[Update]\treg: %d ; use: %d\n", blk.Adesc[v].Reg, nu)
	}
}

// Next use info evaluation
// ~~~~~~~~~~~~~~~~~~~~~~~~
func (blk Blk) GetUseInfo() {
	// table to track next use info
	liveness := make(map[string]int)

	for i := len(blk.Stmts) - 1; i >= 0; i-- {
		// TODO continue if op == label ;
		if strings.Compare(blk.Stmts[i].Op, "label") == 0 {
			continue
		}

		s := []string{blk.Stmts[i].Dst}
		for _, v := range blk.Stmts[i].Src {
			if strings.Compare(v.Typ, "string") == 0 {
				s = append(s, v.Val)
			}
		}
		for _, v := range s {
			// step 1 : attach to line i info abt vars
			if _, ok := liveness[v]; !ok {
				liveness[v] = 1024
			}
			blk.Table[i] = append(blk.Table[i], UseInfo{v, liveness[v], len(blk.Stmts) - 1 - i})
		}
		// step 2 : update liveness info
		for _, v := range blk.Stmts[i].Src {
			liveness[v.Val] = i
		}
		// step 3: kill dst
		liveness[s[0]] = 1024
	}
}

// GenTAC generates the three-address code (in-memory) data structure
// from the input file. The format of each statement in the input file
// is a tuple of the form -
// 	<line-number, operation, destination-variable, source-variable(s)>
func GenTAC(file *os.File) (tac Tac) {
	var blk *Blk = nil
	re := regexp.MustCompile("(^-?[0-9]*$)") // integers
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		record := strings.Split(scanner.Text(), ",")
		// Sanitize the records
		for i := 0; i < len(record); i++ {
			record[i] = strings.TrimSpace(record[i])
		}
		// A basic block starts:
		//	* at label instruction
		//	* after jump instruction
		// and ends:
		//	* before label instruction
		//	* at jump instruction
		switch record[1] {
		case "label":
			if blk != nil {
				blk.EmptyDesc = make(map[int]bool)
				tac = append(tac, *blk) // end the previous block
			}
			blk = new(Blk) // start a new block
			// label statement is the part of the newly created block
			line, _ := strconv.Atoi(record[0])
			blk.Stmts = append(blk.Stmts, Stmt{line, record[1], record[2], []SrcVar{}})
		case "jmp":
			blk.EmptyDesc = make(map[int]bool)
			tac = append(tac, *blk) // end the previous block
			blk = new(Blk)          // start a new block
			fallthrough             // move into next section to update blk.Src
		default:
			// Prepare a slice of source variables.
			var sv []SrcVar
			for i := 3; i < len(record); i++ {
				var typ string = "string"
				if re.MatchString(record[i]) {
					typ = "int"
				}
				sv = append(sv, SrcVar{typ, record[i]})
			}
			line, _ := strconv.Atoi(record[0])
			blk.Stmts = append(blk.Stmts, Stmt{line, record[1], record[2], sv})
		}
	}
	// Push the last allocated basic block
	blk.EmptyDesc = make(map[int]bool)
	tac = append(tac, *blk)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}
