// This file implements routines for peephole optimizations on a three-address
// code IR.

package tac

import "sort"

// The following constant declarations determine the confidence in dropping a
// basic block from the three-address code data structure.
const (
	NIL   = iota
	FALSE // cannot be dropped
	MAYBE // **might** be dropped
	TRUE  // will be dropped
)

// PeepHole performs peephole optimizations on the generated three-address code
// data structure.
func (tac Tac) PeepHole() Tac {
	// TODO: Figure out how to use a reference to a slice.
	tac = tac.JumpsOverJumps()
	tac = tac.ControlFlow()
	return tac
}

// --- [ Eliminating jumps over jumps ] ----------------------------------------
// Consider the following code sequence -
//
// 	if debug == 1 goto L1
// 	goto L2
// 	L1: print debugging information
// 	L2:
//
// After eliminating jumps over jumps, the transformed code looks as -
//
// 	if debug != 1 goto L2
// 	print debugging information
// 	L2:
//
// JumpsOverJumps performs jump-over-jump peephole optimization.
func (tac Tac) JumpsOverJumps() Tac {
	// Find a sequence of statements where "beq" is followed by "jmp".
	foundBEQ := false
	for k, blk := range tac {
		for _, stmt := range blk.Stmts {
			switch stmt.Op {
			case BEQ:
				foundBEQ = true
			case JMP:
				if foundBEQ {
					// Valid sequence encountered. Update the
					// instruction and jump target of the
					// last statement in the previous block.
					index := len(tac[k-1].Stmts) - 1
					branchOp := tac[k-1].Stmts[index].Op
					switch branchOp {
					case BNE:
						tac[k-1].Stmts[index].Op = BEQ
					case BEQ:
						tac[k-1].Stmts[index].Op = BNE
					default:
						panic("PeepHole: invalid branch")
					}
					tac[k-1].Stmts[index].Dst = stmt.Dst

					// Replace the jump instruction in the
					// current basic block by all the
					// instructions from the next block. The
					// first and last statements, namely label
					// and jump respectively are not copied.
					if len(tac[k+1].Stmts) > 0 {
						tac[k].Stmts = tac[k+1].Stmts[1:]
					}

					// Drop the next basic block.
					tac = append(tac[:k+1], tac[k+2:]...)
					foundBEQ = false
				}
			default:
				foundBEQ = false
			}
		}
	}

	tac.FixLineNumbers()
	return tac
}

// --- [ Flow-of-Control Optimizations ] ---------------------------------------
// Consider the following code sequence -
//
//	goto L1
//	...
//	L1: goto L2
//
// The unnecessary jump to L1 can be eliminated as follows -
//
//	goto L2
//	...
//	L1: goto L2
//
// If there are no more jumps to L1, the entire label can be dropped provided
// that it is immediately preceded by an unconditional jump.
// Similar logic can be extended to branch statements.
//
// A notation which is used in the following segment is a "fallthrough label".
// Fallthrough labels are the blocks which are not preceded by an unconditional
// jump. Consider the following code -
//
//	a = 0
//	L1: a = 1
//
// In the above code, L1 is a fallthrough label.
//
// ControlFlow performs flow of control optimizations.
func (tac Tac) ControlFlow() Tac {
	// dropIndices contains the indices of blocks which are to be dropped.
	dropIndices := make(map[int]bool)
	labelMap := tac.LabelInfo()

	// In the second pass, inspect the jump statements.
	for k, blk := range tac {
		for _, stmt := range blk.Stmts {
			if IsBranchOp(stmt.Op) {
				if labelMap[stmt.Dst].canDrop == MAYBE {
					// The block which contained only jump/branch
					// statement were marked MAYBE earlier. Since
					// we've found a block referencing the current
					// block (marked MAYBE), it can be dropped.
					labelMap[stmt.Dst].canDrop = TRUE
					dropIndices[labelMap[stmt.Dst].index] = true
				} else {
					// Update inward references of the
					// blocks to which the current block
					// references.
					labelMap[stmt.Dst].inRef = append(labelMap[stmt.Dst].inRef, k)
				}
			}
		}
	}

	// The blocks which are not referenced by any other block can be dropped.
	for k, label := range labelMap {
		if len(label.inRef) == 0 && labelMap[k].canDrop != FALSE {
			labelMap[k].canDrop = TRUE // flip MAYBE to TRUE
			dropIndices[label.index] = true
		}
	}

	// The blocks are supposed be dropped in asceding order of their
	// occurrence to enable proper index calculations while updating the TAC
	// data structure.
	keys := []int{}
	for k := range dropIndices {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// The blocks which reference the to-be-dropped blocks should be updated
	// to point to the corresponding next valid block (which won't be
	// dropped).
	// Consider the code -
	//
	//	L1: a = 0
	//	    goto L2
	//	L2: goto L3
	//	L3: a = 1
	//
	// After eliminating L2, L1 should be modified as follows -
	//
	//	L1: a = 0
	//	    goto L3
	//	L3: a = 1
	for i := 0; i < len(keys); i++ {
		blk := tac[keys[i]]
		index := len(blk.Stmts) - 1
		nextBlk := 0
		toBeDropped := TRUE
		// Keep following the references to the next blocks until we find
		// one which will not be dropped.
		for toBeDropped != FALSE {
			if IsBranchOp(blk.Stmts[index].Op) {
				// Follow the destination of jump/branch
				// statement.
				nextBlk = labelMap[blk.Stmts[index].Dst].index
				toBeDropped = labelMap[tac[nextBlk].Stmts[0].Dst].canDrop
			} else {
				// Fallthough block is the next valid block.
				// TODO: Check if nextBlk can reach past the
				// boundary of the tac data structure.
				nextBlk = keys[i] + 1
				toBeDropped = labelMap[blk.Stmts[nextBlk].Dst].canDrop
			}
		}
		// Update the next valid blocks for all the ones referencing the
		// current block.
		for _, prevBlk := range labelMap[blk.Stmts[0].Dst].inRef {
			index = len(tac[prevBlk].Stmts) - 1
			if IsBranchOp(tac[prevBlk].Stmts[index].Op) {
				if _, ok := labelMap[tac[nextBlk].Stmts[0].Dst]; ok {
					tac[prevBlk].Stmts[index].Dst = tac[nextBlk].Stmts[0].Dst
				}
			}
		}
		// Else the next valid block will be reached via a fallthrough.
	}

	// Drop the blocks, and update the tac data structure accordingly.
	dropCount := 0
	for _, i := range keys {
		tac = append(tac[:i-dropCount], tac[i+1-dropCount:]...)
		dropCount++
	}

	return tac
}

// IsBranchOp verifies whether the given operator is jump/branch.
func IsBranchOp(op string) bool {
	switch op {
	case JMP, BEQ, BNE, BLT, BLE, BGT, BGE:
		return true
	}
	return false
}
