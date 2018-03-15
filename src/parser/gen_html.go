package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func FindNonTerminal(input []string) int {
	index := -1
	for i := len(input) - 1; i >= 0; i-- {
		if input[i][0] >= 'A' && input[i][0] <= 'Z' {
			index = i
			break
		}
	}
	return index
}

func GenHTML(file *os.File) {
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	record := strings.Split(scanner.Text(), " ")
	for i := 0; i < len(record); i++ {
		record[i] = strings.TrimSpace(record[i])
	}
	stmt := record
	fmt.Printf("<b><u>%s</u></b><br><br>\n", stmt[0])
	for scanner.Scan() {
		s := scanner.Text()
		record = strings.Split(s, " ")
		for i := 0; i < len(record); i++ {
			record[i] = strings.TrimSpace(record[i])
		}
		index := FindNonTerminal(stmt)
		// Insert all entries of record into stmt at index'th position.
		temp := []string{}
		if index != -1 {
			temp = append(temp, stmt[0:index]...)
			temp = append(temp, record...)
			temp = append(temp, stmt[index+1:]...)
			stmt = []string{}
			stmt = append(stmt, temp...)
		}
		index = FindNonTerminal(stmt)
		for k, v := range stmt {
			if k == index {
				fmt.Printf("<b><u>%s</u></b> ", v)
			} else if strings.Compare(v, "empty") != 0 {
				fmt.Printf("%s ", v)
			}
		}
		fmt.Println("<br><br>")
	}
}

func main() {
	file, err := os.Open("src/parser/output.txt")
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	GenHTML(file)
}
