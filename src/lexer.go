package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gogo/src/lexer"
	"gogo/src/token"
)

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatalf("Usage: ./lexer <filename>")
	}

	src, err := ioutil.ReadFile(args[1])
	if err != nil {
		log.Fatal(err)
	}

	type TokInfo struct {
		freq   int            // frequency of a lexeme
		litMap map[string]int // stores "unique" lexemes of same type
	}

	freqMap := make(map[token.Type]*TokInfo)
	s := lexer.NewLexer(src)
	for {
		tok := s.Scan()
		if tok.Type == token.EOF {
			break
		}
		lexeme := string(tok.Lit[:])
		if freqMap[tok.Type] == nil {
			freqMap[tok.Type] = &TokInfo{
				1,
				map[string]int{lexeme: 1},
			}
		} else {
			freqMap[tok.Type].litMap[lexeme] = 1
			freqMap[tok.Type].freq++
		}
	}

	fmt.Println("   Token         Occurrences        Lexemes")
	fmt.Println("----------------------------------------------")
	for k, v := range freqMap {
		fmt.Printf("%12s%12d", token.TokMap.Id(k), v.freq)
		for lexeme := range v.litMap {
			fmt.Printf("%18s\n\t\t\t", lexeme)
		}
		fmt.Printf("\r")
	}
}
