package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/shivansh/gogo/src/cmd"
	"github.com/shivansh/gogo/src/runtime"
)

func main() {
	args := os.Args
	if len(args) == 2 && args[1] == "-runtime" {
		// Regenerate runtime code. It doesn't seem necessary to
		// provide the end user info about this flag in the description
		// defined below.
		runtime.Regenerate()
		os.Exit(0)
	}

	asm := flag.Bool("s", false, "Generates MIPS assembly from go program")
	ir := flag.Bool("r", false, "Generates IR instructions from go program")
	ir2asm := flag.Bool("r2s", false, "Generates the MIPS assembly from IR")
	prod := flag.Bool("p", false, "Generates rightmost derivations used in bottom-up parsing")
	// optimize := flag.Bool("O", false, "Turn on optimizations")
	flag.Parse()

	if len(args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: gogo (-p | -r | -r2s | -s) <filename>\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var err error
	ErrInvalidFile := errors.New("invalid filename")
	index := strings.LastIndex(args[2], ".")
	if index == -1 {
		err = ErrInvalidFile
	}

	// If control flow has reached here, then we are not compiling runtime
	// code.
	runtime := false

	if *asm {
		dst := fmt.Sprintf("%s.asm", args[2][:index])
		cmd.GenAsm(args[2], dst, runtime)
	} else if *ir {
		cmd.GenIR(args[2])
	} else if *ir2asm {
		// TODO: Verify if the input is indeed in a valid IR format.
		dst := fmt.Sprintf("%s.asm", args[2][:index])
		cmd.GenAsmFromIR(args[2], dst, runtime)
	} else if *prod {
		if err = cmd.GenHTML(args[2]); err == nil {
			index := strings.LastIndex(args[2], ".")
			genFileName := args[2][:index]
			fmt.Printf("Successfully created %s.html\n", genFileName)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}
