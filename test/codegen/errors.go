// test increments the arguments by 1 and returns them.
func test(a, b int) (int, int) {
	return a + 1, b + 1
}

func main() {
	// Variable usage withoug declaration.
	a = 1

	// Variable redeclarations.
	a := 4

	// Unequal values in LHS and RHS.
	a, b := 1, 2, 3

	// Function return values.
	a := test(1, 2)

	return
}
