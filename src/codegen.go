package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	src, err := ioutil.ReadFile("test/test1.ir")
	if err != nil {
		log.Fatal(err)
	}

	r := csv.NewReader(strings.NewReader(string(src)))
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		GenAsm(record)
	}
}

func GenAsm(record []string) {
	// <line-number, operation, destination-variable, source-variable(s)>
	for i := 0; i < len(record); i++ {
		record[i] = strings.TrimSpace(record[i])
	}

	switch record[1] {
	case "=":
		out := fmt.Sprintf("li $t0, %s", record[3])
		fmt.Println(out)
	}
}
