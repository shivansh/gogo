// This file implements routines for peephole optimizations on a three-address
// code IR.

package tac

// PeepHole performs peephole optimizations on the generated three-address code
// data structure.
func PeepHole(tac Tac) *Tac {
	// --- [ Eliminating jumps over jumps ] --------------------------------
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
	// 	print debugging inform ation
	// 	L2:

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
	return &tac
}
