// This file declares the types used to represent the syntax tree of a source
// program.

package ast

import "strings"

// Prefix declarations used as symbol table metadata. The use of these constants
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
	STRCT  = "struct"
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
func GetType(kind symkind) string {
	switch kind {
	case ARRAYINT:
		return ARRINT // reusing prefix values
	case ARRAYSTR:
		return ARRSTR
	case INTEGER:
		return INT
	case STRING:
		return STR
	default:
		panic("GetType: invalid type")
	}
}

// GetPrefix returns the prefix from a place value.
func GetPrefix(place string) string {
	if i := strings.Index(place, ":"); i != -1 {
		return place[:i]
	}
	return place
}

// StripPrefix strips the prefix from a place value.
func StripPrefix(place string) string {
	i := strings.Index(place, ":")
	return place[i+1:]
}

// AstNode defines a node in the AST of a given program.
type AstNode interface {
	place() string
	code() []string
}

// Node implements the common parts of AstNode.
type Node struct {
	// If the AST node represents an expression, then place stores the name
	// of the variable storing the value of the expression.
	Place string
	Code  []string // IR instructions
}

func (node *Node) place() string { return node.Place }

func (node *Node) code() []string { return node.Code }

type (
	// StructType represents an AST node of a struct.
	StructType struct {
		Node
		Name string // name of the struct
		Len  int    // number of members
	}

	// ArrayType represents an AST node of an array.
	// TODO: ArrayType hasn't been used yet.
	ArrayType struct {
		Node
		// TODO: Update type of Type.
		Type string // type of elements
		Len  int    // array size
	}

	// FuncType represents an AST node of a function.
	FuncType struct {
		Node
	}
)

// --- [ Statements ] ----------------------------------------------------------

type (
	// LabeledStmt represents an AST node of a label statement.
	LabeledStmt struct {
		Node
	}

	// ReturnStmt represents an AST node for a return statement.
	ReturnStmt struct {
		Node
	}

	// SwitchStmt represents an AST node for a switch statement.
	SwitchStmt struct {
		Node
	}

	// ForStmt represents an AST node for a for statement.
	ForStmt struct {
		Node
	}

	// DeferStmt represents an AST node for a defer statement.
	DeferStmt struct {
		Node
	}
)
