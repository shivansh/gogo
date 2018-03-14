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
	const x int = 8
	func temp() {
		const z int = 10
		const w int = 11
	}
	func main() {
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
