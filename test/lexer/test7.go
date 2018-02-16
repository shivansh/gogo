package main

import "fmt"

func main() {
	x := 0
	y := 0
	fmt.Printf("%d\n", x+y)
	fmt.Printf("%d\n", x&y)
	fmt.Printf("%v\n", x|y)
	fmt.Printf("%d\n", x^y)
	if (x > 0) || (y > 0) {
		fmt.Println("Either of them is true")
	} else if !((x > 0) && (y > 0)) {
		fmt.Println("Both of them false")
	}
}
