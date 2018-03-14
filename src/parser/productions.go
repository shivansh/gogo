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
	const x [2]int = 2
	func main() {
		{
			const b int = 2
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
