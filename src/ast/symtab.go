// This file implements routines for creating and manipulating symbol tables.

package ast

type symTabType map[string]SymTabEntry

type SymInfo struct {
	symTab symTabType
	parent *SymInfo // outer scope
}

type SymTabEntry struct {
	kind    symkind
	symbols []string
}

var (
	// funcSymtabCreated keeps track whether a symbol table corresponding to
	// a function declaration has to be instantiated. This is because
	// usually a new symbol table is created when the corresponding block
	// begins. However, in case of functions the arguments also need to be
	// added to the symbol table. Thus the symbol table is instantiated when
	// the production rule corresponding to the arguments is reached and not
	// when the block begins.
	funcSymtabCreated bool
	// currScope keeps track of the currently active symbol table
	// depending on scope.
	currScope *SymInfo
	// globalSymTab keeps track of the global struct and function
	// declarations. Structs and functions can only be declared globally.
	globalSymTab symTabType
)

// InsertSymbol creates a symbol table entry corresponding to a key in the
// symbol table.
func InsertSymbol(key string, kind symkind, vals ...interface{}) {
	var values []string
	for _, v := range vals {
		switch v := v.(type) {
		case string:
			values = append(values, v)
		case []string:
			values = append(values, v...)
		default:
			panic("InsertSymbol: type not supported")
		}
	}
	currScope.symTab[key] = SymTabEntry{
		kind:    kind,
		symbols: values,
	}
}

// GetSymbol returns the symbol table entry in current scope for a key.
func GetSymbol(key string) (SymTabEntry, bool) {
	entry, ok := currScope.symTab[key]
	return entry, ok
}

// Lookup returns the symbol table entry for a given variable in the current
// scope. If not found, the parent symbol table is looked up until the topmost
// symbol table is reached. If not found in all these symbol tables, then the
// global symbol table is looked up which contains the entries corresponding to
// structs and functions.
func Lookup(v string) (*SymTabEntry, bool) {
	for scope := currScope; scope != nil; scope = scope.parent {
		if entry, ok := scope.symTab[v]; ok {
			return &entry, true
		}
	}
	// Lookup in global scope in case the variable corresponds to a struct
	// or a function name.
	for k, entry := range globalSymTab {
		if k == v {
			return &entry, true
		}
	}
	return nil, false
}

// NewScope modifies the symbol table hierarchy, creating a new scope.
func NewScope() {
	newSymTab := SymInfo{make(symTabType), currScope}
	// Update the current symbol table to point to the newly created one.
	currScope = &newSymTab
}
