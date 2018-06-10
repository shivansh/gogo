package ast

import (
	"errors"
	"fmt"
)

var (
	ErrDeclArr    = errors.New("use short declaration for declaring arrays")
	ErrDeclStruct = errors.New("use short declaration for declaring structs")
	ErrShortDecl  = errors.New("no new variables on left side of :=")
)

// ErrUndefined returns an undefined variable error.
func ErrUndefined(varName string) error {
	return fmt.Errorf("undefined: %s", varName)
}

// ErrCountMismatch returns an assignment count mismatch error.
func ErrCountMismatch(leftCount, rightCount int) error {
	return fmt.Errorf("assignment count mismatch: %d = %d", leftCount, rightCount)
}

// ErrInvalidFunc returns an non-function call error.
func ErrInvalidFunc(name string, typ string) error {
	return fmt.Errorf("cannot call non-function %s (type %s)", name, typ)
}

// ErrIndirection returns an invalid indirection error.
func ErrIndirection(varName, varType string) error {
	return fmt.Errorf("invalid indirect of %s (type %s)", varName, varType)
}
