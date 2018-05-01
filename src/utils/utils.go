// Package utils implements general utility functions used by other packages.

package utils

import "strings"

// SplitAndSanitize creates a slice after splitting a string at a separator. The
// entries in resulting slice are trimmed of any whitespace and the resulting
// entries which are empty are removed (originally containing only whitspaces).
func SplitAndSanitize(str string, sep string) []string {
	ret := []string{}
	for _, v := range strings.Split(str, sep) {
		entry := strings.TrimSpace(v)
		if entry != "" {
			ret = append(ret, entry)
		}
	}
	return ret
}

// AppendToSlice is a variadic function which takes multiple strings as
// arguments and appends them to a slice.
func AppendToSlice(slice []string, args ...string) []string {
	for _, v := range args {
		slice = append(slice, v)
	}
	return slice
}
