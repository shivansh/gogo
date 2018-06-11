package gc

const (
	// INITIAL_GC_THRESHOLD determines the number of objects required
	// initially to trigger the GC (it is currently chosen arbitrarily).
	// This threshold is modified at the end of each GC cycle to be twice
	// the number of live objects. This ensures that the GC is not triggered
	// frequently or infrequently.
	INITIAL_GC_THRESHOLD = 36
	// STACK_MAX determines the maximum number of live objects at any given
	// instant, preventing stack overflow.
	// TODO: adjust this to a nice value.
	STACK_MAX = 1024
)

// VM represents a minimal stack-based virtual machine structure. It's sole
// purpose at the moment is to keep track of all the live objects.
type VM struct {
	numObjects  int
	maxObjects  int
	firstObject *Object
	stack       [STACK_MAX]*Object
	stackSize   int
}

// NewVM creates a new VM object.
func NewVM() *VM {
	vm := new(VM)
	vm.maxObjects = INITIAL_GC_THRESHOLD
	return vm
}

// NewObject creates a new Object.
func (vm *VM) NewObject(typ ObjectType) *Object {
	object := new(Object)
	object.typ = typ

	// Insert the current object into the list of allocated objects.
	object.next = vm.firstObject
	vm.firstObject = object
	vm.numObjects++

	return object
}

// Push inserts a new live object into the VM stack.
func (vm *VM) Push(value *Object) {
	if vm.stackSize >= STACK_MAX {
		panic("Stack overflow!")
	}
	vm.stack[vm.stackSize] = value
	vm.stackSize++
}

// Pop removes a live object, effectively marking it as dead.
func (vm *VM) Pop() *Object {
	if vm.stackSize <= 0 {
		panic("Stack underflow!")
	}
	vm.stackSize--
	return vm.stack[vm.stackSize]
}

// PushInt pushes an object into the stack.
func (vm *VM) PushInt(intval int) {
	object := vm.NewObject(OBJ_INT)
	object.value = intval
	vm.Push(object)
}

// PushPair pushes a pair of objects (references) into the stack.
func (vm *VM) PushPair() *Object {
	object := vm.NewObject(OBJ_PAIR)
	object.objPair.tail = vm.Pop()
	object.objPair.head = vm.Pop()
	vm.Push(object)
	return object
}

// MarkAll marks all the objects as live.
func (vm *VM) MarkAll() {
	for i := 0; i < vm.stackSize; i++ {
		vm.stack[i].mark()
	}
}

// FreeVM collects all the live objects in the VM.
func (vm *VM) FreeVM() {
	// Setting stacksize to zero will avoid Markall from marking all the
	// objects in the next GC run.
	vm.stackSize = 0
	vm.gc()
}
