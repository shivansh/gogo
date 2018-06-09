package main

func firstFunc() {
	str := "First function call!\n"
	printStr str
	return
}

func midFunc(a, c int) {
	str := "Second function call! Result: "
	newline := "\n"
	sum := a + c
	printStr str
	printInt sum
	printStr newline
	return
}

func lastFunc() {
	str := "Last function call!\n"
	printStr str
	return
}

func main() {
	defer lastFunc()
	defer midFunc(1+2, 3)
	defer firstFunc()
	return
}
