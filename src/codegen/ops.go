package codegen

import "github.com/shivansh/gogo/src/tac"

// ConvertOp returns an assembly operator from the corresponding IR operator.
func ConvertOp(irOp string) (asmOp string) {
	switch irOp {
	case tac.OR:
		asmOp = "or"
	case tac.AND:
		asmOp = "and"
	case tac.NOR:
		asmOp = "nor"
	case tac.XOR:
		asmOp = "xor"
	case tac.MUL:
		asmOp = "mul"
	case tac.DIV:
		asmOp = "div"
	case tac.SUB:
		asmOp = "sub"
	case tac.REM:
		asmOp = "rem"
	case tac.RST:
		asmOp = "srl"
	case tac.LST:
		asmOp = "sll"
	case tac.BEQ:
		asmOp = "beq"
	case tac.BNE:
		asmOp = "bne"
	case tac.BGT:
		asmOp = "bgt"
	case tac.BGE:
		asmOp = "bge"
	case tac.BLT:
		asmOp = "blt"
	case tac.BLE:
		asmOp = "ble"
	}
	return asmOp
}
