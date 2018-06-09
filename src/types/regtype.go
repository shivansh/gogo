// Package types defines type information used by the register allocator and the
// three-address code data structure. The usage of this package arises when type
// information collectd from the three-address code has to be propogated to the
// register allocator.
package types

// RegType determines the type of the variable stored by a register.
type RegType int

const (
	NIL RegType = iota
	INT
	STR
	ARR
)
