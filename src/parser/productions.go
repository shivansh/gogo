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
	`
	s := lexer.NewLexer([]byte(stmt))
	p := parser.NewParser()
	_, err := p.Parse(s)
	if err != nil {
		log.Fatal(err)
	}
}
