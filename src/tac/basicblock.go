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
	Rdesc      map[int]string
	NextUseTab [][]UseInfo
	Pq         PriorityQueue
}

// NewBlock creates a new basic block and initializes its line number to 0.
func NewBlock(blk *Blk, tac *Tac) (*Blk, int) {
	if blk != nil && len(blk.Stmts) > 0 {
		*tac = append(*tac, *blk) // end the previous block
		// Create a new block only if the current one is not empty. If
		// it is empty, use it. This case arises when a label statement
		// is encountered after a jump statement. The new block created
		// by the jump statement stays empty as the label statement
		// creates a new block of its own.
		blk = new(Blk)
	}
	return blk, 0
}
