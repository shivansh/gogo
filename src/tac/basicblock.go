// This file defines the structure of a basic block.

package tac

import (
	"container/heap"
	"strconv"
)

// Blk represents the structure of a basic block.
type Blk struct {
	Stmts []Stmt
	// Address descriptor
	//	* Keeps track of location where current value of the
	//	  name can be found at compile time.
	//	* The location can be either one or a set of -
	//		- register
	//		- memory address
	//		- stack (TODO)
	Adesc map[string]Addr
	// Register descriptor
	//	* Keeps track of what is currently in each register.
	//	* Initially all registers are empty.
	Rdesc      map[int]RegDesc
	NextUseTab [][]UseInfo
	Pq         PriorityQueue
	child      Child
	pred       map[int]bool // immediate predecessors of the basic block
	dataflow   DataFlowInfo
}

// Child determines the index of the left and right child of a basic block in a
// flow graph.
type Child struct {
	left  int
	right int
}

// set represents the type of data sets used in data-flow analysis.
type set map[string]bool

// DataFlowInfo represents the data structures used in the data-flow analysis.
type DataFlowInfo struct {
	Gen  set
	Kill set
	In   set
	Out  set
}

type Label struct {
	// inRef determines the inward references to the current block. It stores
	// the indices of blocks which reference it. A block references another
	// block via a jump/branch statement.
	inRef   []int
	index   int // index of the basic block this label belongs to
	canDrop int
}

type labelinfotype map[string]*Label

// LabelInfo traverses across the tac DS and collects details about label
// statements.
func (tac Tac) LabelInfo() labelinfotype {
	// labelMap maps a label name with its relevant details.
	labelMap := make(map[string]*Label)
	for blkIndex, blk := range tac {
		if len(blk.Stmts) == 0 {
			continue
		}
		if op := blk.Stmts[0].Op; op == LABEL {
			labelName := blk.Stmts[0].Dst
			// Update inward references of all the blocks to which
			// the current block references.
			for _, stmt := range blk.Stmts {
				if IsBranchOp(stmt.Op) {
					if _, ok := labelMap[stmt.Dst]; !ok {
						labelMap[stmt.Dst] = &Label{}
					}
					labelMap[stmt.Dst].inRef = append(labelMap[stmt.Dst].inRef, blkIndex)
				}
			}

			// If the basic block contains only a jump statement, it
			// **might** be dropped in case another block references
			// it. Only if it is preceded by an unconditional jump,
			// it will be **definite** that the block can be dropped.
			canDrop := FALSE
			if len(blk.Stmts) == 2 && IsBranchOp(blk.Stmts[1].Op) {
				canDrop = MAYBE
			}

			// Avoid dropping fallthrough labels. Check if the last
			// statement of the previous block is a jump/branch.
			if blkIndex >= 1 {
				prevBlk := tac[blkIndex-1]
				index := len(prevBlk.Stmts) - 1
				if !IsBranchOp(prevBlk.Stmts[index].Op) {
					// The block is not preceded by an
					// unconditional jump, hence cannot be
					// dropped in any case.
					canDrop = FALSE
				}
			}

			// Update the current block details.
			if _, ok := labelMap[labelName]; !ok {
				labelMap[labelName] = &Label{[]int{}, blkIndex, canDrop}
			} else {
				labelMap[labelName].index = blkIndex
				labelMap[labelName].canDrop = canDrop
			}
		}
	}

	return labelMap
}

// InitBlock initializes the basic block data structure.
func InitBlock() *Blk {
	blk := new(Blk)
	blk.pred = make(map[int]bool)
	blk.dataflow.Gen = make(set)
	blk.dataflow.Kill = make(set)
	blk.dataflow.In = make(set)
	blk.dataflow.Out = make(set)
	return blk
}

// NewBlock creates a new basic block and initializes its line number to 0.
func (blk *Blk) NewBlock(tac *Tac) (*Blk, int) {
	if blk != nil && len(blk.Stmts) > 0 {
		*tac = append(*tac, *blk) // end the previous block
		// Create a new block only if the current one is not empty. If
		// it is empty, use it. This case arises when a jump statement
		// is encountered after a label statement. The new block created
		// by the jump statement stays empty as the label statement
		// creates a new block of its own.
		blk = InitBlock()
	}
	return blk, 0
}

// InitRegDS initializes the basic block data structures relevant to registers.
func (blk *Blk) InitRegDS() {
	blk.Rdesc = make(map[int]RegDesc)
	blk.Adesc = make(map[string]Addr)
	blk.Pq = make(PriorityQueue, RegLimit)
	blk.NextUseTab = make([][]UseInfo, len(blk.Stmts), len(blk.Stmts))
}

// InitHeap initializes the heap used during register allocation.
func (blk *Blk) InitHeap() {
	// Initialize the priority-queue with all the available free
	// registers with their next-use set to infinity.
	// NOTE: Register $1 is reserved by assembler for pseudo
	// instructions and hence is not assigned to variables.
	for i := 0; i < RegLimit; i++ {
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
			blk.Pq[i] = &UseInfo{
				Name:    strconv.Itoa(i),
				Nextuse: MinInt,
			}
		default:
			blk.Pq[i] = &UseInfo{
				Name: strconv.Itoa(i),
				// A higher priority is given to registers
				// with lower index, resulting in a
				// deterministic allocation. In case all
				// the registers have their Nextuse value
				// initialized to MaxInt, Pop() returns
				// one non-deterministically.
				Nextuse: MaxInt - i,
			}
		}
	}
	heap.Init(&blk.Pq)
}

// MarkDirty marks the given register as dirty. A register whose contents have
// been modified after being loaded from memory is marked dirty.
func (blk *Blk) MarkDirty(reg int) {
	isDirty := true
	blk.Rdesc[reg] = RegDesc{blk.Rdesc[reg].Name, isDirty, blk.Rdesc[reg].Loaded}
}

// UnmarkDirty marks the given register as free.
func (blk *Blk) UnmarkDirty(reg int) {
	isDirty := false
	blk.Rdesc[reg] = RegDesc{blk.Rdesc[reg].Name, isDirty, blk.Rdesc[reg].Loaded}
}

// IsDirty determines if the given register is dirty.
func (blk *Blk) IsDirty(reg int) bool {
	if entry, ok := blk.Rdesc[reg]; ok {
		return entry.Dirty
	} else {
		panic("IsDirty: register not allocated yet")
	}
}

// MarkLoaded is invoked after loading a variable from memory, and it marks the
// corresponding register's entry as loaded.
func (blk *Blk) MarkLoaded(reg int) {
	loaded := true
	blk.Rdesc[reg] = RegDesc{blk.Rdesc[reg].Name, blk.Rdesc[reg].Dirty, loaded}
}

// UnmarkLoaded unmarks the register entry as loaded.
func (blk *Blk) UnmarkLoaded(reg int) {
	loaded := false
	blk.Rdesc[reg] = RegDesc{blk.Rdesc[reg].Name, blk.Rdesc[reg].Dirty, loaded}
}

// IsLoaded determines whether the given register stores the value of the given
// variable and the value has been loaded from memory.
func (blk *Blk) IsLoaded(reg int, varName string) bool {
	if entry, ok := blk.Rdesc[reg]; ok {
		return entry.Name == varName && entry.Loaded
	} else {
		panic("IsLoaded: register not allocated yet")
	}
}
