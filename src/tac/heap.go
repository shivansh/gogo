// This file implements operations supported by priority queue used during
// register allocation by GetReg() routine.

package tac

type PriorityQueue []*UseInfo

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// Pop is expected to return the entity with highest nextuse.
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
