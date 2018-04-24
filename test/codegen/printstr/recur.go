func recur(x, y, z int) int {
        if (x < 0 && y < 0 && z < 0) {
        	return 0
        } else if (y < 0 && z < 0) {
        	temp := recur(x-1, y, z)
        	temp++
        	return temp
        } else if (z < 0) {
        	temp := recur(x-1, y-1, z)
        	temp++
        	return temp
        } else {
        	temp := recur(x-1, y-1, z-1)
        	temp++
        	return temp
        }
}

func main() {
	var arg1, arg2, arg3 int
	str := "Enter 3 integers :\n"
	printStr str
	scanInt arg1
	scanInt arg2
	scanInt arg3
	str1 := "Retuen value of function = "
	printStr str1
	x := recur(arg1, arg2, arg3)
	printInt x
	newline := "\n"
	printStr newline
	return
}
