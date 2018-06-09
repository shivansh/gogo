package main

func temp() (int, int) {
	return 1, 2
}

func main() {
	const b = 1 + 2 + 3
	var d int = 1 + 2
	x := 1 + 2
	a := d
	printInt a
	c, e := temp()
	var f, g int
	f, g = temp()
	printInt f
	printInt g
	return
}
