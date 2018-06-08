package gc

import "testing"

func TestGCPreserve(t *testing.T) {
	vm := NewVM()
	vm.PushInt(1)
	vm.gc()
	if vm.numObjects != 1 {
		t.Error("Should have preserved all the objects.")
	}
}

func TestGCCollect(t *testing.T) {
	vm := NewVM()
	vm.PushInt(1)
	vm.Pop() // mark the only live object as dead
	vm.gc()
	if vm.numObjects != 0 {
		t.Error("Should have collected all the objects.")
	}
}
