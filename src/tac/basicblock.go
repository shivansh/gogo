// This file defines the structure of a basic block.

package tac

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
}

// NewBlock creates a new basic block and initializes its line number to 0.
func NewBlock(blk *Blk, tac *Tac) (*Blk, int) {
	if blk != nil && len(blk.Stmts) > 0 {
		*tac = append(*tac, *blk) // end the previous block
		// Create a new block only if the current one is not empty. If
		// it is empty, use it. This case arises when a jump statement
		// is encountered after a label statement. The new block created
		// by the jump statement stays empty as the label statement
		// creates a new block of its own.
		blk = new(Blk)
	}
	return blk, 0
}

// MarkDirty marks the given register as dirty. A register whose contents have
// been modified after being loaded from memory is marked dirty.
func (blk Blk) MarkDirty(reg int) {
	isDirty := true
	blk.Rdesc[reg] = RegDesc{blk.Rdesc[reg].Name, isDirty, blk.Rdesc[reg].Loaded}
}

// UnmarkDirty marks the given register as free.
func (blk Blk) UnmarkDirty(reg int) {
	isDirty := false
	blk.Rdesc[reg] = RegDesc{blk.Rdesc[reg].Name, isDirty, blk.Rdesc[reg].Loaded}
}

// IsDirty determines if the given register is dirty.
func (blk Blk) IsDirty(reg int) bool {
	if entry, ok := blk.Rdesc[reg]; ok {
		return entry.Dirty
	} else {
		panic("IsDirty: register not allocated yet")
	}
}

// MarkLoaded is invoked after loading a variable from memory, and it marks the
// corresponding register's entry as loaded.
func (blk Blk) MarkLoaded(reg int) {
	loaded := true
	blk.Rdesc[reg] = RegDesc{blk.Rdesc[reg].Name, blk.Rdesc[reg].Dirty, loaded}
}

// UnmarkLoaded unmarks the register entry as loaded.
func (blk Blk) UnmarkLoaded(reg int) {
	loaded := false
	blk.Rdesc[reg] = RegDesc{blk.Rdesc[reg].Name, blk.Rdesc[reg].Dirty, loaded}
}

// IsLoaded determines whether the given register stores the value of the given
// variable and the value has been loaded from memory.
func (blk Blk) IsLoaded(reg int, varName string) bool {
	if entry, ok := blk.Rdesc[reg]; ok {
		return entry.Name == varName && entry.Loaded
	} else {
		panic("IsLoaded: register not allocated yet")
	}
}
