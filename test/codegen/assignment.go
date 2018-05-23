func main() {
	const a, b = 1, 3
	var c, d int = 2, 4
	e, f := 4, 8
	var g int
	// IR instruction for the following should not be "declInt".
	g = a

	// A failing testcase
	// var h string
	// h = "Hello world!"
	return
}
