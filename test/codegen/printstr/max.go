func max(x, y int) int {
	if (x < y) {
		return y
	} else {
		return x
	}
}

func maximum(x, y, z int) int {
	str2 := "Maximum value of first two integers = "
	printStr str2
	var temp int = max(x, y)
	printInt temp
	newline := "\n"
	printStr newline
	var maxim int = max(temp, z)
	return maxim
}

func main() {
	var arg1, arg2, arg3 int
	str := "Enter 3 integers :\n"
	printStr str
	scanInt arg1
	scanInt arg2
	scanInt arg3
	str1 := "Maximum value of these three integers = "
	x := maximum(arg1,arg2,arg3)
	printStr str1
	printInt x
	newline := "\n"
	printStr newline
	return
}