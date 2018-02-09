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

type Addr struct {
	Reg int
	Mem int
}

type Stmt struct {
	Line int
	Op   string
	Dst  string
	Src  []*SymInfo
}

type SymInfo struct {
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
	Nextuse int    // MaxInt if dead
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
	Adesc map[string]Addr
	// Register descriptor:
	//	* Keeps track of what is currently in each register.
	//	* Initially all registers are empty.
	Rdesc      map[int]string
	NextUseTab [][]UseInfo
	Pq         PriorityQueue
}

const (
	// RegLimit determines the upper bound on the number of free registers at any
	// given instant supported by the concerned architecture (MIPS in this case).
	// Currently, for testing purposes the value is set "too" low.
	RegLimit = 4
	// Variables which are dead have their next-use set to MaxInt.
	MaxInt = int(^uint(0) >> 1)
	// CommentLit is MIPS character sequence for comment initialization.
	CommentLit = ';' // can be '#' or ';'
)

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest nextuse.
	return pq[i].Nextuse > pq[j].Nextuse
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*UseInfo)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// Register allocator
// ~~~~~~~~~~~~~~~~~~
// Arguments:
//	* stmt: The allocator ensures that all the variables available in Stmt
//		object have been allocated a register.
//	* ts: If a register had to be spilled when GetReg() was called, the text
//	      segment should be updated with an equivalent statement (store-word).
//
// GetReg handles all the side-effects induced due to register allocation, namely -
//	* Updating lookup tables.
//	* Additional instructions resulting due to register spilling.
func (blk Blk) GetReg(stmt *Stmt, ts *TextSec) {
	// allocReg is a slice of all the register DS which are popped from the heap
	// and have been assigned a variable's data. These DS are updated with the
	// newly assigned variable's next-use info and after all the variables (x,y,z)
	// are assigned a register, all entities in allocReg are pushed back into the
	// heap. This ensures that the source variables' registers don't spill each other.
	var allocReg []*UseInfo
	var srcVars []string

	// Collect all "variables" available in stmt. Register allocation is first
	// done for the source variables and then for the destination variable.
	for _, v := range stmt.Src {
		if strings.Compare(v.Typ, "string") == 0 {
			srcVars = append(srcVars, v.Val)
		}
	}
	srcVars = append(srcVars, stmt.Dst)

	for k, v := range srcVars {
		if _, hasReg := blk.Adesc[v]; !hasReg {
			item := heap.Pop(&blk.Pq).(*UseInfo) // element with highest next-use
			reg, _ := strconv.Atoi(item.Name)
			if _, ok := blk.Rdesc[reg]; ok {
				comment := fmt.Sprintf("%c spilled %s, freed $t%s", CommentLit, blk.Rdesc[reg], item.Name)
				ts.Stmts = append(ts.Stmts, fmt.Sprintf("\tsw $t%s, %s\t\t%s", item.Name, blk.Rdesc[reg], comment))
			}
			allocReg = append(allocReg, &UseInfo{strconv.Itoa(reg), blk.FindNextUse(stmt.Line, v)})
			if k < 2 {
				delete(blk.Adesc, blk.Rdesc[reg])
				delete(blk.Rdesc, reg)
			}
			blk.Rdesc[reg] = v
			blk.Adesc[v] = Addr{reg, blk.Adesc[v].Mem}
		}
	}

	// Push the popped items with updated priorities back into heap.
	for _, v := range allocReg {
		heap.Push(&blk.Pq, v)
	}
}

// Next-use allocation heuristic
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Traverse the statements in a basic block from bottom-up, while updating
// the next use symbol table using the following algorithm (x = y op z) -
// 	Step 1: attach to i'th line in NextUseTab information currently
// 		in symbol table (nuSymTab) about variables x, y and z.
// 	Step 2: mark x as dead and no next use in nuSymTab.
// 	Step 3: mark y and z to be live and set their next use to i in nuSymTab.
func (blk Blk) EvalNextUseInfo() {
	// nuSymTab is a symbol table to track next use
	// information corresponding to all the variables.
	nuSymTab := make(map[string]int)
	for i := len(blk.Stmts) - 1; i >= 0; i-- {
		if strings.Compare(blk.Stmts[i].Op, "label") == 0 {
			continue
		}
		s := []string{blk.Stmts[i].Dst}
		for _, v := range blk.Stmts[i].Src {
			if strings.Compare(v.Typ, "string") == 0 {
				s = append(s, v.Val)
			}
		}
		// Step 1
		for _, v := range s {
			if _, ok := nuSymTab[v]; !ok {
				nuSymTab[v] = MaxInt
			}
			blk.NextUseTab[i] = append(blk.NextUseTab[i], UseInfo{v, nuSymTab[v]})
		}
		// Step 2
		nuSymTab[s[0]] = MaxInt
		// Step 3
		for _, v := range blk.Stmts[i].Src {
			nuSymTab[v.Val] = i
		}
	}
}

// GenTAC generates the three-address code (in-memory) data structure
// from the input file. The format of each statement in the input file
// is a tuple of the form -
// 	<line-number, operation, destination-variable, source-variable(s)>
func GenTAC(file *os.File) (tac Tac) {
	// var blk *Blk = nil
	blk := new(Blk)
	var line int
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
			// label statement is part of the newly created block.
			if blk != nil {
				tac = append(tac, *blk) // end the previous block
			}
			blk = new(Blk) // start a new block
			line = 0
			blk.Stmts = append(blk.Stmts, Stmt{line, record[1], record[2], []*SymInfo{}})
			line++
		case "jmp":
			tac = append(tac, *blk) // end the previous block
			blk = new(Blk)          // start a new block
			line = 0
			fallthrough // move into next section to update blk.Src
		default:
			// Prepare a slice of source variables.
			var sv []*SymInfo
			for i := 3; i < len(record); i++ {
				var typ string = "string"
				if re.MatchString(record[i]) {
					typ = "int"
				}
				sv = append(sv, &SymInfo{typ, record[i]})
			}
			blk.Stmts = append(blk.Stmts, Stmt{line, record[1], record[2], sv})
			line++
		}
	}
	// Push the last allocated basic block
	tac = append(tac, *blk)
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return
}

// FindNextUse returns the next use information corresponding to
// a variable "name" available in line number "line" of the table.
func (blk Blk) FindNextUse(line int, name string) int {
	for _, v := range blk.NextUseTab[line] {
		if strings.Compare(v.Name, name) == 0 {
			return v.Nextuse
		}
	}
	return MaxInt
}
