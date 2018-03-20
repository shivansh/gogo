// Package gentoken generate tokens and the corresponding lexemes extracted
// from a source file.

package gentoken

import (
	"fmt"
	"io/ioutil"
	"log"

	"gogo/tmp/lexer"
	"gogo/tmp/token"
)

func PrintTokens(file string) {
	src, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	type TokInfo struct {
		freq   int             // frequency of a token
		litMap map[string]bool // stores "unique" lexemes of same type
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
			// Allocate on first touch.
			freqMap[tok.Type] = &TokInfo{
				1,
				map[string]bool{lexeme: true},
			}
		} else {
			freqMap[tok.Type].freq++
			if !freqMap[tok.Type].litMap[lexeme] {
				freqMap[tok.Type].litMap[lexeme] = true
			}
		}
	}

	fmt.Println("   Token         Occurrences        Lexemes")
	fmt.Println("----------------------------------------------")
	for key, value := range freqMap {
		fmt.Printf("%12s%12d", token.TokMap.Id(key), value.freq)
		for lexeme := range value.litMap {
			fmt.Printf("%18s\n\t\t\t", lexeme)
		}
		fmt.Printf("\r")
	}
}
