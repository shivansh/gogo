package runtime

import "github.com/shivansh/gogo/src/cmd"

// Regenerate regenerates the runtime code.
func Regenerate() {
	runtimeSrc := "src/runtime/dummy.go"
	runtimeDst := "src/runtime/runtime.asm"
	runtime := true
	cmd.GenAsm(runtimeSrc, runtimeDst, runtime)
}
