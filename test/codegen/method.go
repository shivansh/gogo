type Node struct {
	x int
}

func (n Node) Getter() {
	a := n.x
	return a
}

func main() {
	y := Node{1}
	// printInt y.x
	z := y.Getter()
	// printInt z
	return
}
