package main

// test increments the arguments by 1 and returns them.
func test(a, c int) (int, int) {
	return a + 1, c + 1
}

func main() {
	// NOTE: When testing, modify below statements to erroneous states.
	// Variable usage withoug declaration.
	a := 1

	// Variable redeclarations.
	a = 4

	// Unequal values in LHS and RHS.
	b, c, d := 1, 2, 3

	// Function return values.
	e, f := test(1, 2)

	return
}
