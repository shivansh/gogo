package main

func main() {
	a := 1
	switch (a) {
	case 2:
		a = a + 1
	default:
		a = a + 4
		printInt a
	}
	return
}
