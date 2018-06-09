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

// ErrCountMismatch returns an assignment count mismatch error.
func ErrCountMismatch(leftCount, rightCount int) error {
	return fmt.Errorf("assignment count mismatch: %d = %d", leftCount, rightCount)
}
