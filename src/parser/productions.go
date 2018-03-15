package main

import (
	"log"

	"gogo/tmp/lexer"
	"gogo/tmp/parser"
)

func main() {
	stmt := `package main
	func main() {
		const a int = 1
		switch a {
		case 1:
			const a int = 1
		case 2:
			const a int = 1
		default:
			const a int = 1
		}
	}
	`

	s := lexer.NewLexer([]byte(stmt))
	p := parser.NewParser()
	_, err := p.Parse(s)
	if err != nil {
		log.Fatal(err)
	}
}
