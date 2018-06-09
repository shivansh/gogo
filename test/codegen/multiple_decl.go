package main

func main() {
	a := [3]int{}
	a[0], a[1], a[2] = 0, 1, 2
	var x int = a[0]
	var y int = a[1]
	var z int = a[2]
	printInt x
	printInt y
	printInt z
	return
}
