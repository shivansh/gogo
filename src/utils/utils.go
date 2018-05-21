// Package utils implements general utility functions used by other packages.

package utils

import (
	"log"
	"strings"
)

// SplitAndSanitize creates a slice after splitting a string at a separator. The
// entries in resulting slice are trimmed of any whitespace and the resulting
// entries which are empty are removed (originally containing only whitspaces).
func SplitAndSanitize(str string, sep string) (retVal []string) {
	for _, v := range strings.Split(str, sep) {
		entry := strings.TrimSpace(v)
		if entry != "" {
			retVal = append(retVal, entry)
		}
	}
	return retVal
}

// AppendCode is a variadic function which takes multiple strings as arguments
// and appends them to a slice of IR codes.
func AppendCode(slice []string, args ...interface{}) []string {
	for _, v := range args {
		switch v.(type) {
		case string:
			slice = append(slice, v.(string))
		case []string:
			slice = append(slice, v.([]string)...)
		default:
			log.Fatalf("AppendCode: unsupported type %v", v)
		}
	}
	return slice
}
