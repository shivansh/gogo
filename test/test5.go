package main

import "fmt"

type course struct {
	assign  int
	remarks string
}

func main() {
	var CS335 course
	var x int
	fmt.Println("Enter the marks obtained by the student")
	fmt.Scanf("%d", &x)
	CS335.assign = x
	switch {
	case (x > 90):
		CS335.remarks = "excellent"
	case (x > 80):
		CS335.remarks = "good"
	case (x > 70):
		CS335.remarks = "average"
	default:
		CS335.remarks = "poor"
	}

	fmt.Printf("The student has been %s in the course \n", CS335.remarks)
}
