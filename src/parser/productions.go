package main

import (
	"io/ioutil"
	"log"

	"gogo/tmp/lexer"
	"gogo/tmp/parser"
)

func main() {
	content, err := ioutil.ReadFile("src/parser/input.go")
	if err != nil {
		log.Fatal(err)
	}
	s := lexer.NewLexer(content)
	p := parser.NewParser()
	_, err = p.Parse(s)
	if err != nil {
		log.Fatal(err)
	}
}
