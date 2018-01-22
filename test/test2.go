package test

import (
	"fmt"
	"os"
)

func main() {
	// Identifiers
	_, err := os.Stat("noFile") // blank identifier
	if err != nil {
		log.Fatal(err)
	}
	// 'identifier' should not be ambiguous with 'blank identifier'
	_varIdent := "ident"

	// Floating-point literals
	varDec := 2 + 4
	varFloat := 2.5
	varExp := 6.67428e-11

	// String literals
	strLit := "String"
	strRaw := `Raw string`

	// Rune literals
	runeLit := 'a'

	// Array types
	var buffer [256]byte
}
