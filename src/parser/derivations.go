// Package parser generates the rightmost derivations used in the bottom-up
// parsing of a given go program and pretty-prints them in HTML format,
// highlighting important characteristics.

package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	parseError "github.com/shivansh/gogo/src/goccgen/errors"
	"github.com/shivansh/gogo/src/goccgen/lexer"
	"github.com/shivansh/gogo/src/goccgen/parser"
	"github.com/shivansh/gogo/src/utils"
)

// GenProductions generates the RHS of productions in the reverse order of
// rightmost derivations used in the bottom-up parsing of the input program.
// The routine prints to stdout by default.
func GenProductions(file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	s := lexer.NewLexer(content)
	p := parser.NewParser()
	// When parsing is finished, a final call to the routine PrintIR() (in
	// package ast) is made. This prints the generated IR instructins.
	_, err = p.Parse(s)
	if err != nil {
		e := err.(*parseError.Error)
		return fmt.Errorf("%s:%d: %s\n", file, e.ErrorToken.Pos.Line, e.Err)
	}
	return nil
}

// FindNonTerminal returns the index of the rightmost non-terminal in the given
// string of terminals and non-terminals. In the current usecase, this string
// is the RHS of the production being used in a reduce step of bottom-up parsing.
func FindNonTerminal(input []string) int {
	index := -1
	for i := len(input) - 1; i >= 0; i-- {
		if input[i] != "" && input[i][0] >= 'A' && input[i][0] <= 'Z' {
			index = i
			break
		}
	}
	return index
}

// RightmostDerivation generates the HTML showing the rightmost derivations
// used in bottom-up parsing of the input program.
func RightmostDerivation(file string) error {
	// Create a pipe with stdout mapped to its write end.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	outChan := make(chan []string)

	// Buffer the output generated by GenProductions into a pipe. The output
	// is copied in a separate goroutine so that printing doesn't block.
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outChan <- utils.Tac(buf.String())
	}()

	// The output of GenProductions will be buffered in the pipe.
	if err := GenProductions(file); err != nil {
		return err
	}
	w.Close()
	// Restore the original state (before pipe was created).
	os.Stdout = old

	productions := <-outChan

	// Find the start symbol (first non-empty string).
	index := 0
	for k, v := range productions {
		if v != "" {
			index = k
			break
		}
	}
	record := strings.Split(productions[index], " ")
	for i := 0; i < len(record); i++ {
		record[i] = strings.TrimSpace(record[i])
	}
	// str will hold the string of terminals generated by the parser after
	// all the rightmost-derivations for the given input program.
	str := record

	output := fmt.Sprintf("%s.html", strings.TrimSuffix(file, ".go"))
	f, err := os.Create(output)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	_, err = writer.WriteString(fmt.Sprintf("<b><u>%s</u></b><br><br>\n", str[0]))
	if err != nil {
		log.Fatal(err)
	}

	// The rightmost non-terminal in str (currently the start symbol) will
	// be replaced by the RHS of the next production in "productions" until
	// no more non-terminals are left.
	for _, s := range productions[index+1:] {
		record = strings.Split(s, " ")
		for i := 0; i < len(record); i++ {
			record[i] = strings.TrimSpace(record[i])
		}
		index := FindNonTerminal(str)
		// Insert all entries of record into str at index'th position.
		temp := []string{}
		if index != -1 {
			temp = append(temp, str[0:index]...)
			temp = append(temp, record...)
			temp = append(temp, str[index+1:]...)
			str = append([]string{}, temp...)
		}
		startIndex := index
		endIndex := index + len(record) - 1
		index = FindNonTerminal(str)
		for k, v := range str {
			if k == startIndex {
				_, err = writer.WriteString(fmt.Sprintf("<font color=\"blue\">"))
				if err != nil {
					log.Fatal(err)
				}
			}
			if k == index {
				_, err = writer.WriteString(fmt.Sprintf("<b><u>%s</u></b> ", v))
				if err != nil {
					log.Fatal(err)
				}
			} else if v != "empty" {
				_, err = writer.WriteString(fmt.Sprintf("%s ", v))
				if err != nil {
					log.Fatal(err)
				}
			}
			if k == endIndex {
				_, err = writer.WriteString(fmt.Sprintf("</font>"))
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		_, err = writer.WriteString("<br><br>\n")
		if err != nil {
			log.Fatal(err)
		}
	}
	writer.Flush()
	return nil
}
