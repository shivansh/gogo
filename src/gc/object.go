package gc

type ObjectType uint8

const (
	OBJ_INT = iota
	OBJ_PAIR
)

type ObjInt int

type ObjPair struct {
	head *Object
	tail *Object
}

type Object struct {
	marked bool
	typ    ObjectType
	// next determines the next object in the list of all the objects. The
	// next object maybe live or dead.
	next  *Object
	value int
	// TODO: Use union instead of a separate allocation for INT and PAIR.
	objInt ObjInt
	// objPair denotes references, where the head object references the tail
	// object.
	objPair ObjPair
}

// isMarked verifies whether the object has been marked.
func (object *Object) isMarked() bool {
	if object.marked {
		return true
	}
	return false
}

// mark marks an object live.
func (object *Object) mark() {
	// If the object is marked, we're done. This avoids recursing on cycles
	// in the object graph.
	if object.isMarked() {
		return
	}
	object.marked = true

	// Handle references in objects.
	if object.typ == OBJ_PAIR {
		object.objPair.head.mark()
		object.objPair.tail.mark()
	}
}
