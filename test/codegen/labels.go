package main

func main() {
	a := 1
	b := 2
	newline := "\n"

	goto l2

l1:
	printInt a
	printStr newline
	return

l2:
	a = 4
	printInt b
	printStr newline
	goto l1
}
