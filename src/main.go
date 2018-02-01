package main

import (
	"fmt"
	"gogo/src/tac"
)

func main() {
	inp := tac.GenTAC("../test/test1.ir")
	fmt.Println(inp.Stmts[0])
}
