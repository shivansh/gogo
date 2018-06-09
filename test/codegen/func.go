package main

func temp(x, y int) (int, int) {
	return x + 1, y + 1
}

func main() {
	newline := "\n"

	a, b := temp(1, 2)
	printInt a
	printStr newline
	printInt b

	return
}
