package tac

// EvalFlowGraph generates a flow graph from a three-address code DS. When a
// jump/branch is encountered, the target and the next basic block is marked as
// the right and left child respectively. The immediate predecessors of all the
// basic blocks are also updated.
func (tac Tac) EvalFlowGraph(labelMap labelinfotype) {
	for k, blk := range tac {
		targetBlk := -1
		// TODO: Avoid the following computation in each iteration.
		if k+1 < len(tac) {
			targetBlk = k + 1
			// Left child is always the next block.
			tac[k].child.left = k + 1
			// Update predecessors of the target block.
			if tac[targetBlk].pred == nil {
				tac[targetBlk].pred = make(map[int]bool)
			}
			tac[targetBlk].pred[k] = true
		} else {
			tac[k].child.left = -1
		}
		index := len(blk.Stmts) - 1
		if index < 0 {
			continue
		}
		switch blk.Stmts[index].Op {
		case JMP:
			// unconditional jump
			label := blk.Stmts[index].Dst
			targetBlk = labelMap[label].index
			tac[k].child.left = targetBlk
		case BEQ, BNE, BLT, BLE, BGT, BGE:
			label := blk.Stmts[index].Dst
			targetBlk = labelMap[label].index
			tac[k].child.right = targetBlk
		}
		// Update predecessors of the target block.
		if targetBlk >= len(tac) || targetBlk == -1 {
			continue
		}
		tac[targetBlk].pred[k] = true
	}
}
