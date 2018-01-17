package main

import (
	"fmt"
	"github.com/goccmack/gocc/example/calc/lexer"
	"io/ioutil"
	"log"
)

func main() {
	src, err := ioutil.ReadFile("test/calcfile")
	if err != nil {
		log.Fatal(err)
	}

	s := lexer.NewLexer(src)
	fmt.Println("   offset   line   column   lexeme")
	fmt.Println("-------------------------------------")
	for {
		tok := s.Scan()
		if tok.Pos.Offset >= len(src) {
			break
		}
		fmt.Printf("%6d %7d %7d %8s\n", tok.Pos.Offset, tok.Pos.Line,
			tok.Pos.Column, string(tok.Lit[:]))
	}
}
