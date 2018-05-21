package ast

// Prefix decarations used as symbol table metadata. The use of these constants
// is to act as meta information stored in the place attribute of a node.
// Currently these are simply prepended to the place attribute, and checked for
// presence when required.
const (
	FNC = "func"
	DRF = "deref"
	PTR = "pointer"
	ARR = "array"
	INT = "int"
	STR = "string"
)

// symkind determines the kind of symbol table entry.
type symkind uint8

// The following declarations determine the values which can be taken by symkind.
const (
	NIL = iota
	FUNCTION
	DEREF
	POINTER
	ARRAY
	INTEGER
	STRING
	STRUCT
)
