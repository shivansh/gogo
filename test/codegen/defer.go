func helloWorld(a, b int) {
	sum := a + b
}

func temp() {
	a := 1
}

func main() {
	// temp() will be called first and then helloWorld() will be called.
	defer helloWorld(1+2, 3)
	defer temp()
	return
}
