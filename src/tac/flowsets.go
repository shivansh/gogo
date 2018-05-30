package tac

import "reflect"

// EvalDataFlowSets computes the data-flow sets GEN and KILL.
func (tac Tac) EvalDataFlowSets() {
	for k, blk := range tac {
		for _, stmt := range blk.Stmts {
			switch stmt.Op {
			case DECLInt, DECLSTR:
				tac[k].dataflow.Gen[stmt.Dst] = true
			case EQ:
				tac[k].dataflow.Kill[stmt.Dst] = true
			}
		}
	}
	tac.EvalTransferFuncs()
}

// Transfer functions evaluation
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// The following equations are used for evaluating the transfer functions -
//   * IN[n] = ∪ p∈pred(n) OUT[p]
//   * OUT[n] = GEN[n] ∪ (IN[n] − KILL[n])
// These transfer functions are evaluated using a fixpoint computation.
func (tac Tac) EvalTransferFuncs() {
	// Initialize Out[n] to GEN[n] for all the basic blocks.
	for k, blk := range tac {
		for v := range blk.dataflow.Gen {
			tac[k].dataflow.Out[v] = true
		}
	}

	// Fixpoint computation.
	again := true
	for again {
		again = false
		for k, blk := range tac {
			// Compute IN[n].
			newIn := make(set)
			for p := range blk.pred {
				for v := range tac[p].dataflow.Out {
					newIn[v] = true
				}
			}
			// Verify if the newly computed IN set is same as the
			// one computed in the previous iteration. If it is,
			// the fix-point computation stops.
			if !reflect.DeepEqual(newIn, tac[k].dataflow.In) {
				again = true
			}
			tac[k].dataflow.In = newIn

			// Compute OUT[n].
			// TODO: Use bitmap to optimize set operations.
			live := make(set)
			for v := range tac[k].dataflow.In {
				live[v] = true
			}
			for v := range tac[k].dataflow.Kill {
				live[v] = false
			}
			for v := range live {
				tac[k].dataflow.Out[v] = true
			}
		}
	}
}
