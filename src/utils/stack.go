// An implementation of stack data structure. This implementation is used for
// handling function calls invoked by `defer`.

package utils

type Stack struct {
	top *Item
	Len int
}

// Item defines a single element of the stack.
type Item struct {
	val  interface{} // value can be of any type
	prev *Item
}

// CreateStack creates a new empty stack.
func CreateStack() *Stack {
	return &Stack{nil, 0}
}

// Push pushes a value of any type on top of the stack.
func (stk *Stack) Push(val interface{}) {
	stk.Len++
	stk.top = &Item{val, stk.top}
}

// Pop removes and returns the value on top of the stack.
func (stk *Stack) Pop() interface{} {
	if stk.Len > 0 {
		retval := stk.top.val
		stk.top = stk.top.prev
		stk.Len--
		return retval
	}
	return nil // TODO Error handling
}

// Peek returns the value on top of the stack.
func (stk *Stack) Peek() interface{} {
	if stk.Len > 0 {
		return stk.top.val
	}
	return nil // TODO Error handling
}
