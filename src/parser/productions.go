package main

import (
	"log"

	"gogo/tmp/lexer"
	"gogo/tmp/parser"
)

func main() {
	stmt := `package main
	import (
	"fmt"
	"os"
	)
	func main() {
	const b int = 2
	x := 4
	x++
	{
	const z = y * 2
	}
	End:
	const a int = 1
	}
	`

	s := lexer.NewLexer([]byte(stmt))
	p := parser.NewParser()
	_, err := p.Parse(s)
	if err != nil {
		log.Fatal(err)
	}
}
