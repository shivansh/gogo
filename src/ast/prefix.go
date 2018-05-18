package ast

// Prefix decarations used as symbol table metadata.
const (
	FNC = "func"
	DRF = "deref"
	PTR = "pointer"
	ARR = "array"
	INT = "int"
	STR = "string"
)

type symkind uint8

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
