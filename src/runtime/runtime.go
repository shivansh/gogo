package runtime

import (
	"fmt"

	"github.com/shivansh/gogo/src/cmd"
)

// Regenerate regenerates the runtime code.
func Regenerate() {
	fmt.Println("--- [ Regenerating runtime ] ----------------------------")
	runtimeSrc := "src/runtime/dummy.go"
	runtimeDst := "src/runtime/runtime.asm"
	runtime := true
	cmd.GenAsm(runtimeSrc, runtimeDst, runtime)
}
