// This file defines operators supported by the three-address code IR.

package tac

const (
	// arithmetic operators
	ADD = "+"
	SUB = "-"
	MUL = "*"
	DIV = "/"
	REM = "%"
	EQ  = "="

	// branch operators
	BGT = "bgt"
	BGE = "bge"
	BLT = "blt"
	BLE = "ble"
	BEQ = "beq"
	BNE = "bne"
	JMP = "jmp"

	// array operators
	FROM = "from"
	INTO = "into"

	// binary operators
	OR  = "or"
	AND = "and"
	NOR = "nor"
	XOR = "xor"
	NOT = "not"

	// shift operators
	RST = ">>"
	LST = "<<"

	// function operators
	FUNC  = "func"
	LABEL = "label"
	RET   = "ret"
	CALL  = "call"
	STORE = "store"

	CMT = "#" // comments

	// declaration operators
	DECL    = "decl"
	DECLInt = "declInt"
	DECLSTR = "declStr"

	// I/O operators
	SCANINT  = "scanint"
	PRINTINT = "printint"
	PRINTSTR = "printstr"
)
