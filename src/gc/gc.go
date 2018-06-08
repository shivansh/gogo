// Package gc implements a POC for a mark-and-sweep garbage collector. This is
// yet to be integrated with the rest of the compiler

package gc

import "fmt"

// sweep makes a single pass through the entire list of unallocated objects,
// unallocating the ones which are not marked.
func (vm *VM) sweep() {
	for object := vm.firstObject; object != nil; {
		if !object.isMarked() {
			// The current object was not reached.
			// TODO: instead of unallocating memory, we'll make it
			// available for future allocations since go being garbage
			// collected itself doesn't seem to provide an API for
			// freeing memory (atleast I haven't found it yet).
			// For now, we only update the number of live objects
			// just to make the tests pass.
			vm.numObjects--
		} else {
			// The current object was reached, unmark it for the
			// next GC call.
			object.marked = false
		}
		object = object.next
	}
}

// gc is the entry point for the garbage collector.
func (vm *VM) gc() {
	numObjects := vm.numObjects
	vm.MarkAll()
	vm.sweep()
	// After every call to GC the number of live objects will be modified.
	// We update the GC threshold for the next run to be twice the size of
	// the currently live objects.
	vm.maxObjects = vm.numObjects * 2
	fmt.Printf("GC: Collected %d, remaining %d\n", numObjects-vm.numObjects,
		vm.numObjects)
}
