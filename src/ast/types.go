package ast

// Prefix decarations used as symbol table metadata. The use of these constants
// is to act as meta information stored in the place attribute of a node.
// Currently these are simply prepended to the place attribute, and checked for
// presence when required.
const (
	FNC    = "func"
	DRF    = "deref"
	PTR    = "pointer"
	ARR    = "arr"
	ARRINT = "arrint"
	ARRSTR = "arrstr"
	INT    = "int"
	STR    = "string"
)

// symkind determines the kind of symbol table entry.
type symkind uint8

// The following declarations determine the values which can be taken by symkind.
const (
	NIL symkind = iota
	FUNCTION
	DEREF
	POINTER
	ARRAYINT
	ARRAYSTR
	INTEGER
	STRING
	STRUCT
)

// GetType returns the type information from a symkind variable.
func GetType(kind symkind) (retVal string) {
	switch kind {
	case ARRAYINT:
		retVal = ARRINT // reusing prefix values
	case ARRAYSTR:
		retVal = ARRSTR
	case INTEGER:
		retVal = INT
	case STRING:
		retVal = STR
	default:
		panic("GetType: invalid type")
	}
	return retVal
}

// GetPrefix returns the prefix from a place value.
func GetPrefix(place string) string {
	i := 0
	for ; i < len(place) && place[i] != ':'; i++ {
	}
	return place[:i]
}

// StripPrefix strips the prefix from a place value.
func StripPrefix(place string) string {
	i := 0
	for ; i < len(place) && place[i] != ':'; i++ {
	}
	if i < len(place) {
		return place[i+1:]
	}
	return place
}

// StructType represents an AST node of type struct.
type StructType struct {
	Node
}
