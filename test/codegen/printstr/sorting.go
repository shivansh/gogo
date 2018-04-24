func main() {
	a := [1000]int{}
	var length int
	str1 := "Enter length of the array : "
	printStr str1
	scanInt length
	newline := "\n"
	var x int
	str3 := "Enter "
	printStr str3
	printInt length
	str4 := "integers :\n"
	printStr str4
	for i := 0; i<length; i++ {
		scanInt x
		a[i] = x
	}
	var min, minidx, temp int
	for i = 0; i<length; i++ {
		min = a[i]
		minidx = i
		for j := i; j<length; j++ {
			if (a[j] < min) {
				min = a[j]
				minidx = j
			}
		}
		temp = a[minidx]
		a[minidx] = a[i]
		a[i] = temp
	}
	str := "\n"
	for i = 0; i<length; i++ {
		y := a[i]
		printInt y
		printStr str
	}
	return
}