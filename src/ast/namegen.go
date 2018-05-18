// This file implements routines for generating unique names for labels,
// temporaries and variables.

package ast

import "fmt"

var (
	tmpIndex   int // index used for naming temporaries
	labelIndex int // index used for naming labels
	varIndex   int // index used for renaming variables
)

// RealName extracts the original name of a variable from its renamed version.
func RealName(s string) string {
	i := len(s) - 1
	for ; i > 0 && s[i] != '.'; i-- {
	}
	return s[:i]
}

// NewTmp generates a unique temporary variable name.
func NewTmp() string {
	t := fmt.Sprintf("t%d", tmpIndex)
	tmpIndex++
	return t
}

// NewLabel generates a unique label name.
func NewLabel() string {
	l := fmt.Sprintf("l%d", labelIndex)
	labelIndex++
	return l
}

// NewVar generates a unique variable name used for renaming. A variable named
// var will be renamed to 'var.int_lit' where int_lit is an integer. Since
// variable names cannot contain a '.', this will not result in a naming
// conflict with an existing variable. The renamed variable will only occur in
// the IR (there is no constraint on variable names in IR as of now).
func RenameVariable(v string) string {
	ret := fmt.Sprintf("%s.%d", v, varIndex)
	varIndex++
	return ret
}
