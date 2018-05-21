// This file defines operators supported by the source language.

package ast

const (
	// arithmetic operators
	ADD = "+"
	SUB = "-"
	// An asterisk can be to define a multiplication and dereference
	// operator, hence its name.
	AST = "*"
	DIV = "/"
	REM = "%"

	// boolean operators
	NOT = "!"
	OR  = "||"
	AND = "&&"

	// comparison operators
	EQ  = "=="
	NEQ = "!="
	LEQ = "<="
	LT  = "<"
	GEQ = ">="
	GT  = ">"

	// increment/decrement operators
	INC = "++"
	DEC = "--"

	AMP = "&"
)
