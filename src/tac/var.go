// This file implements operations supported by a source variable in the three
// address code IR representation.

package tac

import (
	"log"
	"strconv"
)

type (
	I32 int
	Str string
)

// SrcVar defines the properties of a source variable.
type SrcVar interface {
	IntVal() int
	StrVal() string
}

func (U I32) IntVal() int {
	return int(U)
}

func (U I32) StrVal() string {
	return strconv.Itoa(U.IntVal())
}

func (U Str) IntVal() (i int) {
	i, err := strconv.Atoi(U.StrVal())
	if err != nil {
		log.Fatal(err)
	}
	return
}

func (U Str) StrVal() string {
	return string(U)
}
