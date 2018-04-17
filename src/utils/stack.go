// An implementation of stack data structure. This implementation is used for
// handling function calls invoked by `defer`.

package utils

type Stack struct {
	top *Node
	Len int
}

// Node defines a single element of the stack.
type Node struct {
	// For the current usecase, each item of the stack for deferred calls
	// stores the code for each function call.
	val  []string
	prev *Node
}

// CreateStack creates a new empty stack.
func CreateStack() *Stack {
	return &Stack{nil, 0}
}

// Push pushes a value of any type on top of the stack.
func (stk *Stack) Push(val []string) {
	stk.Len++
	stk.top = &Node{val, stk.top}
}

// Pop removes and returns the value on top of the stack.
func (stk *Stack) Pop() []string {
	if stk.Len > 0 {
		retval := stk.top.val
		stk.top = stk.top.prev
		stk.Len--
		return retval
	}
	return nil // TODO Error handling
}

// Peek returns the value on top of the stack.
func (stk *Stack) Peek() []string {
	if stk.Len > 0 {
		return stk.top.val
	}
	return nil // TODO Error handling
}
